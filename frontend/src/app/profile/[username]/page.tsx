"use client"

import { useEffect, useState } from "react"
import { useParams } from "next/navigation"
import Link from "next/link"
import { getUserProfile, getUserStats } from "@/lib/api"
import { Card } from "@/components/ui"
import { formatDate } from "@/lib/utils"

interface Profile {
  id: string
  username: string
  display_name: string
  bio?: string
  email_notify_comments: boolean
  article_count: number
  created_at: string
}

interface Stats {
  article_count: number
  total_claps: number
  total_comments: number
  total_views: number
}

export default function ProfilePage() {
  const { username } = useParams<{ username: string }>()
  const [profile, setProfile] = useState<Profile | null>(null)
  const [stats, setStats] = useState<Stats | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")

  useEffect(() => {
    if (!username) return
    setLoading(true)
    Promise.all([
      getUserProfile(username as string),
      getUserProfile(username as string).then((p) => getUserStats(p.id)),
    ])
      .then(([profileData, statsData]) => {
        setProfile(profileData)
        if (statsData.data) setStats(statsData.data)
      })
      .catch(() => setError("User not found"))
      .finally(() => setLoading(false))
  }, [username])

  if (loading) return <div className="max-w-2xl mx-auto px-4 py-16"><ProfileSkeleton /></div>
  if (error) return <div className="max-w-2xl mx-auto px-4 py-16 text-center text-red-500">{error}</div>
  if (!profile) return null

  return (
    <div className="max-w-2xl mx-auto px-4 py-16">
      <div className="flex items-center gap-6 mb-8">
        <div className="w-20 h-20 rounded-full bg-gray-200 dark:bg-gray-700 flex items-center justify-center text-2xl font-bold text-gray-500">
          {(profile.display_name || profile.username)[0].toUpperCase()}
        </div>
        <div>
          <h1 className="text-3xl font-bold">{profile.display_name || profile.username}</h1>
          <p className="text-gray-500 dark:text-gray-400">@{profile.username}</p>
          {profile.bio && <p className="mt-2 text-gray-600 dark:text-gray-300">{profile.bio}</p>}
          <p className="text-xs text-gray-400 mt-1">Joined {formatDate(profile.created_at)}</p>
        </div>
      </div>

      <div className="grid grid-cols-4 gap-4 mb-8">
        <StatCard label="Articles" value={profile.article_count} />
        <StatCard label="Claps" value={stats?.total_claps || 0} />
        <StatCard label="Comments" value={stats?.total_comments || 0} />
        <StatCard label="Views" value={stats?.total_views || 0} />
      </div>

      <Link
        href={`/articles?author=${profile.id}`}
        className="inline-block px-6 py-2 bg-black dark:bg-white text-white dark:text-black rounded-full font-medium hover:opacity-80 transition"
      >
        View Articles →
      </Link>
    </div>
  )
}

function StatCard({ label, value }: { label: string; value: number }) {
  return (
    <Card className="text-center p-4">
      <div className="text-2xl font-bold">{value}</div>
      <div className="text-xs text-gray-500">{label}</div>
    </Card>
  )
}

function ProfileSkeleton() {
  return (
    <div className="animate-pulse">
      <div className="flex items-center gap-6 mb-8">
        <div className="w-20 h-20 rounded-full bg-gray-200 dark:bg-gray-700" />
        <div className="flex-1">
          <div className="h-8 bg-gray-200 dark:bg-gray-700 rounded w-48 mb-2" />
          <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-32" />
        </div>
      </div>
      <div className="grid grid-cols-4 gap-4">
        {[1, 2, 3, 4].map((i) => (
          <div key={i} className="h-20 bg-gray-200 dark:bg-gray-700 rounded" />
        ))}
      </div>
    </div>
  )
}
