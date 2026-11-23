import { request } from "@/shared/http-client.api"
import { toPaginated, toVersion } from "@/types/template.mappers"
import type { VersionWire } from "@/types/template.wire"
import type { ApiContext, Paginated, Version } from "@/types/template"

// Path defined by the backend contract.
const path = (id: string) => `/api/v1/plantillas/${encodeURIComponent(id)}/versiones`

export const listVersions = async (
  ctx: ApiContext,
  id: string,
  page: number,
  limit: number,
): Promise<Paginated<Version>> =>
  toPaginated(await request<Paginated<VersionWire>>(ctx, `${path(id)}?page=${page}&limit=${limit}`), toVersion)

export const getVersion = async (ctx: ApiContext, id: string, number: number): Promise<Version> =>
  toVersion(await request<VersionWire>(ctx, `${path(id)}/${number}`))
