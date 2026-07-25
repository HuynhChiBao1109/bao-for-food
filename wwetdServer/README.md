# WWETD Server

Golang backend scaffold using Gin, MongoDB, Redis, and WebSocket.

## Architecture

```text
cmd/server              Application composition root
internal/config         Environment config loading
internal/domain         Domain models and domain errors
internal/repository     Database repositories
internal/service        Business logic and interfaces
internal/handler        HTTP/WebSocket handlers
internal/router         Gin route registration and middleware
internal/infrastructure MongoDB, Redis, realtime infrastructure
internal/cache          Cache/pub-sub abstraction
```

Handlers depend on services. Services depend on repository/cache interfaces. Repositories own database access. Infrastructure clients are wired only in `cmd/server/main.go`.

## Run locally

```bash
docker compose up -d
go run ./cmd/server
```

## Debug and hot reload

VS Code:

- Use `Debug Server` to run the Go server with the debugger.
- Run task `server: hot reload` to start the server with Air. When you save a `.go` or `.env` file, Air rebuilds and restarts the server like nodemon.
- If Air is missing, run task `server: install air` once.

Terminal:

```powershell
go install github.com/air-verse/air@latest
.\scripts\dev.ps1
```

Useful endpoints:

```text
GET  /api/v1/health
GET  /api/v1/auth/me
PATCH /api/v1/auth/profile
POST /api/v1/auth/avatar
POST /api/v1/users
GET  /api/v1/users/:id
GET  /api/v1/users?page=1&limit=20
GET  /api/v1/restaurants/nearby?query=quán%20ăn&limit=20
GET  /api/v1/restaurants/today?lat=10.776889&lng=106.700806
GET  /api/v1/ws
```

Example create user:

```bash
curl -X POST http://localhost:8090/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Ada Lovelace","email":"ada@example.com"}'
```

When a user is created, the service publishes a `user.created` event to Redis. The Redis bridge forwards that event to connected WebSocket clients.
