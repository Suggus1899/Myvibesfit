# SDD-002: Fotos de Progreso y Capa de Almacenamiento
> **Tipo:** Feature Full-Stack (Backend + Móvil)
> **Fecha:** 2026-08-31
> **Autor / Responsable:** Gustavo Colina
> **Estado:** `Draft`

---

## 1. Contexto y Objetivos de Producto

### 1.1 Problema del Usuario

La báscula miente durante una recomposición: el peso no se mueve y el cuerpo sí. La foto
mensual es la única evidencia que el cliente acepta cuando el número no acompaña, y hoy la app
no tiene dónde ponerla.

`progress_photo` está modelada desde `0001_init.sql:155` con **cero queries**. La columna
`storage_key` deja claro qué falta: no hay ninguna capa de almacenamiento de objetos en el
proyecto, y este es el primer caso que la necesita de verdad.

### 1.2 Objetivos y Criterios de Éxito

- [ ] **O1:** El cliente sube una foto desde la app y la ve en su línea de tiempo.
- [ ] **O2:** Una foto es **privada por defecto**; el coach solo la ve si el cliente la comparte.
- [ ] **O3:** Ningún byte de imagen atraviesa la API: se sube y se descarga con URLs firmadas de
      vida corta.
- [ ] **O4:** Revocar el compartido con el coach surte efecto en menos de lo que dura una URL
      firmada.
- [ ] **O5:** Sin credenciales de almacenamiento la API arranca igual y los endpoints de foto
      devuelven `503`, sin romper el resto del backend.

### 1.3 Fuera de Alcance (Non-goals)

- **Media de ejercicios.** Verificado contra el esquema: `exercise.video_url` y
  `exercise.thumbnail_url` son `text` con URLs externas. No necesitan esta capa, y meterlos aquí
  duplicaría trabajo por una falsa simetría. El plan original agrupaba ambos; el esquema dice que
  son problemas distintos.
- Comparación lado a lado, superposición, detección de pose, marcas de agua.
- Álbumes, etiquetas o comentarios sobre la foto.
- Compartir fuera de la app (link público, redes).

---

## 2. Modelo de Dominio e Invariantes

```go
type ProgressPhoto struct {
    ID              uuid.UUID
    UserID          uuid.UUID
    StorageKey      string
    Pose            string
    TakenOn         time.Time
    SharedWithCoach bool
    CreatedAt       time.Time
}

type ProgressPhotoRepository interface {
    Create(ctx context.Context, p ProgressPhoto) (ProgressPhoto, error)
    ListByUser(ctx context.Context, userID uuid.UUID) ([]ProgressPhoto, error)
    ListSharedWithCoach(ctx context.Context, clientUserID uuid.UUID) ([]ProgressPhoto, error)
    SetShared(ctx context.Context, id, userID uuid.UUID, shared bool) error
    Delete(ctx context.Context, id, userID uuid.UUID) (string, error) // devuelve la storage_key
}

// ObjectStore es el port de almacenamiento. Igual que PushSender y
// SuggestionProposer: el domain no sabe si detras hay S3, R2 o disco.
type ObjectStore interface {
    SignedUpload(ctx context.Context, key, contentType string, ttl time.Duration) (string, error)
    SignedDownload(ctx context.Context, key string, ttl time.Duration) (string, error)
    Delete(ctx context.Context, key string) error
}
```

### Invariantes

1. **Privada por defecto.** `shared_with_coach` es `false` en el esquema y ninguna ruta lo
   invierte implícitamente. Compartir es siempre un acto explícito del cliente.
2. **La `storage_key` la genera el servidor**, nunca el cliente:
   `progress/{user_id}/{uuid}.{ext}`. Aceptarla del cliente permitiría escribir en la carpeta de
   otro usuario o pisar una foto existente.
3. **El coach lee, nunca escribe.** No hay ruta que permita a un coach subir o borrar una foto de
   su cliente.
4. **Borrar es borrar de verdad.** A diferencia del resto del proyecto —donde se decidió
   soft-delete en todo—, una foto del cuerpo que el usuario pide eliminar se borra de la base y
   del bucket. Es la excepción, y es deliberada.
5. **TTL corto de las URLs firmadas** (5 minutos de subida, 5 de descarga). Es lo que hace
   cumplible **O4**: una URL vieja caduca sola en vez de sobrevivir a la revocación.

---

## 3. Esquema y Persistencia

**No hace falta migración.** La tabla existe con la forma necesaria e índice
`(user_id, taken_on DESC)`. Solo faltan las queries en `api/db/queries/progress_photo.sql`.

Las de listado ordenan por `taken_on DESC`, que es exactamente el índice que ya está.

---

## 4. Contratos de API REST

| Método | Ruta | Middleware | Roles | Descripción |
|---|---|---|---|---|
| POST | `/v1/me/progress-photos/upload-url` | `auth` | client | Devuelve URL firmada y `storage_key` |
| POST | `/v1/me/progress-photos` | `auth` | client | Confirma la subida y crea la fila |
| GET | `/v1/me/progress-photos` | `auth` | client | Línea de tiempo con URLs de descarga |
| PATCH | `/v1/me/progress-photos/{id}/share` | `auth` | client | Comparte o deja de compartir |
| DELETE | `/v1/me/progress-photos/{id}` | `auth` | client | Borra fila y objeto |
| GET | `/v1/coach/clients/{id}/progress-photos` | `auth`, `RequireOrg`, `RequireRole` | coach, owner | Solo las compartidas |

