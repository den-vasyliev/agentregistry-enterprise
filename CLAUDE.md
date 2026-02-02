# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/claude-code) when working with code in this repository.

## Project Overview

Agent Registry is a centralized registry to securely curate, discover, deploy, and manage agentic infrastructure including MCP servers, agents, and skills. It provides governance and control for AI artifacts, enabling developers to build and deploy AI applications with confidence.

## Architecture

The project consists of three main components:

1. **CLI (`arctl`)** - Command-line tool built with Cobra for registry management, resource discovery, installation, and IDE configuration
2. **Server** - HTTP API server built with Gin/Huma that serves the REST API and embedded Next.js UI
3. **Controller** - Kubernetes controller using controller-runtime for CRD-based storage and deployments

### Directory Structure

- `cmd/` - Entry points (cli, server, controller)
- `internal/` - Core implementation
  - `cli/` - CLI command implementations
  - `controller/` - Kubernetes controller logic
  - `httpapi/` - HTTP API handlers
  - `registry/` - Registry business logic and API with embedded UI
  - `runtime/` - Runtime deployment logic
- `pkg/` - Public packages
- `ui/` - Next.js 14 frontend (TypeScript, Tailwind, shadcn/ui)
- `charts/` - Helm charts for Kubernetes deployment
- `api/` - CRD API definitions

### Data Storage

- **Local mode**: SQLite database at `~/.arctl/arctl.db`
- **Kubernetes mode**: CRD-based storage (MCPServerCatalog, AgentCatalog, SkillCatalog)

## Build Commands

```bash
# Full build (UI + CLI)
make build

# Build CLI only (faster iteration)
go build -o bin/arctl cmd/cli/main.go

# Build server
make build-server

# Build controller
make build-controller

# Build UI only
make build-ui

# Development UI with hot reload
make dev-ui

# Run tests
make test

# Run controller tests with envtest
make test-controller

# Format code
make fmt

# Lint code
make lint
```

## Testing

```bash
# Run all Go tests
go test ./...

# Run with coverage
go test -cover ./...

# Run controller tests (requires envtest)
make test-controller

# Run integration tests
go test -ldflags "$(LDFLAGS)" -tags=integration -v ./...
```

## Code Style

- Go: Use `gofmt` and `golangci-lint` (config in `.golangci.yaml`)
- TypeScript/React: Follow Next.js and React best practices
- Commits: Follow conventional commits (feat, fix, docs, style, refactor, test, chore)

## Key Dependencies

- **CLI**: github.com/spf13/cobra, github.com/charmbracelet/bubbletea
- **API**: github.com/danielgtaylor/huma/v2
- **Kubernetes**: sigs.k8s.io/controller-runtime, k8s.io/client-go
- **MCP**: github.com/modelcontextprotocol/go-sdk

## Important Patterns

- The UI is embedded into the Go binary at compile time via `//go:embed`
- The server auto-starts on first CLI command and imports built-in seed data
- CRDs follow the pattern `agentregistry.dev/v1alpha1`
- Version info is injected via LDFLAGS at build time

## Environment

- Go 1.25+
- Node.js 18+ (for UI development)
- Docker (for container builds)
- Kubernetes 1.27+ (for deployment)
