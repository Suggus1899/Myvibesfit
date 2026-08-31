# SDD-001: Notificaciones Push
> **Tipo:** Feature Full-Stack (Backend + Móvil)
> **Fecha:** 2026-08-31
> **Autor / Responsable:** Gustavo Colina
> **Estado:** `Approved` — backend implementado y verificado; pendientes el adapter FCM y la integración Flutter (ver §9)

---

## 1. Contexto y Objetivos de Producto

### 1.1 Problema del Usuario

El cliente no se entera de nada fuera de la app. Cuando el coach le asigna un programa, el
plan aparece en silencio: hay que abrir la app y mirar. Y la racha —el único mecanismo de
retención que ya está construido y funcionando— se rompe sin aviso, que es exactamente el
momento en que un recordatorio vale algo.

La tabla `device_token` está modelada desde `0003_execution.sql:231` y **no tiene una sola
query**: hoy solo existe el struct que generó sqlc (`repository/db/models.go:1186`). Este SDD
la pone en uso.

Roles impactados: `client` (recibe). El `coach` queda fuera de este incremento — ver 1.3.

### 1.2 Objetivos y Criterios de Éxito

- [ ] **O1:** Un cliente con la app instalada recibe una push al minuto de que su coach le
      asigna un programa, con la app cerrada.
- [ ] **O2:** Un cliente con racha activa que no registró su hábito recibe un recordatorio en
      la ventana horaria configurada, una sola vez por día.
- [x] **O3:** Registrar el mismo dispositivo N veces deja **una** fila en `device_token`.
- [ ] **O4:** Un token que FCM reporta como muerto se borra solo; la tabla no crece sin techo.
- [x] **O5:** Sin credenciales de FCM configuradas, la API arranca igual y el envío es un no-op.
      El desarrollo local no depende de Firebase.

### 1.3 Fuera de Alcance (Non-goals)

- **Notificar al coach** (sugerencia de IA nueva, cliente que abandona). El coach trabaja con
  el panel abierto en un escritorio; la push le aporta poco y duplica el trabajo de decidir
  copy y deep-link. Se agrega cuando haya una queja real, reusando el mismo `PushSender`.
- **Web push.** El `CHECK (platform IN ('ios','android'))` excluye `web` a propósito: el build
  Flutter web es la herramienta de verificación local, no un canal de producto.
- Centro de notificaciones in-app, notificaciones con acciones, imágenes o agrupación.
- Preferencias por tipo de notificación (silenciar solo rachas). Se difiere hasta tener los dos
  tipos vivos; hoy sería configuración para un valor que nadie cambió todavía.

---

## 2. Modelo de Dominio e Invariantes (`api/internal/domain/`)

### 2.1 Entidades y Puertos

```go
package domain

type DeviceToken struct {
    ID         uuid.UUID
    UserID     uuid.UUID
    Token      string
    Platform   string // "ios" | "android"
    LastSeenAt time.Time
    CreatedAt  time.Time
}

const (
    PlatformIOS     = "ios"
    PlatformAndroid = "android"
)

type DeviceTokenRepository interface {
    Upsert(ctx context.Context, userID uuid.UUID, token, platform string) (DeviceToken, error)
    ListByUser(ctx context.Context, userID uuid.UUID) ([]DeviceToken, error)
    Delete(ctx context.Context, userID uuid.UUID, token string) (bool, error)
}

// PushSender es el port hacia el proveedor de push, con la misma forma que
// SuggestionProposer: el domain declara la intencion, el adapter conoce el SDK.
type PushSender interface {
    Send(ctx context.Context, tokens []string, n Notification) (dead []string, err error)
}

type Notification struct {
    Title string
    Body  string
    Data  map[string]string // deep-link: {"route": "/workout"}
}
```

`Send` devuelve los tokens que el proveedor reportó como inválidos. Esa devolución es lo que
hace cumplible **O4** sin un job de limpieza aparte.

### 2.2 Invariantes

1. **Un dispositivo, una fila.** `device_token.token` es `UNIQUE` **global**, no por usuario
   (`0003_execution.sql:234`). Si el usuario A cierra sesión y el B entra en el mismo teléfono,
   FCM entrega el mismo token: la fila debe **cambiar de dueño**, no chocar. El upsert resuelve
   `ON CONFLICT (token) DO UPDATE SET user_id = EXCLUDED.user_id`. Resolverlo contra
   `(user_id, token)` haría fallar el registro del segundo usuario con un 500 — es la misma
   clase de error que la regla de oro 4 describe para `habit_log`.
