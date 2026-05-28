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

// Comments
export function listComments(slug: string) {
  return request<{ success: boolean; data: any[] }>(`/articles/${slug}/comments`)
}
export function createComment(slug: string, body: { content: string; parent_id?: string }, token: string) {
  return request<{ success: boolean; data: any }>(`/articles/${slug}/comments`, { method: "POST", body: JSON.stringify(body), token })
}
export function deleteComment(slug: string, commentId: string, token: string) {
  return request(`/articles/${slug}/comments/${commentId}`, { method: "DELETE", token })
}

// Claps
export function getClapInfo(slug: string, token?: string) {
  return request<{ total_claps: number; user_claps: number; article_id: string }>(`/articles/${slug}/clap${token ? "" : ""}`, token ? { token } : {})
}
export function clapArticle(slug: string, count: number, token: string) {
  return request<{ total_claps: number; user_claps: number; article_id: string }>(`/articles/${slug}/clap`, { method: "POST", body: JSON.stringify({ count }), token })
}

// Bookmarks
export function getBookmarks(token: string, page = 1, limit = 20) {
  const offset = (page - 1) * limit
  return request<{ success: boolean; data: any[] }>(`/bookmarks?limit=${limit}&offset=${offset}`, { token })
}
export function bookmarkArticle(slug: string, token: string) {
  return request(`/articles/${slug}/bookmark`, { method: "POST", token })
}
export function unbookmarkArticle(slug: string, token: string) {
  return request(`/articles/${slug}/bookmark`, { method: "DELETE", token })
}

// Search
export function searchArticles(q: string, page = 1, limit = 20) {
  const offset = (page - 1) * limit
  return request<{ success?: boolean; data: any[] }>(`/search?q=${encodeURIComponent(q)}&limit=${limit}&offset=${offset}`)
}

// User profile
export function getUserProfile(username: string) {
  return request<{ id: string; username: string; display_name: string; bio?: string; avatar_url?: string; email_notify_comments: boolean; article_count: number; created_at: string }>(`/users/${username}`)
}
export function updateUserProfile(body: { display_name?: string; bio?: string; avatar_url?: string; email_notify_comments?: boolean }, token: string) {
  return request(`/users/me`, { method: "PATCH", body: JSON.stringify(body), token })
}

// User stats
export function getUserStats(userId: string) {
  return request<{ success?: boolean; data: { article_count: number; total_claps: number; total_comments: number; total_views: number } }>(`/users/${userId}/stats`)
}
