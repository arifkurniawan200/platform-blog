"use client"

import { useState, useEffect } from "react"
import Link from "next/link"
import { useAuthStore } from "@/stores/auth"
import { getBookmarks, unbookmarkArticle } from "@/lib/api"
import { Card, Button } from "@/components/ui"

export default function BookmarksPage() {
  const { token } = useAuthStore()
  const [bookmarks, setBookmarks] = useState<any[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!token) { setLoading(false); return }
    getBookmarks(token)
      .then(res => setBookmarks(res.data || []))
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [token])

  async function handleRemove(slug: string) {
    if (!token) return
    try {
      await unbookmarkArticle(slug, token)
      setBookmarks(prev => prev.filter(b => b.slug !== slug))
    } catch {}
  }

  if (!token) {
    return (
      <div className="max-w-2xl mx-auto py-20 text-center">
        <h1 className="text-3xl font-bold mb-4">Your Bookmarks</h1>
        <p className="text-neutral-500 mb-6">Sign in to save and view your bookmarked stories.</p>
        <Link href="/login"><Button>Sign In</Button></Link>
      </div>
    )
  }

  if (loading) {
    return (
      <div className="max-w-3xl mx-auto">
        <h1 className="text-3xl font-bold mb-8">Bookmarks</h1>
        <div className="space-y-4">
          {[1, 2, 3, 4].map(i => (
            <div key={i} className="h-20 bg-neutral-100 dark:bg-neutral-900 rounded-xl animate-pulse" />
          ))}
        </div>
      </div>
    )
  }

  return (
    <div className="max-w-3xl mx-auto">
      <h1 className="text-3xl font-bold mb-8">Your Bookmarks</h1>

      {bookmarks.length === 0 ? (
        <div className="text-center py-16">
          <div className="text-5xl mb-4">📖</div>
          <h2 className="text-xl font-semibold mb-2">No bookmarks yet</h2>
          <p className="text-neutral-500 mb-6">Click the bookmark icon on any article to save it here.</p>
          <Link href="/"><Button variant="outline">Discover Stories</Button></Link>
        </div>
      ) : (
        <div className="space-y-4">
          {bookmarks.map(bm => (
            <Card key={bm.article_id} className="flex items-start justify-between gap-4 hover:border-neutral-400 dark:hover:border-neutral-600 transition-colors">
              <Link href={`/article/${bm.slug}`} className="flex-1 min-w-0">
                <h2 className="font-semibold text-lg mb-1 hover:text-neutral-600 dark:hover:text-neutral-300 transition-colors truncate">
                  {bm.title}
                </h2>
                <p className="text-sm text-neutral-400">Saved</p>
              </Link>
              <button
                onClick={() => handleRemove(bm.slug)}
                className="text-neutral-400 hover:text-red-500 transition-colors p-1 shrink-0"
              >
                <svg width="18" height="18" viewBox="0 0 24 24">
                  <path d="M6 7v14a2 2 0 002 2h8a2 2 0 002-2V7M9 7V5a2 2 0 012-2h2a2 2 0 012 2v2M3 7h18" stroke="currentColor" strokeWidth="1.5" fill="none" strokeLinecap="round" strokeLinejoin="round" />
                </svg>
              </button>
            </Card>
          ))}
        </div>
      )}
    </div>
  )
}