2. **Sin `org_id`.** `device_token` cuelga de `app_user`, no de la organización. Una push nunca
   lleva datos del gimnasio en el cuerpo, solo un deep-link; así un usuario que ya salió de la
   org no recibe contenido que no le corresponde.
3. **El envío nunca rompe la operación de negocio.** Un fallo de FCM se registra y se sigue.
   Asignar un programa no puede fallar porque Google esté caído.
4. **Un recordatorio de racha por usuario y día.** El batch filtra por hábitos sin log de hoy;
   la ventana horaria evita que dos corridas del cron dupliquen el aviso.

---

## 3. Esquema y Persistencia (`api/db/`)

**No hace falta migración.** La tabla existe desde `0003_execution.sql:231-239` con la forma que
esta feature necesita, índice sobre `user_id` incluido. Solo faltan las queries.

### 3.1 Queries SQLC (`api/db/queries/device_token.sql`)

```sql
-- name: UpsertDeviceToken :one
INSERT INTO device_token (user_id, token, platform)
VALUES ($1, $2, $3)
ON CONFLICT (token) DO UPDATE
  SET user_id = EXCLUDED.user_id,
      platform = EXCLUDED.platform,
      last_seen_at = now()
RETURNING *;

-- name: ListDeviceTokensByUser :many
SELECT * FROM device_token WHERE user_id = $1 ORDER BY last_seen_at DESC;

-- name: DeleteDeviceToken :execrows
DELETE FROM device_token WHERE token = $1 AND user_id = $2;
```

`DeleteDeviceToken` es `:execrows` para poder distinguir "no estaba" de "se borró" sin una
lectura previa.

---

## 4. Contratos de API REST (`api/internal/transport/http/`)

| Método | Ruta | Middleware | Roles | Descripción |
|---|---|---|---|---|
| POST | `/v1/me/device-tokens` | `auth` | cualquiera | Registra o refresca el token del dispositivo |
| DELETE | `/v1/me/device-tokens` | `auth` | cualquiera | Baja el token (logout) |

Cuelgan del grupo `/me` ya existente (`router.go:68-88`). No llevan `RequireOrg`: el registro
del dispositivo es anterior e independiente de pertenecer a un gimnasio.

```go
type RegisterDeviceTokenRequest struct {
    Token    string `json:"token"`
    Platform string `json:"platform"` // "ios" | "android"
}

type DeviceTokenResponse struct {
    Token      string    `json:"token"`
    Platform   string    `json:"platform"`
    LastSeenAt time.Time `json:"last_seen_at"`
}
```

Validación en el handler: `token` no vacío, `platform` dentro del enum. Cualquier otro valor es
`400`, no un `500` de la constraint.

---

## 5. Diseño Hexagonal (Backend Go)

### 5.1 Adapter (`internal/adapter/push/`)

`fcm.go` implementa `domain.PushSender` sobre `firebase.google.com/go/v4/messaging`. Traduce la respuesta
de FCM: `UNREGISTERED` e `INVALID_ARGUMENT` sobre el token → lista `dead`. Es el único archivo
que importa el SDK de Firebase, igual que `adapter/anthropic/` es el único que importa el de
Anthropic.

**`noop.go`** en el mismo paquete: implementa el port y registra en log lo que habría
enviado. Es lo que se cablea cuando no hay credenciales (**O5**).

### 5.2 Configuración (`internal/config/config.go`)

```go
FCMProjectID       string // FCM_PROJECT_ID
FCMCredentialsJSON string // FCM_CREDENTIALS_JSON (ruta al service account)
```

Ambas **opcionales**, con `getEnv(key, "")`. A diferencia de `DATABASE_URL` y
`JWT_ACCESS_SECRET`, su ausencia no aborta el arranque: `cmd/api` cablea `NoopSender` y deja una
línea de log diciéndolo. Nunca hardcodeadas, nunca commiteadas.

### 5.3 Servicio (`internal/service/notification.go`)

```go
func (s *NotificationService) NotifyUser(ctx context.Context, userID uuid.UUID, n domain.Notification)
```

Lee los tokens del usuario, llama al sender, borra los que volvieron muertos. **No devuelve
error al caller**: registra y sigue (invariante 3).

### 5.4 Disparadores

**Programa asignado** — en `AssignmentService.Assign`, **después** de que `uow.Execute` retorne
sin error, nunca adentro. Una llamada HTTP a Google dentro de una transacción de Postgres
mantiene la transacción abierta a merced de la latencia de red.

