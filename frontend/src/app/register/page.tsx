"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"
import Link from "next/link"
import { authRegister } from "@/lib/api"
import { useAuthStore } from "@/stores/auth"
import { Button, Input, Card } from "@/components/ui"

export default function RegisterPage() {
  const [username, setUsername] = useState("")
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [loading, setLoading] = useState(false)
  const router = useRouter()
  const setTokens = useAuthStore((s) => s.setTokens)

  const validate = () => {
    const errs: Record<string, string> = {}
    if (!username || username.length < 3) errs.username = "Min 3 characters"
    if (!email || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) errs.email = "Valid email required"
    if (!password || password.length < 8) errs.password = "Min 8 characters"
    setErrors(errs)
    return Object.keys(errs).length === 0
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!validate()) return
    setLoading(true)
    try {
      const res = await authRegister({ username, email, password })
      setTokens(res.data.access_token, res.data.refresh_token)
      router.push("/")
    } catch (err: any) {
      setErrors({ form: err.message })
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="max-w-md mx-auto pt-16">
      <Card>
        <h1 className="text-2xl font-bold mb-6">Create an account</h1>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div><label className="text-sm font-medium mb-1.5 block">Username</label><Input value={username} onChange={(e: any) => setUsername(e.target.value)} error={errors.username} placeholder="johndoe" /></div>
          <div><label className="text-sm font-medium mb-1.5 block">Email</label><Input type="email" value={email} onChange={(e: any) => setEmail(e.target.value)} error={errors.email} placeholder="john@example.com" /></div>
          <div><label className="text-sm font-medium mb-1.5 block">Password</label><Input type="password" value={password} onChange={(e: any) => setPassword(e.target.value)} error={errors.password} placeholder="Min 8 characters" /></div>
          {errors.form && <p className="text-sm text-red-500">{errors.form}</p>}
          <Button type="submit" className="w-full" disabled={loading}>{loading ? "Creating..." : "Create Account"}</Button>
        </form>
        <p className="mt-4 text-sm text-center text-neutral-500">Already have an account? <Link href="/login" className="font-medium text-neutral-900 dark:text-white hover:underline">Login</Link></p>
      </Card>
    </div>
  )
}
