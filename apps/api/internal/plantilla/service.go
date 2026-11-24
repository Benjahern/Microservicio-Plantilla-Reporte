package plantilla

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/example/microservicio-plantilla-reporte/internal/api"
)

type Service struct{ repo Repository }

func NewService(r Repository) *Service { return &Service{r} }
func validate(name string, content map[string]any) error {
	if name != strings.TrimSpace(name) || name == "" || utf8.RuneCountInString(name) > 150 {
		return fmt.Errorf("%w: nombre debe tener de 1 a 150 caracteres sin espacios en los extremos", ErrValidation)
	}
	if len(content) == 0 {
		return fmt.Errorf("%w: contenido debe ser un objeto JSON no vacío", ErrValidation)
	}
	return nil
}
func (s *Service) Create(ctx context.Context, user string, in CreateRequest, key string) (Response, error) {
	if e := validate(in.Nombre, in.Contenido); e != nil {
		return Response{}, e
	}
	if utf8.RuneCountInString(in.Descripcion) > 1000 {
		return Response{}, fmt.Errorf("%w: descripcion supera 1000 caracteres", ErrValidation)
	}
	return s.repo.Create(ctx, user, in, key)
}
func (s *Service) List(ctx context.Context, user string, all bool, p, l int) ([]Response, int64, error) {
	return s.repo.List(ctx, user, all, p, l)
}
func (s *Service) Get(ctx context.Context, user string, all bool, id string) (Response, error) {
	return s.repo.Get(ctx, user, all, id)
}
func (s *Service) Update(ctx context.Context, user string, all bool, id string, in UpdateRequest, key string) (Response, error) {
	if e := validate(in.Nombre, in.Contenido); e != nil {
		return Response{}, e
	}
	if in.Activo == nil {
		return Response{}, fmt.Errorf("%w: activo es obligatorio", ErrValidation)
	}
	if in.VersionEsperada < 1 || in.VersionEsperada > api.MaxInt {
		return Response{}, fmt.Errorf("%w: version_esperada fuera de rango", ErrValidation)
	}
	if utf8.RuneCountInString(in.Descripcion) > 1000 {
		return Response{}, fmt.Errorf("%w: descripcion supera 1000 caracteres", ErrValidation)
	}
	return s.repo.Update(ctx, user, all, id, in, key)
}
func (s *Service) Delete(ctx context.Context, user string, all bool, id string) error {
	return s.repo.Delete(ctx, user, all, id)
}
func (s *Service) Versions(ctx context.Context, user string, all bool, id string, p, l int) ([]VersionResponse, int64, error) {
	return s.repo.Versions(ctx, user, all, id, p, l)
}
func (s *Service) Version(ctx context.Context, user string, all bool, id string, n int) (VersionResponse, error) {
	return s.repo.Version(ctx, user, all, id, n)
}
