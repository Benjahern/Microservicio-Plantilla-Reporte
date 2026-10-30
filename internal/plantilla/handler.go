package plantilla

import (
	"encoding/json"
	"errors"
	"github.com/example/microservicio-plantilla-reporte/internal/auth"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct{ service *Service }

func NewHandler(s *Service) *Handler { return &Handler{s} }
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	u := auth.User(r.Context())
	p := strings.Trim(u.Perfil, " ")
	all := p == "administrador"
	if r.URL.Path == "/api/v1/plantillas" {
		switch r.Method {
		case http.MethodPost:
			if p == "lector" {
				writeError(w, 403, "forbidden", "Permiso insuficiente")
				return
			}
			h.create(w, r, u.ID)
		case http.MethodGet:
			page, limit, ok := pagination(w, r)
			if !ok {
				return
			}
			d, total, e := h.service.List(r.Context(), u.ID, all, page, limit)
			if e != nil {
				fail(w, e)
				return
			}
			writeJSON(w, 200, ListResponse{Data: d, Pagination: Pagination{Page: page, Limit: limit, Total: total}})
		default:
			method(w)
		}
		return
	}
	const prefix = "/api/v1/plantillas/"
	if !strings.HasPrefix(r.URL.Path, prefix) {
		writeError(w, 404, "not_found", "Ruta no encontrada")
		return
	}
	parts := strings.Split(strings.Trim(strings.TrimPrefix(r.URL.Path, prefix), "/"), "/")
	id := parts[0]
	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			d, e := h.service.Get(r.Context(), u.ID, all, id)
			if e != nil {
				fail(w, e)
				return
			}
			writeJSON(w, 200, d)
		case http.MethodPut:
			if p == "lector" {
				writeError(w, 403, "forbidden", "Permiso insuficiente")
				return
			}
			var in UpdateRequest
			if !decode(w, r, &in) {
				return
			}
			d, e := h.service.Update(r.Context(), u.ID, all, id, in)
			if e != nil {
				fail(w, e)
				return
			}
			writeJSON(w, 200, d)
		case http.MethodDelete:
			if p == "lector" {
				writeError(w, 403, "forbidden", "Permiso insuficiente")
				return
			}
			if e := h.service.Delete(r.Context(), u.ID, all, id); e != nil {
				fail(w, e)
				return
			}
			w.WriteHeader(204)
		default:
			method(w)
		}
		return
	}
	if len(parts) == 2 && parts[1] == "versiones" && r.Method == http.MethodGet {
		page, limit, ok := pagination(w, r)
		if !ok {
			return
		}
		d, total, e := h.service.Versions(r.Context(), u.ID, all, id, page, limit)
		if e != nil {
			fail(w, e)
			return
		}
		writeJSON(w, 200, map[string]any{"data": d, "pagination": Pagination{Page: page, Limit: limit, Total: total}})
		return
	}
	if len(parts) == 3 && parts[1] == "versiones" && r.Method == http.MethodGet {
		n, e := strconv.Atoi(parts[2])
		if e != nil || n < 1 {
			writeError(w, 400, "bad_request", "Número de versión inválido")
			return
		}
		d, e := h.service.Version(r.Context(), u.ID, all, id, n)
		if e != nil {
			fail(w, e)
			return
		}
		writeJSON(w, 200, d)
		return
	}
	writeError(w, 404, "not_found", "Ruta no encontrada")
}
func (h *Handler) create(w http.ResponseWriter, r *http.Request, user string) {
	var in CreateRequest
	if !decode(w, r, &in) {
		return
	}
	d, e := h.service.Create(r.Context(), user, in)
	if e != nil {
		fail(w, e)
		return
	}
	writeJSON(w, 201, d)
}
func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if e := d.Decode(v); e != nil {
		writeError(w, 400, "bad_request", "JSON inválido: "+e.Error())
		return false
	}
	var extra any
	if e := d.Decode(&extra); e != io.EOF {
		writeError(w, 400, "bad_request", "Se permite un único objeto JSON")
		return false
	}
	return true
}
func pagination(w http.ResponseWriter, r *http.Request) (int, int, bool) {
	p, l := 1, 20
	var e error
	if v := r.URL.Query().Get("page"); v != "" {
		p, e = strconv.Atoi(v)
	}
	if e != nil || p < 1 {
		writeError(w, 400, "bad_request", "page inválido")
		return 0, 0, false
	}
	if v := r.URL.Query().Get("limit"); v != "" {
		l, e = strconv.Atoi(v)
	}
	if e != nil || l < 1 || l > 100 {
		writeError(w, 400, "bad_request", "limit debe estar entre 1 y 100")
		return 0, 0, false
	}
	return p, l, true
}
func fail(w http.ResponseWriter, e error) {
	var max *http.MaxBytesError
	switch {
	case errors.As(e, &max):
		writeError(w, 413, "payload_too_large", "Cuerpo excede 1 MiB")
	case errors.Is(e, ErrNotFound):
		writeError(w, 404, "not_found", "Plantilla no encontrada")
	case errors.Is(e, ErrConflict):
		writeError(w, 409, "conflict", "La versión cambió; consulta la versión actual")
	case strings.Contains(e.Error(), "inválido") || strings.Contains(e.Error(), "obligatorio") || strings.Contains(e.Error(), "caracteres") || strings.Contains(e.Error(), "contenido"):
		writeError(w, 400, "bad_request", e.Error())
	default:
		writeError(w, 500, "internal_error", "Error interno")
	}
}
func method(w http.ResponseWriter) {
	writeError(w, 405, "method_not_allowed", "Método no permitido")
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}
