package plantilla

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/example/microservicio-plantilla-reporte/internal/api"
	q "github.com/example/microservicio-plantilla-reporte/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool  *pgxpool.Pool
	query *q.Queries
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool, query: q.New(pool)}
}

var _ Repository = (*PostgresRepository)(nil)

var snapshot = pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly}

func (r *PostgresRepository) Create(ctx context.Context, owner string, in CreateRequest) (out Response, err error) {
	actor, id := uuid(owner), uuid(api.NewUUID())
	err = pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		queries := r.query.WithTx(tx)
		p, err := queries.InsertPlantilla(ctx, q.InsertPlantillaParams{ID: id, PropietarioID: actor})
		if err != nil {
			return err
		}
		out, err = writeVersion(ctx, queries, p, actor, in, "crear")
		return err
	})
	return out, err
}

func (r *PostgresRepository) Get(ctx context.Context, owner string, all bool, id string) (Response, error) {
	row, err := r.query.GetPlantilla(ctx, q.GetPlantillaParams{ID: uuid(id), ActorID: uuid(owner), VerTodas: all})
	if err != nil {
		return Response{}, missing(err)
	}
	return templateResponse(row.Plantillas, row.PlantillaVersiones)
}

func (r *PostgresRepository) List(ctx context.Context, owner string, all bool, page, limit int) (out []Response, total int64, err error) {
	actor := uuid(owner)
	err = pgx.BeginTxFunc(ctx, r.pool, snapshot, func(tx pgx.Tx) error {
		queries := r.query.WithTx(tx)
		rows, err := queries.ListPlantillas(ctx, q.ListPlantillasParams{ActorID: actor, VerTodas: all, Limite: int32(limit), Salto: offset(page, limit)})
		if err != nil {
			return err
		}
		out = make([]Response, 0, len(rows))
		for _, row := range rows {
			item, err := templateResponse(row.Plantillas, row.PlantillaVersiones)
			if err != nil {
				return err
			}
			out = append(out, item)
		}
		total, err = queries.CountPlantillas(ctx, q.CountPlantillasParams{ActorID: actor, VerTodas: all})
		return err
	})
	return out, total, err
}

func (r *PostgresRepository) Update(ctx context.Context, owner string, all bool, id string, in UpdateRequest) (out Response, err error) {
	actor, target := uuid(owner), uuid(id)
	err = pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		queries := r.query.WithTx(tx)
		p, err := queries.AdvanceVersion(ctx, q.AdvanceVersionParams{
			ID: target, ActorID: actor, VerTodas: all, VersionEsperada: int32(in.VersionEsperada),
		})
		if errors.Is(err, pgx.ErrNoRows) {
			_, err = queries.GetPlantilla(ctx, q.GetPlantillaParams{ID: target, ActorID: actor, VerTodas: all})
			if err != nil {
				return missing(err)
			}
			return ErrConflict
		}
		if err != nil {
			return err
		}
		out, err = writeVersion(ctx, queries, p, actor, in.CreateRequest, "actualizar")
		return err
	})
	return out, err
}

func (r *PostgresRepository) Delete(ctx context.Context, owner string, all bool, id string) error {
	actor, target := uuid(owner), uuid(id)
	return pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		queries := r.query.WithTx(tx)
		p, err := queries.SoftDelete(ctx, q.SoftDeleteParams{ID: target, ActorID: actor, VerTodas: all})
		if err != nil {
			return missing(err)
		}
		return audit(ctx, queries, p, actor, "eliminar")
	})
}

func (r *PostgresRepository) Versions(ctx context.Context, owner string, all bool, id string, page, limit int) (out []VersionResponse, total int64, err error) {
	target, actor := uuid(id), uuid(owner)
	err = pgx.BeginTxFunc(ctx, r.pool, snapshot, func(tx pgx.Tx) error {
		queries := r.query.WithTx(tx)
		if _, err := queries.GetPlantilla(ctx, q.GetPlantillaParams{ID: target, ActorID: actor, VerTodas: all}); err != nil {
			return missing(err)
		}
		rows, err := queries.ListVersions(ctx, q.ListVersionsParams{PlantillaID: target, Limite: int32(limit), Salto: offset(page, limit)})
		if err != nil {
			return err
		}
		out = make([]VersionResponse, 0, len(rows))
		for _, row := range rows {
			item, err := versionResponse(row)
			if err != nil {
				return err
			}
			out = append(out, item)
		}
		total, err = queries.CountVersions(ctx, target)
		return err
	})
	return out, total, err
}

func (r *PostgresRepository) Version(ctx context.Context, owner string, all bool, id string, n int) (VersionResponse, error) {
	v, err := r.query.GetVersion(ctx, q.GetVersionParams{ID: uuid(id), ActorID: uuid(owner), VerTodas: all, Numero: int32(n)})
	if err != nil {
		return VersionResponse{}, missing(err)
	}
	return versionResponse(v)
}

func writeVersion(ctx context.Context, queries *q.Queries, p q.Plantillas, actor pgtype.UUID, in CreateRequest, action string) (Response, error) {
	content, err := json.Marshal(in.Contenido)
	if err != nil {
		return Response{}, fmt.Errorf("%w: contenido inválido", ErrValidation)
	}
	v, err := queries.InsertVersion(ctx, q.InsertVersionParams{
		PlantillaID: p.ID, Numero: p.VersionActual, Nombre: in.Nombre, Descripcion: in.Descripcion,
		Contenido: content, Activo: in.Activo == nil || *in.Activo, AutorID: actor,
	})
	if err != nil {
		var invalid *pgconn.PgError
		if errors.As(err, &invalid) && (invalid.Code == "22P05" || invalid.Code == "22003") {
			err = fmt.Errorf("%w: texto o número fuera del rango admitido", ErrValidation)
		}
		return Response{}, err
	}
	if err := audit(ctx, queries, p, actor, action); err != nil {
		return Response{}, err
	}
	return templateResponse(p, v)
}

func audit(ctx context.Context, queries *q.Queries, p q.Plantillas, actor pgtype.UUID, action string) error {
	return queries.InsertAudit(ctx, q.InsertAuditParams{PlantillaID: p.ID, ActorID: actor, Accion: action, Version: p.VersionActual})
}

func templateResponse(p q.Plantillas, v q.PlantillaVersiones) (Response, error) {
	version, err := versionResponse(v)
	if err != nil {
		return Response{}, err
	}
	version.CreatedAt = p.CreatedAt.Time.UTC()
	return Response{ID: p.ID.String(), PropietarioID: p.PropietarioID.String(), VersionActual: int(p.VersionActual),
		VersionResponse: version, UpdatedAt: p.UpdatedAt.Time.UTC()}, nil
}

func versionResponse(v q.PlantillaVersiones) (VersionResponse, error) {
	var content map[string]any
	d := json.NewDecoder(bytes.NewReader(v.Contenido))
	d.UseNumber()
	if err := d.Decode(&content); err != nil {
		return VersionResponse{}, err
	}
	return VersionResponse{Numero: int(v.Numero), Nombre: v.Nombre, Descripcion: v.Descripcion, Contenido: content,
		Activo: v.Activo, AutorID: v.AutorID.String(), CreatedAt: v.CreatedAt.Time.UTC()}, nil
}

func offset(page, limit int) int64 { return int64(page-1) * int64(limit) }

func uuid(id string) pgtype.UUID {
	// Both the authenticated identity and the HTTP path have been validated.
	var value pgtype.UUID
	_ = value.Scan(id)
	return value
}

func missing(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}
