package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/example/microservicio-plantilla-reporte/internal/api"
	"github.com/example/microservicio-plantilla-reporte/internal/auth"
	"github.com/example/microservicio-plantilla-reporte/internal/plantilla"
	"github.com/example/microservicio-plantilla-reporte/internal/usuarios"
)

const owner = "11111111-1111-4111-8111-111111111111"
const other = "22222222-2222-4222-8222-222222222222"
const base = "/api/v1/plantillas"

type identities map[string]usuarios.UsuarioDTO

func (i identities) Authenticate(_ context.Context, token string) (usuarios.UsuarioDTO, error) {
	user, ok := i[token]
	if !ok {
		return usuarios.UsuarioDTO{}, auth.ErrInvalidToken
	}
	return user, nil
}

func router(dev bool) http.Handler {
	return NewRouter(identities{
		"lector": {ID: owner, Perfil: "lector"},
		"editor": {ID: owner, Perfil: "editor"},
		"otro":   {ID: other, Perfil: "editor"},
		"admin":  {ID: other, Perfil: "administrador"},
	}, plantilla.NewHandler(plantilla.NewService(NewMockRepository())), func(context.Context) error { return nil }, dev)
}

func call(t *testing.T, h http.Handler, method, path, token, body string, want int) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != want {
		t.Fatalf("%s %s: status = %d, esperado %d: %s", method, path, rec.Code, want, rec.Body)
	}
	return rec
}

func TestMatrizAcceso(t *testing.T) {
	const create = `{"nombre":"Reporte","contenido":{"titulo":"Muestra"}}`
	const update = `{"nombre":"Reporte nuevo","contenido":{"titulo":"Cambio"},"activo":true,"version_esperada":1}`
	target := base + "/" + owner
	cases := []struct {
		method, path, body string
		want               [5]int
	}{
		{"GET", base, "", [5]int{401, 200, 200, 200, 200}},
		{"POST", base, create, [5]int{401, 403, 201, 201, 201}},
		{"GET", target, "", [5]int{401, 200, 200, 404, 200}},
		{"PUT", target, update, [5]int{401, 403, 200, 404, 200}},
		{"DELETE", target, "", [5]int{401, 403, 204, 404, 204}},
		{"GET", target + "/versiones", "", [5]int{401, 200, 200, 404, 200}},
		{"GET", target + "/versiones/1", "", [5]int{401, 200, 200, 404, 200}},
	}
	for _, c := range cases {
		for idx, token := range []string{"", "lector", "editor", "otro", "admin"} {
			t.Run(c.method+c.path+"/"+token, func(t *testing.T) {
				rec := call(t, router(false), c.method, c.path, token, c.body, c.want[idx])
				if c.want[idx] >= 400 && rec.Header().Get("Content-Type") != "application/problem+json" {
					t.Fatal(rec.Header())
				}
			})
		}
	}
}

func TestContratoHTTP(t *testing.T) {
	h := router(true)
	call(t, h, "GET", "/healthz", "", "", 200)
	call(t, h, "GET", "/readyz", "", "", 200)
	call(t, h, "GET", "/health/live", "", "", 404)
	call(t, h, "GET", "/api/docs", "", "", 200)
	doc := call(t, h, "GET", "/api/docs/openapi.json", "", "", 200)
	if !json.Valid(doc.Body.Bytes()) || doc.Body.String() != api.OpenAPI {
		t.Fatal("Swagger no sirve el contrato generado")
	}
	call(t, router(false), "GET", "/api/docs", "", "", 404)
	call(t, h, "GET", base+"/no-es-un-uuid", "editor", "", 400)
	call(t, h, "GET", base+"?limit=101", "editor", "", 400)
	call(t, h, "POST", base, "editor", `{"nombre":"A","contenido":{}}`, 400)
	created := call(t, h, "POST", base, "editor", `{"nombre":"A","contenido":{"titulo":"A"}}`, 201)
	var out api.Response
	if err := json.Unmarshal(created.Body.Bytes(), &out); err != nil || !api.UUIDPattern.MatchString(out.ID) || out.VersionActual != 1 || !out.Activo {
		t.Fatalf("respuesta inválida: %s (%v)", created.Body, err)
	}
}
