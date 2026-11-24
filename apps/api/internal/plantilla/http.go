package plantilla

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"github.com/example/microservicio-plantilla-reporte/internal/api"
	"github.com/example/microservicio-plantilla-reporte/internal/auth"
)

type Handler struct{ service *Service }

func NewHandler(s *Service) *Handler { return &Handler{s} }

func (h *Handler) Operations() map[string]http.HandlerFunc {
	return map[string]http.HandlerFunc{
		"crearPlantilla": h.create, "listarPlantillas": h.list,
		"obtenerPlantilla": h.get, "actualizarPlantilla": h.update, "eliminarPlantilla": h.delete,
		"listarVersiones": h.versions, "obtenerVersion": h.version,
	}
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var in CreateRequest
	if !decode(w, r, &in) {
		return
	}
	out, err := h.service.Create(r.Context(), auth.User(r.Context()).ID, in)
	respond(w, 201, out, err)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	page, limit, ok := pagination(w, r)
	if !ok {
		return
	}
	owner, all := auth.Scope(r.Context())
	items, total, err := h.service.List(r.Context(), owner, all, page, limit)
	respond(w, 200, api.ListResponse[Response]{Data: items, Pagination: api.Pagination{Page: page, Limit: limit, Total: total}}, err)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	owner, all := auth.Scope(r.Context())
	out, err := h.service.Get(r.Context(), owner, all, id(r))
	respond(w, 200, out, err)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	var in UpdateRequest
	if !decode(w, r, &in) {
		return
	}
	owner, all := auth.Scope(r.Context())
	out, err := h.service.Update(r.Context(), owner, all, id(r), in)
	respond(w, 200, out, err)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	owner, all := auth.Scope(r.Context())
	if err := h.service.Delete(r.Context(), owner, all, id(r)); err != nil {
		respond(w, 204, nil, err)
		return
	}
	w.WriteHeader(204)
}

func (h *Handler) versions(w http.ResponseWriter, r *http.Request) {
	page, limit, ok := pagination(w, r)
	if !ok {
		return
	}
	owner, all := auth.Scope(r.Context())
	items, total, err := h.service.Versions(r.Context(), owner, all, id(r), page, limit)
	respond(w, 200, api.ListResponse[VersionResponse]{Data: items, Pagination: api.Pagination{Page: page, Limit: limit, Total: total}}, err)
}

func (h *Handler) version(w http.ResponseWriter, r *http.Request) {
	n, err := positiveInt(r.PathValue("numero"), api.MaxInt)
	if err != nil {
		api.Error(w, 400, "bad_request", "Número de versión inválido")
		return
	}
	owner, all := auth.Scope(r.Context())
	out, err := h.service.Version(r.Context(), owner, all, id(r), n)
	respond(w, 200, out, err)
}

func id(r *http.Request) string { return strings.ToLower(r.PathValue("id")) }

func decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	media, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if media != "application/json" {
		api.Error(w, 415, "unsupported_media_type", "Se requiere application/json")
		return false
	}
	data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, api.MaxBody))
	var limit *http.MaxBytesError
	if errors.As(err, &limit) {
		api.Error(w, 413, "payload_too_large", "Cuerpo excede 1 MiB")
		return false
	}
	var fields map[string]json.RawMessage
	if err != nil || json.Unmarshal(data, &fields) != nil || fields == nil {
		api.Error(w, 400, "bad_request", "Se requiere un objeto JSON válido")
		return false
	}
	_, isUpdate := dst.(*UpdateRequest)
	for key, value := range fields {
		switch key {
		case "nombre", "descripcion", "contenido", "activo":
		case "version_esperada":
			if isUpdate && !bytes.Equal(value, []byte("null")) {
				continue
			}
			fallthrough
		default:
			api.Error(w, 400, "bad_request", "El cuerpo contiene campos inválidos")
			return false
		}
		if bytes.Equal(value, []byte("null")) {
			api.Error(w, 400, "bad_request", "Los campos no admiten valores nulos")
			return false
		}
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	if d.Decode(dst) != nil {
		api.Error(w, 400, "bad_request", "Los campos no coinciden con el contrato")
		return false
	}
	return true
}

func positiveInt(value string, max int) (int, error) {
	n, err := strconv.Atoi(value)
	if err != nil || n < 1 || n > max {
		return 0, ErrValidation
	}
	return n, nil
}

func pagination(w http.ResponseWriter, r *http.Request) (int, int, bool) {
	page, limit := 1, 20
	values, err := urlQuery(r)
	if err != nil {
		api.Error(w, 400, "bad_request", "Paginación inválida")
		return 0, 0, false
	}
	if values["page"] != 0 {
		page = values["page"]
	}
	if values["limit"] != 0 {
		limit = values["limit"]
	}
	return page, limit, true
}

func urlQuery(r *http.Request) (map[string]int, error) {
	out := make(map[string]int, 2)
	for key, max := range map[string]int{"page": api.MaxInt, "limit": 100} {
		if v, found := r.URL.Query()[key]; found {
			if len(v) != 1 {
				return nil, ErrValidation
			}
			n, err := positiveInt(v[0], max)
			if err != nil {
				return nil, err
			}
			out[key] = n
		}
	}
	return out, nil
}

func respond(w http.ResponseWriter, status int, v any, err error) {
	switch {
	case err == nil:
		api.JSON(w, status, v)
	case errors.Is(err, ErrNotFound):
		api.Error(w, 404, "not_found", "Plantilla o versión no encontrada")
	case errors.Is(err, ErrConflict):
		api.Error(w, 409, "conflict", "La versión cambió; consulta la versión actual")
	case errors.Is(err, ErrValidation):
		api.Error(w, 400, "bad_request", err.Error())
	default:
		api.Error(w, 500, "internal_error", "Error interno")
	}
}
