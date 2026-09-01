# Myvibesfit

App de entrenamiento multi-tenant para gimnasios: coach diseña y supervisa, cliente
entrena desde el móvil. Motor de progresión determinista + sugerencias de IA
supervisadas por el coach.

## Documentación

- [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md) — decisiones, capas, motor de
  progresión, IA, sincronización offline, endpoints
- [docs/DESIGN.md](./docs/DESIGN.md) — tokens de color, tipografía, componentes,
  mapeo a Flutter `ThemeExtension`
- [docs/PHASES.md](./docs/PHASES.md) — plan de desarrollo por fases
- [docs/DEPLOY.md](./docs/DEPLOY.md) — puesta en marcha en un VPS con docker compose: orden de arranque, lo que hay que cambiar del `.env` y qué queda cerrado a propósito
- [docs/SDD_METHODOLOGY.md](./docs/SDD_METHODOLOGY.md) — **metodología SDD actualizada** (5 fases con Definition of Ready/Done, gates obligatorios GitNexus, cobertura de tests, tooling)
- [docs/sdd/README.md](./docs/sdd/README.md) — catálogo de plantillas + SDDs activos (SDD-001/002/003)

## Estado actual

Fases 0-10 de [PHASES.md](./docs/PHASES.md) construidas, **más el cierre de
la cadena de producto** (ver abajo). Lo que falta para producción real
requiere un VPS.

**Este repo trabaja spec-first (2026-08-31).** Una feature no trivial se especifica en
`docs/sdd/active/` antes de escribirla. La metodología define 5 fases con Definition of
Ready/Done y gates de GitNexus.

Los gates se revisaron contra el código el mismo día, y **los que el repo no podía pasar
bajaron a objetivo** en vez de quedar como requisitos que todo el mundo ignora: la auditoría
cubre 4 mutaciones, no 8, y los tests de handler son meta (hoy hay cero; el gate real es el
middleware). Lo aspiracional vive en §6 con su trigger. Ver
[docs/SDD_METHODOLOGY.md](./docs/SDD_METHODOLOGY.md) y [docs/sdd/README.md](./docs/sdd/README.md).

`api/internal/docs` contrasta tres afirmaciones de la metodología con el código en cada
`go test ./...`. Existe porque el documento ya afirmó tres veces cosas que el código desmentía.

### La cadena de producto ya está cerrada

Hasta hace poco el backend estaba muy por delante de sus dos UIs: **27 de 47
endpoints no tenían ningún consumidor** y el producto no se podía usar de
punta a punta sin SQL manual (nadie podía ser coach, no había UI de
programas, y `coach_client` no tenía ni un INSERT en todo el repo, así que el
dashboard del coach siempre estaba vacío). Eso ya no es así:

- `POST /v1/orgs` — crear gimnasio desde la UI; el creador queda `owner` y se
  le reemite el JWT con `org_id`+`role` (mismo patrón que `JoinOrganization`).
- `GET /v1/org`, `GET /v1/org/members`, `PATCH /v1/org/members/{id}/role` —
  gestión de miembros; solo `owner` cambia roles, y el endpoint **no** permite
  otorgar `owner`. Respeta el índice `membership_one_active_role_per_org_uq`
  de la migración `0004` (es un UPDATE del rol, nunca una segunda fila activa).
- **El vínculo coach↔cliente se crea solo**: `AssignmentService.Assign` llama
  a `repos.Coaches.LinkClient` dentro del mismo `uow.Execute` que crea la
  asignación. Por eso `TxRepos` ahora tiene un 5º campo (`Coaches`).
- **Constructor de programas completo en el panel** (`/programs`,
  `/programs/{id}`, `/programs/{id}/workouts/{workoutId}`): crear, editar,
  grilla semana×día, ejercicios con series/reps/RPE/descanso, publicar,
  archivar y asignar. Consume los 20 endpoints de programas que antes nadie
  tocaba.
- `GET|PUT /v1/me/profile` — onboarding del cliente (`client_profile`, que
  tenía cero queries): objetivo, experiencia, días/minutos, equipamiento,
  limitaciones y sistema de unidades. **Alimenta el resumen que se le manda a
  Claude** (`AssignmentSummary` ahora incluye ese contexto).
