-- name: InsertPlantilla :one
INSERT INTO plantillas (id, propietario_id, version_actual)
VALUES (sqlc.arg(id), sqlc.arg(propietario_id), 1)
RETURNING *;

-- name: InsertVersion :one
INSERT INTO plantilla_versiones (plantilla_id, numero, nombre, descripcion, contenido, activo, autor_id)
VALUES (sqlc.arg(plantilla_id), sqlc.arg(numero), sqlc.arg(nombre), sqlc.arg(descripcion),
        sqlc.arg(contenido), sqlc.arg(activo), sqlc.arg(autor_id))
RETURNING *;

-- name: InsertAudit :exec
INSERT INTO auditoria_plantillas (plantilla_id, actor_id, accion, version)
VALUES (sqlc.arg(plantilla_id), sqlc.arg(actor_id), sqlc.arg(accion), sqlc.arg(version));

-- name: GetPlantilla :one
SELECT sqlc.embed(p), sqlc.embed(v)
FROM plantillas p JOIN plantilla_versiones v ON v.plantilla_id = p.id AND v.numero = p.version_actual
WHERE p.id = sqlc.arg(id) AND p.deleted_at IS NULL
  AND (sqlc.arg(ver_todas)::boolean OR p.propietario_id = sqlc.arg(actor_id));

-- name: ListPlantillas :many
SELECT sqlc.embed(p), sqlc.embed(v)
FROM plantillas p JOIN plantilla_versiones v ON v.plantilla_id = p.id AND v.numero = p.version_actual
WHERE p.deleted_at IS NULL AND (sqlc.arg(ver_todas)::boolean OR p.propietario_id = sqlc.arg(actor_id))
ORDER BY p.created_at DESC, p.id DESC
LIMIT sqlc.arg(limite)::integer OFFSET sqlc.arg(salto)::bigint;

-- name: CountPlantillas :one
SELECT count(*) FROM plantillas p
WHERE p.deleted_at IS NULL AND (sqlc.arg(ver_todas)::boolean OR p.propietario_id = sqlc.arg(actor_id));

-- name: AdvanceVersion :one
UPDATE plantillas p SET version_actual = p.version_actual + 1, updated_at = now()
WHERE p.id = sqlc.arg(id) AND p.deleted_at IS NULL
  AND (sqlc.arg(ver_todas)::boolean OR p.propietario_id = sqlc.arg(actor_id))
  AND p.version_actual < 2147483647
  AND p.version_actual = sqlc.arg(version_esperada)
RETURNING p.*;

-- name: SoftDelete :one
UPDATE plantillas p SET deleted_at = now(), updated_at = now()
WHERE p.id = sqlc.arg(id) AND p.deleted_at IS NULL
  AND (sqlc.arg(ver_todas)::boolean OR p.propietario_id = sqlc.arg(actor_id))
RETURNING p.*;

-- name: CountVersions :one
SELECT count(*) FROM plantilla_versiones WHERE plantilla_id = sqlc.arg(plantilla_id);

-- name: ListVersions :many
SELECT * FROM plantilla_versiones WHERE plantilla_id = sqlc.arg(plantilla_id)
ORDER BY numero DESC LIMIT sqlc.arg(limite)::integer OFFSET sqlc.arg(salto)::bigint;

-- name: GetVersion :one
SELECT v.* FROM plantilla_versiones v JOIN plantillas p ON p.id = v.plantilla_id
WHERE v.plantilla_id = sqlc.arg(id) AND v.numero = sqlc.arg(numero) AND p.deleted_at IS NULL
  AND (sqlc.arg(ver_todas)::boolean OR p.propietario_id = sqlc.arg(actor_id));
