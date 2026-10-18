CREATE TABLE IF NOT EXISTS images (
    id           TEXT    PRIMARY KEY,
    storage_key  TEXT    NOT NULL UNIQUE,
    title        TEXT    NOT NULL,
    content_type TEXT    NOT NULL,
    width        INTEGER NOT NULL,
    height       INTEGER NOT NULL,
    size_bytes   INTEGER NOT NULL,
    created_at   TEXT    NOT NULL
) STRICT;

CREATE TABLE IF NOT EXISTS tags (
    id   INTEGER PRIMARY KEY,
    name TEXT    NOT NULL UNIQUE
) STRICT;

CREATE TABLE IF NOT EXISTS image_tags (
    image_id TEXT    NOT NULL REFERENCES images (id) ON DELETE CASCADE,
    tag_id   INTEGER NOT NULL REFERENCES tags (id) ON DELETE CASCADE,
    PRIMARY KEY (image_id, tag_id)
) STRICT, WITHOUT ROWID;

CREATE INDEX IF NOT EXISTS image_tags_tag_idx ON image_tags (tag_id);