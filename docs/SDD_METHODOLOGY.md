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
 └────────────────────────────────────────────────────────┘
```

---

## 2. Las 5 Fases del Ciclo de Vida SDD

### Fase 1: Redacción del SDD (Software Design Document)
Para cualquier épica, nueva feature o refactor significativo, se crea o actualiza un documento en `docs/sdd/active/SDD-<NUMERO>-<nombre>.md` siguiendo el [Catálogo de Plantillas SDD](sdd/README.md) (ver detalle en §3 y en `docs/sdd/templates/`).

**Criterios de salida de Fase 1:**
- Modelo de dominio e invariantes de negocio definidos.
- Contratos de datos (DDL SQL, DTOs de petición/respuesta) especificados.
- Impact analysis previo evaluado (riesgos, tablas afectadas, consumidores en Web y App).

### Fase 2: Especificación de Esquemas y Puertos
Se definen las fuentes de verdad de datos y contratos en el backend:
1. **Migración SQL (Goose):** En `api/db/migrations/`, respetando multi-tenancy (`org_id`), llaves foráneas e índices únicos.
2. **Queries SQLC:** En `api/db/queries/<modulo>.sql`.
3. **Generación SQLC:** Ejecución de `sqlc generate` en `api/` para generar código en `internal/repository/db/`.
4. **Domain Ports & Entities:** Definición de interfaces de repositorio (`*Repository`) y entidades puras en `internal/domain/` sin dependencias externas (solo stdlib + `google/uuid`).
5. **Unit of Work:** Si el caso de uso involucra múltiples agregados de forma transaccional, se declara en `domain.UnitOfWork` y `domain.TxRepos`.

### Fase 3: Pruebas Unitarias de Dominio y Contratos
Antes o en paralelo a la lógica de persistencia:
- Escribir tests unitarios para reglas de negocio complejas en `internal/domain/` o `internal/progression/`.
- Crear o actualizar mocks/fakes de los puertos de dominio para probar los casos de uso en `internal/service/` sin levantar base de datos real.

### Fase 4: Implementación Hexagonal
Siguiendo el flujo de capas desacopladas:
1. **Adapters Secundarios (`internal/adapter/postgres/`):** Implementar los puertos de dominio mapeando desde/hacia los tipos de `db` (sqlc). Traducir `pgx.ErrNoRows` a `domain.ErrNotFound`.
2. **Adapters Externos (`internal/adapter/anthropic/` u otros):** Implementar puertos de integración externa (IA, storage, etc.).
3. **Services (`internal/service/`):** Implementar la orquestación del caso de uso. Depender únicamente de los puertos de dominio.
4. **Transport / Handlers / DTOs (`internal/transport/http/`):**
   - Definir DTOs explícitos de Request y Response en `dto/`.
   - Implementar Handlers en `handler/` (decodificación, validación de forma, invocación de service, mapeo de errores HTTP).
   - Registrar rutas y middlewares (Auth JWT, RBAC `RequireRole`, Tenant) en `router.go`.
5. **Consumidores Frontend / Móvil:**
   - **Web (`web/`):** Crear/actualizar servicios API en `src/lib/api.ts`, componentes UI en `src/components/`, y páginas en `src/app/`.
   - **Flutter (`app/`):** Modelos en `core/network/models.dart`, queries locales Drift en `core/storage/`, StateNotifier / Riverpod providers, y vistas UI.

### Fase 5: Verificación, Impact Analysis y Auditoría
1. **Impact Analysis:** Verificar callers y posibles roturas con GitNexus / pruebas estáticas.
2. **Pipeline de Verificación:**
   - Backend: `go vet ./... && go test ./... && go build ./... && gofmt -l .`
   - Web: `pnpm lint && pnpm build`
   - Flutter: `flutter analyze && flutter test`
3. **Auditoría:** Asegurar que mutaciones críticas queden registradas mediante `domain.AuditRepository` / `AuditLogger`.

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
