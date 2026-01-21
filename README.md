# Gaming-Network

Backend Go + frontend (future) in a split architecture with Docker support.

## Backend
- Entry point: `backend/cmd/api/main.go`
- Migrations: `backend/migrations/sqlite`

## Dev (backend)
```
cd backend
go run ./cmd/api
```

## Docker
```
docker compose up --build
```
