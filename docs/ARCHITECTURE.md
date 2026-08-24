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
| IA | Claude API (`claude-sonnet-5`), propone → coach aprueba | Trazabilidad y responsabilidad clínica |
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
│   │   ├── domain/               Entidades, errores tipados, value objects
│   │   ├── transport/http/
│   │   │   ├── router.go         chi, montaje de rutas
│   │   │   ├── middleware/       auth, tenant, ratelimit, requestid, recover
│   │   │   ├── dto/              Request/Response explícitos, nunca entidades
│   │   │   └── handler/          Un fichero por módulo
│   │   ├── service/               Lógica de negocio y transacciones
│   │   ├── repository/            Interfaces + implementación sobre sqlc
│   │   ├── progression/           Motor determinista (sin dependencias externas)
│   │   ├── ai/                    Cliente Claude, prompts, validación de salida
│   │   └── platform/              jwt, hash, storage S3, push, logger, clock
│   ├── db/
│   │   ├── migrations/            goose: 0001_init, 0002_training, 0003_execution
│   │   ├── queries/                .sql fuente de sqlc
│   │   └── seed/                   Catálogo de ejercicios, hábitos, logros
│   ├── sqlc.yaml
│   └── openapi.yaml               Contrato: genera el cliente Dart y los tipos TS
├── web/                           Panel del coach (Next.js)
├── app/                           Cliente Flutter
├── docs/
│   ├── ARCHITECTURE.md
│   └── DESIGN.md
├── docker-compose.yml
└── AGENTS.md
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

Un job nocturno recorre las asignaciones activas y, para las que tienen suficiente
historial (mínimo 3 sesiones del bloque), llama a Claude con un resumen estructurado:
adherencia, RPE medio por patrón de movimiento, sesiones saltadas, tendencia de volumen
y racha de hábitos.

La respuesta se valida contra un esquema estricto antes de tocar la BD. Se guarda en
`ai_suggestion` con `input_snapshot` — los datos exactos que la generaron — y llega al
coach como tarjeta aprobable en su panel. **Nada se aplica solo.**

Reglas duras:
- La IA no escribe en `assigned_exercise`; escribe una propuesta.
- Sugerencias fuera de rango razonable (>10 % de salto de carga, >30 % de volumen) se
  descartan en el servidor antes de mostrarse.
- El cliente nunca ve una sugerencia sin aprobar.

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

## 6. Endpoints principales

```
POST   /v1/auth/register
POST   /v1/auth/login
POST   /v1/auth/refresh
POST   /v1/orgs/join                    Vincular con código de gimnasio

GET    /v1/me
PUT    /v1/me/profile
GET    /v1/me/assignment/current        Entrenamiento activo + semana en curso
GET    /v1/me/home                      Anillo del día, racha, próximo entreno, hábitos

GET    /v1/exercises                    Globales + los del gimnasio del usuario
GET    /v1/exercises/{id}

POST   /v1/sync/sessions                Lote idempotente de sesiones y series
POST   /v1/sync/habits                  Lote idempotente de registros de hábito

GET    /v1/progress/exercises/{id}      Serie temporal para la gráfica
GET    /v1/progress/records
GET    /v1/progress/volume              Volumen semanal por grupo muscular

# Coach (panel web)
GET    /v1/coach/clients
GET    /v1/coach/clients/{id}/overview  Adherencia, PRs, alertas
POST   /v1/coach/programs
POST   /v1/coach/programs/{id}/assign
PATCH  /v1/coach/assigned-exercises/{id}
GET    /v1/coach/suggestions?status=pending
POST   /v1/coach/suggestions/{id}/approve
POST   /v1/coach/suggestions/{id}/reject

# Admin del gimnasio
GET    /v1/org/members
POST   /v1/org/invitations
PATCH  /v1/org/branding
```

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
