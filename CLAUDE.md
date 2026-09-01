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
go test ./internal/docs/...                                        # fails if docs/SDD_METHODOLOGY.md drifts from the code

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
  - `internal/domain/` — entities, **13** repository ports plus two integration ports (`SuggestionProposer`, `PushSender`), and `UnitOfWork`/`TxRepos` (6 fields) for the **4** multi-aggregate transactions: `Assign`, `SyncSessions`, `LogHabit`, `AISuggestionService.Review`. No `pgx`/sqlc/anthropic-sdk imports — stdlib + `google/uuid` only.
  - `internal/service/` — use cases; depend only on `domain` ports (verified with grep to have zero `pgx`/db imports).
  - `internal/adapter/postgres/` — implements the ports over `internal/repository/db` (sqlc-generated, the only layer touching `pgtype`); translates `pgx.ErrNoRows` → `domain.ErrNotFound`.
  - `internal/adapter/anthropic/` — implements `domain.SuggestionProposer` (Claude tool-use call for AI suggestions).
  - `internal/adapter/push/` — implements `domain.PushSender`. `fcm.go` speaks the FCM HTTP v1 API over `net/http` + `oauth2/google` (the firebase-admin SDK was rejected: it pulled grpc, protobuf and appengine to send one message). `noop.go` is wired when `FCM_CREDENTIALS_JSON` is unset, so local dev needs no Firebase account.
  - `internal/transport/http/` — chi router, handlers, DTOs, JWT auth + RBAC middleware.
  - `internal/progression/` — pure, DB-free progression engine (4 strategies, table-driven tests). **Wired into `SyncService`** (`internal/service/sync.go:11`): each synced session progresses the next occurrence of that exercise. An `override_source = ai_suggestion` on the target is honoured once and then consumed.
  - `cmd/api` (HTTP server) and `cmd/worker` (single-pass AI suggestion generator, meant for cron — no embedded scheduler).
- **`web/`** — Next.js App Router coach panel; talks to the API via `src/lib/api.ts` (JWT in `localStorage`, auto-refresh on 401).
- **`app/`** — Flutter client app; go_router with an auth-session guard, Drift for offline persistence on native (in-memory fallback on web).
- **`docs/ARCHITECTURE.md`**, **`docs/DESIGN.md`**, **`docs/PHASES.md`** — decision log, design tokens/theming, phased build plan.
- **`docs/SDD_METHODOLOGY.md`** + **`docs/sdd/`** — this repo works spec-first. Non-trivial features get an SDD in `docs/sdd/active/` **before** code (template `01_FEATURE_SDD.md`, or `02_ADR_ARCHITECTURE.md` for a decision without a feature). Three exist: SDD-001 push notifications, SDD-002 progress photos, SDD-003 worker operation.

## When the GitNexus runner fails to start

`.gitnexus/run.cjs` reinstalls its dependencies on every invocation. When the **gitnexus MCP
server is running**, it holds `lbugjs.node` open, the reinstall cannot overwrite it, and every
command dies with `EBUSY` / `EPERM`. The same hung process is usually why the MCP tools return
`CONNECT_TIMEOUT`, so both paths appear broken at once from a single cause.

**Do not kill the MCP server for this.** The package is already extracted on disk; invoke its CLI
directly and no install step runs:

```bash
# find it once, then reuse the path
ls -d "$LOCALAPPDATA"/npm-cache/_npx/*/node_modules/gitnexus 2>/dev/null || \
  ls -d F:/Proyectos/.caches/npm/_npx/*/node_modules/gitnexus

G="F:/Proyectos/.caches/npm/_npx/5e786f48223a616c/node_modules/gitnexus/dist/cli/index.js"
node "$G" analyze --index-only .                                    # path is POSITIONAL here
node "$G" detect-changes --scope all --repo Myvibesfit               # --repo is REQUIRED: 11 repos indexed
node "$G" impact "SymbolName" --direction upstream --repo Myvibesfit
```

Two flag differences from `run.cjs`: `analyze` takes the path as a positional argument (no
`--repo`), and every other command **requires** `--repo Myvibesfit` because this machine has
eleven repositories in one index.

If even that fails, then and only then:

1. **Say so, out loud.** In the reply, and in the commit message if you commit. Never present a
   skipped analysis as a clean one, and never invent a risk level — a fabricated impact analysis
   is worse than an absent one.
2. **Fall back to a text search** (`grep -rn "SymbolName" api/`) and label it as such. It answers
   "who mentions this", not "what breaks".
3. **Judge by what the change touches.** Docs and `_test.go` files change no production symbol —
   say that and proceed. Anything else: ask the user before committing.

<!-- gitnexus:start -->
# GitNexus — Code Intelligence

This project is indexed by GitNexus as **Myvibesfit** (5,691 nodes, 13,887 relationships, 408 execution flows).

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
