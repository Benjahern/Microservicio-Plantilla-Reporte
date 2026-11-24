import { afterEach, describe, expect, it, vi } from "vitest"
import { ApiError, request } from "./http-client.api"

const ctx = { baseUrl: "", token: "abc" }

afterEach(() => vi.unstubAllGlobals())

describe("request", () => {
  it("sends the Bearer token and parses JSON", async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify({ ok: true }), { status: 200 }))
    vi.stubGlobal("fetch", fetchMock)
    await expect(request(ctx, "/api/v1/plantillas")).resolves.toEqual({ ok: true })
    expect(fetchMock.mock.calls[0]?.[1]?.headers).toMatchObject({ Authorization: "Bearer abc" })
  })

  it("translates the API error into ApiError", async () => {
    const body = { code: "forbidden", title: "Insufficient permission" }
    vi.stubGlobal("fetch", vi.fn().mockImplementation(() => Promise.resolve(new Response(JSON.stringify(body), { status: 403 }))))
    await expect(request(ctx, "/x")).rejects.toMatchObject({ status: 403, code: "forbidden" })
    await expect(request(ctx, "/x")).rejects.toBeInstanceOf(ApiError)
  })

  it("returns undefined on 204", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(null, { status: 204 })))
    await expect(request(ctx, "/x", { method: "DELETE" })).resolves.toBeUndefined()
  })
})
