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

Fases 0-4 de [PHASES.md](./docs/PHASES.md) completas y verificadas end-to-end
contra Postgres real (migraciones + seed + servidor + curl):

- **Esquema de BD** (`api/db/migrations/`, goose): `0001_init` (tenencia/identidad),
  `0002_training` (catálogo/programas), `0003_execution` (sesiones/hábitos/
  gamificación/IA).
- **API Go** (`api/`): auth JWT (access + refresh con rotación), vinculación a
  gimnasio por código, RBAC por rol, catálogo de ejercicios (CRUD + filtro
  global/gym, 40 ejercicios semilla), CRUD de programas/días/ejercicios del
  coach (`program`, `program_workout`, `program_exercise`), motor de
  progresión puro en `internal/progression/` (4 estrategias con tests, sin
  BD), 4 reglas de progresión semilla, y asignación con copia transaccional
  programa→cliente (`assignment`/`assigned_workout`/`assigned_exercise`,
  independiente de la plantilla tras copiarse).
- **No incluido aún:** el panel Next.js (`web/`) donde el coach usaría esta
  API con UI — por ahora solo existe el backend, probado por API directa.
- **Siguiente paso:** Fase 5 (Flutter: esqueleto + registro de entrenamiento).

Notas del entorno:
- `docker-compose.yml` mapea Postgres a `5433:5432` en este equipo porque el
  5432 nativo ya estaba ocupado por otra instancia local del usuario.
- El almacenamiento de Docker Desktop (imágenes, contenedores, volúmenes)
  vive físicamente en `F:\Docker\wsl\`, enlazado por junction NTFS desde
  `%LOCALAPPDATA%\Docker\wsl\` (Docker Desktop no expone esta ruta como
  setting editable; se relocalizó a mano). Ver estructura en `F:\Docker\`.

## Stack

Backend: Go 1.25 + chi + sqlc + PostgreSQL 16, capas
`handler → service → repository`. Panel coach: Next.js 16 + TS + Tailwind + shadcn.
App cliente: Flutter + Riverpod + Drift (cache offline) + dio. Monorepo, Docker
Compose en VPS.

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
```

Verificación antes de commit: `go vet ./... && go test ./... && go build ./...`
(API) · `pnpm lint && pnpm build` (panel) · `dart analyze && flutter test` (app).
