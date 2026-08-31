# Metodología SDD (Spec-Driven Development & Software Design Document)
## Myvibesfit

Este documento establece la metodología de ingeniería estándar para el desarrollo, evolución y refactorización de funcionalidades en **Myvibesfit**.

---

## 1. ¿Qué es SDD en Myvibesfit?

La metodología **SDD** en Myvibesfit unifica dos principios fundamentales:

1. **Software Design Document (Documento de Diseño Formal):** Antes de escribir código para cualquier funcionalidad o cambio de arquitectura, se redacta una especificación estructurada que valida decisiones, invariantes de negocio, riesgos y contratos técnicos.
2. **Spec / Schema-Driven Development (Desarrollo Guiado por Especificación y Esquemas):** El código se deriva rigurosamente a partir de contratos y esquemas formales:
   - **Base de datos:** Migraciones Goose (`api/db/migrations/`) + queries SQLC (`api/db/queries/`).
   - **Dominio:** Entidades, Value Objects y Puertos (`internal/domain/`).
   - **API / Red:** DTOs explícitos (`internal/transport/http/dto/`) y contratos de sync/idempotencia.
   - **Frontend / Móvil:** Modelos tipados y esquemas de persistencia local (Drift SQLite en Flutter, TypeScript interfaces en Next.js).

```
 ┌────────────────────────────────────────────────────────┐
 │ 1. SOFTWARE DESIGN DOCUMENT (SDD)                      │
 │    Requerimientos, invariantes, flujos y contratos     │
 └───────────────────────────┬────────────────────────────┘
                              │
                              ▼
 ┌────────────────────────────────────────────────────────┐
 │ 2. SCHEMA & PORT SPECIFICATION                         │
 │    Goose DDL ──> SQLC Queries ──> Domain Entities/Ports│
 └───────────────────────────┬────────────────────────────┘
                              │
                              ▼
 ┌────────────────────────────────────────────────────────┐
 │ 3. CONTRACT & UNIT TESTS (TDD)                         │
 │    Tests de dominio puro, progresión y casos de uso    │
 └───────────────────────────┬────────────────────────────┘
                              │
                              ▼
 ┌────────────────────────────────────────────────────────┐
 │ 4. HEXAGONAL IMPLEMENTATION                            │
 │    Adapters ──> Use Case Services ──> HTTP Handlers/DTO│
 │    ──> Frontend (Web Coach / Flutter Client)           │
 └───────────────────────────┬────────────────────────────┘
                              │
                              ▼
 ┌────────────────────────────────────────────────────────┐
 │ 5. VERIFICATION & AUDIT                                │
 │    Impact Analysis, CI (Go/TS/Flutter tests), Audit Log│
 └───────────────────────────┬────────────────────────────┘
```

---

## 2. Las 5 Fases del Ciclo de Vida SDD

### Fase 1: Redacción del SDD (Software Design Document)
Para cualquier épica, nueva feature o refactor significativo, se crea o actualiza un documento en `docs/sdd/active/SDD-<NUMERO>-<nombre>.md` siguiendo el [Catálogo de Plantillas SDD](sdd/README.md) (ver detalle en §3 y en `docs/sdd/templates/`).

**Criterios de salida de Fase 1 (Definition of Ready):**
- [ ] Modelo de dominio e invariantes de negocio definidos.
- [ ] Contratos de datos (DDL SQL, DTOs de petición/respuesta) especificados.
- [ ] **Impact analysis previo evaluado** con GitNexus `impact()` (riesgos, tablas afectadas, consumidores en Web y App).
- [ ] Anotado qué anexos aplican (A: API contract, B: extended schema, C: offline sync, D: AI flow, E: bugfix/RCA).
- [ ] Revisado por al menos un segundo ingeniero o validado con checklist arquitectónico.

### Fase 2: Especificación de Esquemas y Puertos
Se definen las fuentes de verdad de datos y contratos en el backend:
1. **Migración SQL (Goose):** En `api/db/migrations/`, respetando multi-tenancy (`org_id`), llaves foráneas e índices únicos.
2. **Queries SQLC:** En `api/db/queries/<modulo>.sql`.
3. **Generación SQLC:** Ejecución de `sqlc generate` en `api/` para generar código en `internal/repository/db/`.
4. **Domain Ports & Entities:** Definición de interfaces de repositorio (`*Repository`) y entidades puras en `internal/domain/` sin dependencias externas (solo stdlib + `google/uuid`).
5. **Unit of Work:** Si el caso de uso involucra múltiples agregados de forma transaccional, se declara en `domain.UnitOfWork` y `domain.TxRepos`.

