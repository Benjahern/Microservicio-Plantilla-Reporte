-- +goose Up
CREATE TABLE idempotencia_plantillas (
    actor_id uuid NOT NULL,
    operacion text NOT NULL CHECK (operacion IN ('crear', 'actualizar')),
    clave_hash bytea NOT NULL CHECK (octet_length(clave_hash) = 32),
    solicitud_hash bytea NOT NULL CHECK (octet_length(solicitud_hash) = 32),
    respuesta bytea,
    PRIMARY KEY (actor_id, operacion, clave_hash)
);

-- +goose Down
DROP TABLE idempotencia_plantillas;
