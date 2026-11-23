package usuarios

// UsuarioDTO representa la identidad configurada para un consumidor del servicio.
type UsuarioDTO struct {
	ID     string `json:"usuario_id"`
	Perfil string `json:"perfil"`
}
