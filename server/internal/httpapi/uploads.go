package httpapi

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"bowatt/internal/media"
	"bowatt/internal/store"

	"github.com/oklog/ulid/v2"
)

func (s *Server) handleCreateUpload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, media.MaxUploadBytes+64<<10)

	file, _, err := r.FormFile("file")
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}
	if errors.Is(err, http.ErrMissingFile) {
		s.writeError(w, r, missingFile())
		return
	}
	if err != nil {
		s.writeError(w, r, bodyReadError(err))
		return
	}
	defer func() { _ = file.Close() }()

	decoded, width, height, err := media.DecodeFormat(file)
	if err != nil {
		s.writeError(w, r, unsupportedMedia("invalid_image", "the file could not be decoded as an image"))
		return
	}

	title, err := media.NormalizeTitle(r.FormValue("title"))
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	tags, err := media.NormalizeTags(formTags(r.MultipartForm))
	if err != nil {
		s.writeError(w, r, err)
		return
	}

	id := ulid.Make().String()
	storageKey := id + decoded.Extension
	finalPath := filepath.Join(s.cfg.UploadsDir(), storageKey)

	// File before metadata: an orphaned file is unreachable, a metadata row
	// without its file would break the feed.
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		s.writeError(w, r, fmt.Errorf("rewind upload: %w", err))
		return
	}
	size, err := copyToUploads(file, finalPath, media.MaxUploadBytes)
	if err != nil {
		s.writeError(w, r, fmt.Errorf("store image file: %w", err))
		return
	}
	if size > media.MaxUploadBytes {
		if removeErr := os.Remove(finalPath); removeErr != nil {
			s.log.Error("could not remove oversized image file",
				slog.String("error", removeErr.Error()), slog.String("key", storageKey))
		}
		s.writeError(w, r, tooLarge("file_too_large",
			"images must be at most %d bytes", media.MaxUploadBytes))
		return
	}

	image, err := s.store.Create(r.Context(), store.NewImage{
		ID:          id,
		StorageKey:  storageKey,
		Title:       title,
		Tags:        tags,
		ContentType: decoded.ContentType,
		Width:       width,
		Height:      height,
		SizeBytes:   size,
		CreatedAt:   time.Now().UTC(),
	})
	if err != nil {
		if removeErr := os.Remove(finalPath); removeErr != nil {
			s.log.Error("could not remove orphaned image file",
				slog.String("error", removeErr.Error()), slog.String("key", storageKey))
		}
		s.writeError(w, r, err)
		return
	}

	w.Header().Set("Location", image.URL)
	s.writeJSON(w, http.StatusCreated, image)
}

func (s *Server) handleServeUpload(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	// Anchoring the key shape here is what makes path traversal structurally
	// impossible instead of a filter to remember.
	if !media.ValidStorageKey(key) {
		s.writeError(w, r, notFound("image_not_found", "no image with that key"))
		return
	}

	path := filepath.Join(s.cfg.UploadsDir(), key)
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		s.writeError(w, r, notFound("image_not_found", "no image with that key"))
		return
	}

	// Stored images are immutable, so they can be cached forever.
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeFile(w, r, path)
}

func bodyReadError(err error) error {
	if maxBytesErr, ok := errors.AsType[*http.MaxBytesError](err); ok {
		return tooLarge("file_too_large", "upload exceeds the %d byte limit", maxBytesErr.Limit)
	}
	return badRequest("invalid_multipart", "could not read the multipart body: %v", err)
}

func missingFile() *apiError {
	return badRequest("missing_file", `a non-empty file part named "file" is required`)
}

func formTags(form *multipart.Form) []string {
	var tags []string
	for _, value := range form.Value["tags"] {
		for part := range strings.SplitSeq(value, ",") {
			if part = strings.TrimSpace(part); part != "" {
				tags = append(tags, part)
			}
		}
	}
	return tags
}

func copyToUploads(src io.Reader, path string, limit int64) (int64, error) {
	dst, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return 0, err
	}

	size, err := io.Copy(dst, io.LimitReader(src, limit+1))
	if err != nil {
		_ = dst.Close()
		return size, err
	}
	return size, dst.Close()
}
