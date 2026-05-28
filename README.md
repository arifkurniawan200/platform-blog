# Platform Blog

Blogging platform modern dengan arsitektur microservice — terinspirasi Medium. Backend Go (clean architecture) + frontend Next.js 14.

## ✨ Fitur

### Autentikasi & Profil
- **Register / Login** — JWT access token + refresh token
- **User Profile** — Display name, bio, avatar, stats
- **Edit Profile** — Update display name, bio, avatar URL

### Articles
- **Rich Text Editor** — Tiptap editor dengan toolbar (bold, italic, heading, quote, code block, image, link)
- **Write & Publish** — Buat artikel langsung published
- **Article View** — Halaman baca dengan tampilan bersih
- **List Articles** — Halaman home menampilkan semua artikel terbaru

### Interaksi
- **Claps** 👏 — Apresiasi artikel (multiple claps per user)
- **Comments** 💬 — Komentar di artikel, dengan thread
- **Bookmarks** 🔖 — Simpan artikel favorit untuk dibaca nanti

### Discovery
- **Full-text Search** — Cari artikel berdasarkan judul, subtitle, dan konten (PostgreSQL FTS)
- **Author Stats** — Total articles, total claps, total comments per penulis

### Notifikasi
- **Email Notification** — Notifikasi email saat artikel dikomentari (via Himalaya CLI)

## 🛠 Tech Stack

| Layer | Teknologi |
|-------|-----------|
| **Backend** | Go 1.22, chi-compatible `net/http`, pgx, golang-jwt, go-playground/validator |
| **Frontend** | Next.js 14 (App Router), React 18, TypeScript, Tailwind CSS |
| **Editor** | Tiptap 3 (React) — rich text editor dengan extensions |
| **State** | Zustand (auth store), TanStack React Query (data fetching) |
| **UI** | class-variance-authority, clsx, tailwind-merge, next-themes (dark mode), Lucide icons |
| **Database** | PostgreSQL 16 — GIN index untuk full-text search |
| **Cache** | Redis 7 — refresh token blacklist |
| **Container** | Docker + Docker Compose |
| **Architecture** | Clean architecture per service: **handler → usecase → repository** |

## 🏗 Arsitektur

```text
Browser (Next.js :3000)
        ↓
Gateway (:8080) — CORS + JWT proxy + routing
  ├── Auth Service (:8081) — user, profile, follows
  └── Article Service (:8082) — articles, comments, claps, bookmarks, search
        ↓
  PostgreSQL :5432 + Redis :6379
```

## 🚀 Quick Start

### Prasyarat
- Docker & Docker Compose
- Node.js 18+ (untuk frontend dev)

### Backend

```bash
# Clone repo
git clone https://github.com/arifkurniawan200/platform-blog.git
cd platform-blog

# Start semua service
docker compose up -d

# Run migrasi
docker compose exec auth go run cmd/migrate/main.go up
docker compose exec article go run cmd/migrate/main.go up
```

### Frontend

```bash
cd frontend
cp .env.example .env.local   # atau set NEXT_PUBLIC_API_URL
npm install
npm run dev                   # → http://localhost:3000
```

### Test Cepat

```bash
# Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"johndoe","email":"john@example.com","password":"secret123"}'

# Login (simpan access_token)
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"secret123"}' | jq -r '.data.access_token')

# Create article
curl -X POST http://localhost:8080/api/v1/articles \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"title":"Hello World","subtitle":"My first post","content":"<p>Hello!</p>"}'

# List articles
curl http://localhost:8080/api/v1/articles

# Search
curl "http://localhost:8080/api/v1/search?q=hello"
```

## 📡 API Endpoints

### Auth Service

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/api/v1/auth/register` | Public | Register user baru |
| `POST` | `/api/v1/auth/login` | Public | Login → access + refresh token |
| `POST` | `/api/v1/auth/refresh` | Public | Refresh access token |
| `GET` | `/api/v1/users/{username}` | Public | Public profile user |
| `GET` | `/api/v1/users/me` | JWT | Profile sendiri |
| `PATCH` | `/api/v1/users/me` | JWT | Update profile (name, bio, avatar) |

### Article Service

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/api/v1/articles` | Public | List published articles |
| `GET` | `/api/v1/search?q=` | Public | Full-text search |
| `GET` | `/api/v1/articles/{slug}` | Public | Article detail by slug |
| `GET` | `/api/v1/articles/{slug}/comments` | Public | List comments |
| `POST` | `/api/v1/articles` | JWT | Create article |
| `PUT` | `/api/v1/articles/{slug}` | JWT | Update article |
| `DELETE` | `/api/v1/articles/{slug}` | JWT | Delete article |
| `POST` | `/api/v1/articles/{slug}/clap` | JWT | Clap article |
| `GET` | `/api/v1/articles/{slug}/clap` | Public | Get clap count |
| `POST` | `/api/v1/articles/{slug}/comments` | JWT | Add comment |
| `DELETE` | `/api/v1/articles/{slug}/comments/{id}` | JWT | Delete own comment |
| `POST` | `/api/v1/articles/{slug}/bookmark` | JWT | Bookmark article |
| `DELETE` | `/api/v1/articles/{slug}/bookmark` | JWT | Unbookmark |
| `GET` | `/api/v1/bookmarks` | JWT | List bookmarks user |
| `GET` | `/api/v1/users/{id}/stats` | Public | Author stats |

