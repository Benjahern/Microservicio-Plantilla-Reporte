# Hito 2 · Refinamiento y calidad

La API persiste plantillas y auditoría en PostgreSQL, autentica contra JWKS y valida ediciones concurrentes.

- SHA de main asociado a `hito-2`.
- Run verde y Quality Gate del mismo SHA.
- Adjuntar `openapi.yaml` de ese tag y `sbom.cdx.json` del run.
- Enlazar el resultado de oasdiff contra `hito-1`.
- Enlazar `docs/modos-de-falla.md`, `e2e/testdata/concurrencia.golden.json` y el run rojo de la mutación.
