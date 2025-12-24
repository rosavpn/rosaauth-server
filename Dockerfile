FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# runtime
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/web ./web
COPY --from=builder /app/internal/database/migrations ./internal/database/migrations

EXPOSE 3000

ENTRYPOINT ["./server"]