- `GET /v1/habits/logs?date=` — los hábitos ya marcados en una fecha; sin
  esto la app no podía saber cuáles estaban cumplidos y duplicaba logs.

Verificado end-to-end **desde el navegador real** (no solo curl): registrar →
crear gimnasio → crear/publicar programa → cliente se une con el código →
asignar → el cliente aparece en el dashboard, y `GET /v1/me/assignment/current`
devuelve el plan (antes daba 404 siempre).

**Backend (`api/`)** — verificado end-to-end contra Postgres real (migraciones +
seed + servidor + curl):
- Esquema de BD (`api/db/migrations/`, goose): `0001_init` (tenencia/identidad),
  `0002_training` (catálogo/programas), `0003_execution` (sesiones/hábitos/
  gamificación/IA).
- Auth JWT (access + refresh con rotación), vinculación a gimnasio por código,
  RBAC por rol, catálogo de ejercicios (40 semilla), CRUD de programas/días/
  ejercicios del coach, motor de progresión puro en `internal/progression/`
  (4 estrategias con tests, sin BD, 4 reglas semilla), asignación con copia
  transaccional programa→cliente independiente de la plantilla.
- Sync de entrenamientos (`POST /v1/sync/sessions`, idempotente por
  `client_local_id`, detecta PRs automáticamente), progreso (`/v1/progress/*`,
  métricas corporales), hábitos (`/v1/habits/*`, suscripción + check-in) y
  gamificación (XP, rachas, logros — 5 hábitos y 6 logros semilla) en
  `internal/service/{sync,progress,habit,gamification}.go`.
- CORS habilitado (`github.com/go-chi/cors`) para que Flutter Web y el panel
  Next.js llamen a la API desde el navegador en dev.
- `GET /v1/coach/clients` (fase 8): overview por cliente — plan activo,
  última sesión, racha, PRs de los últimos 14 días y `needs_attention`
  (true si hay plan activo y pasaron >3 días sin entrenar). Un solo query en
  `db/queries/coach.sql` con subqueries escalares envueltas en `COALESCE`
  para forzar tipos no-nulos seguros de escanear (sqlc no infiere
  nullability a través de casts/subqueries en columnas de `LEFT JOIN`).
  Lógica de "necesita atención" vive en `internal/service/coach.go`, no en
  SQL. Verificado con coach+2 clientes reales vía curl (con/sin plan,
  sesión vieja/reciente, racha, PR) y RBAC (403 rol client, 401 sin token).

**App Flutter (`app/`)** — `flutter analyze` sin hallazgos, `flutter test` 4/4:
- Tema desde `docs/DESIGN.md` vía `ThemeExtension<AppColors>`, Riverpod,
  go_router con guard de sesión, cliente dio con refresh automático de token.
- Persistencia offline con Drift (`core/storage/`) en plataformas nativas —
  imports condicionales, con un store en memoria como fallback en web (ver
  nota `ponytail` en `workout_store.dart`: SQLite-WASM no aporta nada a una
  vista previa web de una app que es mobile-first).
- Pantallas: login/registro, home (racha, XP, plan del día), entrenamiento
  activo (registro de series + temporizador de descanso con haptics), cola de
  sync que se dispara sola al terminar un entrenamiento, progreso (records +
  gráfica con `fl_chart`), hábitos (check-in diario), logros.
- Compilado y corrido en `flutter run -d web-server` — carga, el guard de auth
  redirige correctamente, cero errores de CORS. **No se pudo verificar
  visualmente ni interactuar con la UI en esta sesión**: el panel del
  navegador no compone frames en este entorno (limitación de la herramienta).
  Corre `flutter run -d chrome` (o abre `http://localhost:5555` con el backend
  arriba) para verlo de verdad antes de dar la UI por buena.

**Panel del coach (`web/`)** — Next.js 16 (App Router, Turbopack) + TS +
Tailwind v4 + shadcn/ui, `pnpm lint` y `pnpm build` limpios:
- `src/lib/api.ts`: login, refresh automático en 401 (el access token dura
  15 min), tokens en `localStorage`.
- `/login`: formulario email/password contra `POST /v1/auth/login`.
- `/`: tabla de clientes del coach — plan, último entreno, racha, PRs y
  badge "Necesita atención" — consumiendo `GET /v1/coach/clients`; redirige
  a `/login` sin sesión.
