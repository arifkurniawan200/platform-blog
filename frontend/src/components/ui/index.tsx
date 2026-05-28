"use client"

import Link from "next/link"
import { useAuthStore } from "@/stores/auth"
import { cn } from "@/lib/utils"

export function Button({ className, variant = "default", size = "md", ...props }: any) {
  const base = "inline-flex items-center justify-center rounded-lg font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary disabled:pointer-events-none disabled:opacity-50"
  const variants: Record<string, string> = {
    default: "bg-neutral-900 text-white hover:bg-neutral-800 dark:bg-white dark:text-neutral-900 dark:hover:bg-neutral-100",
    ghost: "hover:bg-neutral-100 dark:hover:bg-neutral-800",
    outline: "border border-neutral-200 dark:border-neutral-700 hover:bg-neutral-50 dark:hover:bg-neutral-800",
  }
  const sizes: Record<string, string> = { sm: "h-8 px-3 text-sm", md: "h-10 px-4 text-sm", lg: "h-12 px-6 text-base" }
  return <button className={cn(base, variants[variant], sizes[size], className)} {...props} />
}

export function Input({ className, error, ...props }: any) {
  return (
    <div>
      <input className={cn("w-full h-10 px-3 rounded-lg border border-neutral-200 dark:border-neutral-700 bg-white dark:bg-neutral-900 text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900 dark:focus:ring-white transition-colors", error && "border-red-500", className)} {...props} />
      {error && <p className="mt-1 text-xs text-red-500">{error}</p>}
    </div>
  )
}

export function Card({ className, children, ...props }: any) {
  return <div className={cn("rounded-xl border border-neutral-200 dark:border-neutral-800 bg-white dark:bg-neutral-950 p-6", className)} {...props}>{children}</div>
}

export function Badge({ className, children }: any) {
  return <span className={cn("inline-flex items-center rounded-full bg-neutral-100 dark:bg-neutral-800 px-2.5 py-0.5 text-xs font-medium text-neutral-600 dark:text-neutral-300", className)}>{children}</span>
}

export function Navbar() {
  const { token, logout } = useAuthStore()
  return (
    <header className="sticky top-0 z-50 border-b border-neutral-200 dark:border-neutral-800 bg-white/80 dark:bg-neutral-950/80 backdrop-blur">
      <div className="max-w-6xl mx-auto flex items-center justify-between h-14 px-4">
        <div className="flex items-center gap-4">
          <Link href="/" className="font-bold text-lg tracking-tight">Platform Blog</Link>
          <Link href="/search" className="text-sm text-neutral-500 hover:text-neutral-900 dark:hover:text-white transition-colors">🔍 Search</Link>
        </div>
        <nav className="flex items-center gap-3">
          {token ? (
            <>
              <Link href="/write" className="text-sm text-neutral-600 dark:text-neutral-400 hover:text-neutral-900 dark:hover:text-white transition-colors">Write</Link>
              <Link href="/bookmarks" className="text-sm text-neutral-600 dark:text-neutral-400 hover:text-neutral-900 dark:hover:text-white transition-colors">Bookmarks</Link>
              <Button variant="ghost" size="sm" onClick={logout}>Logout</Button>
            </>
          ) : (
            <>
              <Link href="/login" className="text-sm text-neutral-600 dark:text-neutral-400 hover:text-neutral-900 dark:hover:text-white transition-colors">Login</Link>
              <Link href="/register"><Button size="sm">Get Started</Button></Link>
            </>
          )}
        </nav>
      </div>
    </header>
  )
}
