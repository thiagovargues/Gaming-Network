FROM golang:1.22-alpine AS builder

WORKDIR /app
RUN apk add --no-cache gcc musl-dev

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o /bin/server ./cmd/server

FROM alpine:3.20
RUN addgroup -S app && adduser -S app -G app
WORKDIR /app

COPY --from=builder /bin/server /app/server
COPY --from=builder /app/pkg /app/pkg

USER app
EXPOSE 8080
ENV DB_PATH=/app/data/app.db

CMD ["/app/server"]
