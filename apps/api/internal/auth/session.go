package auth

import (
	"net/http"
	"strings"

	"github.com/example/microservicio-plantilla-reporte/internal/api"
)

// IssueSession runs behind JWT authentication. A session exchange requires an
// explicit Bearer token even when the browser already holds a session cookie.
func IssueSession(w http.ResponseWriter, r *http.Request) {
	scheme, token, ok := strings.Cut(r.Header.Get("Authorization"), " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") || token == "" {
		api.Error(w, 401, "unauthorized", "Se requiere un token Bearer")
		return
	}
	if !sameOrigin(r) {
		api.Error(w, 403, "forbidden", "Origen inválido para la sesión")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	http.SetCookie(w, &http.Cookie{
		Name: "sesion", Value: token, Path: "/api",
		HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(http.StatusNoContent)
}
