-- +goose Up
CREATE TABLE plantillas (
    id uuid PRIMARY KEY,
    propietario_id uuid NOT NULL,
    version_actual integer NOT NULL CHECK (version_actual > 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);
CREATE INDEX plantillas_propietario_activas
    ON plantillas (propietario_id, created_at DESC, id DESC) WHERE deleted_at IS NULL;
CREATE INDEX plantillas_activas
    ON plantillas (created_at DESC, id DESC) WHERE deleted_at IS NULL;

CREATE TABLE plantilla_versiones (
    plantilla_id uuid NOT NULL REFERENCES plantillas (id),
    numero integer NOT NULL CHECK (numero > 0),
    nombre varchar(150) NOT NULL CHECK (length(nombre) > 0),
    descripcion varchar(1000) NOT NULL,
    contenido jsonb NOT NULL CHECK (jsonb_typeof(contenido) = 'object' AND contenido <> '{}'::jsonb),
    activo boolean NOT NULL,
    autor_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (plantilla_id, numero)
);

CREATE TABLE auditoria_plantillas (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    plantilla_id uuid NOT NULL REFERENCES plantillas (id),
    actor_id uuid NOT NULL,
    accion text NOT NULL CHECK (accion IN ('crear', 'actualizar', 'eliminar')),
    version integer NOT NULL,
    ocurrido_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX auditoria_plantillas_objeto ON auditoria_plantillas (plantilla_id, id);

-- +goose Down
DROP TABLE auditoria_plantillas;
DROP TABLE plantilla_versiones;
DROP TABLE plantillas;
