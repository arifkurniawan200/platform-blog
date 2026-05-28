import type { Metadata } from "next"
import { Navbar } from "@/components/ui"
import "./globals.css"

export const metadata: Metadata = {
  title: "Platform Blog",
  description: "A modern blogging platform",
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className="min-h-screen bg-neutral-50 dark:bg-neutral-950 text-neutral-900 dark:text-neutral-100 font-sans antialiased">
        <Navbar />
        <main className="mx-auto max-w-6xl px-4 py-8">{children}</main>
      </body>
    </html>
  )
}
