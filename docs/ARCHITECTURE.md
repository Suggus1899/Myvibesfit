# Myvibesfit — Arquitectura

Plataforma multi-tenant de entrenamiento para gimnasios. El coach diseña y supervisa
desde un panel web; el cliente entrena desde el móvil. Un motor determinista calcula la
progresión de cargas y una capa de IA propone ajustes que el coach aprueba.

## 1. Decisiones tomadas

| Decisión | Elección | Motivo |
|---|---|---|
| Tenencia | Multi-tenant por `organization` (gimnasio) | El cliente que paga es el gimnasio |
| Identidad | Registro abierto + vinculación por código | El usuario existe sin gimnasio (`org_id` nullable) |
| Backend | Go 1.25 + chi + sqlc + PostgreSQL 16 | SQL explícito, bajo consumo en VPS, `set_log` crecerá a millones de filas |
| Panel coach | Next.js 16 App Router + TS + Tailwind + shadcn | Construir un mesociclo de 12 semanas necesita pantalla grande |
| App cliente | Flutter + Riverpod + go_router + Drift + dio | Único frontend móvil; Drift da el cache offline |
| Offline | Híbrido: lectura cacheada + cola de mutaciones | En un gimnasio sin cobertura hay que poder registrar series |
| Progresión | Determinista (doble progresión / RPE) | Es una fórmula. Un LLM decidiendo cargas sin historial es un riesgo de lesión |
| IA | Claude API, propone → coach aprueba. Modelo por `ANTHROPIC_SUGGESTION_MODEL` (default `claude-opus-5`) | Trazabilidad y responsabilidad clínica |
| Repos | Monorepo + Docker Compose en VPS | Un solo sitio, control de costes |
| v1 | Entrenamiento + hábitos | Nutrición queda para v2, el esquema ya la contempla |

**Arquitectura elegida: monolito modular.** Un solo binario Go con módulos aislados
(`identity`, `catalog`, `programming`, `training`, `habits`, `gamification`, `insights`).
No hay microservicios: un gimnasio de 500 socios cabe entero en una instancia. Los módulos
se comunican por interfaces de servicio, así que si alguno necesita salir después, sale sin
reescribir el resto.

## 2. Estructura del monorepo

```
myvibesfit/
├── api/                          Backend Go
│   ├── cmd/api/main.go           Composición de dependencias y arranque
│   ├── cmd/worker/main.go        Jobs: rachas, sugerencias IA, push
│   ├── internal/
│   │   ├── config/               Todo por variable de entorno, cero hardcode
│   │   ├── domain/               Entidades, los 12 ports, UnitOfWork,
│   │   │                         GamificationService. Sin imports de pgx/sqlc
│   │   ├── service/              Casos de uso: dependen solo de ports
│   │   ├── adapter/
│   │   │   ├── postgres/         Implementa los ports sobre sqlc; único
│   │   │   │                     lugar con pgtype. Traduce pgx.ErrNoRows
│   │   │   └── anthropic/        Implementa domain.SuggestionProposer
│   │   ├── transport/http/
│   │   │   ├── router.go         chi, montaje de rutas
│   │   │   ├── middleware/       auth.go (JWT) y rbac.go (RequireRole,
│   │   │   │                     RequireOrg). Rate limit, request-id y
│   │   │   │                     recover son los de chi, montados en router.go
│   │   │   ├── dto/              Request/Response explícitos, nunca entidades
│   │   │   └── handler/          Un fichero por módulo
│   │   ├── repository/db/        Generado por sqlc — NO editar a mano
│   │   ├── progression/          Motor determinista (sin dependencias
│   │   │                         externas). Testeado, todavía sin llamadores
│   │   └── platform/             jwt, hash, refreshtoken
│   ├── db/
│   │   ├── migrations/           goose: 0001_init, 0002_training,
│   │   │                         0003_execution, 0004_data_integrity
│   │   ├── queries/               .sql fuente de sqlc
│   │   └── seed/                  Catálogo de ejercicios, hábitos, logros
│   └── sqlc.yaml
├── web/                           Panel del coach + landing pública (Next.js)
├── app/                           Cliente Flutter
├── docs/
│   ├── ARCHITECTURE.md
│   ├── DESIGN.md
│   └── PHASES.md
├── docker-compose.yml
└── AGENTS.md

> **No existe `openapi.yaml`.** Este documento lo describía como el contrato
> que genera el cliente Dart y los tipos TS; nunca se construyó. Hoy los DTOs
> se escriben a mano en los tres lados (`transport/http/dto/`,
> `app/lib/core/network/models.dart`, tipos inline en `web/src/`). Si se
> retoma, es una decisión abierta, no algo que falte terminar.
```