**Gate de salida Fase 2:**
- [ ] `sqlc generate` compila sin errores.
- [ ] `go build ./...` pasa en `api/` (verifica que ports no rompieron).
- [ ] Migración aplicable con `goose up` contra BD local.

### Fase 3: Pruebas Unitarias de Dominio y Contratos (TDD)
Antes o en paralelo a la lógica de persistencia:
- Escribir tests unitarios para reglas de negocio complejas en `internal/domain/` o `internal/progression/`.
- Crear o actualizar mocks/fakes de los puertos de dominio para probar los casos de uso en `internal/service/` sin levantar base de datos real.
- **Cobertura mínima obligatoria:**
  - Lógica de dominio pura (progresión, gamificación, validación IA): **100% funciones públicas**.
  - Services (`internal/service/`): **casos feliz + 1 error + 1 edge case** por método público.
  - Handlers HTTP: **objetivo**, no gate. Hoy hay **cero** tests de handler; exigirlos
    bloquearia todo PR y el gate se ignoraria entero. Lo que si es gate: el middleware
    de auth/RBAC, cubierto en `internal/transport/http/middleware` (21 casos).
  - Flutter: router/redirects, `AuthController`, cola de sync, `WorkoutStore`.

**Gate de salida Fase 3:**
- [ ] `go test ./internal/domain/... ./internal/progression/... ./internal/service/...` pasa.
- [ ] `flutter test` pasa (incluye tests nuevos).

### Fase 4: Implementación Hexagonal
Siguiendo el flujo de capas desacopladas:
1. **Adapters Secundarios (`internal/adapter/postgres/`):** Implementar los puertos de dominio mapeando desde/hacia los tipos de `db` (sqlc). Traducir `pgx.ErrNoRows` a `domain.ErrNotFound`.
2. **Adapters Externos (`internal/adapter/anthropic/` u otros):** Implementar puertos de integración externa (IA, storage, push, etc.).
3. **Services (`internal/service/`):** Implementar la orquestación del caso de uso. Depender únicamente de los puertos de dominio.
4. **Transport / Handlers / DTOs (`internal/transport/http/`):**
   - Definir DTOs explícitos de Request y Response en `dto/`.
   - Implementar Handlers en `handler/` (decodificación, validación de forma, invocación de service, mapeo de errores HTTP).
   - Registrar rutas y middlewares (Auth JWT, RBAC `RequireRole`, Tenant) en `router.go`.
5. **Consumidores Frontend / Móvil:**
   - **Web (`web/`):** Crear/actualizar servicios API en `src/lib/api.ts`, componentes UI en `src/components/`, y páginas en `src/app/`.
   - **Flutter (`app/`):** Modelos en `core/network/models.dart`, queries locales Drift en `core/storage/`, StateNotifier / Riverpod providers, y vistas UI.

**Gate de salida Fase 4:**
- [ ] `go vet ./... && go test ./... && go build ./... && gofmt -l .` pasa.
- [ ] `pnpm lint && pnpm build` pasa en `web/`.
- [ ] `flutter analyze && flutter test` pasa en `app/`.
- [ ] **GitNexus `impact()`** ejecutado sobre símbolos modificados — riesgo **LOW/MEDIUM** o HIGH/CRITICAL con aprobación documentada.

### Fase 5: Verificación, Impact Analysis y Auditoría
1. **Impact Analysis (obligatorio antes de commit/PR):**
   - Ejecutar `gitnexus impact()` sobre cada símbolo público modificado (funciones, structs, interfaces, métodos de handler).
   - Ejecutar `gitnexus detect_changes({scope: "all"})` — debe reportar `risk_level` conocido (no `unknown` por `partial: true`).
   - Si `risk: HIGH` o `CRITICAL` → requerir revisión explícita y plan de migración.
   - Si `risk: UNKNOWN` → **no se permite merge**; confirmar con búsqueda textual (grep) que no hay callers ocultos.
2. **Pipeline de Verificación (CI):**
   - Backend: `go vet ./... && go test ./... && go build ./... && gofmt -l .`
   - Web: `pnpm lint && pnpm build`
   - Flutter: `flutter analyze && flutter test`
