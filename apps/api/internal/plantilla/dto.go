package plantilla

import "time"

type CreateRequest struct {
	Nombre      string         `json:"nombre"`
	Descripcion string         `json:"descripcion"`
	Contenido   map[string]any `json:"contenido"`
	Activo      *bool          `json:"activo"`
}
type UpdateRequest struct {
	Nombre          string         `json:"nombre"`
	Descripcion     string         `json:"descripcion"`
	Contenido       map[string]any `json:"contenido"`
	Activo          *bool          `json:"activo"`
	VersionEsperada int            `json:"version_esperada"`
}
type VersionResponse struct {
	Numero      int            `json:"numero"`
	Nombre      string         `json:"nombre"`
	Descripcion string         `json:"descripcion"`
	Contenido   map[string]any `json:"contenido"`
	Activo      bool           `json:"activo"`
	AutorID     string         `json:"autor_id"`
	CreatedAt   time.Time      `json:"created_at"`
}
type Response struct {
	ID            string `json:"id"`
	PropietarioID string `json:"propietario_id"`
	VersionActual int    `json:"version_actual"`
	VersionResponse
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
type ListResponse struct {
	Data       []Response `json:"data"`
	Pagination Pagination `json:"pagination"`
}
type Pagination struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}