### Flujo de una petición

```
HTTP → middleware (requestID → recover → ratelimit → auth → tenant)
     → handler   (decodifica DTO, valida forma)
     → service   (reglas de negocio, transacción)
     → repository (sqlc)
     → PostgreSQL
```

El handler nunca toca la BD. El service nunca conoce HTTP. El repository nunca decide.

### Aislamiento entre gimnasios

El middleware `tenant` resuelve la organización activa desde el JWT y la inyecta en el
contexto. **Toda** query de repository que toque datos de organización recibe `org_id`
como parámetro obligatorio — no hay filtrado opcional. Un usuario sin organización
(registro abierto) opera solo sobre sus propios datos: entrenamiento libre, hábitos y
progreso, sin acceso a programas ni coach.

## 3. Motor de progresión

Vive en `internal/progression/`, es Go puro, sin BD ni red. Entrada: el histórico de
series del ejercicio y la regla asignada. Salida: los objetivos de la próxima sesión.

```
double_progression  Si completa todas las series en el tope del rango de reps
                    → sube la carga el incremento mínimo y vuelve al suelo del rango.
                    Si no → repite objetivo.

linear_load         Suma un incremento fijo por sesión. Tras N fallos consecutivos,
                    descarga al 90 % y reinicia.

rpe_autoregulated   Compara el RPE real con el objetivo y ajusta la carga por la
                    tabla RPE→%1RM. Es lo que hace un coach a ojo, escrito.

percentage_1rm      Carga = %1RM estimado (Epley sobre el mejor set reciente).
```

Determinista, testeable sin infraestructura, y explicable al coach. La IA no lo sustituye:
lo lee.

## 4. Capa de IA

Un job nocturno (`cmd/worker`, una sola pasada — el scheduler es cron, no está
embebido) recorre las asignaciones activas y llama a Claude con un resumen
estructurado: adherencia, tendencia de peso/RPE por ejercicio, PRs recientes,
racha, hábitos, y el perfil del cliente (objetivo, experiencia, equipamiento
disponible, limitaciones).

La respuesta se valida contra un esquema estricto antes de tocar la BD. Se guarda en
`ai_suggestion` con `input_snapshot` — los datos exactos que la generaron — y llega al
coach como tarjeta aprobable en su panel. **Nada se aplica solo.**

Reglas duras (implementadas en `domain.Suggestion.Validate`, con tests en
`domain/ai_suggestion_test.go`):
- La IA no escribe en `assigned_exercise`; escribe una propuesta.
- Sugerencias fuera de rango razonable (>10 % de salto de carga, >30 % de volumen) se
  descartan en el servidor antes de mostrarse.
- El cliente nunca ve una sugerencia sin aprobar.

**Sin umbral mínimo de historial.** Versiones anteriores de este documento
decían "mínimo 3 sesiones del bloque"; ese filtro no existe en el código — el
worker procesa toda asignación activa que tenga coach. Si se quiere, va en
`SuggestionWorkerService.ProcessAssignment`.

## 5. Sincronización offline

El móvil cachea en Drift el entrenamiento asignado, el catálogo de ejercicios usado y los
hábitos activos. Todo lo que el usuario **escribe** durante una sesión va primero a una cola
local (`pending_mutation`) y luego al servidor.

Idempotencia: el móvil genera un `client_local_id` (UUID) por sesión, por serie y por
registro de hábito. Las tres tablas tienen `UNIQUE (user_id, client_local_id)`, así que
reenviar la cola tras un corte de red no duplica nada. El endpoint de sync es un `POST`
por lotes que devuelve el resultado por elemento.

No se resuelven conflictos de escritura concurrente: un cliente entrena desde un
dispositivo a la vez. **Lo que el móvil escribió gana.**

