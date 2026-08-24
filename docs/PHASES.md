# Myvibesfit — Plan por fases

Orden de construcción. Cada fase termina en algo que corre y se verifica, no en código
suelto. Ver decisiones y motivos en [ARCHITECTURE.md](./ARCHITECTURE.md).

## Fase 0 — Cimientos (½–1 semana)

Monorepo, Docker Compose (Postgres + API + adminer), goose + sqlc funcionando,
esqueleto Go con `chi` respondiendo `/health`, CI mínimo (lint + build + test).

**Listo cuando:** `docker compose up` levanta Postgres, las 3 migraciones ya escritas
(`0001_init`, `0002_training`, `0003_execution`) corren limpias, y `sqlc generate`
produce tipos sin error.

## Fase 1 — Identidad y tenencia

`app_user`, `organization`, `membership`, registro/login/refresh JWT, middleware
`auth` + `tenant`, endpoint `POST /v1/orgs/join` (vincular por código). Seed de una
organización de prueba.

**Listo cuando:** un usuario se registra, se vincula a un gimnasio por código, y el
JWT trae `org_id` + rol resuelto por el middleware. Tests de servicio sobre auth.

## Fase 2 — Catálogo de ejercicios

CRUD de `exercise` (admin/coach), endpoint público `GET /v1/exercises` con filtro por
patrón/músculo/equipo, seed de 200-300 ejercicios propios con media. Sube imágenes a
S3/R2 vía `platform/storage`.

**Listo cuando:** el catálogo global + el del gimnasio se listan juntos, con media
servida por URL firmada o pública según corresponda.

## Fase 3 — Programación (coach)

`program`, `program_workout`, `program_exercise`, `progression_rule`. CRUD completo
desde el panel Next.js: crear programa, añadir semanas/días, añadir ejercicios con
sus objetivos. Motor de progresión (`internal/progression/`) con tests unitarios de
las 4 estrategias, sin BD.

**Listo cuando:** un coach arma un mesociclo de 4 semanas en el panel web y queda
persistido en `program_status = draft`.

## Fase 4 — Asignación

`assignment`, `assigned_workout`, `assigned_exercise` — la copia del programa al
publicarlo para un cliente. Endpoint de asignación desde el panel, endpoint
`GET /v1/me/assignment/current` para el móvil.

**Listo cuando:** publicar un programa genera la copia completa para el cliente, y
editar la copia no toca la plantilla original.

## Fase 5 — App Flutter: esqueleto + registro de entrenamiento

Setup Flutter (Riverpod, go_router, dio, Drift), theme desde `docs/DESIGN.md` con
`ThemeExtension<AppColors>`, auth flow, pantalla de entrenamiento activo, registro de
serie (peso/reps/RPE), temporizador de descanso, cierre de sesión.

Cola de sincronización: `client_local_id` por sesión/serie, `POST /v1/sync/sessions`
en lote, reintento al recuperar red.

**Listo cuando:** se puede completar un entrenamiento asignado de principio a fin en
el móvil, en avión, y sincroniza al volver la señal sin duplicar filas.

## Fase 6 — Progreso y récords

`GET /v1/progress/exercises/{id}` (serie temporal), `GET /v1/progress/records`
(cálculo de PRs sobre `set_log`), `GET /v1/progress/volume`. Pantallas de gráficas en
Flutter, métricas corporales (`body_metric`) con registro manual.

**Listo cuando:** tras varias sesiones, el usuario ve su curva de peso levantado y sus
PRs se actualizan solos.

## Fase 7 — Hábitos y gamificación

`habit`, `client_habit`, `habit_log` con check diario. `user_streak`, `user_stats`,
`xp_event`, `achievement` — job que calcula racha y XP tras cada sync. Anillo de
progreso diario y tarjeta de racha en el home.

**Listo cuando:** entrenar y marcar hábitos suma XP y mantiene una racha visible, con
al menos 5 logros desbloqueables.

## Fase 8 — Panel del coach: seguimiento

`GET /v1/coach/clients`, vista de overview por cliente (adherencia, PRs recientes,
alertas de sesiones saltadas). Es lo que convierte el panel de "editor de planes" a
"herramienta de trabajo diaria" del coach.

**Listo cuando:** un coach abre su panel y en una pantalla ve qué cliente necesita
atención hoy, sin entrar cliente por cliente.

## Fase 9 — IA: sugerencias supervisadas

Job nocturno (`cmd/worker`) que arma el resumen estructurado por asignación, llama a
Claude, valida la respuesta contra el esquema, guarda `ai_suggestion`. Tarjetas de
aprobación en el panel del coach (`POST .../approve|reject`).

**Listo cuando:** tras una semana de datos simulados, el worker genera sugerencias
razonables y ninguna se aplica sin que el coach pulse Aprobar.

## Fase 10 — Endurecimiento y lanzamiento

Rate limiting, auditoría (`audit_log`) en mutaciones sensibles, push notifications
(`device_token`), onboarding pulido, políticas de privacidad para fotos de progreso,
backups automáticos de Postgres, monitoreo (logs + métricas básicas) en el VPS.

**Listo cuando:** pasa una checklist de seguridad básica y sobrevive una prueba de
carga ligera (100 usuarios concurrentes registrando series).

---

## Qué NO entra en este plan (v2+)

Nutrición/macros · mensajería coach↔cliente · pagos y suscripciones · wearables
(HealthKit/Health Connect) · feed social · marketplace de coaches. El esquema ya
las contempla sin romper lo construido; se activan cuando haya demanda real.

## Orden de dependencia (resumen)

```
Fase 0 → 1 → 2 → 3 → 4 → 5 → 6
                          ↘ 7 (paralelo a 6, no depende de 6)
                        4 → 8 (paralelo a 5/6/7)
              6 + 7 + 8 → 9
                     9 → 10
```

Fases 5, 6 y 7 pueden solaparse una vez existe el esqueleto Flutter (fin de fase 5).
La fase 8 (panel de seguimiento) puede arrancar en paralelo tan pronto exista
asignación (fin de fase 4) — no depende de que el móvil esté terminado.
