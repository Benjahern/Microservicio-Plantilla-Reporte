package httpapi

import (
	"github.com/example/microservicio-plantilla-reporte/internal/auth"
	"github.com/example/microservicio-plantilla-reporte/internal/plantilla"
	"net/http"
	"time"
)

func NewRouter(reg *auth.Registry, h *plantilla.Handler, ping func() error) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/live", func(w http.ResponseWriter, r *http.Request) { JSON(w, 200, map[string]string{"status": "ok"}) })
	mux.HandleFunc("GET /health/ready", func(w http.ResponseWriter, r *http.Request) {
		if err := ping(); err != nil {
			Error(w, 503, "not_ready", "Base de datos no disponible")
			return
		}
		JSON(w, 200, map[string]string{"status": "ready"})
	})
	mux.Handle("/api/v1/plantillas", auth.Middleware(reg, h))
	mux.Handle("/api/v1/plantillas/", auth.Middleware(reg, h))
	return logging(mux)
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) { w.status = code; w.ResponseWriter.WriteHeader(code) }
func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: 200}
		defer func() {
			if recover() != nil {
				Error(sw, 500, "internal_error", "Error interno")
			}
		}()
		next.ServeHTTP(sw, r)
		_ = start
	})
}
