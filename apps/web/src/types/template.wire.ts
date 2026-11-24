import type { components } from "./contract.gen"

type Schemas = components["schemas"]
export type VersionWire = Schemas["Version"]
export type TemplateWire = Schemas["Plantilla"]
export type CreateTemplateWire = Schemas["Crear"]
export type UpdateTemplateWire = Schemas["Actualizar"]
