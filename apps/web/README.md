# Frontend – Plantillas de reporte

Vue 3 (Composition API y `<script setup>`) y Vite.

## Desarrollo

```bash
cd apps/web
npm ci
npm run dev        # http://localhost:5173
```

Vite dirige `/api` a `http://localhost:5100`. Para otra dirección, configura `BACKEND_URL` en el entorno de Vite.

## Con Docker

Desde la raíz:

```bash
docker compose up -d --build
```

El frontend queda en `http://localhost:5101`. Nginx sirve la aplicación y dirige `/api` al backend.

## Uso

La aplicación usa la cookie de sesión emitida por la API en el mismo origen HTTPS. El integrador envía primero el JWT del proveedor como Bearer a `POST /api/v1/sesiones`. El perfil `lector` puede consultar sus recursos. Las escrituras intentadas por ese perfil muestran el 403 del backend.

## Estructura

```
src/
  shared/http-client.api.ts  Fetch, sesión y errores
  types/contract.gen.ts     Tipos y rutas generados desde OpenAPI
  components/              Listado, editor e historial
  dev-app/                  Aplicación de desarrollo
```
