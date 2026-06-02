# Lite Dokploy

A lightweight self-hosted deployment platform for low-resource devices (TV Boxes, mini PCs, VPS, homelabs).

Inspired by [Dokploy](https://github.com/Dokploy/dokploy), but designed for single-user, single-server use with minimal resource consumption.

## Features

- **Git Integration** — Clone and deploy from GitHub/GitLab repositories
- **Git Webhooks** — Auto-deploy on push events
- **Docker Compose** — Deploy, restart, stop, start your compose stacks
- **Environment Variables** — Per-application KEY=value management
- **Traefik Reverse Proxy** — Automatic SSL via Let's Encrypt
- **Status Dashboard** — Application status, deployment history, logs
- **Single Binary** — Go backend, SQLite database, minimal dependencies

## Quick Start

```bash
# Clone
git clone https://github.com/ahsansuryana/lite-dokploy.git
cd lite-dokploy

# Start with Traefik
docker compose up -d

# Open in browser
open http://localhost:3000
```

## Configuration

Copy `.env.example` to `.env` and adjust:

| Variable | Default | Description |
|----------|---------|-------------|
| `DOMAIN` | `localhost` | Your domain for Traefik routing & SSL |
| `TRAEFIK_USER` | `admin` | Basic auth username for Traefik dashboard |
| `TRAEFIK_PASSWORD_HASH` | — | bcrypt hash for Traefik dashboard password |

## Architecture

```
Frontend (React + Vite + TypeScript)
        │
        ▼
REST API (Go + Chi Router)
        │
        ├── SQLite (modernc.org — pure Go, no CGo)
        ├── Docker CLI (docker compose)
        ├── Git (go-git)
        └── Traefik (reverse proxy + SSL)
```

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.23 |
| Frontend | React 18, Vite, TypeScript, Tailwind CSS |
| Database | SQLite (pure Go driver) |
| Proxy | Traefik v3 |
| Container | Docker Compose |

## Development

```bash
# Backend
cd backend
go run ./cmd/server --port 3000 --data ../data --repos ../repositories

# Frontend (separate terminal)
cd frontend
npm install
npm run dev
```

## License

MIT
