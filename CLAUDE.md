# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

**`AGENTS.md` (repo root) is the authoritative doc** — stack, hexagonal architecture layers, current phase status, known debt, and full command list. Read it before working here; it's kept current by whoever last touched a given area, so prefer it over this summary if they disagree.

## Commands

Monorepo, three independently-run parts. Root scripts `run-api.cmd` (:8080), `run-coach-web.cmd` (:3000), `run-flutter-web.cmd` (:5555) already set the needed env vars for local native Postgres.

```bash
# Backend (api/) — Go, needs Postgres at DATABASE_URL
cd api && goose -dir db/migrations postgres "$DATABASE_URL" up   # after adding a migration
sqlc generate                                                     # after touching db/queries/*.sql
go run ./cmd/api                                                  # server, :8080
go run ./cmd/worker                                               # one-shot AI-suggestion pass, needs ANTHROPIC_API_KEY
go vet ./... && go test ./... && go build ./... && gofmt -l .    # verify before commit
go test ./internal/progression/...                                # single package, e.g.

# Coach panel (web/) — Next.js 16 + TS + Tailwind + shadcn, pnpm
cd web && pnpm dev      # :3000
pnpm lint && pnpm build # verify before commit

# Client app (app/) — Flutter + Riverpod + Drift
cd app && flutter analyze && flutter test  # verify before commit
flutter run -d chrome                       # or web-server; visually confirm UI changes here
```

## Architecture

Multi-tenant gym training app: coach designs/supervises programs, client trains from mobile. A deterministic progression engine drives set-by-set load/rep suggestions; an AI worker layers supervised suggestions on top (never auto-applied — coach approves/rejects).

- **`api/`** — Go, **hexagonal (ports & adapters)**, migrated from a coupled `handler → service → repository` deliberately at the user's request:
  - `internal/domain/` — entities, 11 repository ports, `UnitOfWork`/`TxRepos` for the 3 multi-aggregate transactions (`Assign`, `SyncSessions`, `LogHabit`). No `pgx`/sqlc/anthropic-sdk imports — stdlib + `google/uuid` only.
  - `internal/service/` — use cases; depend only on `domain` ports (verified with grep to have zero `pgx`/db imports).
  - `internal/adapter/postgres/` — implements the ports over `internal/repository/db` (sqlc-generated, the only layer touching `pgtype`); translates `pgx.ErrNoRows` → `domain.ErrNotFound`.
  - `internal/adapter/anthropic/` — implements `domain.SuggestionProposer` (Claude tool-use call for AI suggestions).
  - `internal/transport/http/` — chi router, handlers, DTOs, JWT auth + RBAC middleware.
  - `internal/progression/` — pure, DB-free progression engine (4 strategies, table-driven tests); not yet wired into any service flow.
  - `cmd/api` (HTTP server) and `cmd/worker` (single-pass AI suggestion generator, meant for cron — no embedded scheduler).
- **`web/`** — Next.js App Router coach panel; talks to the API via `src/lib/api.ts` (JWT in `localStorage`, auto-refresh on 401).
- **`app/`** — Flutter client app; go_router with an auth-session guard, Drift for offline persistence on native (in-memory fallback on web).
- **`docs/ARCHITECTURE.md`**, **`docs/DESIGN.md`**, **`docs/PHASES.md`** — decision log, design tokens/theming, phased build plan.

<!-- gitnexus:start -->
# GitNexus — Code Intelligence

This project is indexed by GitNexus as **Myvibesfit** (4531 symbols, 10744 relationships, 359 execution flows).

> Index stale? Run `node .gitnexus/run.cjs analyze --index-only` from the project root — it auto-selects an available runner. No `.gitnexus/run.cjs` yet? Bootstrap with `npx`, `bunx`, or `pnpm dlx` — e.g. `bunx gitnexus@latest analyze` (npm 11 npx crash; #1939).

## Always Do

- **MUST run impact analysis before editing.** Use `impact({target: "symbolName", direction: "upstream"})` (MCP) or `node .gitnexus/run.cjs impact "symbolName" --direction upstream --repo .` (CLI fallback); report callers, processes, and risk. Never substitute grep for graph analysis.
- **MUST analyze graph changes before committing.** Use `detect_changes({scope: "all"})` (MCP) or `node .gitnexus/run.cjs detect-changes --scope all --repo .` (CLI fallback). `partial: true` or `truncated: true` is not a clean check — a zero means unseen, not unaffected; re-run it. For regression review: `detect_changes({scope: "compare", base_ref: "main"})` or `node .gitnexus/run.cjs detect-changes --scope compare --base-ref "main" --repo .`.
- **MUST warn the user** if impact analysis returns HIGH or CRITICAL risk before proceeding with edits.
- **MUST treat `risk: UNKNOWN` as unresolved, not as low.** An empty caller set is not evidence the symbol is unused — it can also mean the callers are not resolvable by the index (plain-object property access, dynamic dispatch, cross-language calls). `impact` pairs `UNKNOWN` with a `riskNote` saying so. Confirm with a text search before treating the symbol as safe to change or delete; do not proceed on the strength of a zero.
- When exploring unfamiliar code, use `query({search_query: "concept"})` to find execution flows instead of grepping. It returns process-grouped results ranked by relevance.
- When you need full context on a specific symbol — callers, callees, which execution flows it participates in — use `context({name: "symbolName"})`.
- For security review, `explain({target: "fileOrSymbol"})` lists taint findings (source→sink flows; needs `analyze --pdg`).

## Never Do

- NEVER edit a function, class, or method before MCP/CLI impact analysis.
- NEVER ignore HIGH or CRITICAL risk warnings from impact analysis, and never read `UNKNOWN` as an all-clear — it means the walk could not answer, which is the one verdict that requires confirming by other means.
- NEVER rename symbols with find-and-replace — use `rename` which understands the call graph.
- NEVER commit before MCP/CLI graph change analysis.

## Resources

| Resource | Use for |
| --- | --- |
| `gitnexus://repo/Myvibesfit/context` | Codebase overview, check index freshness |
| `gitnexus://repo/Myvibesfit/clusters` | All functional areas |
| `gitnexus://repo/Myvibesfit/processes` | All execution flows |
| `gitnexus://repo/Myvibesfit/process/{name}` | Step-by-step execution trace |

## CLI

| Task | Read this skill file |
| --- | --- |
| Understand architecture / "How does X work?" | `.claude/skills/gitnexus-exploring/SKILL.md` |
| Blast radius / "What breaks if I change X?" | `.claude/skills/gitnexus-impact-analysis/SKILL.md` |
| Trace bugs / "Why is X failing?" | `.claude/skills/gitnexus-debugging/SKILL.md` |
| Rename / extract / split / refactor | `.claude/skills/gitnexus-refactoring/SKILL.md` |
| Tools, resources, schema reference | `.claude/skills/gitnexus-guide/SKILL.md` |
| Index, status, clean, wiki CLI commands | `.claude/skills/gitnexus-cli/SKILL.md` |

<!-- gitnexus:end -->
