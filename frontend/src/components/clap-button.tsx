"use client"

import { useState, useEffect, useCallback, useRef } from "react"
import { useAuthStore } from "@/stores/auth"
import { getClapInfo, clapArticle } from "@/lib/api"

export function ClapButton({ slug }: { slug: string }) {
  const { token } = useAuthStore()
  const [totalClaps, setTotalClaps] = useState(0)
  const [userClaps, setUserClaps] = useState(0)
  const [animating, setAnimating] = useState(false)
  const [pendingClaps, setPendingClaps] = useState<number[]>([])
  const clapTimer = useRef<ReturnType<typeof setTimeout>>()

  useEffect(() => {
    getClapInfo(slug, token || undefined)
      .then(res => {
        const d = res.data || res
        setTotalClaps(d.total_claps || 0)
        setUserClaps(d.user_claps || 0)
      })
      .catch(() => {})
  }, [slug, token])

  const handleClap = useCallback(() => {
    if (!token) return
    const count = 1
    setAnimating(true)
    setPendingClaps(prev => [...prev, count])

    // Fire and forget — update optimistically
    clapArticle(slug, count, token)
      .then(res => {
        const d = res.data || res
        setTotalClaps(d.total_claps || 0)
        setUserClaps(d.user_claps || 0)
      })
      .catch(() => {})

    // Clear animation
    if (clapTimer.current) clearTimeout(clapTimer.current)
    clapTimer.current = setTimeout(() => {
      setAnimating(false)
      setPendingClaps([])
    }, 700)
  }, [slug, token])

  // Floating clap particles
  const particles = pendingClaps.map((_, i) => (
    <span
      key={i}
      className="absolute text-lg select-none pointer-events-none"
      style={{
        left: `${Math.random() * 20 - 10}px`,
        animation: `clapFloat ${0.6 + Math.random() * 0.3}s ease-out forwards`,
        opacity: 0,
      }}
    >
      👏
    </span>
  ))

  return (
    <div className="fixed left-8 bottom-8 flex flex-col items-center gap-1 z-30">
      {/* Floating particles */}
      <div className="relative">
        {particles}
        <button
          onClick={handleClap}
          className={`relative w-16 h-16 rounded-full border-2 border-neutral-300 dark:border-neutral-600 flex items-center justify-center transition-all duration-150
            ${animating ? "scale-110 border-neutral-900 dark:border-white bg-neutral-100 dark:bg-neutral-800" : "hover:border-neutral-500 dark:hover:border-neutral-400 bg-white dark:bg-neutral-950"}
            ${!token ? "opacity-50 cursor-not-allowed" : "cursor-pointer hover:scale-105 hover:shadow-lg"}
          `}
          disabled={!token}
          aria-label="Clap"
        >
          <svg
            width="24"
            height="24"
            viewBox="0 0 24 24"
            className={`transition-transform duration-150 ${animating ? "scale-110" : ""}`}
          >
            <path
              d="M11.5 3.5c-.5-.8-1.5-1-2.3-.5S8 4.8 8.5 5.6l1.2 2c-.5.1-1 .3-1.5.6l-2-3.4c-.5-.8-1.5-1-2.3-.5s-1 1.5-.5 2.3l2.5 4.2-2.7.1c-.9 0-1.7.8-1.7 1.7s.8 1.7 1.7 1.7l4.8-.2.2.1c.4.3.9.4 1.4.4h4c1.7 0 3-1.3 3-3v-1c0-.6-.4-1-1-1h-1.5l.3-1.5c.2-.8-.3-1.6-1.1-1.8l-2.4-.7.5-3.2c.1-.9-.5-1.7-1.4-1.8-.9-.2-1.7.4-1.9 1.3l-.6 3.5c-.2-.1-.4-.2-.6-.2h-.4l1-3.8z"
              fill={userClaps > 0 ? "currentColor" : "none"}
              stroke="currentColor"
              strokeWidth="1.5"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
            <path
              d="M4.5 11.5l-1 4c-.3 1.1.4 2.2 1.5 2.5l6 1.5c.3.1.7.1 1 .1h3.5c1.4 0 2.5-1.1 2.5-2.5v-1"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.5"
              strokeLinecap="round"
            />
          </svg>
        </button>
      </div>
      <span className={`text-sm font-semibold transition-colors ${animating ? "text-neutral-900 dark:text-white" : "text-neutral-500 dark:text-neutral-400"}`}>
        {totalClaps}
      </span>
    </div>
  )
}
