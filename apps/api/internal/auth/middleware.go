package auth

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/example/microservicio-plantilla-reporte/internal/api"
	"github.com/example/microservicio-plantilla-reporte/internal/usuarios"
)

type contextKey struct{}

func User(ctx context.Context) usuarios.UsuarioDTO {
	u, _ := ctx.Value(contextKey{}).(usuarios.UsuarioDTO)
	return u
}

func Scope(ctx context.Context) (string, bool) {
	u := User(ctx)
	return u.ID, u.Perfil == "administrador"
}

type Authenticator interface {
	Authenticate(context.Context, string) (usuarios.UsuarioDTO, error)
}

func Middleware(authenticator Authenticator, roles []string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scheme, token, ok := strings.Cut(r.Header.Get("Authorization"), " ")
		if r.Header.Get("Authorization") == "" {
			cookie, err := r.Cookie("sesion")
			if err == nil {
				if !sameOrigin(r) {
					api.Error(w, 403, "forbidden", "Origen inválido para la sesión")
					return
				}
				scheme, token, ok = "Bearer", cookie.Value, true
			}
		}
		if !ok || !strings.EqualFold(scheme, "Bearer") || token == "" {
			api.Error(w, 401, "unauthorized", "Token inválido o ausente")
			return
		}
		user, err := authenticator.Authenticate(r.Context(), token)
		if errors.Is(err, ErrIdentityUnavailable) {
			api.Error(w, 503, "identity_unavailable", "Proveedor de identidad no disponible")
			return
		}
		if err != nil {
			api.Error(w, 401, "unauthorized", "Token inválido o ausente")
			return
		}
		if !slices.Contains(roles, user.Perfil) {
			api.Error(w, 403, "forbidden", "Permiso insuficiente")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), contextKey{}, user)))
	})
}

func sameOrigin(r *http.Request) bool {
	if r.Method == http.MethodGet || r.Method == http.MethodHead {
		return true
	}
	if r.Header.Get("Sec-Fetch-Site") == "cross-site" {
		return false
	}
	if origin := r.Header.Get("Origin"); origin != "" {
		parsed, err := url.Parse(origin)
		return err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && strings.EqualFold(parsed.Host, r.Host)
	}
	return true
}
