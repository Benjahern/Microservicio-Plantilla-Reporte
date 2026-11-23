# Frontend – Plantillas de reporte

Vue 3 (Composition API, `<script setup>`) + Vue Router + Vite. Sin librerías de UI.

## Desarrollo

```bash
cd apps/web
npm install
npm run dev        # http://localhost:5173
```

Vite redirige `/api` a `http://localhost:8080` (la API no tiene CORS). Para otra dirección:
`VITE_API_TARGET=http://host:puerto npm run dev`.

## Con Docker

Desde `apps/api`:

```bash
docker compose --env-file deploy/.env -f deploy/compose.yaml up --build
```

Frontend en `http://localhost:8081` (nginx sirve la SPA y hace proxy de `/api` hacia el servicio `api`).

## Uso

Se ingresa con uno de los tokens definidos en `AUTH_IDENTITIES_JSON` (`deploy/.env`).
El perfil `lector` solo consulta; si intenta crear/editar/eliminar, la API responde 403 y se muestra el mensaje.

## Estructura

```
src/
  api/client.js              fetch + token Bearer + manejo de errores
  composables/useAuth.js     token en sessionStorage
  composables/usePlantillas.js  listado, detalle, versiones, crear/actualizar/eliminar
  views/                     Login, Listado, Detalle (con historial), Formulario
  components/                AppHeader, AlertMessage, PaginationBar
```
