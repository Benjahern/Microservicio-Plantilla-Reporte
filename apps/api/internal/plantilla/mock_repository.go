package plantilla

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/example/microservicio-plantilla-reporte/internal/api"
)

type mockTemplate struct {
	current  Response
	versions []VersionResponse
}

// MockRepository keeps simulated data in memory and resets on restart.
type MockRepository struct {
	mu   sync.Mutex
	data map[string]*mockTemplate
}

func NewMockRepository() *MockRepository {
	r := &MockRepository{data: make(map[string]*mockTemplate)}
	r.create("11111111-1111-4111-8111-111111111111", "11111111-1111-4111-8111-111111111111",
		CreateRequest{Nombre: "Reporte de ejemplo", Contenido: map[string]any{"titulo": "Muestra"}},
		time.Date(2026, 11, 23, 12, 0, 0, 0, time.UTC))
	return r
}

func (r *MockRepository) create(id, owner string, in CreateRequest, now time.Time) Response {
	v := VersionResponse{Numero: 1, Nombre: in.Nombre, Descripcion: in.Descripcion, Contenido: in.Contenido,
		Activo: in.Activo == nil || *in.Activo, AutorID: owner, CreatedAt: now}
	out := Response{ID: id, PropietarioID: owner, VersionActual: 1, VersionResponse: v, UpdatedAt: now}
	r.data[id] = &mockTemplate{current: out, versions: []VersionResponse{v}}
	return out
}

func (r *MockRepository) Create(_ context.Context, owner string, in CreateRequest) (Response, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.create(api.NewUUID(), owner, in, time.Now().UTC()), nil
}

func (r *MockRepository) accessible(owner string, all bool, id string) (*mockTemplate, error) {
	d, ok := r.data[id]
	if !ok || (!all && d.current.PropietarioID != owner) {
		return nil, ErrNotFound
	}
	return d, nil
}

func (r *MockRepository) Get(_ context.Context, owner string, all bool, id string) (Response, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	d, err := r.accessible(owner, all, id)
	if err != nil {
		return Response{}, err
	}
	return d.current, nil
}

func slicePage[T any](items []T, page, limit int) ([]T, int64) {
	start := min((page-1)*limit, len(items))
	return items[start:min(start+limit, len(items))], int64(len(items))
}

func (r *MockRepository) List(_ context.Context, owner string, all bool, page, limit int) ([]Response, int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	items := make([]Response, 0, len(r.data))
	for _, d := range r.data {
		if all || d.current.PropietarioID == owner {
			items = append(items, d.current)
		}
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].UpdatedAt.Equal(items[j].UpdatedAt) {
			return items[i].ID > items[j].ID
		}
		return items[i].UpdatedAt.After(items[j].UpdatedAt)
	})
	selected, total := slicePage(items, page, limit)
	return selected, total, nil
}

func (r *MockRepository) Update(_ context.Context, owner string, all bool, id string, in UpdateRequest) (Response, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	d, err := r.accessible(owner, all, id)
	if err != nil {
		return Response{}, err
	}
	if d.current.VersionActual != in.VersionEsperada || in.VersionEsperada == api.MaxInt {
		return Response{}, ErrConflict
	}
	now := time.Now().UTC()
	v := VersionResponse{Numero: in.VersionEsperada + 1, Nombre: in.Nombre, Descripcion: in.Descripcion,
		Contenido: in.Contenido, Activo: *in.Activo, AutorID: owner, CreatedAt: now}
	d.versions = append(d.versions, v)
	d.current.VersionActual, d.current.VersionResponse, d.current.UpdatedAt = v.Numero, v, now
	return d.current, nil
}

func (r *MockRepository) Delete(_ context.Context, owner string, all bool, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, err := r.accessible(owner, all, id); err != nil {
		return err
	}
	delete(r.data, id)
	return nil
}

func (r *MockRepository) Versions(_ context.Context, owner string, all bool, id string, page, limit int) ([]VersionResponse, int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	d, err := r.accessible(owner, all, id)
	if err != nil {
		return nil, 0, err
	}
	items := make([]VersionResponse, len(d.versions))
	for i, v := range d.versions {
		items[len(items)-1-i] = v
	}
	selected, total := slicePage(items, page, limit)
	return selected, total, nil
}

func (r *MockRepository) Version(_ context.Context, owner string, all bool, id string, n int) (VersionResponse, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	d, err := r.accessible(owner, all, id)
	if err != nil || n < 1 || n > len(d.versions) {
		return VersionResponse{}, ErrNotFound
	}
	return d.versions[n-1], nil
}
