"use client"

import { useEffect, useState, Suspense } from "react"
import { useSearchParams, useRouter } from "next/navigation"
import Link from "next/link"
import { searchArticles } from "@/lib/api"
import { Card } from "@/components/ui"
import { formatDate } from "@/lib/utils"

interface SearchResult {
  id: string
  title: string
  slug: string
  subtitle?: string
  cover_image?: string
  reading_time: number
  published_at?: string
  clap_count: number
  comment_count: number
  author_username: string
  author_display_name: string
  rank: number
}

function SearchInner() {
  const searchParams = useSearchParams()
  const router = useRouter()
  const q = searchParams.get("q") || ""
  const [query, setQuery] = useState(q)
  const [results, setResults] = useState<SearchResult[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (!q) return
    setLoading(true)
    searchArticles(q)
      .then((res) => setResults(res.data || []))
      .catch(() => setResults([]))
      .finally(() => setLoading(false))
  }, [q])

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    if (query.trim()) router.push(`/search?q=${encodeURIComponent(query.trim())}`)
  }

  return (
    <div className="max-w-2xl mx-auto px-4 py-16">
      <h1 className="text-3xl font-bold mb-8">Search</h1>

      <form onSubmit={handleSearch} className="mb-8">
        <div className="flex gap-2">
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search articles..."
            className="flex-1 px-4 py-3 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 focus:outline-none focus:ring-2 focus:ring-black dark:focus:ring-white"
          />
          <button
            type="submit"
            className="px-6 py-3 bg-black dark:bg-white text-white dark:text-black rounded-lg font-medium hover:opacity-80 transition"
          >
            Search
          </button>
        </div>
      </form>

      {!q && <p className="text-gray-500 text-center">Enter a search term to find articles.</p>}

      {loading && <SearchSkeleton />}

      {!loading && q && results.length === 0 && (
        <p className="text-gray-500 text-center">No results found for &quot;{q}&quot;.</p>
      )}

      {!loading && results.length > 0 && (
        <div className="space-y-4">
          {results.map((r) => (
            <Link key={r.id} href={`/article/${r.slug}`}>
              <Card className="hover:shadow-md transition cursor-pointer p-5">
                <div className="flex gap-4">
                  {r.cover_image && (
                    <img src={r.cover_image} alt="" className="w-20 h-14 object-cover rounded" />
                  )}
                  <div className="flex-1 min-w-0">
                    <h2 className="font-semibold text-lg truncate">{r.title}</h2>
                    {r.subtitle && <p className="text-gray-500 text-sm truncate">{r.subtitle}</p>}
                    <div className="flex items-center gap-3 mt-1 text-xs text-gray-400">
                      <span>By {r.author_display_name}</span>
                      {r.published_at && <span>{formatDate(r.published_at)}</span>}
                      <span>{r.reading_time} min read</span>
                      <span>👏 {r.clap_count}</span>
                      <span>💬 {r.comment_count}</span>
                    </div>
                  </div>
                </div>
              </Card>
            </Link>
          ))}
        </div>
      )}
    </div>
  )
}

export default function SearchPage() {
  return (
    <Suspense fallback={<SearchSkeleton />}>
      <SearchInner />
    </Suspense>
  )
}

function SearchSkeleton() {
  return (
    <div className="space-y-4 animate-pulse">
      {[1, 2, 3].map((i) => (
        <div key={i} className="h-24 bg-gray-200 dark:bg-gray-700 rounded" />
      ))}
    </div>
  )
}
