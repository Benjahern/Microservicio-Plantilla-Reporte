package httpapi

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/example/microservicio-plantilla-reporte/internal/api"
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	if w.status != 0 {
		return
	}
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(p)
}

func Observability(next http.Handler, log *slog.Logger, version string) http.Handler {
	log = log.With("servicio", "plantillas-reporte", "version", version)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		id := r.Header.Get(api.RequestIDHeader)
		if !api.UUIDPattern.MatchString(id) {
			id = api.NewUUID()
		}
		r = r.WithContext(api.WithRequestID(r.Context(), id))
		w.Header().Set(api.RequestIDHeader, id)
		out := &statusWriter{ResponseWriter: w}
		defer func() {
			if recover() != nil && out.status == 0 {
				api.Error(out, 500, "internal_error", "Error interno")
			}
			status := out.status
			if status == 0 {
				status = http.StatusOK
			}
			route := r.Pattern
			if _, path, ok := strings.Cut(route, " "); ok {
				route = path
			}
			if route == "" {
				route = "(sin ruta)"
			}
			method := r.Method
			switch method {
			case "GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS", "PATCH":
			default:
				method = "(otro)"
			}
			log.Info("petición HTTP", "request_id", id, "metodo", method, "ruta", route,
				"estado", status, "duracion_ms", time.Since(start).Milliseconds())
		}()
		next.ServeHTTP(out, r)
	})
}
