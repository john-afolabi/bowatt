package store

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schema string

const timestampLayout = time.RFC3339Nano

type SQLite struct {
	db *sql.DB
}

func Open(ctx context.Context, path string) (*SQLite, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	dsn := fmt.Sprintf(
		"file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)&_pragma=synchronous(NORMAL)",
		path,
	)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite %q: %w", path, err)
	}

	db.SetMaxOpenConns(1)

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("connect to sqlite %q: %w", path, err)
	}
	if _, err := db.ExecContext(ctx, schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}
	return &SQLite{db: db}, nil
}

func (s *SQLite) Close() error {
	return s.db.Close()
}

func (s *SQLite) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *SQLite) Create(ctx context.Context, in NewImage) (Image, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Image{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	createdAt := in.CreatedAt.UTC()
	_, err = tx.ExecContext(ctx, `
		INSERT INTO images (id, storage_key, title, content_type, width, height, size_bytes, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		in.ID, in.StorageKey, in.Title, in.ContentType, in.Width, in.Height, in.SizeBytes,
		createdAt.Format(timestampLayout),
	)
	if err != nil {
		return Image{}, fmt.Errorf("insert image: %w", err)
	}

	for _, tag := range in.Tags {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO tags (name) VALUES (?) ON CONFLICT (name) DO NOTHING`, tag,
		); err != nil {
			return Image{}, fmt.Errorf("insert tag %q: %w", tag, err)
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO image_tags (image_id, tag_id) SELECT ?, id FROM tags WHERE name = ?`,
			in.ID, tag,
		); err != nil {
			return Image{}, fmt.Errorf("link tag %q: %w", tag, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return Image{}, fmt.Errorf("commit image: %w", err)
	}

	return Image{
		ID:          in.ID,
		Title:       in.Title,
		Tags:        append([]string{}, in.Tags...),
		URL:         URLPathPrefix + in.StorageKey,
		Width:       in.Width,
		Height:      in.Height,
		SizeBytes:   in.SizeBytes,
		ContentType: in.ContentType,
		StorageKey:  in.StorageKey,
		CreatedAt:   createdAt,
	}, nil
}

func (s *SQLite) List(ctx context.Context, q ListQuery) ([]Image, error) {
	if q.Limit <= 0 {
		return nil, errors.New("list: limit must be positive")
	}

	var (
		conditions []string
		args       []any
	)

	if q.Cursor != "" {
		conditions = append(conditions, "i.id < ?")
		args = append(args, q.Cursor)
	}

	if len(q.Tags) > 0 {
		// AND semantics: only images carrying every requested tag match.
		conditions = append(conditions, fmt.Sprintf(`
			(SELECT COUNT(DISTINCT t.name)
			   FROM image_tags it
			   JOIN tags t ON t.id = it.tag_id
			  WHERE it.image_id = i.id AND t.name IN (%s)) = ?`, placeholders(len(q.Tags))))
		for _, tag := range q.Tags {
			args = append(args, tag)
		}
		args = append(args, len(q.Tags))
	}

	query := `SELECT i.id, i.title, i.storage_key, i.content_type, i.width, i.height, i.size_bytes, i.created_at
	            FROM images i`
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY i.id DESC LIMIT ?"
	args = append(args, q.Limit)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query images: %w", err)
	}
	defer rows.Close()

	images := make([]Image, 0, q.Limit)
	for rows.Next() {
		var (
			img       Image
			createdAt string
		)
		if err := rows.Scan(
			&img.ID, &img.Title, &img.StorageKey, &img.ContentType,
			&img.Width, &img.Height, &img.SizeBytes, &createdAt,
		); err != nil {
			return nil, fmt.Errorf("scan image: %w", err)
		}
		img.CreatedAt, err = time.Parse(timestampLayout, createdAt)
		if err != nil {
			return nil, fmt.Errorf("parse stored timestamp %q: %w", createdAt, err)
		}
		img.URL = URLPathPrefix + img.StorageKey
		images = append(images, img)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read images: %w", err)
	}

	tagsByImage, err := s.tagsFor(ctx, images)
	if err != nil {
		return nil, err
	}
	for i := range images {
		images[i].Tags = append([]string{}, tagsByImage[images[i].ID]...)
	}
	return images, nil
}

func (s *SQLite) tagsFor(ctx context.Context, images []Image) (map[string][]string, error) {
	result := make(map[string][]string, len(images))
	if len(images) == 0 {
		return result, nil
	}

	args := make([]any, 0, len(images))
	for _, img := range images {
		args = append(args, img.ID)
	}

	query := fmt.Sprintf(`
		SELECT it.image_id, t.name
		  FROM image_tags it
		  JOIN tags t ON t.id = it.tag_id
		 WHERE it.image_id IN (%s)
		 ORDER BY t.name`, placeholders(len(args)))

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query image tags: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var imageID, name string
		if err := rows.Scan(&imageID, &name); err != nil {
			return nil, fmt.Errorf("scan image tag: %w", err)
		}
		result[imageID] = append(result[imageID], name)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read image tags: %w", err)
	}
	return result, nil
}

func placeholders(n int) string {
	return strings.TrimSuffix(strings.Repeat("?,", n), ",")
}
