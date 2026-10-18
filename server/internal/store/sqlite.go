package store

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

var schema string

const timestampLayout = time.RFC3339Nano

type SQLite struct {
	db *sql.DB
}

func Open(ctx context.Context, path string) (*SQLite, error) {
	dsn := fmt.Sprintf(
		"file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)&_pragma=synchronous(NORMAL)",
		path,
	)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite %q: %w", path, err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

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
	createdAt := in.CreatedAt.UTC()
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO images (id, storage_key, title, content_type, width, height, size_bytes, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		in.ID, in.StorageKey, in.Title, in.ContentType, in.Width, in.Height, in.SizeBytes,
		createdAt.Format(timestampLayout),
	)
	if err != nil {
		return Image{}, fmt.Errorf("insert image: %w", err)
	}

	return Image{
		ID:          in.ID,
		Title:       in.Title,
		URL:         URLPathPrefix + in.StorageKey,
		Width:       in.Width,
		Height:      in.Height,
		SizeBytes:   in.SizeBytes,
		ContentType: in.ContentType,
		StorageKey:  in.StorageKey,
		CreatedAt:   createdAt,
	}, nil
}

func (s *SQLite) List(ctx context.Context, limit int) ([]Image, error) {
	if limit <= 0 {
		return nil, errors.New("list: limit must be positive")
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, title, storage_key, content_type, width, height, size_bytes, created_at
		  FROM images
		 ORDER BY id DESC
		 LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("query images: %w", err)
	}
	defer rows.Close()

	images := make([]Image, 0, limit)
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
		if img.CreatedAt, err = parseTimestamp(createdAt); err != nil {
			return nil, err
		}
		img.URL = URLPathPrefix + img.StorageKey
		images = append(images, img)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read images: %w", err)
	}
	return images, nil
}

func parseTimestamp(value string) (time.Time, error) {
	t, err := time.Parse(timestampLayout, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse stored timestamp %q: %w", value, err)
	}
	return t, nil
}
