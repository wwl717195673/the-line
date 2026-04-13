# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

The Line (虾线) is an AI-native workflow collaboration console. It manages workflow processes with node-based execution, agent integration, and deliverable tracking. The UI and documentation are in Chinese. Currently implements a fixed "teacher class transfer" (教师调课) workflow template with 9 nodes.

## Development Commands

### Backend (Go + Gin + GORM)

```bash
cd backend
go run ./cmd/api          # Start dev server on :8080
go build ./cmd/api        # Compile binary
go test ./...             # Run all tests
go test ./internal/service/...  # Run tests for a specific package
```

**Required**: MySQL 8 instance. Quick setup:
```bash
docker run -d --name the-line-mysql -e MYSQL_ROOT_PASSWORD=root -e MYSQL_DATABASE=the_line -p 3306:3306 mysql:8.0
```

**Environment variables** (all have defaults):
- `MYSQL_DSN` — default: `root:root@tcp(127.0.0.1:3306)/the_line?charset=utf8mb4&parseTime=True&loc=Local`
- `APP_PORT` — default: `8080`
- `GIN_MODE` — default: `debug`
- `AUTO_MIGRATE` — default: `true` (auto-creates tables and seeds the fixed template)

### Frontend (React 18 + TypeScript + Vite)

```bash
cd frontend
npm install
npm run dev       # Dev server on :5173, proxies /api and /uploads to :8080
npm run build     # Type-check + production build to dist/
npm run preview   # Serve built artifacts
```

No linter or formatter is configured.

## Architecture

### Monorepo Layout

- `frontend/` — React SPA
- `backend/` — Go REST API
- `docs/` — PRD (`docs/prd/`), tech plans (`docs/tech_plan/`), and tech details (`docs/tech_detail/`)

### Backend — Clean Layered Architecture

`cmd/api/main.go` → `internal/app/` (router/server) → `internal/handler/` → `internal/service/` → `internal/repository/` → `internal/model/`

Supporting packages:
- `internal/config/` — env-based configuration
- `internal/db/` — MySQL setup, auto-migration, seed data
- `internal/domain/` — business constants and the fixed template definition (`fixed_template.go`)
- `internal/dto/` — request/response DTOs
- `internal/response/` — standard JSON response wrapper

Key patterns:
- Constructor-based dependency injection (handler ← service ← repository)
- Actor identity resolved from HTTP headers (`X-Person-ID`, `X-Role-Type`) per request
- GORM datatypes for JSON columns in workflow node input/output schemas
- Fixed template is defined in `internal/domain/fixed_template.go` and synced to DB on every migration via `internal/db/fixed_template_seed.go`

### Frontend Structure

- `src/pages/` — route-level components (Dashboard, Runs, Templates, Deliverables, Resources)
- `src/components/` — shared UI components
- `src/api/` — API client functions organized by domain
- `src/hooks/` — custom data-fetching hooks (useRuns, useRunDetail, useTemplates, etc.)
- `src/types/api.ts` — TypeScript types mirroring backend DTOs
- `src/lib/http.ts` — fetch wrapper that injects actor headers
- `src/lib/actor.ts` — localStorage-based actor/identity management

Key patterns:
- Actor identity stored in localStorage, sent as `X-Person-ID` and `X-Role-Type` headers on every API call
- Custom CSS with no framework; styles in `styles.css`
- React Router v6 for routing (`src/App.tsx`)

### Core Domain Concepts

- **FlowTemplate / FlowTemplateNode**: Workflow blueprint with ordered nodes. Each node has a type (manual/review/notify/archive/execute), input/output schemas, and owner rules.
- **FlowRun / FlowRunNode**: Runtime instances of a template. Runs progress through nodes sequentially.
- **Actor system**: Requests carry person ID + role type. Test actors: ID 1 (leader), ID 2 (middle_office), ID 3 (operation).
- **Deliverable**: Output artifacts attached to runs/nodes, with a review workflow (pending → approved/rejected).
- **Agent**: AI agent configurations that can be bound to "execute" type nodes.