3. **Auditoría:** Asegurar que **todas** las mutaciones críticas queden registradas mediante `domain.AuditRepository` / `AuditLogger`.
   - Mutaciones auditadas **hoy** (verificado con grep sobre `internal/service/`):
     `assignment.assign`, `assignment.cancel`, `program.set_status`, `ai_suggestion.review`.
     El gate es: **toda mutacion nueva o modificada** se audita, y ninguna de esas cuatro
     pierde su registro. Las que faltan estan en §6 con su trigger, no aqui.
   - Cada entrada debe incluir: `actor_user_id`, `org_id`, `action`, `entity_type`, `entity_id`, `metadata` JSON.

**Gate de salida Fase 5 (Definition of Done):**
- [ ] Todos los checks de CI en verde.
- [ ] `detect_changes()` sin `partial: true` ni `truncated: true` sin explicación.
- [ ] AuditLog verificado en mutaciones nuevas/modificadas.
- [ ] SDD movido a `docs/sdd/completed/`.
- [ ] `AGENTS.md` y `ARCHITECTURE.md` actualizados si la feature cambia arquitectura, endpoints o deuda conocida.

---

## 3. Plantilla Oficial de SDD (Software Design Document)

Toda nueva funcionalidad se documenta con la plantilla canónica
[`docs/sdd/templates/01_FEATURE_SDD.md`](sdd/templates/01_FEATURE_SDD.md): **9 secciones
obligatorias** (contexto, dominio, esquema, API, hexagonal, web, móvil, seguridad, verificación)
y **5 anexos opcionales** que se completan solo si la feature los toca — A contrato de API
detallado, B esquema extendido, C sync offline, D flujo de IA, E bugfix/RCA.

Para una decisión de arquitectura o un refactor sin feature nueva, usar
[`docs/sdd/templates/02_ADR_ARCHITECTURE.md`](sdd/templates/02_ADR_ARCHITECTURE.md) en su lugar.
Esas dos son las únicas plantillas; ver [docs/sdd/README.md](sdd/README.md) para el flujo de uso.

### Checklist de completitud del SDD (usar en revisión)
| Sección | Requerida | Verificación |
|---------|-----------|--------------|
| 1. Contexto y Objetivos | Sí | Problema de usuario, criterios de éxito medibles, non-goals explícitos |
| 2. Modelo de Dominio | Sí | Entidades, invariantes, puertos (Go interfaces), enums constantes |
| 3. Esquema y Persistencia | Sí | Migración Goose (o "no requerida"), queries SQLC con nombres `:one/:many/:exec` |
| 4. Contratos API | Sí | Tabla método/ruta/middleware/roles, DTOs request/response con validaciones |
| 5. Diseño Hexagonal | Sí | Adapters, Services, Handlers, wiring en `main.go`, config env vars |
| 6. Frontend Web | Sí | Páginas/componentes, llamadas API, estados loading/error/empty |
| 7. Móvil Flutter | Sí | Modelos, providers, storage local, navegación, permisos |
| 8. Seguridad y Multi-tenancy | Sí | RBAC, org_id obligatorio, PII, secrets, CORS |
| 9. Plan de Verificación | Sí | Tests por capa, verificación manual, device real si push/FCM |
| Anexo A: API Detail | Si cambia contrato | OpenAPI fragment o tabla exhaustiva |
| Anexo B: Extended Schema | Si migración compleja | DDL completo, índices, FK, RLS si aplica |
| Anexo C: Offline Sync | Si toca móvil | Drift tables, cola, idempotencia, conflictos |
| Anexo D: AI Flow | Si usa IA | Prompt structure, tool schema, validación, topes |
| Anexo E: Bugfix/RCA | Si es bugfix | 5 whys, fix, regression test, prevention |

---

## 4. Reglas de Oro Arquitectónicas en SDD

1. **Domain no conoce infraestructura:** `internal/domain/` jamás importa `pgx`, `sqlc`, `net/http` o SDKs externos. Solo stdlib y `google/uuid`.
2. **Services dependen solo de Ports:** `internal/service/` no accede a la base de datos directamente; opera mediante interfaces de dominio.
3. **Traducción obligatoria de errores:** Todo adapter de persistencia debe traducir `pgx.ErrNoRows` a `domain.ErrNotFound`.
4. **Idempotencia de cliente — de transporte y de negocio son cosas distintas:** toda escritura
   que venga del móvil incluye `client_local_id` (UUID) protegido con `UNIQUE (user_id, client_local_id)`.
   Eso cubre **reenviar el mismo lote**, y nada más. Si la tabla tiene además una restricción de
   negocio más estrecha, el `ON CONFLICT` del upsert debe apuntar a **esa** restricción, no a
   `client_local_id`. Confundirlas fue lo que hizo que un segundo check-in del mismo hábito y día,
   con otro `client_local_id`, devolviera un **500 filtrando el nombre de la constraint**
   (`habit_log`, `UNIQUE (client_habit_id, log_date)`). Todo upsert de cliente devuelve además
   `(xmax = 0) AS inserted`, y los premios (XP, rachas, logros) se otorgan **solo** con ese flag
   en verdadero.
