// Mirror of the API contract (internal/plantilla/dto.go).
// Field names are defined by the backend: do not rename them here.
import type { Content } from "./template"

export interface VersionWire {
  numero: number
  nombre: string
  descripcion: string
  contenido: Content
  activo: boolean
  autor_id: string
  created_at: string
}

export interface TemplateWire extends VersionWire {
  id: string
  propietario_id: string
  version_actual: number
  updated_at: string
}

export interface CreateTemplateWire {
  nombre: string
  descripcion: string
  contenido: Content
  activo: boolean
}

export interface UpdateTemplateWire extends CreateTemplateWire {
  version_esperada: number
}
