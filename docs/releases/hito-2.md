# Hito 2 · Refinamiento y calidad

La API persiste plantillas y auditoría en PostgreSQL, admite reintentos idempotentes de creaciones y actualizaciones, emite sesiones con JWT validado contra JWKS y protege las ediciones concurrentes.

- SHA de main asociado a `hito-2`.
- Run verde y Quality Gate del mismo SHA.
- El borrador creado por CI adjunta `openapi.yaml`, `sbom.cdx.json` y `oasdiff-hito-1.txt`.
- Enlazar `docs/modos-de-falla.md`, `e2e/testdata/concurrencia.golden.json` y el run rojo de la mutación.
