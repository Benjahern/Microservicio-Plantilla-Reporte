package auth

import (
	"context"
	"github.com/example/microservicio-plantilla-reporte/internal/api"
	"github.com/example/microservicio-plantilla-reporte/internal/usuarios"
	"net/http"
	"slices"
	"strings"
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
	Authenticate(context.Context, string) (usuarios.UsuarioDTO, bool)
}

func Middleware(authenticator Authenticator, roles []string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scheme, token, ok := strings.Cut(r.Header.Get("Authorization"), " ")
		if r.Header.Get("Authorization") == "" {
			cookie, err := r.Cookie("sesion")
			if err == nil {
				scheme, token, ok = "Bearer", cookie.Value, true
			}
		}
		if !ok || !strings.EqualFold(scheme, "Bearer") || token == "" {
			api.Error(w, 401, "unauthorized", "Token inválido o ausente")
			return
		}
		user, valid := authenticator.Authenticate(r.Context(), token)
		if !valid {
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
