# Bowatt Assessment

Anonymous image uploads with a live feed.

Images can be uploaded anonymously with Title, Tags, and every open browser receives new uploads over a websocket without refreshing.

The API lives in `server/`, the web app in `client/`.

- **`client/`** — React + TypeScript SPA built with [Vite+](https://viteplus.dev) (Vite/Rolldown/Vitest/Oxlint/Oxfmt), styled with Tailwind CSS v4 and shadcn/ui components.
- **`server/`** — Go 1.27 API that stores image files on disk, keeps metadata in SQLite, and fans new uploads out to websocket subscribers.


## Quick start (Docker)

```bash
docker compose up --build
```


## Local development

```bash
cd server && go run ./cmd/api
```

OR

With [Air](https://github.com/air-verse/air) for live reloads.

```bash
cd server && go run github.com/air-verse/air@latest
```


```bash
cd client && npm install && npm run dev
```


## API

All responses are JSON. Errors always use one envelope.

```json
{ "error": { "code": "invalid_title", "message": "title must not be empty" } }
```

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/api/images` | One page of the feed, newest first. Query: `limit` (1–100, default 24), `cursor` (an image id from `nextCursor`), `tag` (repeatable, combined with **AND**). |
| `POST` | `/uploads` | `multipart/form-data` with `file`, `title`, and `tags` (comma separated). Returns `201` with the stored image and a `Location` header. |
| `GET` | `/uploads/{key}` | The stored image bytes, `Cache-Control: immutable`. |
| `GET` | `/api/ws` | Websocket subscription for new uploads. |
| `GET` | `/healthz` | Liveness: `200` while the database answers, `503` otherwise. |


### Websocket protocol

Server to client only; each frame is `{ "type": ..., "data": ... }`:

```json
{
  "type": "image.created",
  "data": {
    "id": "01J…",
    "title": "Sunset",
    "tags": [
      "sunset"
    ],
    "url": "/uploads/01J….png",
    "width": 800,
    "height": 600,
    "sizeBytes": 204800,
    "createdAt": "2026-10-17T07:00:00Z"
  }
}
```


## Data model

Metadata lives in SQLite (`$DB_PATH`, default `./data/app.db`). The schema in `server/internal/store/schema.sql` is applied on startup. Image bytes are at `$DATA_DIR/uploads/{storage_key}`.

An image can have many tags, and a tag is shared by many images. `image_tags` is the join table

```
images ──< image_tags >── tags
```

### `images`

| Column | Type | Meaning |
| --- | --- | --- |
| `id` | `TEXT` primary key | ULID. Feed order is `ORDER BY id DESC`; the cursor is `WHERE id < ?`. |
| `storage_key` | `TEXT` unique | `{id}` plus `.jpg`, `.png`, or `.gif`. The public URL is `/uploads/{storage_key}`. |
| `title` | `TEXT` | Normalised title, 1–120 characters. |
| `content_type` | `TEXT` | `image/jpeg`, `image/png`, or `image/gif`. |
| `width`, `height` | `INTEGER` | Dimensions from the decoded header. |
| `size_bytes` | `INTEGER` | Stored file size. |
| `created_at` | `TEXT` | UTC timestamp in nanoseconds. Returned as `createdAt`. |

### `tags`

| Column | Type | Meaning |
| --- | --- | --- |
| `id` | `INTEGER` primary key | Surrogate key. Callers only ever use `name`. |
| `name` | `TEXT` unique | Normalised tag, at most 24 characters. |

### `image_tags`

| Column | Type | Meaning |
| --- | --- | --- |
| `image_id` | `TEXT` | References `images(id)` `ON DELETE CASCADE`. |
| `tag_id` | `INTEGER` | References `tags(id)` `ON DELETE CASCADE`. |


## Possible next steps/improvements
- Local disk used for storing the image at the moment. S3 compatible storage would be better for this as this allows image blob upload to be entirely client side.
- Images are currently stored unscaled and unstripped. Thumbnails or different optimized version of the images can be generated with EXIF metadata stripped.
- New uploads are broadcast from an in-memory hub, so only browsers connected to that process see them. More than one node needs a shared pub/sub (Redis, for example) so each process can forward the event to its own sockets.
