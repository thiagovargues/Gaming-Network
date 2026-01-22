'use client'

import Link from 'next/link'
import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { API_URL, apiFetch } from '../../src/lib/api'

type User = {
  id: number
  email: string
  first_name: string
  last_name: string
  nickname?: string | null
  avatar_path?: string | null
  auth_providers?: string[]
}

function GoogleBadge() {
  return (
    <span className="nav__badge" title="Connected via Google" aria-label="Connected via Google">
      <svg viewBox="0 0 24 24" aria-hidden="true" className="nav__badge-icon">
        <path
          fill="#EA4335"
          d="M12 10.2v3.9h5.5c-.2 1.3-1.6 3.8-5.5 3.8-3.3 0-6-2.7-6-5.9s2.7-5.9 6-5.9c1.9 0 3.2.8 3.9 1.5l2.6-2.5C16.9 3.6 14.7 2.5 12 2.5 7 2.5 3 6.5 3 11.9S7 21.3 12 21.3c6 0 9-4.2 9-7.9 0-.5-.1-.9-.2-1.3H12z"
        />
        <path fill="#34A853" d="M3.8 8.4l3.2 2.4C7.8 9.1 9.7 7.5 12 7.5c1.9 0 3.2.8 3.9 1.5l2.6-2.5C16.9 3.6 14.7 2.5 12 2.5c-3.5 0-6.6 2-8.2 5.9z" />
        <path fill="#FBBC05" d="M3 11.9c0 1.1.3 2.1.8 3l3.2-2.4c-.2-.5-.3-1-.3-1.6s.1-1.1.3-1.6L3.8 8.4C3.3 9.3 3 10.5 3 11.9z" />
        <path fill="#4285F4" d="M12 21.3c3.5 0 6.4-1.2 8.5-3.3l-3.7-2.9c-1 .7-2.3 1.2-4.8 1.2-2.9 0-5.3-1.9-6.1-4.5l-3.2 2.4c1.6 3.9 4.7 6.1 9.3 6.1z" />
      </svg>
    </span>
  )
}

export default function Nav() {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)
  const router = useRouter()

  useEffect(() => {
    let cancelled = false
    async function load() {
      try {
        const res = await fetch(`${API_URL}/api/me`, { credentials: 'include' })
        if (!res.ok) {
          if (!cancelled) setUser(null)
          return
        }
        const data = await res.json()
        if (!cancelled) setUser(data)
      } finally {
        if (!cancelled) setLoading(false)
      }
    }
    load()
    return () => {
      cancelled = true
    }
  }, [])

  async function onLogout() {
    try {
      await apiFetch('/api/auth/logout', { method: 'POST' })
    } finally {
      setUser(null)
      router.push('/login')
      router.refresh()
    }
  }

  const displayName = user?.nickname || [user?.first_name, user?.last_name].filter(Boolean).join(' ') || user?.email
  const isGoogle = user?.auth_providers?.includes('google')

  return (
    <nav className="nav">
      <Link href="/">Home</Link>
      <Link href="/feed">Feed</Link>
      <Link href="/groups">Groups</Link>
      <Link href="/chat">Chat</Link>
      <Link href="/notifications">Notifications</Link>
      {!loading && user ? (
        <div className="nav__user">
          <Link href={`/profile/${user.id}`} className="nav__user-link">
            Logged in as: {displayName}
          </Link>
          {isGoogle && <GoogleBadge />}
          <button type="button" onClick={onLogout}>Logout</button>
        </div>
      ) : (
        <>
          <Link href="/register">Register</Link>
          <Link href="/login">Login</Link>
        </>
      )}
    </nav>
  )
}
