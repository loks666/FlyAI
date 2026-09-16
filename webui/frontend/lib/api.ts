const API_BASE = process.env.NEXT_PUBLIC_API_BASE_URL ?? ""

export async function api<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...init?.headers,
    },
  })
  const body = await response.json().catch(() => ({}))
  if (!response.ok) {
    throw new Error(body.message || body.msg || `请求失败（${response.status}）`)
  }
  return body as T
}

export const jsonBody = (value: unknown): RequestInit => ({
  body: JSON.stringify(value),
})
