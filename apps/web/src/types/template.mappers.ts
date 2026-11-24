// Translation layer between the backend contract (Spanish names) and the UI domain (English names).
// If the backend renames its fields, only this file and template.wire.ts need to change.
import type { CreateTemplate, Paginated, Template, UpdateTemplate, Version } from "./template"
import type { CreateTemplateWire, TemplateWire, UpdateTemplateWire, VersionWire } from "./template.wire"

export const toVersion = (w: VersionWire): Version => ({
  number: w.numero,
  name: w.nombre,
  description: w.descripcion,
  content: w.contenido,
  active: w.activo,
  authorId: w.autor_id,
  createdAt: w.created_at,
})

export const toTemplate = (w: TemplateWire): Template => ({
  ...toVersion(w),
  id: w.id,
  ownerId: w.propietario_id,
  currentVersion: w.version_actual,
  updatedAt: w.updated_at,
})

export const toCreateWire = (t: CreateTemplate): CreateTemplateWire => ({
  nombre: t.name,
  descripcion: t.description,
  contenido: t.content,
  activo: t.active,
})

export const toUpdateWire = (t: UpdateTemplate): UpdateTemplateWire => ({
  nombre: t.name,
  descripcion: t.description,
  contenido: t.content,
  activo: t.active,
  version_esperada: t.expectedVersion,
})

export const toPaginated = <W, T>(page: Paginated<W>, map: (item: W) => T): Paginated<T> => ({
  data: page.data.map(map),
  pagination: page.pagination,
})
