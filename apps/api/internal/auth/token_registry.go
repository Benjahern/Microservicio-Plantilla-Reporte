package auth

import (
	"crypto/subtle"
	"github.com/example/microservicio-plantilla-reporte/internal/config"
	"github.com/example/microservicio-plantilla-reporte/internal/usuarios"
)

type Registry struct{ identities []entry }
type entry struct {
	token string
	user  usuarios.UsuarioDTO
}

func NewRegistry(ids []config.Identity) *Registry {
	r := &Registry{}
	for _, id := range ids {
		r.identities = append(r.identities, entry{id.Token, usuarios.UsuarioDTO{ID: id.UserID, Perfil: id.Profile}})
	}
	return r
}

func (r *Registry) Resolve(token string) (usuarios.UsuarioDTO, bool) {
	for _, e := range r.identities {
		if len(token) == len(e.token) && subtle.ConstantTimeCompare([]byte(token), []byte(e.token)) == 1 {
			return e.user, true
		}
	}
	return usuarios.UsuarioDTO{}, false
}
