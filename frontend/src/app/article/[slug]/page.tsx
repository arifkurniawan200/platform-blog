"use client"

import { useEffect, useState } from "react"
import { useParams } from "next/navigation"
import { getArticle } from "@/lib/api"
import { formatDate, readingTime } from "@/lib/utils"
import { Badge } from "@/components/ui"

export default function ArticlePage() {
  const { slug } = useParams<{ slug: string }>()
  const [article, setArticle] = useState<any>(null)
  const [loading, setLoading] = useState(true)
  const [scrollProgress, setScrollProgress] = useState(0)

  useEffect(() => {
    getArticle(slug).then((r: any) => { setArticle(r.data); setLoading(false) }).catch(() => setLoading(false))
  }, [slug])

  useEffect(() => {
    const handleScroll = () => {
      const scrollTop = window.scrollY
      const docHeight = document.documentElement.scrollHeight - window.innerHeight
      setScrollProgress(docHeight > 0 ? (scrollTop / docHeight) * 100 : 0)
    }
    window.addEventListener("scroll", handleScroll)
    return () => window.removeEventListener("scroll", handleScroll)
  }, [])

  if (loading) return <div className="flex justify-center py-20"><div className="animate-spin h-8 w-8 border-2 border-neutral-900 dark:border-white border-t-transparent rounded-full" /></div>
  if (!article) return <div className="text-center py-20 text-neutral-400">Article not found.</div>

  return (
    <>
      {/* Reading progress bar */}
      <div className="fixed top-14 left-0 h-0.5 bg-neutral-900 dark:bg-white z-40 transition-all duration-150" style={{ width: `${scrollProgress}%` }} />

      <article className="max-w-[680px] mx-auto">
        {/* Header */}
        <header className="mb-10 pt-8">
          <h1 className="font-serif text-4xl md:text-5xl font-bold leading-tight mb-6">{article.title}</h1>
          <div className="flex items-center gap-3 text-sm text-neutral-500 dark:text-neutral-400">
            <span className="font-medium text-neutral-700 dark:text-neutral-300">{article.author?.display_name || "Anonymous"}</span>
            <span>·</span>
            <span>{formatDate(article.published_at || article.created_at)}</span>
            <span>·</span>
            <span>{readingTime(article.content)} min read</span>
          </div>
          {article.tags?.length > 0 && (
            <div className="flex flex-wrap gap-2 mt-4">
              {article.tags.map((t: any) => <Badge key={t.id || t.name}>{t.name || t}</Badge>)}
            </div>
          )}
        </header>

        {/* Content */}
        <div
          className="prose prose-lg prose-neutral dark:prose-invert max-w-none
            prose-headings:font-serif prose-headings:font-bold prose-headings:tracking-tight
            prose-h2:text-2xl prose-h2:mt-12 prose-h2:mb-4
            prose-p:leading-relaxed prose-p:text-[18px] prose-p:mb-6
            prose-blockquote:border-l-2 prose-blockquote:border-neutral-900 dark:prose-blockquote:border-white
            prose-blockquote:italic prose-blockquote:text-neutral-600 dark:prose-blockquote:text-neutral-300
            prose-code:text-sm prose-code:bg-neutral-100 dark:prose-code:bg-neutral-800 prose-code:px-1.5 prose-code:py-0.5 prose-code:rounded
            prose-pre:bg-neutral-900 dark:prose-pre:bg-black prose-pre:text-neutral-100 prose-pre:rounded-xl prose-pre:text-sm
            prose-img:rounded-xl prose-img:shadow-lg
            prose-a:text-neutral-900 dark:prose-a:text-white prose-a:underline prose-a:decoration-neutral-400
            prose-strong:text-neutral-900 dark:prose-strong:text-white
          "
          dangerouslySetInnerHTML={{ __html: article.content }}
        />

        {/* Footer */}
        <footer className="mt-16 pt-8 border-t border-neutral-200 dark:border-neutral-800">
          <div className="flex items-center gap-4">
            <div className="w-12 h-12 rounded-full bg-neutral-200 dark:bg-neutral-800 flex items-center justify-center text-lg font-medium">
              {(article.author?.display_name || "A")[0].toUpperCase()}
            </div>
            <div>
              <p className="font-medium">{article.author?.display_name || "Anonymous"}</p>
              <p className="text-sm text-neutral-500">Published on {formatDate(article.published_at || article.created_at)}</p>
            </div>
          </div>
        </footer>
      </article>
    </>
  )
}
