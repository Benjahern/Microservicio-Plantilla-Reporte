# Microservicio de plantillas de reporte

API REST en Go y MongoDB para crear, consultar, versionar y eliminar lógicamente plantillas.

## Inicio

1. Copia `deploy/.env.example` a `deploy/.env` y configura tokens únicos por consumidor.
2. Ejecuta desde la raíz: `docker compose --env-file deploy/.env -f deploy/compose.yaml up --build`.

MongoDB se inicia como replica set para soportar transacciones. API en `http://localhost:8080`.

## Autenticación

Usa `Authorization: Bearer <token>`. `AUTH_IDENTITIES_JSON` configura un arreglo de `{ "token", "usuario_id", "perfil" }`; perfiles admitidos: `lector`, `editor`, `administrador`. Lector consulta; editor crea/modifica sus recursos; administrador administra todos. Tokens compartidos equivalen a compartir identidad. En producción usar un gestor de secretos y rotar las credenciales.

## API

- `POST /api/v1/plantillas` crear (versión 1).
- `GET /api/v1/plantillas?page=1&limit=20` listar las propias; administrador lista todas.
- `GET /api/v1/plantillas/{id}` consultar versión actual.
- `PUT /api/v1/plantillas/{id}` crear nueva versión. Requiere `version_esperada` para control optimista.
- `DELETE /api/v1/plantillas/{id}` baja lógica.
- `GET /api/v1/plantillas/{id}/versiones?page=1&limit=20` historial.
- `GET /api/v1/plantillas/{id}/versiones/{numero}` consultar instantánea.
- `GET /health/live`, `GET /health/ready`.

Requests de creación/actualización usan `nombre`, `descripcion`, `contenido` (objeto JSON) y `activo`. IDs de propietario y autor se derivan del token. Actualizar no altera versiones previas. Errores usan `{ "error": { "code", "message" } }`.

## Desarrollo

Requiere Go 1.23+. `go test ./...` y `go vet ./...`. La integración local requiere `deploy/compose.yaml`.
