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
infrastructure          MongoDB, Redis, realtime infrastructure
integrations            Piso and MapVina API clients
internal/strategy       Map provider selection strategy
internal/cache          Cache/pub-sub abstraction
```

Handlers depend on services. Services depend on repository/cache interfaces. Repositories own database access. Infrastructure clients are wired only in `cmd/server/main.go`.

## Map provider

Restaurant services use a provider-neutral map client. Select the implementation
through `.env` without changing service code:

```env
MAP_PROVIDER=piso
```

Supported values are `piso` and `vinamap`. Configure the corresponding API key:

```env
PISO_API_KEY=your-piso-key

# Or use MapVina
MAP_PROVIDER=vinamap
VINA_API_KEY=your-mapvina-key
VINA_BASE_URL=https://maps.mapvina.com
VINA_RADIUS=5000
VINA_PLACE_TYPE=restaurant
```

The strategy selects one client at startup. The MapVina adapter normalizes its
Google-compatible response to the restaurant schema consumed by the mobile app.

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
