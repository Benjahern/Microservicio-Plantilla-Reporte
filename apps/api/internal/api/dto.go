package api

import (
	"crypto/rand"
	"fmt"
	"regexp"
	"time"
)

const MaxBody = 1 << 20
const MaxInt = 2147483647

var UUIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func NewUUID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = b[6]&0x0f | 0x40
	b[8] = b[8]&0x3f | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[:4], b[4:6], b[6:8], b[8:10], b[10:])
}

type CreateRequest struct {
	Nombre      string         `json:"nombre" validate:"required,max=150" example:"Reporte mensual"`
	Descripcion string         `json:"descripcion" validate:"max=1000" example:"Resumen de operaciones"`
	Contenido   map[string]any `json:"contenido" validate:"required,min=1"`
	Activo      *bool          `json:"activo" example:"true"`
}
type UpdateRequest struct {
	CreateRequest
	VersionEsperada int `json:"version_esperada" validate:"min=1,max=2147483647" example:"1"`
}
type VersionResponse struct {
	Numero      int            `json:"numero" example:"1"`
	Nombre      string         `json:"nombre" example:"Reporte mensual"`
	Descripcion string         `json:"descripcion" example:"Resumen de operaciones"`
	Contenido   map[string]any `json:"contenido"`
	Activo      bool           `json:"activo" example:"true"`
	AutorID     string         `json:"autor_id" example:"11111111-1111-4111-8111-111111111111"`
	CreatedAt   time.Time      `json:"created_at" example:"2026-11-23T12:00:00Z"`
}
type Response struct {
	ID            string `json:"id" example:"11111111-1111-4111-8111-111111111111"`
	PropietarioID string `json:"propietario_id" example:"11111111-1111-4111-8111-111111111111"`
	VersionActual int    `json:"version_actual" example:"1"`
	VersionResponse
	UpdatedAt time.Time `json:"updated_at" example:"2026-11-23T12:00:00Z"`
}
type ListResponse[T any] struct {
	Data       []T        `json:"data"`
	Pagination Pagination `json:"pagination"`
}
type Pagination struct {
	Page  int   `json:"page" example:"1"`
	Limit int   `json:"limit" example:"20"`
	Total int64 `json:"total" example:"1"`
}
type Estado struct {
	Status string `json:"status" example:"ok"`
}
type Problema struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Code   string `json:"code"`
	Detail string `json:"detail"`
}
