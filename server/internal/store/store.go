package store

import "time"

const URLPathPrefix = "/uploads/"

type Image struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Width       int       `json:"width"`
	Height      int       `json:"height"`
	SizeBytes   int64     `json:"sizeBytes"`
	ContentType string    `json:"-"`
	StorageKey  string    `json:"-"`
	CreatedAt   time.Time `json:"createdAt"`
}

type NewImage struct {
	ID          string
	StorageKey  string
	Title       string
	ContentType string
	Width       int
	Height      int
	SizeBytes   int64
	CreatedAt   time.Time
}
