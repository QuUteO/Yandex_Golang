# Stage 1: Builder
FROM golang:1.24 AS builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/notification ./cmd/main.go

# Stage 2: Run
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/bin/notification /app/bin/notification
COPY --from=builder /app/db /app/db

CMD ["/app/bin/notification"]