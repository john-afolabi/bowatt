package media

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	MaxTitleLength  = 120
	MaxTagsPerImage = 10
	MaxTagLength    = 24
)

var (
	tagSep  = regexp.MustCompile(`[\s_./]+`)
	tagJunk = regexp.MustCompile(`[^a-z0-9-]`)
	tagDash = regexp.MustCompile(`-{2,}`)
)

type InvalidInputError struct {
	Field  string
	Reason string
}

func (e *InvalidInputError) Error() string {
	return fmt.Sprintf("%s %s", e.Field, e.Reason)
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
