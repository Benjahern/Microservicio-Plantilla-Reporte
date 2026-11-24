# Modos de falla

| Falla | Protección | Verificación |
| --- | --- | --- |
| Cliente desalineado | Cliente y rutas generados del mismo contrato | `pnpm codegen` y `git diff --exit-code` |
| Contrato incompatible | oasdiff contra main | Job calidad |
| Entrada inválida | DTO, UUID, paginación, tipo de contenido y máximo 1 MiB | Enviar entrada inválida y constatar 400, 413 o 415 |
| Rol o propietario incorrecto | Roles del contrato y filtro de propietario | Enviar una escritura de lector o una consulta de otro usuario |
| Dos ediciones tienen la misma versión esperada | PostgreSQL comprueba la versión al actualizar la fila en la misma transacción | E2E concurrente con resultado 200 y 409. La mutación aislada debe dar 200 y 200 y fallar |
| Una escritura queda sin su evento de auditoría | Escritura, versión y evento comparten transacción | E2E confirma tres eventos con UUID, acción y fecha |
| Reinicio de la API | Las versiones y la auditoría residen en PostgreSQL | E2E reinicia el binario y consulta historial y versión actual |
| PostgreSQL no disponible | Compose espera al healthcheck al arrancar y /readyz hace ping con timeout | Con PostgreSQL caído, /readyz responde 503 y /healthz sigue en 200 |
| Fallo de migración | Goose se ejecuta como contenedor previo y Compose exige código 0 | Inspeccionar el estado y los logs de migrate |
| JWT expirado o emitido para otra audiencia | Firma RS256 y claims validados | E2E verifica 401 en ambos casos |
| JWKS inaccesible sin clave válida en caché | Timeout de 5 s y respuesta 503 sin detalles internos | Arrancar sin proveedor y enviar JWT con kid válido |
| Escritura de navegador desde origen ajeno | Cookie SameSite=Lax y comprobación de origen para métodos de escritura | Prueba de middleware con cookie y origen cruzado |
| Fecha interpretada en zona incorrecta | PostgreSQL usa timestamptz y las respuestas se normalizan a UTC | Ejecutar el E2E con TZ=UTC y TZ=Pacific/Auckland |
| Exposición de credenciales en logs | Registro de campos fijos y patrones de ruta sin cuerpo, query ni credenciales | Prueba de middleware e inspección de stdout JSON |
| Swagger expuesto fuera de dev | Registro de ruta condicionado por API_ENV | Probar /api/docs con API_ENV=prod |
