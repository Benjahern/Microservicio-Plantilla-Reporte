package plantilla

import (
	"context"
	"errors"

	"github.com/example/microservicio-plantilla-reporte/internal/api"
)

type CreateRequest = api.CreateRequest
type UpdateRequest = api.UpdateRequest
type Response = api.Response
type VersionResponse = api.VersionResponse

var ErrNotFound = errors.New("plantilla no encontrada")
var ErrConflict = errors.New("versión esperada no coincide")
var ErrValidation = errors.New("entrada inválida")
var ErrIdempotencyConflict = errors.New("clave de idempotencia asociada a otra petición")

type Repository interface {
	Create(context.Context, string, CreateRequest, string) (Response, error)
	List(context.Context, string, bool, int, int) ([]Response, int64, error)
	Get(context.Context, string, bool, string) (Response, error)
	Update(context.Context, string, bool, string, UpdateRequest, string) (Response, error)
	Delete(context.Context, string, bool, string) error
	Versions(context.Context, string, bool, string, int, int) ([]VersionResponse, int64, error)
	Version(context.Context, string, bool, string, int) (VersionResponse, error)
}
