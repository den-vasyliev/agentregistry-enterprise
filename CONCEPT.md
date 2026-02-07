# Agent Registry: AI Inventory for Enterprise

Unified discovery layer for agentic infrastructure - automatically indexes MCP servers, agents, skills, and models across clusters. No publish step, no separate marketplace - if it's running, it's in the catalog.

## Problem

Enterprise teams building with AI face a fragmented landscape: agents scattered across clusters, MCP servers deployed by different teams, models running in different environments - with no central view of what exists, what's available for reuse, or how resources relate to each other.

## Architecture

Agent Registry is the inventory layer that sits on top of the AI runtime stack:

```
┌──────────────────────────────────────┐
│  Agent Registry (Inventory)          │  discovery + catalog + UI
├────────────┬───────────┬─────────────┤
│   kagent   │ kgateway  │    llm-d    │  AI runtime
├────────────┴───────────┴─────────────┤
│           Kubernetes                 │  infrastructure
└──────────────────────────────────────┘
```

Agent Registry does NOT replace the runtime - it indexes everything running across the stack into a single catalog.

## Core Differentiators

- **Zero-config discovery** - auto-indexes resources from kagent, kgateway, and llm-d. No manual registration required.
- **Multi-runtime** - MCP servers (kgateway), agents (kagent), models (llm-d) all visible in one UI.
- **Multi-cluster** - discovers resources across clusters using workload identity federation.
- **GitOps-native** - CRDs are the source of truth. All lifecycle management and governance through Git workflows.
- **No marketplace overhead** - if it's running, it's in the catalog. Running = available for reuse.

## Concept

### 1. Governance & Compliance via GitOps

All lifecycle management, governance, and compliance is handled through GitOps workflows:

- **Audit trail** - Git history provides immutable record of who changed what, when, and why
- **Approval workflows** - pull request reviews and branch protection for deployment changes
- **Supply chain trust** - signed commits, image attestation, and policy-as-code in the Git pipeline
- **CI/CD integration** - test, evaluate, and register agents automatically on merge
- **Deploy authorization** - deploy button in UI gated by OIDC authentication with group-based policy access

### 2. Dependency Graph from Runtime

Agent dependencies are already defined in the runtime layer. A kagent Agent CR includes its skills, tools, and MCP server references. Agent Registry surfaces this data from the catalog:

- Agent detail view shows linked skills, tools, and MCP servers
- Impact analysis: understand what breaks if a resource is removed
- Cross-resource navigation between related catalog entries

### 3. Discovery & Search

Simple, fast filtering without database dependencies:

- **Category filters** - filter by resource type, category, tags
- **Tag-based search** - label resources with team, domain, use-case tags
- **Verified identity** - filter by verified organization and verified publisher
- **Deployment status** - filter by running, external, not deployed, failed

### 4. Reuse Metrics from Observability Stack

Usage statistics pulled from the observability layer (Arize Phoenix) rather than built into the registry:

- Trace-based metrics per agent and MCP server (call count, latency, error rate)
- Resource popularity as a trust signal ("used by 12 agents across 4 teams")
- Cost visibility per model from llm-d metrics

### 5. Multi-tenancy & Access Control

- **OIDC authentication** with group-based access control
- **Namespace scoping** - teams see resources in their authorized namespaces
- **Environment isolation** - separate dev, staging, prod catalogs via DiscoveryConfig

### 6. Local Development

Local dev mode mirrors production catalog with a single command:

```bash
make dev
```

Runs the controller locally with auto-discovery from configured clusters, scoped to the developer's personal access. Full catalog available at `localhost:3000`.

## Benefits

| For Platform Teams | For Developers | For Security |
|---|---|---|
| Single view of all AI resources across clusters | Find and reuse existing agents and MCP servers | GitOps audit trail for all changes |
| No manual catalog maintenance (auto-discovery) | One-click deploy to any environment | OIDC + group policies for access control |
| Runtime-agnostic (kagent, kgateway, llm-d) | Local dev with production catalog | Verified publisher identity |
| Multi-cluster visibility with workload identity | No publish step - build, deploy, it's in the catalog | Namespace-based isolation |
