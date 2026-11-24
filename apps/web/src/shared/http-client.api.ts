import type { ApiContext } from "@/types/template"

export class ApiError extends Error {
  constructor(
    public readonly status: number,
    public readonly code: string,
    message: string,
  ) {
    super(message)
    this.name = "ApiError"
  }
}

interface ErrorBody {
  code?: string
  detail?: string
  title?: string
}

export async function request<T>(
  ctx: ApiContext,
  path: string,
  init: { method?: string; body?: unknown } = {},
): Promise<T> {
  const response = await fetch(`${ctx.baseUrl}${path}`, {
    method: init.method ?? "GET",
    credentials: "same-origin",
    signal: AbortSignal.timeout(5000),
    headers: {
      ...(ctx.token ? { Authorization: `Bearer ${ctx.token}` } : {}),
      ...(init.body !== undefined ? { "Content-Type": "application/json" } : {}),
    },
    body: init.body !== undefined ? JSON.stringify(init.body) : undefined,
  })

  if (!response.ok) {
    const body = (await response.json().catch(() => ({}))) as ErrorBody
    throw new ApiError(
      response.status,
      body.code ?? "unknown",
      body.detail ?? body.title ?? `Error ${response.status}`,
    )
  }
  if (response.status === 204) return undefined as T
  return (await response.json()) as T
}

// Turns any error into a user-facing message (shown in Spanish).
export function errorMessage(e: unknown): string {
  if (e instanceof ApiError) {
    if (e.status === 401) return "Inicia sesión para consultar las plantillas."
    if (e.status === 403) return "Tu perfil no tiene permiso para esta acción."
    return e.message
  }
  if (e instanceof Error && e.name === "TimeoutError") return "La API no respondió a tiempo."
  return e instanceof Error ? e.message : "Error inesperado"
}
