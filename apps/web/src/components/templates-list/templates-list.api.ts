import { request } from "@/shared/http-client.api"
import { endpoint } from "@/types/contract.gen"
import { toPaginated, toTemplate } from "@/types/template.mappers"
import type { TemplateWire } from "@/types/template.wire"
import type { ApiContext, Paginated, Template } from "@/types/template"

export const listTemplates = async (ctx: ApiContext, page: number, limit: number): Promise<Paginated<Template>> =>
  toPaginated(await request<Paginated<TemplateWire>>(ctx, `${endpoint("listarPlantillas").path}?page=${page}&limit=${limit}`), toTemplate)

export const deleteTemplate = (ctx: ApiContext, id: string) => {
  const { path, method } = endpoint("eliminarPlantilla", { id })
  return request<void>(ctx, path, { method })
}
