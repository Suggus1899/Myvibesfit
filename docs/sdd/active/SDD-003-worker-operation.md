# SDD-003: Operación del Worker (cron, rachas y observabilidad)
> **Tipo:** ADR + Feature de Backend
> **Fecha:** 2026-08-31
> **Autor / Responsable:** Gustavo Colina
> **Estado:** `Draft`

---

## 1. Contexto y Problema

`cmd/worker` está construido, tiene un presupuesto de 30 minutos por corrida y sale con código
distinto de cero cuando falla. Lo que no tiene es **quién lo ejecuta**. Hoy corre solo si alguien
lo lanza a mano, así que en un despliegue real la mitad supervisada de la IA —la propuesta que el
coach aprueba— nunca llega a existir.

Y hay un segundo trabajo por lotes ya especificado y sin casa: el recordatorio de racha en riesgo
de [SDD-001](SDD-001-push-notifications.md) §5.4. Este documento decide dónde vive.

### Objetivos y Criterios de Éxito

- [ ] **O1:** El pase de sugerencias corre solo, una vez por noche, sin intervención.
- [ ] **O2:** Dos corridas no se solapan aunque una se cuelgue.
- [ ] **O3:** El recordatorio de racha llega en la tarde del usuario, no en la del servidor.
- [ ] **O4:** Una corrida fallida es visible sin entrar por SSH a leer logs.
- [ ] **O5:** Ejecutar el mismo pase dos veces el mismo día no duplica sugerencias ni avisos.

### Fuera de Alcance

- Reintentos automáticos por asignación. Hoy la que falla se salta y se registra; el pase de
  mañana lo intenta de nuevo. Alcanza mientras la cadencia sea diaria.
- Cola de trabajos (Asynq, River, NATS). Son dos trabajos por lotes, no un sistema de colas.
- Multi-instancia del worker. Un proceso cubre el volumen previsible por bastante tiempo.

---

## 2. Decisión: cron del sistema, no scheduler embebido

**Elegido: `cmd/worker` sigue siendo un binario de una corrida, disparado por cron del sistema
(o el timer del orquestador).**

| Criterio | Scheduler en `cmd/api` | Cron del sistema (elegido) | Cola de trabajos |
|---|---|---|---|
| Complejidad | Media | **Baja** | Alta |
| Aísla un fallo del worker de la API | No | **Sí** | Sí |
| Solapamiento | Hay que resolverlo en código | **`flock`, una línea** | Resuelto |
| Observable con lo que ya hay | Parcial | **Sí, código de salida** | Requiere panel propio |
| Escala a N instancias | No | No | Sí |

Un scheduler dentro de `cmd/api` ataría la generación de sugerencias al ciclo de vida del
servidor HTTP: reiniciar la API por un despliegue cancelaría un pase a media ejecución, y un
worker que agota memoria se llevaría puesta la API. La separación que ya existe es correcta; solo
falta el disparador.

**No solapamiento (O2):** `flock -n` sobre un archivo de lock. Si el pase anterior sigue vivo, el
nuevo no arranca. Una línea, sin estado en la base.

```
# /etc/cron.d/myvibesfit
30 3 * * *   myvibesfit  flock -n /tmp/mvf-suggestions.lock /opt/myvibesfit/worker suggestions
*/30 * * * * myvibesfit  flock -n /tmp/mvf-streaks.lock     /opt/myvibesfit/worker streaks
```

El binario pasa a aceptar un **subcomando** (`suggestions` | `streaks`). Sin subcomando ejecuta
`suggestions`, para no romper la invocación actual.

---

## 3. La pasada de rachas

Corre cada 30 minutos, no una vez al día, y eso es lo que resuelve **O3**: en cada corrida
notifica solo a los usuarios para quienes *en su propia zona* ya pasó la hora configurada.

