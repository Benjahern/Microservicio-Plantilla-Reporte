# Plantillas de reporte

Microservicio Go y frontend Vue para crear, consultar, versionar y dar de baja plantillas de reporte.

## Arranque local

```sh
docker compose up -d --build
```

Los puertos provisionales son API **5100** y frontend **5101**. Para asignar `51N0` y `51N1`, copia `.env.example` a `.env` y ajusta ambos puertos. PostgreSQL queda en la red interna de Compose. La configuración sirve para desarrollo local.

- Frontend en http://localhost:5101.
- API interactiva en http://localhost:5100/api/docs con `API_ENV=dev`.
- Comprobaciones en `GET /healthz` y `GET /readyz`.
- Hito 2 usa PostgreSQL. Compose aplica las migraciones en `migrate` antes de iniciar el backend.

El proveedor de identidad emite el JWT. Para abrir una sesión web, envía ese JWT como Bearer a `POST /api/v1/sesiones` por el mismo origen HTTPS. La API emite la cookie `sesion` de host con `Path=/api`, `HttpOnly`, `Secure` y `SameSite=Lax`. Las integraciones y Swagger también pueden usar Bearer directamente. Configura `AUTH_JWKS_URL`, `AUTH_ISSUER`, `AUTH_AUDIENCE` y `AUTH_ROLE_CLAIM` con el proveedor real. El valor de desarrollo apunta a un JWKS local esperado en el puerto 5102. En producción usa `API_ENV=prod`, HTTPS y credenciales de PostgreSQL inyectadas por el entorno.

Para reintentar una creación o edición, envía el mismo `Idempotency-Key` y el mismo cuerpo en ambas peticiones. PostgreSQL conserva la primera respuesta exitosa en la transacción de negocio. Una clave del mismo actor y operación con otra petición produce 409. La cabecera es opcional para mantener compatibilidad con el Hito 1.

La rama local `entrega/hito-1` conserva la versión simulada para la primera entrega.

## Desarrollo y evidencia

```sh
pnpm install --frozen-lockfile
pnpm codegen
git diff --exit-code
cd apps/api
go build ./...
go test ./...
```

El [contrato](contracts/openapi.yaml) declara permisos por operación. El cliente TypeScript y la tabla de rutas Go se generan desde ese archivo.

Regenera el repositorio SQL con `sqlc generate` (v1.29.0) desde la raíz. Usa el contenedor `migrate` para preparar una base exclusiva de pruebas. Con `TEST_DATABASE_URL` apuntando a esa base, ejecuta `bash scripts/e2e.sh` en Linux. Para una sola entrada de la matriz, usa `bash scripts/e2e.sh normal Pacific/Auckland` o `bash scripts/e2e.sh normal UTC`. El informe del binario se escribe en `coverage/e2e.cover`.

Consulta [arquitectura y propiedad de datos](docs/arquitectura.md), [modos de falla](docs/modos-de-falla.md) y [entrega de hitos](docs/entrega.md). El [pipeline](../../actions/workflows/ci.yml) usa un enlace relativo. El enlace definitivo de SonarQube queda pendiente de la URL del proyecto del curso.
