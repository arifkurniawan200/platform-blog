"use client"

import { useEffect, useState } from "react"
import Link from "next/link"
import { listArticles } from "@/lib/api"
import { Card, Badge } from "@/components/ui"
import { formatDate, readingTime } from "@/lib/utils"

function HomeSkeleton() {
  return (
    <div>
      <div className="h-10 w-48 bg-neutral-100 dark:bg-neutral-800 rounded animate-pulse mb-3" />
      <div className="h-5 w-96 bg-neutral-100 dark:bg-neutral-800 rounded animate-pulse mb-10" />
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {[1,2,3,4,5,6].map(i => (
          <div key={i} className="rounded-xl border border-neutral-200 dark:border-neutral-800 p-6">
            <div className="flex gap-2 mb-3">
              <div className="h-3 w-16 bg-neutral-100 dark:bg-neutral-800 rounded animate-pulse" />
              <div className="h-3 w-12 bg-neutral-100 dark:bg-neutral-800 rounded animate-pulse" />
              <div className="h-3 w-14 bg-neutral-100 dark:bg-neutral-800 rounded animate-pulse" />
            </div>
            <div className="h-6 bg-neutral-100 dark:bg-neutral-800 rounded animate-pulse mb-2" />
            <div className="h-4 w-3/4 bg-neutral-100 dark:bg-neutral-800 rounded animate-pulse mb-1" />
            <div className="h-4 w-1/2 bg-neutral-100 dark:bg-neutral-800 rounded animate-pulse mb-3" />
            <div className="flex gap-1.5">
              <div className="h-5 w-12 bg-neutral-100 dark:bg-neutral-800 rounded-full animate-pulse" />
              <div className="h-5 w-16 bg-neutral-100 dark:bg-neutral-800 rounded-full animate-pulse" />
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

export default function HomePage() {
  const [articles, setArticles] = useState<any[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => { listArticles().then((r: any) => { setArticles(r.data); setLoading(false) }).catch(() => setLoading(false)) }, [])

  if (loading) return <HomeSkeleton />

  return (
    <div>
      <div className="mb-10">
        <h1 className="text-4xl font-bold tracking-tight mb-3">Platform Blog</h1>
        <p className="text-lg text-neutral-500 dark:text-neutral-400 max-w-2xl">Discover stories, thinking, and expertise from writers on any topic.</p>
      </div>
      {articles.length === 0 ? (
        <div className="text-center py-20 text-neutral-400">No articles yet. Be the first to write!</div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {articles.map((a: any) => (
            <Link key={a.id} href={`/article/${a.slug}`} className="group">
              <Card className="h-full hover:border-neutral-400 dark:hover:border-neutral-600 transition-colors">
                <div className="flex items-center gap-2 mb-3 text-xs text-neutral-500">
                  <span>{a.author?.display_name || "Anonymous"}</span>
                  <span>·</span>
                  <span>{formatDate(a.published_at || a.created_at)}</span>
                  <span>·</span>
                  <span>{readingTime(a.content)} min read</span>
                </div>
                <h2 className="font-semibold text-lg mb-2 group-hover:text-neutral-600 dark:group-hover:text-neutral-300 transition-colors leading-snug">{a.title}</h2>
                <p className="text-sm text-neutral-500 dark:text-neutral-400 line-clamp-2 mb-3">{a.content?.replace(/<[^>]*>/g, "").substring(0, 150)}</p>
                <div className="flex flex-wrap gap-1.5">
                  {a.tags?.map((t: any) => <Badge key={t.id || t.name}>{t.name || t}</Badge>)}
                </div>
              </Card>
            </Link>
          ))}
        </div>
      )}
    </div>
  )
}