- **Verificado de verdad en el navegador** (a diferencia de la app Flutter,
  aquí el árbol de accesibilidad sí se pudo leer): login real contra la API,
  tabla renderizando los datos correctos para un cliente con alerta y otro
  al día, logout, cero errores de consola, preflight CORS en 200.
- Falta (fuera de alcance de fase 8): invitar/gestionar coaches desde el
  panel — hoy el rol `coach` solo se otorga a mano en BD, no hay UI para
  editar programas (eso seguía probándose solo por API desde fase 3/4).
- `/suggestions` (fase 9): tarjetas de aprobación — kind traducido a
  español, detalle del payload, razón, badge de confianza, botones
  Aprobar/Rechazar del mismo tamaño (sin sesgo visual, per `docs/DESIGN.md`
  §5). Verificado en el navegador: aprobar saca la tarjeta de la lista y
  persiste en BD (`status=approved`, `reviewed_by` seteado).

**Fase 9 (IA supervisada) — completa con una salvedad importante:**
- `cmd/worker`: por cada asignación activa arma un `AssignmentSummary`
  (sesiones completadas, streak, adherencia a hábitos, PRs recientes,
  tendencia de peso/RPE por ejercicio, más el perfil del cliente —
  `internal/adapter/anthropic/suggester.go` + queries en
  `db/queries/ai_suggestion.sql`) y le pide a Claude una
  sugerencia vía tool use forzado (`propose_suggestion`), validada en Go
  antes de guardarse como `pending`: enum de `suggestion_kind`, rango de
  `confidence`, y **los topes de magnitud que `docs/ARCHITECTURE.md` §4
  documentaba sin implementar** (>10% de salto de carga, >30% de volumen →
  se descarta; tests en `domain/ai_suggestion_test.go`).
  Nunca se auto-aplica — solo el coach puede aprobar/rechazar desde
  `/v1/ai-suggestions/{id}/approve|reject`.
- **Verificado el pipeline completo excepto la llamada real a Claude**: corrí
  el worker contra datos sembrados reales — encontró la asignación activa y
  construyó el resumen sin errores en ninguna de las queries; falló *solo*
  en `client.Messages.New` porque este entorno no tiene `ANTHROPIC_API_KEY`
  ni un perfil de `ant auth login`. El flujo de aprobación (listar,
  aprobar, rechazar, RBAC) se probó de punta a punta insertando sugerencias
  de prueba directamente en `ai_suggestion` para simular lo que el worker
  produciría. **Falta que el usuario configure `ANTHROPIC_API_KEY`** (o
  `ant auth login`) para ver una sugerencia real generada por Claude.
- No hay cron configurado — `cmd/worker` es un binario de una sola pasada,
  pensado para invocarse por `cron`/Task Scheduler fuera de este repo.

**Fase 10 (endurecimiento) — hecho lo que no depende de un VPS:**
- Rate limiting (`github.com/go-chi/httprate`): 300 req/min por IP global,
  10 req/min por IP en `/auth/register|login|refresh` — `router.go`.
- Auditoría: tabla `audit_log` (ya existía en el esquema) ahora se escribe
  de verdad. `internal/service/audit.go` define `AuditLogger.Log(...)`
  (falla local con `log.Printf`, nunca tumba la mutación que audita) e
  inyectado en `AssignmentService`, `ProgramService`, `AISuggestionService`.
  Se registra: `assignment.assign`, `assignment.cancel`,
  `program.set_status` (cubre publish/archive vía el mismo método), y
  `ai_suggestion.review` (cubre approve/reject) — con `actor_user_id` y
  `org_id` reales desde el contexto de auth. Verificado con
  `go build/vet/test` limpios; no se probó con curl porque esta sesión
  corrió sin levantar el servidor (ver nota de entorno).