```go
// ponytail: entrega at-most-once. Si el proceso muere entre el commit y el envio,
// esa push se pierde. Outbox + reintento cuando perder un aviso duela de verdad.
if err := uow.Execute(ctx, fn); err != nil {
    return err
}
s.notifications.NotifyUser(ctx, clientUserID, planAssignedNotification(programName))
```

**Racha en riesgo** — pasada nueva en `cmd/worker`, que ya es un binario de una corrida pensado
para cron. Busca usuarios con racha activa y sin `habit_log` de hoy, y notifica. La hora de la
ventana sale de `STREAK_REMINDER_HOUR` (default `20`).

> Corregido tras verificar contra el esquema: `app_user` **sí** tiene columna `timezone`
> (`0001_init.sql`, default `'UTC'`), y `domain.User.Timezone` ya la transporta. La deuda real es
> otra: **ninguna query la escribe**, así que hoy todos los usuarios leen como `UTC`. El batch
> debe agrupar por `timezone` desde el principio —el esquema lo permite— y queda pendiente que el
> onboarding envíe la zona real del dispositivo.


---

## 6. Frontend: Panel Web Coach (`web/`)

Sin cambios. El coach no recibe push en este incremento (1.3) y el registro de dispositivos es
del cliente móvil.

---

## 7. Móvil: App Flutter (`app/`)

- **Dependencia:** `firebase_messaging`. Requiere `google-services.json` (Android) y
  `GoogleService-Info.plist` (iOS), ambos fuera de git.
- **Permiso:** iOS siempre lo pide; Android 13+ requiere `POST_NOTIFICATIONS` en runtime. Se
  pide **después** del primer entrenamiento completado, no en el arranque: pedirlo en frío es la
  forma más rápida de que lo nieguen para siempre.
- **Ciclo de vida del token:**
  - Tras login/registro exitoso → `POST /v1/me/device-tokens`.
  - `onTokenRefresh` → `POST` de nuevo (el upsert lo absorbe).
  - En logout → `DELETE` **antes** de borrar los tokens JWT, o la petición sale sin auth.
- **Guarda de plataforma:** en `kIsWeb` no registrar nada. El `CHECK` de la tabla rechazaría
  `web` con un 500 y el flujo de desarrollo local se rompería sin motivo.
- **Deep-link:** `Notification.Data["route"]` se pasa a `go_router`. Con la app cerrada,
  `getInitialMessage()` se lee antes del primer `push` de ruta.

---

## 8. Seguridad, Multi-Tenancy y Auditoría

- El token del dispositivo se asocia **siempre** al `user_id` del JWT, nunca a uno del body.
- El cuerpo de la push no lleva PII ni datos del gimnasio: título, texto genérico y ruta.
- `FCM_CREDENTIALS_JSON` apunta a un service account con permiso de envío y nada más.
- **Sin `audit_log`.** Registrar cada push llenaría la tabla de eventos que no son decisiones de
  negocio. La acción auditable (`assignment.create`) ya se registra donde corresponde.

---

## 9. Plan de Verificación y Testing

- [x] **Dominio/servicio con fakes:** `go test ./internal/service/...`
  - `NotifyUser` borra los tokens que el sender devolvió como muertos.
  - Un `PushSender` que devuelve error **no** hace fallar `Assign`.
  - Sin tokens registrados, no se llama al sender.
- [x] **Upsert contra Postgres real:** registrar el mismo token con el usuario A y después con el
  B → **una** fila, `user_id` = B. Es la invariante 1 y ningún test con fakes la prueba.
- [x] **Arranque sin credenciales:** `cmd/api` levanta y loguea que el push está deshabilitado (**O5**).
- [ ] **Verificación de capas:**
  - Backend: `cd api && go vet ./... && go test ./... && go build ./... && gofmt -l .`
  - Móvil: `cd app && flutter analyze && flutter test`
- [ ] **Prueba en dispositivo real:** con la app **cerrada**, asignar un programa desde el panel y
  confirmar que llega y que el tap abre la pantalla del plan. Un emulador sin Play Services no
  prueba esto.

---

## Anexo C — Sync offline y móvil

El registro del token **no** entra en la cola de sincronización. Es una operación idempotente y
barata: si falla, el próximo `onTokenRefresh` o el próximo login la reintenta. Meterla en la cola
de sesiones mezclaría datos de entrenamiento con estado de dispositivo.

El único punto de contacto con el sync existente: el `DELETE` del logout compite con una cola de
sesiones pendientes. **Orden obligatorio:** intentar `syncNow()`, después `DELETE` del token,
después limpiar los JWT. Invertirlo deja sesiones huérfanas sin credenciales para subirlas.
