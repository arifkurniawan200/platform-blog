"use client"

import { useState, useEffect, useCallback } from "react"
import { useAuthStore } from "@/stores/auth"
import { listComments, createComment, deleteComment } from "@/lib/api"
import { formatDate } from "@/lib/utils"
import { Button } from "@/components/ui"

interface Comment {
  id: string
  article_id: string
  user_id: string
  parent_id: string | null
  content: string
  created_at: string
}

export function CommentSection({ slug }: { slug: string }) {
  const { token } = useAuthStore()
  const [comments, setComments] = useState<Comment[]>([])
  const [loading, setLoading] = useState(true)
  const [content, setContent] = useState("")
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState("")

  const loadComments = useCallback(async () => {
    try {
      const res = await listComments(slug)
      setComments(res.data || [])
    } catch { /* no comments yet */ }
    setLoading(false)
  }, [slug])

  useEffect(() => { loadComments() }, [loadComments])

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!content.trim() || !token) return

    setSubmitting(true)
    setError("")
    try {
      const res = await createComment(slug, { content: content.trim() }, token)
      setComments(prev => [res.data, ...prev])
      setContent("")
    } catch (err: any) {
      setError(err.message || "Failed to post comment")
    }
    setSubmitting(false)
  }

  async function handleDelete(commentId: string) {
    if (!token) return
    try {
      await deleteComment(slug, commentId, token)
      setComments(prev => prev.filter(c => c.id !== commentId))
    } catch { /* silently fail */ }
  }

  if (loading) return <CommentSkeleton />

  return (
    <section className="max-w-[680px] mx-auto mt-16 pt-12 border-t border-neutral-200 dark:border-neutral-800">
      <h3 className="text-2xl font-bold mb-8">
        Responses ({comments.length})
      </h3>

      {/* Write comment */}
      {token ? (
        <form onSubmit={handleSubmit} className="mb-10 p-4 rounded-xl border border-neutral-200 dark:border-neutral-800 bg-white dark:bg-neutral-950">
          <textarea
            value={content}
            onChange={e => setContent(e.target.value)}
            placeholder="Write a response..."
            rows={3}
            maxLength={2000}
            className="w-full bg-transparent resize-none text-sm focus:outline-none placeholder:text-neutral-400 mb-3"
          />
          <div className="flex items-center justify-between">
            <span className="text-xs text-neutral-400">{content.length}/2000</span>
            <Button size="sm" disabled={!content.trim() || submitting}>
              {submitting ? "Posting..." : "Respond"}
            </Button>
          </div>
          {error && <p className="mt-2 text-xs text-red-500">{error}</p>}
        </form>
      ) : (
        <div className="mb-10 p-6 rounded-xl border border-neutral-200 dark:border-neutral-800 text-center">
          <p className="text-sm text-neutral-500 dark:text-neutral-400">
            <a href="/login" className="underline hover:text-neutral-900 dark:hover:text-white">Sign in</a> to leave a response.
          </p>
        </div>
      )}

      {/* Comments list */}
      {comments.length === 0 ? (
        <p className="text-sm text-neutral-400 text-center py-8">
          No responses yet. Be the first to share your thoughts.
        </p>
      ) : (
        <div className="space-y-6">
          {comments.map(comment => (
            <CommentCard
              key={comment.id}
              comment={comment}
              onDelete={() => handleDelete(comment.id)}
              isOwner={token ? comment.user_id === "current" : false}
            />
          ))}
        </div>
      )}
    </section>
  )
}

function CommentCard({ comment, onDelete, isOwner }: { comment: Comment; onDelete: () => void; isOwner: boolean }) {
  const [showDelete, setShowDelete] = useState(false)

  return (
    <div
      className="group relative"
      onMouseEnter={() => setShowDelete(true)}
      onMouseLeave={() => setShowDelete(false)}
    >
      <div className="flex items-start gap-3">
        <div className="w-8 h-8 rounded-full bg-neutral-200 dark:bg-neutral-800 flex items-center justify-center text-xs font-medium shrink-0 mt-0.5">
          {(comment.user_id || "U")[0].toUpperCase()}
        </div>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-1">
            <span className="text-sm font-medium">User</span>
            <span className="text-xs text-neutral-400">{formatDate(comment.created_at)}</span>
          </div>
          <p className="text-sm leading-relaxed text-neutral-700 dark:text-neutral-300 whitespace-pre-wrap break-words">
            {comment.content}
          </p>
        </div>
      </div>

      {showDelete && isOwner && (
        <button
          onClick={onDelete}
          className="absolute top-0 right-0 text-xs text-neutral-400 hover:text-red-500 transition-colors"
        >
          Delete
        </button>
      )}
    </div>
  )
}

function CommentSkeleton() {
  return (
    <section className="max-w-[680px] mx-auto mt-16 pt-12 border-t border-neutral-200 dark:border-neutral-800">
      <div className="h-7 w-40 bg-neutral-200 dark:bg-neutral-800 rounded animate-pulse mb-8" />
      <div className="space-y-6">
        {[1, 2, 3].map(i => (
          <div key={i} className="flex gap-3">
            <div className="w-8 h-8 rounded-full bg-neutral-200 dark:bg-neutral-800 animate-pulse shrink-0" />
            <div className="flex-1 space-y-2">
              <div className="h-4 w-24 bg-neutral-200 dark:bg-neutral-800 rounded animate-pulse" />
              <div className="h-4 w-full bg-neutral-200 dark:bg-neutral-800 rounded animate-pulse" />
              <div className="h-4 w-2/3 bg-neutral-200 dark:bg-neutral-800 rounded animate-pulse" />
            </div>
          </div>
        ))}
      </div>
    </section>
  )
}
