# Despliegue en un VPS

Todo el stack vive en `docker-compose.yml`: Postgres, migraciones, API, panel y
Caddy. La app Flutter no se despliega aquí — se distribuye aparte (pendiente).

## Antes de tocar el servidor

**El DNS va primero.** Caddy pide el certificado a Let's Encrypt en el arranque,
y la validación consulta el DNS. Si los dominios todavía no apuntan al VPS,
Caddy arranca, falla la validación y reintenta con backoff: el sitio queda sin
HTTPS y el log no siempre lo dice claro. Apunta ambos registros A y espera a que
propaguen antes de levantar nada.

```
api.tu-dominio.com     A   <ip del vps>
panel.tu-dominio.com   A   <ip del vps>
```

## Primera puesta en marcha

```bash
git clone <repo> && cd Myvibesfit
cp .env.example .env
```

Edita `.env`. Lo que no puede quedar como está:

| Variable | Por qué |
|---|---|
| `JWT_ACCESS_SECRET` | Genera uno real: `openssl rand -base64 48`. La API **se niega a arrancar** con el de desarrollo o con menos de 32 caracteres cuando `ENV=production` |
| `POSTGRES_PASSWORD` | El default `myvibesfit` es público, está en este repo |
| `API_DOMAIN`, `PANEL_DOMAIN`, `PUBLIC_API_URL` | Sin default a propósito: compose falla si faltan |
| `CORS_ALLOWED_ORIGINS` | En producción **no hay default permisivo**. Si no declaras el dominio del panel, el navegador bloqueará sus peticiones |

Después:

```bash
docker compose up -d --build
docker compose logs -f caddy   # confirma que emitió los certificados
curl https://api.tu-dominio.com/health
```

El orden lo resuelve compose solo: `db` sano → `migrate` aplica las migraciones
y sale → `api` arranca. Es imposible servir tráfico contra un esquema viejo,
porque `api` espera a que `migrate` termine con éxito.

## Actualizar

```bash
git pull
docker compose up -d --build
```

Las migraciones nuevas se aplican solas en el mismo paso. Si una falla, `migrate`
sale distinto de cero y `api` **no** arranca — el servicio se queda con la
versión anterior en vez de subir con el esquema a medias.

## El worker

No es un servicio que viva: es una pasada única que el cron del host invoca
(decisión de [SDD-003](sdd/active/SDD-003-worker-operation.md)). Por eso está
tras el profile `cli` y `docker compose up` no lo arranca.

```bash
# a mano
docker compose --profile cli run --rm worker

# en /etc/cron.d/myvibesfit — flock evita que dos pases se solapen
30 3 * * * root cd /opt/Myvibesfit && flock -n /tmp/mvf.lock docker compose --profile cli run --rm worker
```

Necesita `ANTHROPIC_API_KEY` en el `.env`. Sin ella corre y falla al llamar a
Claude; el resto del backend no la necesita.

## Lo que está cerrado a propósito

- **Postgres** solo escucha en `127.0.0.1:5433`. Desde fuera del VPS no se
  alcanza; para conectarte usa un túnel SSH.
- **La API no se publica** al host: solo Caddy la alcanza, por la red interna.
- **Adminer** está tras el profile `debug`. Es un administrador de base de datos
  sin autenticación delante: levantarlo por defecto sería regalar la base.
  `docker compose --profile debug up -d adminer`, y bájalo al terminar.

## Verificación tras desplegar

```bash
curl -sI https://api.tu-dominio.com/health | head -1    # 200 y certificado válido
curl -s  https://api.tu-dominio.com/health              # {"status":"ok"}
docker compose ps                                        # api y web en running, migrate en exited(0)
docker compose logs api | grep -i "push notifications"   # dice si el push está activo o deshabilitado
```

## Todavía no resuelto

- **Backups.** Nada respalda Postgres hoy. Es lo siguiente (Fase 3), y para un
  gimnasio real pesa más que cualquier feature: perder el historial de
  entrenamientos no se arregla.
- **Push real.** El backend está listo, pero sin proyecto Firebase el envío
  queda en no-op. La API lo dice en el log al arrancar.
- **Distribución de la app.** Cómo llegan los teléfonos a tener la app Flutter
  sigue sin decidirse.
