package media

import (
	"fmt"
	"image"
	"io"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"
)

const (
	MaxUploadBytes  = 10 * 1024 * 1024
	MaxTitleLength  = 120
	MaxTagsPerImage = 10
	MaxTagLength    = 24
)

var (
	tagSep  = regexp.MustCompile(`[\s_./]+`)
	tagJunk = regexp.MustCompile(`[^a-z0-9-]`)
	tagDash = regexp.MustCompile(`-{2,}`)
)

var storageKeyPattern = regexp.MustCompile(`^[0-9A-HJKMNP-TV-Z]{26}\.(jpg|png|gif)$`)

type Format struct {
	ContentType string
	Extension   string
}

var formats = map[string]Format{
	"image/jpeg": {ContentType: "image/jpeg", Extension: ".jpg"},
	"image/png":  {ContentType: "image/png", Extension: ".png"},
	"image/gif":  {ContentType: "image/gif", Extension: ".gif"},
}

type InvalidInputError struct {
	Field  string
	Reason string
}

func (e *InvalidInputError) Error() string {
	return fmt.Sprintf("%s %s", e.Field, e.Reason)
}

func DecodeFormat(r io.Reader) (Format, int, int, error) {
	cfg, name, err := image.DecodeConfig(r)
	if err != nil {
		return Format{}, 0, 0, fmt.Errorf("decode image header: %w", err)
	}
	format, ok := formats["image/"+name]
	if !ok {
		return Format{}, 0, 0, fmt.Errorf("decoded format %q is not supported", name)
	}
	if cfg.Width <= 0 || cfg.Height <= 0 {
		return Format{}, 0, 0, fmt.Errorf("decoded image has unusable dimensions %dx%d", cfg.Width, cfg.Height)
	}
	if cfg.Width > 20000 || cfg.Height > 20000 {
		return Format{}, 0, 0, fmt.Errorf("image dimensions %dx%d exceed the 20000px limit", cfg.Width, cfg.Height)
	}
	return format, cfg.Width, cfg.Height, nil
}

func ValidStorageKey(key string) bool {
	return storageKeyPattern.MatchString(key)
}

func NormalizeTitle(raw string) (string, error) {
	title := strings.Join(strings.Fields(raw), " ")
	if title == "" {
		return "", &InvalidInputError{Field: "title", Reason: "must not be empty"}
	}
	if n := utf8.RuneCountInString(title); n > MaxTitleLength {
		return "", &InvalidInputError{
			Field:  "title",
			Reason: fmt.Sprintf("must be at most %d characters, got %d", MaxTitleLength, n),
		}
	}
	return title, nil
}

func NormalizeTag(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", &InvalidInputError{Field: "tag", Reason: "must not be empty"}
	}

	tag := tagSep.ReplaceAllString(strings.ToLower(trimmed), "-")
	tag = tagJunk.ReplaceAllString(tag, "")
	tag = strings.Trim(tagDash.ReplaceAllString(tag, "-"), "-")
	if tag == "" {
		return "", &InvalidInputError{
			Field:  "tag",
			Reason: fmt.Sprintf("%q has no letters or digits to build a tag from", raw),
		}
	}
	if len(tag) > MaxTagLength {
		return "", &InvalidInputError{
			Field:  "tag",
			Reason: fmt.Sprintf("must be at most %d characters, got %d", MaxTagLength, len(tag)),
		}
	}
	return tag, nil
}

func NormalizeTags(raw []string) ([]string, error) {
	tags := make([]string, 0, len(raw))
	for _, candidate := range raw {
		tag, err := NormalizeTag(candidate)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	slices.Sort(tags)
	tags = slices.Compact(tags)
	if len(tags) > MaxTagsPerImage {
		return nil, &InvalidInputError{
			Field:  "tags",
			Reason: fmt.Sprintf("at most %d tags per image, got %d", MaxTagsPerImage, len(tags)),
		}
	}
	return tags, nil
}
