import { request } from "@/shared/http-client.api"
import { endpoint } from "@/types/contract.gen"
import { toCreateWire, toTemplate, toUpdateWire } from "@/types/template.mappers"
import type { TemplateWire } from "@/types/template.wire"
import type { ApiContext, CreateTemplate, Template, UpdateTemplate } from "@/types/template"

export const getTemplate = async (ctx: ApiContext, id: string): Promise<Template> =>
  toTemplate(await request<TemplateWire>(ctx, endpoint("obtenerPlantilla", { id }).path))

export const createTemplate = async (ctx: ApiContext, data: CreateTemplate): Promise<Template> => {
  const { path, method } = endpoint("crearPlantilla")
  return toTemplate(await request<TemplateWire>(ctx, path, { method, body: toCreateWire(data) }))
}

export const updateTemplate = async (ctx: ApiContext, id: string, data: UpdateTemplate): Promise<Template> => {
  const { path, method } = endpoint("actualizarPlantilla", { id })
  return toTemplate(await request<TemplateWire>(ctx, path, { method, body: toUpdateWire(data) }))
}
