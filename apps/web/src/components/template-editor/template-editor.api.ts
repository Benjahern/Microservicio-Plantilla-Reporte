import { request } from "@/shared/http-client.api"
import { toCreateWire, toTemplate, toUpdateWire } from "@/types/template.mappers"
import type { TemplateWire } from "@/types/template.wire"
import type { ApiContext, CreateTemplate, Template, UpdateTemplate } from "@/types/template"

// Path defined by the backend contract.
const BASE_PATH = "/api/v1/plantillas"

export const getTemplate = async (ctx: ApiContext, id: string): Promise<Template> =>
  toTemplate(await request<TemplateWire>(ctx, `${BASE_PATH}/${encodeURIComponent(id)}`))

export const createTemplate = async (ctx: ApiContext, data: CreateTemplate): Promise<Template> =>
  toTemplate(await request<TemplateWire>(ctx, BASE_PATH, { method: "POST", body: toCreateWire(data) }))

export const updateTemplate = async (ctx: ApiContext, id: string, data: UpdateTemplate): Promise<Template> =>
  toTemplate(
    await request<TemplateWire>(ctx, `${BASE_PATH}/${encodeURIComponent(id)}`, {
      method: "PUT",
      body: toUpdateWire(data),
    }),
  )
