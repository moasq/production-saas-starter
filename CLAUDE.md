# CLAUDE.md

Guidance for Claude Code when working with this monorepo.

## Monorepo Structure

Production SaaS Starter - Enterprise-grade B2B SaaS boilerplate with Next.js 16 frontend and Go backend.

```
production-saas-starter/
├── go-b2b-starter/        # Go backend (API, auth, billing, AI/RAG)
├── next_b2b_starter/      # Next.js frontend (React 19, TypeScript)
├── setup.sh               # One-line setup script
├── DEVELOPMENT.md         # Development workflow guide
└── README.md              # Project overview
```

## Quick Navigation

**Working on Backend?**
- See: `go-b2b-starter/.claude/CLAUDE.md`
- Key commands: `make server`, `make dev`, `make sqlc`, `make test`
- Tech: Go, Gin, PostgreSQL, SQLC, Stytch, Polar.sh

**Working on Frontend?**
- See: `next_b2b_starter/.claude/CLAUDE.md`
- Key commands: `pnpm dev`, `pnpm build`, `pnpm lint`
- Tech: Next.js 16, React 19, TypeScript, Tailwind, shadcn/ui, TanStack Query

## Getting Started

```bash
# One-line setup
chmod +x setup.sh && ./setup.sh

# Or manual start:

# Backend
cd go-b2b-starter
make run-deps          # Start PostgreSQL
make migrateup         # Apply migrations
make dev               # Run with hot reload

# Frontend
cd next_b2b_starter
pnpm install
pnpm dev               # Start Next.js dev server
```

## Architecture Overview

| Component | Technology | Purpose |
|-----------|------------|---------|
| **Authentication** | Stytch B2B | Magic links, RBAC, multi-tenant |
| **Billing** | Polar.sh | Subscriptions, webhooks, usage metering |
| **Database** | PostgreSQL + SQLC | Type-safe queries, pgvector |
| **Backend** | Go + Gin | Clean Architecture, DI (uber-go/dig) |
| **Frontend** | Next.js 16 + React 19 | Server Actions, TanStack Query |
| **Styling** | Tailwind + shadcn/ui | Consistent design system |

## Key Patterns

### Backend (Go)
- **Clean Architecture**: domain -> app -> infra layers
- **Repository Pattern**: Domain interfaces implemented by SQLC-backed repos
- **Event Bus**: Loose coupling between modules
- **RBAC Middleware**: Permission-based route protection

### Frontend (Next.js)
- **Server Actions**: For mutations with auth/permission guards
- **TanStack Query**: Server state management with caching
- **Repository Pattern**: API client abstractions
- **Type Safety**: Strict TypeScript throughout

## Documentation

- `DEVELOPMENT.md` - Development workflow and setup
- `SETUP.md` - Initial project configuration
- `go-b2b-starter/docs/` - Backend documentation
- `next_b2b_starter/docs/` - Frontend documentation
