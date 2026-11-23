package auth

import (
	"github.com/example/microservicio-plantilla-reporte/internal/config"
	"testing"
)

func TestResolveMapsTokenWithoutReturningToken(t *testing.T) {
	r := NewRegistry([]config.Identity{{Token: "secret", UserID: "user-1", Profile: "editor"}})
	u, ok := r.Resolve("secret")
	if !ok || u.ID != "user-1" || u.Perfil != "editor" {
		t.Fatalf("resolve = %#v, %v", u, ok)
	}
	if _, ok = r.Resolve("wrong"); ok {
		t.Fatal("unexpected token match")
	}
}