- **Fuera de alcance en este entorno** (requieren VPS o cuentas reales):
  backups automáticos de Postgres, monitoreo del VPS, prueba de carga de 100
  usuarios concurrentes, y el cron del worker (especificado en
  [SDD-003](docs/sdd/active/SDD-003-worker-operation.md)).
  Las **push notifications ya no están aquí**: el backend está construido y
  verificado en vivo (queries, repositorio, `PushSender`, adapter FCM sobre
  HTTP v1, dos endpoints y el aviso al asignar un programa). Lo que falta es
  la integración Flutter y la entrega real, que sí necesita un proyecto
  Firebase y un dispositivo físico. Ver
  [SDD-001](docs/sdd/active/SDD-001-push-notifications.md).

## Tests

- **Go**: `internal/progression` (4 estrategias, table-driven),
  `internal/platform` (jwt/hash), `internal/service` (`aggregateTrends`), y
  `internal/domain` — `GamificationService` con un fake del port
  (acumulación de XP y nivel, los 4 casos de borde de `BumpStreak`, y que
  `CheckAchievements` no redesbloquea), `Suggestion.Validate` (incluidos los
  topes de magnitud) y `NeedsAttention`.
- **Cubierto desde entonces**: `SyncService.SyncSessions` (6 tests con un
  `UnitOfWork` falso, incluida la progresión y el no-repetir-recompensas en un
  reenvío), `NotificationService` y el middleware de auth/RBAC
  (`internal/transport/http/middleware`, 21 casos: firma forjada, token
  expirado, JWT sin `org_id`, `RequireRole` con lista vacía).
- **Sin cubrir todavía (requeridos por SDD actualizado):** `AssignmentService.Assign`, **handlers HTTP** (1 test/handler: happy path + 400 + 401 + 403), Flutter: router/redirects, `AuthController`, cola de sync, `WorkoutStore`.
- **Cobertura objetivo (SDD §2 Fase 3)** — lo marcado como gate bloquea; el resto es meta:
  - Dominio puro (progresión, gamificación, validación IA): 100% funciones públicas
  - Services: caso feliz + 1 error + 1 edge case por método público
  - Handlers HTTP: **objetivo, no gate** — hoy hay cero; el gate real es el middleware
  - Flutter: router/redirects, `AuthController`, cola de sync, `WorkoutStore`

- **Docs**: `internal/docs` contrasta tres marcadores de
  `docs/SDD_METHODOLOGY.md` (acciones auditadas, campos de `TxRepos`, tablas
  sin uso) contra el código. Si la metodología se desactualiza,
  `go test ./...` falla con el arreglo en el mensaje. No es ceremonia: ese
  documento ya afirmó tres veces cosas que el código desmentía.

## Deuda conocida (diagnosticada, no arreglada)

Auditorías previas (go-reviewer, architect, database-reviewer) dejaron esto
documentado — no hace falta volver a auditar, están confirmados:

- **`TxRepos` pasó el umbral**: 6 campos y 5 flujos, ninguno usa todos. La
  señal para repartirlo ahora es el primer flujo que necesite otro nivel de
  aislamiento, no el conteo.
- **`internal/progression/` ya tiene llamador** (`SyncService.progressPlan`);
  lo que sigue sin usarse es el 1RM real: `OneRepMaxKg` se pasa en cero, así
  que la estrategia `percentage_1rm` mantiene el peso en vez de calcularlo.
- **La app no manda `assigned_exercise_id`** en el sync: la progresión cruza
  por `assigned_workout_id` + `exercise_id`. Si un día del plan repitiera el
  mismo ejercicio dos veces, ambos compartirían el target calculado.
- **`exercise_swap` no se aplica al aprobarse** (decisión de producto, no
  bug): viaja con el nombre del ejercicio en texto libre, sin id resoluble
  contra el catálogo.
- **`AISuggestionRepository` mezcla 4 agregados ajenos** (counts de
  assignment/session/habit) porque el port se derivó del archivo sqlc, no de
  los casos de uso.
- **Ports con métodos muertos**: `AssignmentRepository.GetByID`,
  `SessionRepository.CountCompletedSessionsOnDate`,
  `HabitRepository.GetHabitByID`, `ProgramRepository.GetProgressionRuleByID`,
  `AISuggestionRepository.GetByID`.
- **Tablas modeladas sin uso**: `progress_photo` (ya especificada en
  [SDD-002](docs/sdd/active/SDD-002-progress-photos.md)), `exercise_alternative`,
  `user_identity` (OAuth). Ya no aplica a `device_token`: tiene queries,
  repositorio, endpoints y adapter desde SDD-001. Tampoco a
  `internal/progression/`, que hoy se llama desde `SyncService` (`sync.go:11`).

