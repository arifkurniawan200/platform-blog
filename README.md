# Platform Blog

Blogging platform modern dengan arsitektur microservice — terinspirasi Medium.

## Arsitektur

```
Traefik / Gateway (JWT + routing)
  ├── Auth Service (register, login, users, follows)
  └── Article Service (articles, comments, tags, claps)
        ↓
    PostgreSQL + Redis
```

## Tech Stack

- **Backend:** Go 1.22, chi-compatible net/http, pgx, golang-jwt
- **Database:** PostgreSQL 16
- **Cache:** Redis 7
- **Container:** Docker + Docker Compose
- **Architecture:** Clean architecture per service (handler → usecase → repository)

## Quick Start

```bash
# Start all services
docker-compose up -d

# Run database migrations
docker-compose exec auth go run cmd/migrate/main.go up
docker-compose exec article go run cmd/migrate/main.go up

# Register a user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"johndoe","email":"john@example.com","password":"secret123"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"secret123"}'

# Create article (with JWT)
curl -X POST http://localhost:8080/api/v1/articles \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access_token>" \
  -d '{"title":"Hello World","content":"My first article"}'
```

## Services

| Service | Port | Description |
|---|---|---|
| Gateway | 8080 | Reverse proxy, JWT validation, routing |
| Auth | 8081 | User registration, login, profiles, follows |
| Article | 8082 | Articles CRUD, tags, comments, claps |

## API Endpoints

### Auth (Public)
- `POST /api/v1/auth/register` — Register new user
- `POST /api/v1/auth/login` — Login, returns JWT tokens

### Articles
- `GET /api/v1/articles` — List published articles (public)
- `GET /api/v1/articles/{slug}` — Get article by slug (public)
- `POST /api/v1/articles` — Create article (auth required)
- `PUT /api/v1/articles/{slug}` — Update article (auth required)
- `DELETE /api/v1/articles/{slug}` — Delete article (auth required)

## Project Structure

```
platform-blog/
├── pkg/                          # Shared packages
│   ├── middleware/jwt.go          # JWT middleware + validation
│   ├── response/response.go       # Standard JSON response helpers
│   └── pagination/pagination.go   # Query param parser
├── services/
│   ├── auth/                      # Auth service
│   │   ├── cmd/auth/main.go
│   │   ├── internal/
│   │   │   ├── domain/            # User, Follow entities
│   │   │   ├── handler/           # HTTP handlers
│   │   │   ├── usecase/           # Business logic
│   │   │   ├── repository/        # PostgreSQL via pgx
│   │   │   └── config/            # Environment config
│   │   ├── migrations/
│   │   └── Dockerfile
│   ├── article/                   # Article service
│   │   ├── cmd/article/main.go
│   │   ├── internal/...
│   │   ├── migrations/
│   │   └── Dockerfile
│   └── gateway/                   # API Gateway
│       ├── cmd/gateway/main.go
│       └── Dockerfile
└── docker-compose.yml
```
