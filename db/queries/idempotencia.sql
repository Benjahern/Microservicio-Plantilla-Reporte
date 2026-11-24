-- name: ReserveIdempotency :execrows
INSERT INTO idempotencia_plantillas (actor_id, operacion, clave_hash, solicitud_hash)
VALUES ($1, $2, $3, $4)
ON CONFLICT (actor_id, operacion, clave_hash) DO NOTHING;

-- name: GetIdempotency :one
SELECT solicitud_hash, respuesta FROM idempotencia_plantillas
WHERE actor_id = $1 AND operacion = $2 AND clave_hash = $3;

-- name: CompleteIdempotency :exec
UPDATE idempotencia_plantillas SET respuesta = $4
WHERE actor_id = $1 AND operacion = $2 AND clave_hash = $3;