**Siguiente paso:** conseguir un VPS real para cerrar el resto de fase 10
(push, backups, monitoreo, prueba de carga, cron del worker).
La UI de Flutter sigue pendiente de que el usuario la confirme visualmente
(`flutter run -d chrome` o `http://localhost:5555`) — el panel del navegador
de este entorno no compone el canvas de Flutter web. Falta
`ANTHROPIC_API_KEY` (o `ant auth login`) para ver una sugerencia de IA real
generada por Claude en `cmd/worker` (fase 9).

Notas del entorno:
- **Esta sesión corre contra Postgres local nativo, no Docker**: el usuario
  pidió seguir sin levantar Docker ni el backend. DB en
  `postgres://postgres:1234@localhost:5432/myvibesfit?sslmode=disable`
  (usuario `postgres`, password fija `1234` por instrucción del usuario).
  Migraciones aplicadas con `goose` contra esa base. `docker-compose.yml`
  sigue existiendo para cuando se retome Docker (mapea Postgres a
  `5433:5432` en este equipo porque el 5432 nativo ya estaba ocupado).
- El almacenamiento de Docker Desktop (imágenes, contenedores, volúmenes)
  vive físicamente en `F:\Docker\wsl\`, enlazado por junction NTFS desde
  `%LOCALAPPDATA%\Docker\wsl\` (Docker Desktop no expone esta ruta como
  setting editable; se relocalizó a mano). Ver estructura en `F:\Docker\`.
- `CORS_ALLOWED_ORIGINS` (env var) controla los orígenes permitidos; en
  development cae por defecto a `localhost:5555`/`localhost:3000`.

## Stack

Backend: Go 1.26 + chi + sqlc + PostgreSQL 16, **arquitectura hexagonal
(ports & adapters)** — ver sección abajo. Panel coach: Next.js 16 + TS +
Tailwind + shadcn. App cliente: Flutter + Riverpod + Drift (cache offline) +
dio. Monorepo, Docker Compose en VPS.

## Arquitectura del backend (hexagonal)

Migrado de `handler → service → repository` (acoplado a sqlc/pgx en todas
las capas) a ports & adapters, a pedido explícito del usuario. `internal/domain`
no importa `pgx`/sqlc/`anthropic-sdk` — solo stdlib + `google/uuid`:

- `internal/domain/`: entidades (`Program`, `Assignment`, `WorkoutSession`,
  `DeviceToken`, etc.), **13 ports de repositorio** (`ProgramRepository`,
  `AssignmentRepository`, `SessionRepository`, `GamificationRepository`,
  `HabitRepository`, `ProgressRepository`, `CoachRepository`,
  `AISuggestionRepository`, `ExerciseRepository`, `IdentityRepository`,
  `AuditRepository`, `ProfileRepository`, `DeviceTokenRepository`), dos ports
  de integración (`SuggestionProposer` hacia Claude, `PushSender` hacia FCM),
  el `UnitOfWork` con
  `TxRepos{Assignments,Sessions,Gamification,Habits,Coaches,Suggestions}` —
  para las **4** transacciones multi-agregado: `Assign`, `SyncSessions`,
  `LogHabit` y `AISuggestionService.Review` — y `GamificationService` (domain
  service sin estado propio, recibe el port por parámetro para poder
  participar en la tx de quien lo llama).
- `internal/service/`: casos de uso — dependen solo de ports de `domain`,
  cero imports de `pgx`/`internal/repository/db` (verificado con grep, debe
  seguir dando cero).
- `internal/adapter/postgres/`: implementa los 13 ports envolviendo
  `internal/repository/db` (sqlc, sin tocar — sigue siendo el único lugar
  con `pgtype`). Traduce `pgx.ErrNoRows` → `domain.ErrNotFound` en cada
  adapter — si un adapter nuevo lo olvida, un flujo que tolera "sin fila
  todavía" (usuario nuevo sin `user_stats`) rompe en silencio.
- `internal/adapter/push/`: implementa `domain.PushSender`. `fcm.go` habla el
  HTTP v1 de FCM sobre `net/http` + `oauth2/google`; se descartó
  firebase-admin-go porque arrastraba grpc, protobuf y appengine para enviar
  un mensaje. `noop.go` es lo que se cablea sin `FCM_CREDENTIALS_JSON`, para
  que el desarrollo local no dependa de una cuenta de Firebase.
- `internal/progression/`: motor puro sin BD. **Ya tiene llamador**:
  `SyncService` lo usa (`sync.go:11`) para progresar la próxima ocurrencia
  del ejercicio en cada sesión sincronizada.
- `internal/adapter/anthropic/`: implementa `domain.SuggestionProposer`
  (antes `internal/ai`, ya no existe).
- `internal/transport/http/handler/`: sin cambios estructurales, solo las
  funciones `xDTO()` cambiaron su parámetro de `db.X` a `domain.X`. El
  contrato HTTP (rutas, JSON) es idéntico — cero cambios para el panel
  Next.js o la app Flutter.
- Solo 7 enums Postgres se duplican como tipos en domain (los que gatean
  una rama de control real: `ProgramStatus`, `SessionStatus`,
  `SuggestionStatus`, `SetType`, `RecordType`); el resto
  (`ExperienceLevel`, `TrainingGoal`, `MemberRole`, etc.) quedan `string`
  plano porque ningún service rama sobre su valor.

Verificado con `go build/vet/test` limpios después de cada fase. **Y también
en vivo contra Postgres real y el servidor levantado**, que es donde
aparecieron los bugs que la suite no vio:

- El XP de `personal_record` se otorgaba en *cada* resincronización.
  `estimated_1rm` calculaba 93.3333 pero la columna es `numeric(8,2)` y
  guardaba 93.33, así que el récord se superaba a sí mismo para siempre. El
  test unitario no lo vio porque el fake devolvía una lista de récords vacía.
- `habit_log` devolvía un 500 filtrando el nombre de la constraint al segundo
  check-in del mismo día con otro `client_local_id`.
- El traspaso de un `device_token` entre dos usuarios en el mismo teléfono:
  verificado que deja **una** fila y cambia de dueño, y que un `INSERT` plano
  ahí falla con `23505`.

La lección, que vale para la próxima sesión: **compilar no es verificar**. El
camino crítico (login, crear+publicar programa, asignar, sincronizar sesión
con series, marcar hábito, aprobar sugerencia) se ejercita contra el servidor,
no solo con fakes.

## Convenciones

Igual que el resto de proyectos del workspace (ver `F:\Proyectos\.claude\CLAUDE.md`):
Conventional Commits, sin atribución de IA, código en inglés, comunicación en
español, variables de entorno para toda config, DTOs explícitos, errores nunca
silenciados.

## Comandos

Con Postgres local nativo (lo habitual en esta máquina) hay tres scripts en
la raíz que ya traen las variables de entorno puestas: `run-api.cmd` (:8080),
`run-coach-web.cmd` (:3000) y `run-flutter-web.cmd` (:5555). También están
como entradas en `.claude/launch.json`.

A mano, o con Docker:

```bash
docker compose up -d db          # solo si se usa el Postgres del compose (:5433)
cd api && goose -dir db/migrations postgres "$DATABASE_URL" up
sqlc generate                    # después de tocar db/queries/*.sql
go run ./cmd/api
cd ../web && pnpm dev            # panel del coach + landing en :3000
```

El worker de IA es una **pasada única**, pensado para cron (no tiene
scheduler embebido). Necesita `ANTHROPIC_API_KEY`. El cron en sí, los
subcomandos y la pasada de rachas están especificados en
[SDD-003](./docs/sdd/active/SDD-003-worker-operation.md), todavía sin implementar:

```bash
cd api && go run ./cmd/worker    # una corrida; en el VPS va en crontab nocturno
```

Verificación antes de commit: `go vet ./... && go test ./... && go build ./...`
y `gofmt -l .` (API) · `pnpm lint && pnpm build` (panel) ·
`flutter analyze && flutter test` (app).

<!-- gitnexus:start -->
# GitNexus — Code Intelligence

This project is indexed by GitNexus as **Myvibesfit** (5,729 nodes, 13,947 relationships, 408 execution flows).

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
