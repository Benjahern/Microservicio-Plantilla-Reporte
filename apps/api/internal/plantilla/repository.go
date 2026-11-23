package plantilla

import "context"

type Repository interface {
	Create(context.Context, string, CreateRequest) (Response, error)
	List(context.Context, string, bool, int, int) ([]Response, int64, error)
	Get(context.Context, string, bool, string) (Response, error)
	Update(context.Context, string, bool, string, UpdateRequest) (Response, error)
	Delete(context.Context, string, bool, string) error
	Versions(context.Context, string, bool, string, int, int) ([]VersionResponse, int64, error)
	Version(context.Context, string, bool, string, int) (VersionResponse, error)
}
