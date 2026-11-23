// Domain types used by the UI (English, camelCase).
// The backend contract lives in ./template.wire.ts and is translated by ./template.mappers.ts.

// Values defined by the backend (internal/plantilla/dto.go).
export type Profile = "lector" | "editor" | "administrador"

export type Content = Record<string, unknown>

export interface Version {
  number: number
  name: string
  description: string
  content: Content
  active: boolean
  authorId: string
  createdAt: string
}

export interface Template extends Version {
  id: string
  ownerId: string
  currentVersion: number
  updatedAt: string
}

export interface Pagination {
  page: number
  limit: number
  total: number
}

export interface Paginated<T> {
  data: T[]
  pagination: Pagination
}

export interface CreateTemplate {
  name: string
  description: string
  content: Content
  active: boolean
}

export interface UpdateTemplate extends CreateTemplate {
  expectedVersion: number
}

// Context received by every micro-component: it depends on neither routes nor global state.
export interface ApiContext {
  baseUrl: string
  token: string
}
