const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1"

interface RequestOptions extends RequestInit {
  token?: string
}

async function request<T>(endpoint: string, options: RequestOptions = {}): Promise<T> {
  const { token, ...init } = options
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(init.headers as Record<string, string>),
  }
  if (token) headers["Authorization"] = `Bearer ${token}`

  const res = await fetch(`${API_BASE}${endpoint}`, { ...init, headers })
  const data = await res.json()

  if (!res.ok) throw new Error(data.error?.message || data.message || "Request failed")
  return data
}

// Auth
export function authRegister(body: { username: string; email: string; password: string }) {
  return request<{ success: boolean; data: { access_token: string; refresh_token: string } }>("/auth/register", { method: "POST", body: JSON.stringify(body) })
}
export function authLogin(body: { email: string; password: string }) {
  return request<{ success: boolean; data: { access_token: string; refresh_token: string } }>("/auth/login", { method: "POST", body: JSON.stringify(body) })
}

// Articles
export function listArticles(page = 1, limit = 12) {
  const offset = (page - 1) * limit
  return request<{ success: boolean; data: any[] }>(`/articles?limit=${limit}&offset=${offset}`)
}
export function getArticle(slug: string) {
  return request<{ success: boolean; data: any }>(`/articles/${slug}`)
}
export function createArticle(body: { title: string; content: string; tags?: string[] }, token: string) {
  return request<{ success: boolean; data: any }>("/articles", { method: "POST", body: JSON.stringify(body), token })
}
export function updateArticle(slug: string, body: { title: string; content: string; tags?: string[] }, token: string) {
  return request<{ success: boolean; data: any }>(`/articles/${slug}`, { method: "PUT", body: JSON.stringify(body), token })
}
