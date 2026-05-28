"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"
import { useEditor, EditorContent } from "@tiptap/react"
import StarterKit from "@tiptap/starter-kit"
import ImageExtension from "@tiptap/extension-image"
import LinkExtension from "@tiptap/extension-link"
import Placeholder from "@tiptap/extension-placeholder"
import { createArticle } from "@/lib/api"
import { useAuthStore } from "@/stores/auth"
import { Button, Input, Badge } from "@/components/ui"
import { Bold, Italic, List, ListOrdered, Quote, Code, Link2, Image, Heading1, Heading2 } from "lucide-react"

export default function WritePage() {
  const [title, setTitle] = useState("")
  const [tagInput, setTagInput] = useState("")
  const [tags, setTags] = useState<string[]>([])
  const [publishing, setPublishing] = useState(false)
  const [error, setError] = useState("")
  const token = useAuthStore((s) => s.token)
  const router = useRouter()

  const editor = useEditor({
    extensions: [
      StarterKit,
      ImageExtension,
      LinkExtension.configure({ openOnClick: true }),
      Placeholder.configure({ placeholder: "Tell your story..." }),
    ],
    editorProps: { attributes: { class: "prose prose-neutral dark:prose-invert max-w-none min-h-[300px] focus:outline-none" } },
  })

  const addTag = () => {
    const t = tagInput.trim().toLowerCase()
    if (t && !tags.includes(t) && tags.length < 5) { setTags([...tags, t]); setTagInput("") }
  }

  const handlePublish = async () => {
    if (!title.trim() || !editor?.getHTML()) { setError("Title and content are required"); return }
    if (!token) { setError("Please login first"); return }
    setPublishing(true)
    setError("")
    try {
      const res = await createArticle({ title, content: editor.getHTML(), tags }, token)
      router.push(`/article/${res.data.slug}`)
    } catch (err: any) {
      setError(err.message)
    } finally {
      setPublishing(false)
    }
  }

  const ToolbarButton = ({ onClick, active, children }: any) => (
    <button type="button" onClick={onClick} className={`p-1.5 rounded hover:bg-neutral-100 dark:hover:bg-neutral-800 transition-colors ${active ? "bg-neutral-100 dark:bg-neutral-800" : ""}`}>{children}</button>
  )

  if (!token) return <div className="max-w-3xl mx-auto pt-20 text-center"><h1 className="text-2xl font-bold mb-4">Login required</h1><p className="text-neutral-500">You need to be logged in to write.</p></div>

  return (
    <div className="max-w-3xl mx-auto">
      <Input className="!text-3xl !font-bold !border-none !h-auto !px-0 !py-2 mb-4" value={title} onChange={(e: any) => setTitle(e.target.value)} placeholder="Article title..." />

      {/* Toolbar */}
      <div className="flex items-center gap-0.5 mb-4 p-1 border border-neutral-200 dark:border-neutral-800 rounded-lg bg-neutral-50 dark:bg-neutral-900">
        <ToolbarButton onClick={() => editor?.chain().focus().toggleBold().run()} active={editor?.isActive("bold")}><Bold size={16} /></ToolbarButton>
        <ToolbarButton onClick={() => editor?.chain().focus().toggleItalic().run()} active={editor?.isActive("italic")}><Italic size={16} /></ToolbarButton>
        <ToolbarButton onClick={() => editor?.chain().focus().toggleHeading({ level: 1 }).run()} active={editor?.isActive("heading", { level: 1 })}><Heading1 size={16} /></ToolbarButton>
        <ToolbarButton onClick={() => editor?.chain().focus().toggleHeading({ level: 2 }).run()} active={editor?.isActive("heading", { level: 2 })}><Heading2 size={16} /></ToolbarButton>
        <span className="w-px h-4 bg-neutral-300 dark:bg-neutral-700 mx-1" />
        <ToolbarButton onClick={() => editor?.chain().focus().toggleBulletList().run()} active={editor?.isActive("bulletList")}><List size={16} /></ToolbarButton>
        <ToolbarButton onClick={() => editor?.chain().focus().toggleOrderedList().run()} active={editor?.isActive("orderedList")}><ListOrdered size={16} /></ToolbarButton>
        <ToolbarButton onClick={() => editor?.chain().focus().toggleBlockquote().run()} active={editor?.isActive("blockquote")}><Quote size={16} /></ToolbarButton>
        <ToolbarButton onClick={() => editor?.chain().focus().toggleCodeBlock().run()} active={editor?.isActive("codeBlock")}><Code size={16} /></ToolbarButton>
        <span className="w-px h-4 bg-neutral-300 dark:bg-neutral-700 mx-1" />
        <ToolbarButton onClick={() => { const url = prompt("URL:"); if (url) editor?.chain().focus().setLink({ href: url }).run() }} active={editor?.isActive("link")}><Link2 size={16} /></ToolbarButton>
        <ToolbarButton onClick={() => { const url = prompt("Image URL:"); if (url) editor?.chain().focus().setImage({ src: url }).run() }}><Image size={16} /></ToolbarButton>
      </div>

      <EditorContent editor={editor} className="min-h-[300px] border border-neutral-200 dark:border-neutral-800 rounded-lg p-6 mb-4" />

      {/* Tags */}
      <div className="flex items-center gap-2 mb-6">
        {tags.map((t) => <Badge key={t}><span className="cursor-pointer mr-1" onClick={() => setTags(tags.filter((x) => x !== t))}>×</span>{t}</Badge>)}
        <input className="text-sm border-none bg-transparent focus:outline-none w-24" value={tagInput} onChange={(e) => setTagInput(e.target.value)} onKeyDown={(e) => { if (e.key === "Enter") { e.preventDefault(); addTag() } }} placeholder="Add tag..." />
      </div>

      {error && <p className="text-sm text-red-500 mb-4">{error}</p>}

      <div className="flex gap-3">
        <Button onClick={handlePublish} disabled={publishing}>{publishing ? "Publishing..." : "Publish"}</Button>
        <Button variant="ghost" onClick={() => router.back()}>Cancel</Button>
      </div>
    </div>
  )
}
