package plantilla

import (
	"context"
	"fmt"
	"strings"
)

type Service struct{ repo Repository }

func NewService(r Repository) *Service { return &Service{r} }
func validate(name string, content map[string]any) error {
	if name != strings.TrimSpace(name) || name == "" || len(name) > 150 {
		return fmt.Errorf("nombre es obligatorio y debe tener hasta 150 caracteres")
	}
	if len(content) == 0 {
		return fmt.Errorf("contenido debe ser un objeto JSON no vacío")
	}
	return nil
}
func (s *Service) Create(ctx context.Context, user string, in CreateRequest) (Response, error) {
	if e := validate(in.Nombre, in.Contenido); e != nil {
		return Response{}, e
	}
	if len(in.Descripcion) > 1000 {
		return Response{}, fmt.Errorf("descripcion supera 1000 caracteres")
	}
	return s.repo.Create(ctx, user, in)
}
func (s *Service) List(ctx context.Context, user string, all bool, p, l int) ([]Response, int64, error) {
	return s.repo.List(ctx, user, all, p, l)
}
func (s *Service) Get(ctx context.Context, user string, all bool, id string) (Response, error) {
	return s.repo.Get(ctx, user, all, id)
}
func (s *Service) Update(ctx context.Context, user string, all bool, id string, in UpdateRequest) (Response, error) {
	if e := validate(in.Nombre, in.Contenido); e != nil {
		return Response{}, e
	}
	if in.Activo == nil {
		return Response{}, fmt.Errorf("activo es obligatorio")
	}
	if in.VersionEsperada < 1 {
		return Response{}, fmt.Errorf("version_esperada debe ser positiva")
	}
	if len(in.Descripcion) > 1000 {
		return Response{}, fmt.Errorf("descripcion supera 1000 caracteres")
	}
	return s.repo.Update(ctx, user, all, id, in)
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
