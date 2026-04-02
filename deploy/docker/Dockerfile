FROM golang:1.26.1-alpine3.23 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./cmd/server
# [module:cron:start]
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/cron ./cmd/cron
# [module:cron:end]
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/migrate ./cmd/migrate

FROM alpine:3.23.3

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app

COPY --from=builder /app/server ./server
# [module:cron:start]
COPY --from=builder /app/cron ./cron
# [module:cron:end]
COPY --from=builder /app/migrate ./migrate
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/views ./views
COPY deploy/docker/entrypoint.sh ./entrypoint.sh
RUN chmod +x ./entrypoint.sh

USER appuser

HEALTHCHECK --interval=30s --timeout=3s CMD wget -qO- http://localhost:8080/health/live || exit 1

EXPOSE 8080
ENTRYPOINT ["./entrypoint.sh"]
CMD ["./server"]