## 6. Endpoints

Esta lista refleja `internal/transport/http/router.go`, que es la fuente de
verdad. Todo bajo `/v1`.

```
# Público
GET    /health
POST   /v1/auth/register                rate limit 10/min por IP
POST   /v1/auth/login                   idem
POST   /v1/auth/refresh                 idem, rota el refresh token

# Catálogo (auth opcional: sin token devuelve solo los globales)
GET    /v1/exercises
GET    /v1/exercises/{id}

# Cliente autenticado
GET    /v1/me
GET    /v1/me/profile                   Onboarding: objetivo, equipamiento, unidades
PUT    /v1/me/profile
POST   /v1/orgs                         Crear gimnasio (el creador queda owner)
POST   /v1/orgs/join                    Vincular con código de gimnasio
GET    /v1/me/assignment/current        Mesociclo completo del cliente
GET    /v1/me/stats                     XP, nivel, rachas
GET    /v1/me/achievements

POST   /v1/sync/sessions                Lote idempotente por client_local_id

GET    /v1/progress/exercises/{id}      Serie temporal para la gráfica
GET    /v1/progress/records
GET    /v1/progress/volume
POST   /v1/body-metrics
GET    /v1/body-metrics

GET    /v1/habits
GET    /v1/habits/logs?date=            Qué hábitos ya se marcaron ese día
GET    /v1/me/habits
POST   /v1/habits/subscribe
DELETE /v1/habits/{id}
POST   /v1/habits/{id}/log

# Gimnasio (requiere rol owner/admin/coach + org en el JWT)
GET    /v1/org
GET    /v1/org/members
PATCH  /v1/org/members/{id}/role        Solo owner; no puede otorgar owner

POST   /v1/exercises                    Crear en el catálogo del gimnasio
PATCH  /v1/exercises/{id}               Scoped a la org: no toca los globales
DELETE /v1/exercises/{id}               idem (soft-delete vía is_active)

GET    /v1/progression-rules
POST   /v1/progression-rules

POST   /v1/programs
GET    /v1/programs
GET    /v1/programs/{id}
PATCH  /v1/programs/{id}
POST   /v1/programs/{id}/publish
POST   /v1/programs/{id}/archive
POST   /v1/programs/{id}/assign         Copia la plantilla + vincula coach↔cliente
POST   /v1/programs/{id}/workouts
GET    /v1/programs/{id}/workouts
PATCH  /v1/workouts/{id}
DELETE /v1/workouts/{id}
POST   /v1/workouts/{id}/exercises
GET    /v1/workouts/{id}/exercises
PATCH  /v1/program-exercises/{id}
DELETE /v1/program-exercises/{id}
POST   /v1/assignments/{id}/cancel

GET    /v1/coach/clients                Overview con needs_attention
GET    /v1/coach/suggestions            Pendientes de revisar
POST   /v1/ai-suggestions/{id}/approve
POST   /v1/ai-suggestions/{id}/reject
```

**Planeados y nunca construidos** (estaban en versiones anteriores de este
documento como si existieran): `GET /v1/me/home` (la app compone el home con
`/me/stats` + `/me/assignment/current`), `POST /v1/sync/habits` (el registro
va de a uno por `POST /v1/habits/{id}/log`), `GET /v1/coach/clients/{id}/overview`
(el overview viene embebido en el listado), `PATCH /v1/coach/assigned-exercises/{id}`
(solo se puede editar la plantilla, no la copia asignada),
`POST /v1/org/invitations` y `PATCH /v1/org/branding`.

## 7. Comandos

```bash
docker compose up -d db && cd api && goose -dir db/migrations postgres "$DATABASE_URL" up && sqlc generate && go run ./cmd/api
```

Verificación mínima antes de commit: `go vet ./... && go test ./... && go build ./...`
para la API, `pnpm lint && pnpm build` para el panel, `dart analyze && flutter test`
para la app.

## 8. Fuera del alcance de v1

Nutrición y macros · mensajería coach↔cliente · pagos y suscripciones · wearables
(HealthKit / Health Connect) · feed social · marketplace de coaches.

El esquema ya soporta nutrición y mensajería sin migraciones destructivas cuando toquen.
