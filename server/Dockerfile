FROM golang:1.27-alpine AS build
WORKDIR /src

ENV CGO_ENABLED=0 GOOS=linux

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM alpine:3.22 AS runtime

RUN apk add --no-cache ca-certificates tzdata \
	&& adduser -D -H -u 10001 app \
	&& mkdir -p /data && chown app:app /data

USER app
WORKDIR /app

COPY --from=build /out/api /app/api

ENV LISTEN_ADDR=:8080 \
	DATA_DIR=/data \
	DB_PATH=/data/app.db

EXPOSE 8080
VOLUME ["/data"]

ENTRYPOINT ["/app/api"]