5. **IA supervisada por diseño:** Ninguna sugerencia de IA se auto-aplica en tablas operativas (`assigned_exercise`, `user_stats`). Siempre se almacena en `ai_suggestion` con estado `pending` y requiere aprobación explícita del coach.
6. **Multi-tenancy obligatorio:** Toda consulta que involucre datos del gimnasio debe recibir y filtrar obligatoriamente por `org_id`.
7. **TxRepos se parte por aislamiento, no por tamaño:** hoy tiene 6 campos y lo usan 4 flujos (`assignment`, `sync`, `habit`, `ai_suggestion`); ninguno usa todos, y eso por sí solo no justifica partirlo. El disparador es que un flujo necesite otro nivel de aislamiento transaccional. No se crea `UnitOfWork[T]`.
8. **Progression engine es puro:** `internal/progression/` no importa BD, red, ni tiempo. Entrada = historial + regla, Salida = targets. Cualquier llamador pasa `OneRepMaxKg` real cuando exista.

---

## 5. Tooling en el Ciclo SDD

### GitNexus (análisis de grafo de código)

**El servidor MCP se cae con frecuencia** (`CONNECT_TIMEOUT` en esta misma sesión). Cuando eso
pasa, el gate **no se salta**: se usa el CLI, que es equivalente y es el que documenta
`CLAUDE.md`. Un gate cuyo cumplimiento depende de que un servicio esté en pie no es un gate.

| Cuándo | MCP | Fallback CLI (desde la raíz del repo) |
|---|---|---|
| **Antes de editar** un símbolo público | `impact({target:"X", direction:"upstream"})` | `node .gitnexus/run.cjs impact "X" --direction upstream --repo .` |
| **Antes de commit** | `detect_changes({scope:"all"})` | `node .gitnexus/run.cjs detect-changes --scope all --repo .` |
| **Reindexar** tras varios commits | — | `node .gitnexus/run.cjs analyze --index-only` |
| Entender un flujo | `query({search_query:"concepto"})` | — |
| Contexto de un símbolo | `context({name:"X"})` | — |
| Renombrar | `rename({symbol_name:"Viejo", new_name:"Nuevo", dry_run:true})` | — |

Qué hacer con el resultado: `HIGH`/`CRITICAL` se avisa al usuario antes de seguir (lo exige
`CLAUDE.md`) y se documenta la mitigación. `UNKNOWN` **no** es "bajo": significa que el walk no
pudo responder, y hay que confirmarlo con búsqueda textual antes de tratar el símbolo como
seguro.

### Context7 (documentación de librerías)

Sus tools (`mcp__context7__*`) **no están en la sesión principal**: viven en el agente
`ecc:docs-lookup`. Consultarlo por ahí, o leer `go.mod`/`pubspec.yaml` y la documentación
oficial. Recomendado antes de añadir o subir una dependencia; **no es un gate bloqueante**.

### Skills cargadas en el agente
- `codebase-memory` — grafo de conocimiento del proyecto
- `gitnexus-*` — exploring, impact-analysis, debugging, refactoring, review, pdg-query, plan, work
- `ecc:golang-patterns` — patrones idiomáticos Go
- `ecc:golang-testing` — table-driven, testify, fuzzing, goleak
- `ecc:react-patterns` / `ecc:react-performance` — React/Next.js
- `zod` — validación de esquemas
- `ecc:accessibility` — WCAG 2.2 si toca UI

---

## 6. Deuda Técnica Conocida (Registrada, No Arreglada)

> Las tres afirmaciones de abajo son **verificables por máquina**. `api/internal/docs` tiene un
> test que las contrasta con el código en cada `go test ./...`; si esta sección se desactualiza,
> la suite falla. Es la respuesta al problema de fondo: este documento ya afirmó tres veces cosas
> que el código desmentía, y nadie se dio cuenta hasta releerlo a mano.

<!-- audited-actions: assignment.assign, assignment.cancel, program.set_status, ai_suggestion.review -->
<!-- txrepos-fields: 6 -->
<!-- tables-unused: progress_photo, exercise_alternative, user_identity -->

