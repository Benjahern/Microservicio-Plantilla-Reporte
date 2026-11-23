import { request } from "@/shared/http-client.api"
import { toPaginated, toTemplate } from "@/types/template.mappers"
import type { TemplateWire } from "@/types/template.wire"
import type { ApiContext, Paginated, Template } from "@/types/template"

// Path defined by the backend contract.
const BASE_PATH = "/api/v1/plantillas"

export const listTemplates = async (ctx: ApiContext, page: number, limit: number): Promise<Paginated<Template>> =>
  toPaginated(await request<Paginated<TemplateWire>>(ctx, `${BASE_PATH}?page=${page}&limit=${limit}`), toTemplate)

export const deleteTemplate = (ctx: ApiContext, id: string) =>
  request<void>(ctx, `${BASE_PATH}/${encodeURIComponent(id)}`, { method: "DELETE" })
