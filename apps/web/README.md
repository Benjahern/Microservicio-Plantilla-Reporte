# Frontend – Plantillas de reporte

Vue 3 (Composition API y `<script setup>`) y Vite.

## Desarrollo

```bash
cd apps/web
npm ci
npm run dev        # http://localhost:5173
```

Vite dirige `/api` a `http://localhost:5100`. Configura `BACKEND_URL` para un puerto distinto.

## Con Docker

Desde la raíz:

```bash
docker compose up -d --build
```

El frontend queda en `http://localhost:5101`. Nginx sirve la aplicación y dirige `/api` al backend.

## Uso

Usa `demo-lector`, `demo-editor` o `demo-admin` para trabajar con el repositorio simulado de Hito 1.

## Estructura

```
src/
  shared/http-client.api.ts  Fetch, identidad simulada y errores
  types/contract.gen.ts     Tipos y rutas generados desde OpenAPI
  components/              Listado, editor e historial
  dev-app/                  Aplicación de desarrollo
```
