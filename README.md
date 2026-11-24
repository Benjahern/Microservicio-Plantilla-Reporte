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

El navegador usa la cookie `sesion` emitida por el proveedor de identidad con `HttpOnly`, `Secure`, `SameSite=Lax` y alcance compatible con el proxy web. La API también acepta Bearer para integraciones y Swagger. Configura `AUTH_JWKS_URL`, `AUTH_ISSUER`, `AUTH_AUDIENCE` y `AUTH_ROLE_CLAIM` con los valores del proveedor. El valor de desarrollo apunta a un JWKS local esperado en el puerto 5102. Sin ese proveedor y sus JWT, sólo están disponibles las rutas públicas. En producción usa `API_ENV=prod`, HTTPS y credenciales de PostgreSQL inyectadas por el entorno.

La rama local `codex/hito-1` conserva la versión simulada para la primera entrega.

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

Regenera el repositorio SQL con `sqlc generate` (v1.29.0) desde la raíz. Para ejecutar e2e en Linux, aplica `goose -dir db/migrations postgres "$TEST_DATABASE_URL" up` a una base exclusiva de pruebas y ejecuta `bash scripts/e2e.sh`. El informe de cobertura del binario se escribe en `coverage/e2e.cover`.

Consulta [arquitectura y propiedad de datos](docs/arquitectura.md), [modos de falla](docs/modos-de-falla.md) y [entrega de hitos](docs/entrega.md). El [pipeline](../../actions/workflows/ci.yml) usa un enlace relativo. El enlace definitivo de SonarQube queda pendiente de la URL del proyecto del curso.
