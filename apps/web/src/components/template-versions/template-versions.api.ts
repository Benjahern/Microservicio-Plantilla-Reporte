import { request } from "@/shared/http-client.api"
import { endpoint } from "@/types/contract.gen"
import { toPaginated, toVersion } from "@/types/template.mappers"
import type { VersionWire } from "@/types/template.wire"
import type { ApiContext, Paginated, Version } from "@/types/template"

export const listVersions = async (
  ctx: ApiContext,
  id: string,
  page: number,
  limit: number,
): Promise<Paginated<Version>> =>
  toPaginated(await request<Paginated<VersionWire>>(ctx, `${endpoint("listarVersiones", { id }).path}?page=${page}&limit=${limit}`), toVersion)

export const getVersion = async (ctx: ApiContext, id: string, number: number): Promise<Version> =>
  toVersion(await request<VersionWire>(ctx, endpoint("obtenerVersion", { id, numero: number }).path))
