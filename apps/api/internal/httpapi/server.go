package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/example/microservicio-plantilla-reporte/internal/api"
	"github.com/example/microservicio-plantilla-reporte/internal/auth"
	"github.com/example/microservicio-plantilla-reporte/internal/plantilla"
)

func NewRouter(a auth.Authenticator, h *plantilla.Handler, ping func(context.Context) error, dev bool) http.Handler {
	mux := http.NewServeMux()
	handlers := h.Operations()
	handlers["crearSesion"] = auth.IssueSession
	handlers["salud"] = func(w http.ResponseWriter, r *http.Request) {
		api.JSON(w, 200, api.Estado{Status: "ok"})
	}
	handlers["disponibilidad"] = func(w http.ResponseWriter, r *http.Request) {
		if err := ping(r.Context()); err != nil {
			api.Error(w, 503, "not_ready", "Base de datos no disponible")
			return
		}
		api.JSON(w, 200, api.Estado{Status: "ready"})
	}
	if len(handlers) != len(api.Operations) {
		panic("los handlers y el contrato no coinciden")
	}
	methods := make(map[string][]string)
	for name, op := range api.Operations {
		handler, ok := handlers[name]
		if !ok {
			panic("operación sin handler: " + name)
		}
		var endpoint http.Handler = validatePath(handler)
		if op.Permission != "publico" {
			endpoint = auth.Middleware(a, op.Roles, endpoint)
		}
		mux.Handle(op.Pattern, endpoint)
		method, path, _ := strings.Cut(op.Pattern, " ")
		methods[path] = append(methods[path], method)
	}
	for path, allowed := range methods {
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Allow", strings.Join(allowed, ", "))
			api.Error(w, 405, "method_not_allowed", "Método no permitido")
		})
	}
	if dev {
		mux.HandleFunc("GET /api/docs", serveDocs)
		mux.HandleFunc("GET /api/docs/openapi.json", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, api.OpenAPI)
		})
	}
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		api.Error(w, 404, "not_found", "Ruta no encontrada")
	})
	return mux
}

func validatePath(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if id := r.PathValue("id"); id != "" && !api.UUIDPattern.MatchString(id) {
			api.Error(w, 400, "bad_request", "id debe ser un UUID")
			return
		}
		next(w, r)
	})
}

func serveDocs(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprint(w, `<!doctype html><html lang="es"><head><meta charset="utf-8"><title>Plantillas de reporte</title>
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.29.0/swagger-ui.css"></head>
<body><div id="swagger-ui"></div><script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.29.0/swagger-ui-bundle.js"></script>
<script>SwaggerUIBundle({url:"/api/docs/openapi.json",dom_id:"#swagger-ui",validatorUrl:null});</script></body></html>`)
}
