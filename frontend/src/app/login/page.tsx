"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"
import Link from "next/link"
import { authLogin } from "@/lib/api"
import { useAuthStore } from "@/stores/auth"
import { Button, Input, Card } from "@/components/ui"

export default function LoginPage() {
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const [error, setError] = useState("")
  const [loading, setLoading] = useState(false)
  const router = useRouter()
  const setTokens = useAuthStore((s) => s.setTokens)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!email || !password) { setError("All fields are required"); return }
    setLoading(true)
    setError("")
    try {
      const res = await authLogin({ email, password })
      setTokens(res.data.access_token, res.data.refresh_token)
      router.push("/")
    } catch (err: any) {
      setError(err.message || "Invalid credentials")
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="max-w-md mx-auto pt-16">
      <Card>
        <h1 className="text-2xl font-bold mb-6">Welcome back</h1>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="text-sm font-medium mb-1.5 block">Email</label>
            <Input type="email" value={email} onChange={(e: any) => setEmail(e.target.value)} placeholder="john@example.com" />
          </div>
          <div>
            <label className="text-sm font-medium mb-1.5 block">Password</label>
            <Input type="password" value={password} onChange={(e: any) => setPassword(e.target.value)} placeholder="••••••••" />
          </div>
          {error && <p className="text-sm text-red-500">{error}</p>}
          <Button type="submit" className="w-full" disabled={loading}>{loading ? "Logging in..." : "Login"}</Button>
        </form>
        <p className="mt-4 text-sm text-center text-neutral-500">Don&apos;t have an account? <Link href="/register" className="font-medium text-neutral-900 dark:text-white hover:underline">Register</Link></p>
      </Card>
    </div>
  )
}
