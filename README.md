# Ponte POC — Proof of Concept

> European-first real-time collaboration platform
> "Bridge your team. Own your data."

## What this POC demonstrates

- ✅ User registration & login (JWT auth)
- ✅ Create/join rooms
- ✅ Real-time text chat with message persistence
- ✅ Video/audio calls (WebRTC, group up to ~10)
- ✅ Screen sharing
- ✅ Everything runs in Docker

## What this POC does NOT include (planned for later)

- ❌ E2E encryption (MLS) — Phase 2
- ❌ CRDT collaborative documents — Phase 2
- ❌ Desktop client (Tauri) — Phase 2
- ❌ AI transcription — Phase 3
- ❌ Whiteboard — Phase 3

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend API | Go + Fiber + PostgreSQL |
| Real-time | WebSocket (chat) + WebRTC (video) |
| Signaling | Go WebSocket server |
| SFU | mediasoup (via Node.js worker) |
| Frontend | SvelteKit + TypeScript |
| Auth | JWT (access + refresh tokens) |
| Database | PostgreSQL 16 |
| Deployment | Docker Compose |

## Architecture

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│   SvelteKit  │────▸│  Go Backend  │────▸│ PostgreSQL  │
│   Frontend   │     │  (API + WS)  │     │             │
└──────┬───────┘     └──────┬───────┘     └─────────────┘
       │                    │
       │  WebRTC            │ Signaling
       │                    │
       ▼                    ▼
┌─────────────┐     ┌──────────────┐
│   Browser   │◂───▸│  mediasoup   │
│  (WebRTC)   │     │  SFU Worker  │
└─────────────┘     └──────────────┘
```

## Quick Start

```bash
cd docker
docker compose up -d
```

Then open http://localhost:5173

## Project Structure

```
poc/
├── backend/          # Go API + WebSocket server
│   ├── cmd/          # Entry point
│   ├── internal/     # Business logic
│   │   ├── auth/     # JWT auth
│   │   ├── chat/     # Chat handlers
│   │   ├── room/     # Room management
│   │   ├── user/     # User management
│   │   └── ws/       # WebSocket hub
│   ├── migrations/   # SQL migrations
│   ├── go.mod
│   └── Dockerfile
├── frontend/         # SvelteKit app
│   ├── src/
│   │   ├── lib/      # Components, stores, utils
│   │   └── routes/   # Pages
│   ├── package.json
│   └── Dockerfile
├── signaling/        # mediasoup SFU + signaling
│   ├── src/
│   ├── package.json
│   └── Dockerfile
├── docker/
│   └── docker-compose.yml
└── README.md
```