```sql
-- name: ListStreaksAtRisk :many
-- Racha viva, sin log de hoy, y en la zona del usuario ya paso la hora de aviso.
SELECT s.user_id, s.current_count, u.timezone
FROM user_streak s
JOIN app_user u ON u.id = s.user_id
WHERE s.current_count > 0
  AND u.deleted_at IS NULL
  AND EXTRACT(HOUR FROM now() AT TIME ZONE u.timezone) >= @reminder_hour
  AND NOT EXISTS (
    SELECT 1 FROM habit_log l
    JOIN client_habit ch ON ch.id = l.client_habit_id
    WHERE ch.user_id = s.user_id
      AND l.log_date = (now() AT TIME ZONE u.timezone)::date
  );
```

> **Corrección heredada de SDD-001.** Ese documento afirmaba que `app_user` no guarda zona
> horaria. Es falso: la columna existe desde `0001_init.sql` con default `'UTC'`, y
> `domain.User.Timezone` ya la transporta. La deuda real es que **ninguna query la escribe**, así
> que hoy todo usuario se comporta como UTC. Esta query ya la respeta; poblarla es trabajo del
> onboarding y queda anotado como dependencia de O3.

**Idempotencia (O5):** el `NOT EXISTS` es la guarda natural —si ya registró, no se le avisa—,
pero no evita dos avisos el mismo día a quien nunca registra. Hace falta una marca de "último
aviso enviado". Lo más barato sin tabla nueva es una columna `last_reminder_sent_on date` en
`user_streak`, filtrando `IS DISTINCT FROM` la fecha local. Es la **única migración** que pide
este SDD.

---

## 4. Observabilidad (O4)

Sin VPS todavía, la solución tiene que funcionar igual en una máquina sola:

1. **Código de salida distinto de cero** cuando el pase falla entero. Ya está.
2. **Resumen en JSON** al cerrar cada corrida —`processed`, `created`, `failed`, `duration_ms`—
   en una sola línea, para cazarla con `grep` sin parsear la corrida completa.
3. **Ping a un dead-man's switch** (Healthchecks.io o equivalente) al terminar bien, vía
   `HEALTHCHECK_PING_URL`. Opcional: sin la variable no se pinga y no falla. Es lo que convierte
   "no corrió" en una alerta — un cron que nunca se ejecuta no deja ningún log que mirar.

Se descarta Prometheus por ahora: exige un proceso que raspe métricas, y un binario que vive
treinta segundos por noche es el peor candidato posible para eso.

---

## 5. Cambios por capa

- **`cmd/worker/main.go`:** parseo del subcomando, `runSuggestions` y `runStreaks` separadas,
  resumen final y ping opcional.
- **`internal/service/`:** `StreakReminderService`, que consume `ListStreaksAtRisk` y llama a
  `NotificationService.NotifyUser` con `domain.StreakAtRiskNotification`, ya escrita en SDD-001.
- **`api/db/migrations/0006_streak_reminder.sql`:** la columna `last_reminder_sent_on`.
- **Despliegue:** el archivo de cron y el binario, fuera del repositorio de la app.

---

## 6. Consecuencias

**Positivas:** el disparo no cuesta código de aplicación; un worker colgado no toca la API; el
solapamiento lo resuelve el sistema operativo; la operación se lee sin herramientas nuevas.

**Trade-offs:** el cron vive fuera del repositorio y puede desincronizarse de lo que el binario
espera —se mitiga versionándolo junto al despliegue—, y no hay reintento dentro de la misma
noche.

---

## 7. Plan de Verificación

- [ ] **`ListStreaksAtRisk` contra Postgres real** con usuarios en dos zonas: a las 20:00 de
  Madrid entra el usuario de Madrid y no el de Buenos Aires.
- [ ] **Idempotencia (O5):** dos corridas seguidas producen un solo aviso.
- [ ] **No solapamiento (O2):** dos lanzamientos con `flock -n`; el segundo no arranca.
- [ ] **Subcomando ausente** sigue ejecutando `suggestions`.
- [ ] `cd api && go vet ./... && go test ./... && go build ./...`
