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
- Hito 1 usa un repositorio simulado que se reinicia con el proceso. En la interfaz o en Swagger usa `demo-editor`, `demo-lector` o `demo-admin`. Son identidades ficticias para desarrollo.

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

Consulta [arquitectura y propiedad de datos](docs/arquitectura.md), [modos de falla](docs/modos-de-falla.md) y [entrega de hitos](docs/entrega.md). El [pipeline](../../actions/workflows/ci.yml) usa la dirección relativa del repositorio y conserva el enlace después del traspaso. El enlace definitivo de SonarQube queda pendiente de la URL del proyecto del curso.