| Área | Problema | Trigger para arreglar |
|------|----------|----------------------|
| **TxRepos** | 6 campos, 4 flujos, ninguno usa todos | Primer flujo que necesite otro aislamiento |
| **Progression engine** | `OneRepMaxKg=0` → las reglas `percentage_1rm` no calculan. Ya **no** es cierto que no tenga llamadores: `SyncService` lo usa (`sync.go:11`) | Cuando `SyncService.progressPlan` reciba 1RM real |
| **Sync no manda `assigned_exercise_id`** | Progresión cruza por `assigned_workout_id` + `exercise_id`; colisión si día repite ejercicio | Cuando un programa tenga ejercicio duplicado en un día |
| **`exercise_swap` no se aplica** | Texto libre sin id resoluble contra catálogo (decisión producto) | Si se decide implementar swap real |
| **AISuggestionRepository** | Mezcla 4 agregados ajenos (counts assignment/session/habit) | Nuevo tipo de sugerencia que necesite otro agregado |
| **Ports con métodos muertos** | `GetByID`, `CountCompletedSessionsOnDate`, `GetHabitByID`, `GetProgressionRuleByID`, `GetByID` | Limpieza en próximo ADR de refactor |
| **Tablas sin uso** | `progress_photo` (especificada en SDD-002), `exercise_alternative`, `user_identity` (OAuth). `device_token` ya **no** está aquí: tiene queries, repositorio, endpoints y adapter | Cuando el SDD correspondiente se active |
| **Auditoría incompleta** | Solo 4 mutaciones escriben `audit_log`. Sin cubrir: `habit.log`, `sync.sessions`, `progress.body_metrics`, y el cambio de rol de miembro | Cuando el gimnasio tenga más de un `admin`, o ante la primera disputa sobre quién cambió qué |
| **Cero tests de handler** | `internal/transport/http/handler/` no tiene ni un test; el middleware sí (21 casos) | Primer bug que se escape por serialización de DTO o mapeo de código HTTP |

---

## 7. Estados de SDD y Flujo de Archivos

```
docs/sdd/
├── templates/
│   ├── 01_FEATURE_SDD.md          # Plantilla feature full-stack
│   └── 02_ADR_ARCHITECTURE.md     # Plantilla decisión arquitectura
├── active/
│   ├── SDD-001-push-notifications.md     # Approved
│   ├── SDD-002-progress-photos.md        # Draft
│   └── SDD-003-worker-operation.md       # Draft
└── completed/
    └── README.md                     # Índice de SDDs cerrados
```

**Transiciones de estado:**
- `Draft` → `Approved` : Revisión arquitectónica + Definition of Ready (Fase 1 completa)
- `Approved` → `In Progress` : Inicio Fase 2 (schema/ports)
- `In Progress` → `Done` : Definition of Done (Fase 5 completa) → mover a `completed/`

---

## 8. Referencia Rápida: Comandos de Verificación por Componente

```bash
# Backend (api/)
cd api
go vet ./... && go test ./... && go build ./... && gofmt -l .
goose -dir db/migrations postgres "$DATABASE_URL" up
sqlc generate

# Panel Coach (web/)
cd web
pnpm lint && pnpm build

# App Flutter (app/)
cd app
flutter analyze && flutter test

# GitNexus (desde raíz del repo)
# Impact analysis antes de editar
# detect_changes antes de commit
```

---

## 9. Cambios Recientes en la Metodología (Changelog)

| Fecha | Cambio | Motivo |
|-------|--------|--------|
| 2026-08-31 | Añadidos **Definition of Ready/Done** por fase | Gaps detectados: impact analysis no obligatorio, tests handlers faltantes, audit incompleto |
| 2026-08-31 | **GitNexus `impact()` y `detect_changes()` obligatorios** en Fase 5 | Prevenir roturas silenciosas, reemplazar "grep manual" |
| 2026-08-31 | **Cobertura mínima de tests** especificada en Fase 3 | Solo dominio puro tenía tests; services/handlers/flutter sin red |
| 2026-08-31 | **Checklist de completitud SDD** en §3 | SDDs incompletos (falta impact analysis, consumidores, anexos) |
| 2026-08-31 | **Regla 7 y 8** en §4 (TxRepos, Progression engine) | Deuda documentada en AGENTS.md ahora codificada en metodología |
| 2026-08-31 | **Tooling obligatorio** §5 (GitNexus, Context7, Skills) | Estandarizar cómo se usa el grafo de código y docs externas |

---

*Última actualización: 2026-08-31 — alineado con estado real del código (fases 0-10 completadas, cadena de producto cerrada, 3 SDDs activos).*