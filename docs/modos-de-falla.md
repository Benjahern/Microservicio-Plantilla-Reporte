# Modos de falla

| Falla | Protección | Verificación |
| --- | --- | --- |
| Cliente desalineado | Cliente y rutas generados del mismo contrato | `pnpm codegen` y `git diff --exit-code` |
| Contrato incompatible | oasdiff contra main | Job calidad |
| Entrada inválida | DTO, UUID, paginación, tipo de contenido y máximo 1 MiB | Enviar entrada inválida y constatar 400, 413 o 415 |
| Rol o propietario incorrecto | Roles del contrato y filtro de propietario | Enviar una escritura de lector o una consulta de otro usuario |
| Edición con versión obsoleta | Comparación y actualización bajo el mismo bloqueo del mock | Enviar dos PUT con igual versión esperada |
| Reinicio del mock | Estado simulado recupera datos iniciales | Reiniciar después de una creación |
| PostgreSQL no disponible al arrancar | Compose espera el healthcheck antes de migrate | `docker compose ps -a` |
| Swagger expuesto fuera de dev | Registro de ruta condicionado por API_ENV | Probar /api/docs con API_ENV=prod |
