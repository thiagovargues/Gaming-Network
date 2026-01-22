# Gaming-Network

Monorepo MVP: backend Go + SQLite + WebSocket, frontend Next.js.

## Prereqs
- Go 1.22+
- Node 20+
- Docker + docker compose

## Dev (backend)
```
cd backend
export DB_PATH=storage/app.db
export MEDIA_DIR=storage/media
export CORS_ORIGIN=http://localhost:3000

go run ./cmd/api
```

## Dev (frontend)
```
cd frontend
npm install
npm run dev
```

## Docker
```
docker compose up --build
```

## Seed
```
cd backend
go run ./cmd/seed
```

## Tests (backend)
```
cd backend
go test ./test/api
```

## Env vars
- `PORT` (default 8080)
- `DB_PATH` (default storage/app.db)
- `MEDIA_DIR` (default storage/media)
- `COOKIE_NAME` (default sid)
- `COOKIE_SECURE` (true/false)
- `CORS_ORIGIN` (default http://localhost:3000)
- `FRONTEND_URL` (default empty -> fallback to CORS_ORIGIN)
- `GOOGLE_CLIENT_ID` (OAuth)
- `GOOGLE_CLIENT_SECRET` (OAuth)
- `GOOGLE_REDIRECT_URL` (OAuth callback, e.g. http://localhost:8080/api/auth/google/callback)

## Endpoints principaux
- `POST /api/auth/register`
- `POST /api/auth/login`
- `POST /api/auth/logout`
- `GET /api/me`
- `GET /api/feed`
- `POST /api/posts`
- `POST /api/follows/request`
- `GET /api/notifications`
- `GET /api/ws` (WebSocket)

## Quick test (curl)
```
# register
curl -i -X POST http://localhost:8080/api/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@example.com","password":"password123","first_name":"Demo","last_name":"User","dob":"1990-01-01"}'

# login (save cookie)
curl -i -c cookies.txt -X POST http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@example.com","password":"password123"}'

# create post
curl -i -b cookies.txt -X POST http://localhost:8080/api/posts \
  -H 'Content-Type: application/json' \
  -d '{"text":"hello","visibility":"public"}'

# feed
curl -i -b cookies.txt http://localhost:8080/api/feed

# notifications
curl -i -b cookies.txt http://localhost:8080/api/notifications
```
