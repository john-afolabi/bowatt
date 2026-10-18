package httpapi

import (
	"net/http"
	"strconv"

	"bowatt/internal/media"
	"bowatt/internal/store"

	"github.com/oklog/ulid/v2"
)

var maxPageSize = 100
var defaultPageSize = 24
var filterTagLimit = 5

func (s *Server) handleListImages(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	list := store.ListQuery{Limit: defaultPageSize}

	if raw := query.Get("limit"); raw != "" {
		limit, err := strconv.Atoi(raw)
		if err != nil || limit < 1 {
			s.writeError(w, r, badRequest("invalid_limit", "limit must be a positive integer, got %q", raw))
			return
		}
		if limit > maxPageSize {
			limit = maxPageSize
		}
		list.Limit = limit
	}

	if cursor := query.Get("cursor"); cursor != "" {
		if _, err := ulid.ParseStrict(cursor); err != nil {
			s.writeError(w, r, badRequest("invalid_cursor", "cursor must be an image id"))
			return
		}
		list.Cursor = cursor
	}

	// Repeated tag parameters narrow the feed together (AND), matching the
	// filter semantics of the feed UI.
	rawTags := query["tag"]
	if len(rawTags) > filterTagLimit {
		s.writeError(w, r, badRequest("too_many_tags",
			"at most %d tags can be combined, got %d", filterTagLimit, len(rawTags)))
		return
	}
	for _, raw := range rawTags {
		tag, err := media.NormalizeTag(raw)
		if err != nil {
			s.writeError(w, r, err)
			return
		}
		list.Tags = append(list.Tags, tag)
	}

	items, err := s.store.List(r.Context(), list)
	if err != nil {
		s.writeError(w, r, err)
		return
	}

	page := store.Page{Items: items}
	if len(items) == list.Limit {
		page.NextCursor = items[len(items)-1].ID
	}
	s.writeJSON(w, http.StatusOK, page)
}
