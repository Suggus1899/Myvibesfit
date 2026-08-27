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

## Estado actual

Fases 0-10 de [PHASES.md](./docs/PHASES.md) — fase 10 completa en su
subconjunto realizable sin VPS (ver más abajo).

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
  tendencia de peso/RPE por ejercicio — `internal/ai/suggester.go` +
  queries en `db/queries/ai_suggestion.sql`) y le pide a Claude Opus 5 una
  sugerencia vía tool use forzado (`propose_suggestion`), validada en Go
  contra el enum de `suggestion_kind` antes de guardarse como `pending`.
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
- **Fuera de alcance en este entorno** (requieren VPS real): push
  notifications (`device_token` existe en el esquema pero sin endpoint),
  backups automáticos de Postgres, monitoreo (logs+métricas) del VPS,
  prueba de carga de 100 usuarios concurrentes, políticas de privacidad de
  fotos de progreso (es texto/producto, no código).

**Siguiente paso:** conseguir un VPS real para cerrar el resto de fase 10.
La UI de Flutter sigue pendiente de que el usuario la confirme visualmente
(`flutter run -d chrome` o `http://localhost:5555`). Falta
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
  etc.), los 11 ports (`ProgramRepository`, `AssignmentRepository`,
  `SessionRepository`, `GamificationRepository`, `HabitRepository`,
  `ProgressRepository`, `CoachRepository`, `AISuggestionRepository`,
  `ExerciseRepository`, `IdentityRepository`, `AuditRepository`), el
  `UnitOfWork` (`TxRepos{Assignments,Sessions,Gamification,Habits}` — para
  las 3 transacciones multi-agregado: `Assign`, `SyncSessions`, `LogHabit`),
  y `GamificationService` (domain service sin estado propio, recibe el port
  por parámetro para poder participar en la tx de quien lo llama).
- `internal/service/`: casos de uso — dependen solo de ports de `domain`,
  cero imports de `pgx`/`internal/repository/db` (verificado con grep, debe
  seguir dando cero).
- `internal/adapter/postgres/`: implementa los 11 ports envolviendo
  `internal/repository/db` (sqlc, sin tocar — sigue siendo el único lugar
  con `pgtype`). Traduce `pgx.ErrNoRows` → `domain.ErrNotFound` en cada
  adapter — si un adapter nuevo lo olvida, un flujo que tolera "sin fila
  todavía" (usuario nuevo sin `user_stats`) rompe en silencio.
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

Verificado con `go build/vet/test` limpios después de cada fase (identidad
→ catálogo/programas → ejecución/gamificación → IA/worker → cleanup). **No
se probó end-to-end contra un servidor corriendo** (`go run ./cmd/api`)
porque esta sesión corrió sin levantar el backend a pedido del usuario — la
próxima sesión que levante el servidor debería ejercitar el camino crítico
(login, crear+publicar programa, asignar, sincronizar una sesión con
series, marcar un hábito, aprobar una sugerencia) y confirmar respuestas
JSON idénticas a las de antes del refactor.

## Convenciones

Igual que el resto de proyectos del workspace (ver `F:\Proyectos\.claude\CLAUDE.md`):
Conventional Commits, sin atribución de IA, código en inglés, comunicación en
español, variables de entorno para toda config, DTOs explícitos, errores nunca
silenciados.

## Comandos

```bash
docker compose up -d db
cd api && goose -dir db/migrations postgres "$DATABASE_URL" up
sqlc generate
go run ./cmd/api
cd ../web && pnpm dev  # panel del coach en :3000
```

Verificación antes de commit: `go vet ./... && go test ./... && go build ./...`
(API) · `pnpm lint && pnpm build` (panel) · `dart analyze && flutter test` (app).
