package auth

import (
	"context"
	"github.com/example/microservicio-plantilla-reporte/internal/usuarios"
	"net/http"
	"strings"
)

type contextKey struct{}

func User(ctx context.Context) usuarios.UsuarioDTO {
	u, _ := ctx.Value(contextKey{}).(usuarios.UsuarioDTO)
	return u
}

func Middleware(registry *Registry, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scheme, token, ok := strings.Cut(r.Header.Get("Authorization"), " ")
		user, valid := registry.Resolve(token)
		if !ok || !strings.EqualFold(scheme, "Bearer") || !valid {
			http.Error(w, `{"error":{"code":"unauthorized","message":"Token inválido o ausente"}}`, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), contextKey{}, user)))
	})
}
