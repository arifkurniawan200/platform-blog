"use client"

import { useState } from "react"
import { useAuthStore } from "@/stores/auth"
import { bookmarkArticle, unbookmarkArticle } from "@/lib/api"

export function BookmarkButton({ slug }: { slug: string }) {
  const { token } = useAuthStore()
  const [bookmarked, setBookmarked] = useState(false)
  const [loading, setLoading] = useState(false)

  // Check initial bookmark state via the bookmark toggle
  // Since we don't have a GET single bookmark, we'll use optimistic UI

  async function toggle() {
    if (!token || loading) return
    setLoading(true)
    try {
      if (bookmarked) {
        await unbookmarkArticle(slug, token)
        setBookmarked(false)
      } else {
        await bookmarkArticle(slug, token)
        setBookmarked(true)
      }
    } catch {
      // Revert on error
      setBookmarked(prev => !prev)
    }
    setLoading(false)
  }

  return (
    <button
      onClick={toggle}
      disabled={!token}
      className={`p-2 rounded-full transition-all duration-200 ${
        !token
          ? "opacity-40 cursor-not-allowed"
          : bookmarked
          ? "text-neutral-900 dark:text-white"
          : "text-neutral-400 hover:text-neutral-700 dark:hover:text-neutral-300"
      }`}
      aria-label={bookmarked ? "Remove bookmark" : "Bookmark article"}
      title={!token ? "Sign in to bookmark" : bookmarked ? "Remove bookmark" : "Bookmark"}
    >
      <svg
        width="22"
        height="22"
        viewBox="0 0 24 24"
        className={`transition-transform duration-200 ${loading ? "scale-90" : ""}`}
      >
        <path
          d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z"
          fill={bookmarked ? "currentColor" : "none"}
          stroke="currentColor"
          strokeWidth="1.5"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
    </button>
  )
}