**Subida en dos pasos** a propósito: pedir URL, subir directo al bucket, confirmar. Un solo paso
obligaría a que la imagen atraviese la API (**O3**), y con eso vendrían límites de tamaño de
request y memoria del proceso por algo que el bucket hace mejor.

El coste es un estado intermedio: una URL pedida y nunca confirmada deja un objeto huérfano. Se
acepta y se barre con un job posterior; hoy no hay volumen que lo justifique.

Validaciones: `content_type` en `image/jpeg` o `image/png`; `taken_on` no futura; `pose` dentro
de un enum corto (`front`, `back`, `side`) o vacía.

---

## 5. Diseño Hexagonal

### 5.1 Adapter (`internal/adapter/objectstore/`)

- `s3.go` implementa `ObjectStore` contra cualquier API compatible con S3 (R2, MinIO, S3). Se
  elige S3-compatible y no un SDK propietario para no atarse a un proveedor por una feature.
- `unavailable.go`: implementación que devuelve `domain.ErrUnavailable` en todo. Es lo que se
  cablea sin credenciales, y lo que hace que **O5** se cumpla sin `if` repartidos por los
  handlers.

Mismo patrón que `adapter/push`: el no-op se decide en el cableado, no en el caso de uso.

### 5.2 Configuración

```go
StorageEndpoint  string // STORAGE_ENDPOINT
StorageBucket    string // STORAGE_BUCKET
StorageAccessKey string // STORAGE_ACCESS_KEY
StorageSecretKey string // STORAGE_SECRET_KEY
```

Opcionales, como las de FCM. Presentes a medias → el arranque falla explícitamente en vez de
quedar a medio configurar.

---

## 6. Frontend: Panel Web Coach

Galería de solo lectura dentro de `/clients/[id]`, alimentada por el endpoint de coach. Sin
controles de subida ni de borrado: la invariante 3 tiene que notarse también en la UI.

---

## 7. Móvil: App Flutter

- `image_picker` para cámara y galería; compresión en el dispositivo antes de subir (una foto de
  12 MP no aporta nada frente a 1500 px de lado y cuesta datos móviles).
- Línea de tiempo agrupada por mes en la pantalla de progreso.
- El interruptor de compartir con el coach va **en la foto**, no en ajustes: la decisión es por
  imagen, y esconderla en una preferencia global invita a compartir de más.
- La subida **no** entra en la cola de sync de sesiones (ver Anexo C de la plantilla): son
  megabytes, no filas, y mezclarlas retrasaría el registro del entrenamiento.

---

## 8. Seguridad, Multi-Tenancy y Auditoría

- Toda ruta de cliente filtra por el `user_id` del JWT; el `id` del path nunca alcanza para
  autorizar.
- La ruta de coach verifica `IsClientOf` **y** `shared_with_coach = true`. Dos condiciones, no
  una: ser su coach no basta si el cliente no compartió.
- Se audita `progress_photo.share` y `progress_photo.delete` con `AuditRepository`. La subida no:
  es rutina, y llenaría la tabla.
- El bucket es privado. Sin acceso público ni listado.

---

## 9. Plan de Verificación y Testing

- [ ] **Servicio con fakes:** un `ObjectStore` falso; verificar que la `storage_key` la genera el
  servidor y contiene el `user_id` del token, no el del body.
- [ ] **Autorización cruzada, contra Postgres real:** el coach de A no ve las fotos de B; el coach
  de A no ve las fotos **no compartidas** de A. Es la invariante 3 y ningún test con fakes la
  cubre.
- [ ] **Borrado:** la fila desaparece y `ObjectStore.Delete` recibe la key correcta.
- [ ] **Arranque sin credenciales:** la API levanta y los endpoints de foto devuelven `503`
  mientras el resto responde normal (**O5**).
- [ ] Backend `go vet ./... && go test ./... && go build ./...` · Móvil `flutter analyze && flutter test`

---

## Anexo B — Nota de esquema

`progress_photo` no tiene `org_id`, igual que `device_token`. La foto es del usuario, no del
gimnasio: si cambia de gym, sus fotos lo siguen y el coach anterior deja de verlas porque
`IsClientOf` deja de ser cierto. Es el comportamiento correcto y sale gratis del modelo.

---

## Impact analysis previo (Definition of Ready, Fase 1)

Ejecutado el 2026-08-31 con el CLI de GitNexus, una vez recuperada la herramienta.

| Símbolo | Riesgo | Alcance | Epistémico |
|---|---|---|---|
| `config.Load` | `LOW` | 1 | `exact` |
| `NewCoachService` | `LOW` | 1 | `exact` |

Ambos como se esperaba: un solo caller, `cmd/api/main.go`, y el walk resolvió exacto. El diseño
de §5 no cambia. Queda por evaluar `apphttp.Handlers` cuando se añada el campo del handler, pero
es el mismo struct que ya crece con cada feature sin incidentes.

