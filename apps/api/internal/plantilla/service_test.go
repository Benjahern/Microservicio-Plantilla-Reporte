package plantilla

import (
	"context"
	"testing"
)

type fakeRepository struct{ created string }

func (f *fakeRepository) Create(_ context.Context, user string, _ CreateRequest) (Response, error) {
	f.created = user
	return Response{}, nil
}
func (*fakeRepository) List(context.Context, string, bool, int, int) ([]Response, int64, error) {
	return nil, 0, nil
}
func (*fakeRepository) Get(context.Context, string, bool, string) (Response, error) {
	return Response{}, nil
}
func (*fakeRepository) Update(context.Context, string, bool, string, UpdateRequest) (Response, error) {
	return Response{}, nil
}
func (*fakeRepository) Delete(context.Context, string, bool, string) error { return nil }
func (*fakeRepository) Versions(context.Context, string, bool, string, int, int) ([]VersionResponse, int64, error) {
	return nil, 0, nil
}
func (*fakeRepository) Version(context.Context, string, bool, string, int) (VersionResponse, error) {
	return VersionResponse{}, nil
}

func TestCreateValidatesAndUsesIdentity(t *testing.T) {
	repo := &fakeRepository{}
	s := NewService(repo)
	_, err := s.Create(context.Background(), "user-42", CreateRequest{Nombre: "Reporte", Contenido: map[string]any{"x": 1}})
	if err != nil {
		t.Fatal(err)
	}
	if repo.created != "user-42" {
		t.Fatalf("owner = %q", repo.created)
	}
}
func TestCreateRejectsInvalidName(t *testing.T) {
	s := NewService(&fakeRepository{})
	_, err := s.Create(context.Background(), "u", CreateRequest{Nombre: "  ", Contenido: map[string]any{"x": 1}})
	if err == nil {
		t.Fatal("expected validation error")
	}
}
