package plantilla

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	q "github.com/example/microservicio-plantilla-reporte/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (r *PostgresRepository) write(ctx context.Context, actor pgtype.UUID, action, key string, intent any,
	mutate func(*q.Queries) (Response, error),
) (out Response, err error) {
	err = pgx.BeginTxFunc(ctx, r.pool, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}, func(tx pgx.Tx) error {
		queries := r.query.WithTx(tx)
		var err error
		if key == "" {
			out, err = mutate(queries)
			return err
		}
		out, err = idempotentWrite(ctx, queries, actor, action, key, intent, mutate)
		return err
	})
	return out, err
}

func idempotentWrite(ctx context.Context, queries *q.Queries, actor pgtype.UUID, action, key string, intent any,
	mutate func(*q.Queries) (Response, error),
) (Response, error) {
	payload, err := json.Marshal(intent)
	if err != nil {
		return Response{}, fmt.Errorf("%w: petición inválida", ErrValidation)
	}
	keyHash, requestHash := sha256.Sum256([]byte(key)), sha256.Sum256(payload)
	reserved, err := queries.ReserveIdempotency(ctx, q.ReserveIdempotencyParams{
		ActorID: actor, Operacion: action, ClaveHash: keyHash[:], SolicitudHash: requestHash[:],
	})
	if err != nil {
		return Response{}, err
	}
	if reserved == 0 {
		prior, err := queries.GetIdempotency(ctx, q.GetIdempotencyParams{
			ActorID: actor, Operacion: action, ClaveHash: keyHash[:],
		})
		if err != nil {
			return Response{}, err
		}
		if !bytes.Equal(prior.SolicitudHash, requestHash[:]) {
			return Response{}, ErrIdempotencyConflict
		}
		var out Response
		decoder := json.NewDecoder(bytes.NewReader(prior.Respuesta))
		decoder.UseNumber()
		err = decoder.Decode(&out)
		return out, err
	}
	out, err := mutate(queries)
	if err != nil {
		return Response{}, err
	}
	encoded, err := json.Marshal(out)
	if err != nil {
		return Response{}, err
	}
	err = queries.CompleteIdempotency(ctx, q.CompleteIdempotencyParams{
		ActorID: actor, Operacion: action, ClaveHash: keyHash[:], Respuesta: encoded,
	})
	return out, err
}
