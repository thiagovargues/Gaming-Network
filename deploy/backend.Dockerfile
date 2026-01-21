FROM golang:1.22-alpine AS builder

WORKDIR /app
RUN apk add --no-cache gcc musl-dev

WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o /bin/server ./cmd/api

FROM alpine:3.20
RUN addgroup -S app && adduser -S app -G app
WORKDIR /app

COPY --from=builder /bin/server /app/server
COPY --from=builder /app/backend/migrations /app/migrations

USER app
EXPOSE 8080
ENV DB_PATH=/app/storage/app.db

CMD ["/app/server"]