### Response Format

Semua response mengikuti format:

```json
{
  "data": { ... }
}
```

Collection kosong selalu return `{"data":[]}` (bukan `null`).

## 📂 Project Structure

```text
platform-blog/
├── pkg/                              # Shared packages
│   ├── middleware/jwt.go              # JWT generation, validation, context helpers
│   ├── response/response.go           # Standard JSON response writer
│   └── pagination/pagination.go       # Query param parser (limit, offset)
├── services/
│   ├── auth/                          # Auth microservice
│   │   ├── cmd/auth/main.go           # Entry point, route registration
│   │   ├── cmd/migrate/main.go        # Migration runner
│   │   ├── internal/
│   │   │   ├── domain/                # User, Follow entities
│   │   │   ├── handler/               # HTTP handlers (register, login, profile)
│   │   │   ├── usecase/               # Business logic
│   │   │   ├── repository/            # PostgreSQL repository (pgx)
│   │   │   └── config/                # Environment config loader
│   │   ├── migrations/                # SQL migration files
│   │   └── Dockerfile
│   ├── article/                       # Article microservice
│   │   ├── cmd/article/main.go        # Entry point
│   │   ├── cmd/migrate/main.go        # Migration runner
│   │   ├── internal/
│   │   │   ├── domain/                # Article, Comment, Clap, Bookmark
│   │   │   ├── handler/               # HTTP handlers (CRUD, clap, comment, search)
│   │   │   ├── usecase/               # Business logic
│   │   │   └── repository/            # PostgreSQL repository (pgx)
│   │   ├── migrations/                # SQL migration files (incl. FTS index)
│   │   └── Dockerfile
│   └── gateway/                       # API Gateway
│       ├── cmd/gateway/main.go        # Reverse proxy, CORS, JWT routing
│       └── Dockerfile
├── frontend/                          # Next.js 14 frontend
│   ├── src/
│   │   ├── app/                       # App Router pages
│   │   │   ├── page.tsx               # Home — article list
│   │   │   ├── login/page.tsx         # Login page
│   │   │   ├── register/page.tsx      # Register page
│   │   │   ├── write/page.tsx         # Write article (Tiptap editor)
│   │   │   ├── article/[slug]/page.tsx # Article view
│   │   │   ├── search/page.tsx        # Full-text search
│   │   │   ├── bookmarks/page.tsx     # User bookmarks
│   │   │   └── profile/[username]/page.tsx # User profile + stats
│   │   ├── components/
│   │   │   ├── clap-button.tsx        # Clap interaction
│   │   │   ├── bookmark-button.tsx    # Bookmark toggle
│   │   │   ├── comment-section.tsx    # Comment list + form
│   │   │   └── ui/index.tsx           # Shared UI primitives
│   │   ├── lib/
│   │   │   ├── api.ts                 # API client functions
│   │   │   └── auth.ts                # Auth store (Zustand)
│   │   └── hooks/                     # React hooks
│   ├── package.json
│   └── next.config.js
└── docker-compose.yml
```

## 🔒 Keamanan

- **JWT** access token (15 menit) + refresh token (7 hari)
- **Refresh token blacklist** di Redis — invalidation saat logout
- **CORS middleware** di gateway — configurable origins
- **Password hashing** — bcrypt
- **Protected routes** — middleware JWT di auth & article service

## 🧪 Testing

```bash
# Backend unit tests (via CI — Go compiler via GitHub Actions)
# Test files:
#   services/auth/internal/handler/auth_handler_test.go
#   services/article/internal/handler/article_handler_test.go
```

## 📋 Environment Variables

### Auth Service
| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | — | PostgreSQL connection string |
| `JWT_SECRET` | — | JWT signing secret |
| `REDIS_URL` | `redis:6379` | Redis connection |
| `PORT` | `8080` | HTTP port |

### Article Service
| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | — | PostgreSQL connection string |
| `JWT_SECRET` | — | JWT signing secret |
| `REDIS_URL` | `redis:6379` | Redis connection |
| `PORT` | `8080` | HTTP port |

### Gateway
| Variable | Default | Description |
|----------|---------|-------------|
| `AUTH_SERVICE_URL` | — | Auth service upstream |
| `ARTICLE_SERVICE_URL` | — | Article service upstream |
| `JWT_SECRET` | — | JWT validation secret |
| `PORT` | `8080` | HTTP port |

### Frontend
| Variable | Default | Description |
|----------|---------|-------------|
| `NEXT_PUBLIC_API_URL` | `http://localhost:8080/api/v1` | Backend API base URL |
