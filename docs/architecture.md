# Architecture

## Overview

Home Cooking is a personal home cooking ecosystem — a monorepo containing a shared
Go API backend and multiple purpose-built client applications.

The core idea: separate apps optimized for their context (phone for shopping,
tablet for pantry, desktop/emacs for planning), all backed by a single API and
shared data model.

## Repository Structure

```
homeCooking/
├── proto/                  # Protobuf definitions (source of truth for types)
├── api/                    # Go backend (single server, multiple domains)
│   ├── cmd/server/
│   ├── internal/
│   │   ├── config/
│   │   ├── store/
│   │   ├── middleware/
│   │   ├── router/
│   │   └── <domain>/      # e.g. shopping/, recipes/, pantry/
│   └── migrations/
├── clients/
│   ├── android/            # KMP — shopping list (phone)
│   ├── cooking/            # KMP — cooking assistant (phone)
│   ├── pantry/             # KMP — tablet app
│   ├── desktop/            # Compose Desktop — shopping/planning
│   └── emacs/              # Elisp package
├── docs/
└── .github/workflows/
```

## Backend

- **Language:** Go
- **Framework:** stdlib `net/http` (no external router framework)
- **Database:** PostgreSQL via `pgx` + `sqlx`
- **Migrations:** `golang-migrate` with numbered SQL files
- **Architecture:** Single server process with domain-separated packages
  under `internal/`. One process serving all domains is preferable on a
  resource-constrained server vs. multiple microservices.

## Type Sharing

Protobuf definitions in `proto/` are the single source of truth for API
contracts. `protoc` generates:
- Go structs for the API server
- Kotlin data classes for KMP/Android clients
- (Future) Any other language as needed

This replaces the KMP shared module approach — protobufs give the same type
safety across language boundaries without coupling client and server to the
same language.

## Clients

Each client is an independent application that communicates with the API.
Clients share generated protobuf types but are otherwise decoupled from
each other.

- **Android apps** (shopping, cooking): Kotlin + Compose, offline-first with
  local SQLDelight databases and bidirectional sync
- **Tablet app** (pantry): KMP, optimized for larger screen inventory management
- **Desktop app**: Compose Desktop for shopping list management, deal linking,
  and planning workflows that benefit from keyboard/mouse input
- **Emacs client**: Elisp package interfacing with the API, enabling org-mode
  based list authoring that exports to the backend

## Offline-First Sync

Mobile clients use an offline-first architecture:
- Local SQLDelight database as primary data store
- Dirty/deleted flags for change tracking
- Bidirectional sync engine: push local changes, pull server state
- Client-side conflict resolution (local changes win if newer)

## Deployment

- **Container:** Multi-stage Docker build producing a static Go binary on Alpine
- **Registry:** GHCR (GitHub Container Registry)
- **CI/CD:** GitHub Actions — on release, builds image, pushes to GHCR,
  deploys via SSH to NixOS server
- **Server:** NixOS with OCI containers, Traefik reverse proxy, PostgreSQL + PgBouncer
- See [deployment.md](deployment.md) for operational details
