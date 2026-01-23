'use client'

import { useEffect, useMemo, useState } from 'react'
import { API_URL, apiFetch, uploadMedia } from '../../../src/lib/api'
import PostCard from '../../components/PostCard'

type User = {
  id: number
  email: string
  first_name: string
  last_name: string
  dob?: string
  avatar_path?: string | null
  nickname?: string | null
  about?: string | null
  age?: number | null
  show_first_name?: boolean
  show_last_name?: boolean
  show_age?: boolean
  is_public: boolean
}

type FollowRequest = {
  id: number
  from_user_id: number
  to_user_id: number
  status: string
}

export default function ProfilePage({ params }: { params: { id: string } }) {
  const [user, setUser] = useState<User | null>(null)
  const [posts, setPosts] = useState<any[]>([])
  const [followers, setFollowers] = useState<User[]>([])
  const [following, setFollowing] = useState<User[]>([])
  const [me, setMe] = useState<User | null>(null)
  const [outgoing, setOutgoing] = useState<FollowRequest[]>([])
  const [message, setMessage] = useState('')
  
  const [avatarError, setAvatarError] = useState(false)

  async function load() {
    try {
      const data = await apiFetch(`/api/users/${params.id}`)
      setUser(data)
      const postsData = await apiFetch(`/api/users/${params.id}/posts`)
      setPosts(postsData.posts || [])
      const followersRes = await apiFetch(`/api/users/${params.id}/followers`)
      setFollowers(followersRes.users || [])
      const followingRes = await apiFetch(`/api/users/${params.id}/following`)
      setFollowing(followingRes.users || [])
    } catch (err: any) {
      setMessage(err.message)
    }
  }

  async function loadMe() {
    try {
      const meRes = await apiFetch('/api/me')
      setMe(meRes)
      const outRes = await apiFetch('/api/follows/requests/outgoing')
      setOutgoing(outRes.requests || [])
    } catch {
      setMe(null)
      setOutgoing([])
    }
  }

  useEffect(() => {
    load()
    loadMe()
  }, [params.id])

  useEffect(() => {
    setAvatarError(false)
  }, [user?.avatar_path])

  const isMe = me?.id === user?.id
  const isGoogleMe = isMe && (me?.auth_providers || []).includes('google')
  const isFollowing = useMemo(() => {
    if (!me || !user) return false
    return followers.some((f) => f.id === me.id)
  }, [me, user, followers])

  const isPending = useMemo(() => {
    if (!me || !user) return false
    return outgoing.some((r) => r.to_user_id === user.id)
  }, [me, user, outgoing])

  async function onFollow() {
    if (!user) return
    setMessage('')
    try {
      await apiFetch('/api/follows/request', {
        method: 'POST',
        body: JSON.stringify({ to_user_id: user.id })
      })
      await load()
      await loadMe()
    } catch (err: any) {
      setMessage(err.message)
    }
  }

  async function onUnfollow() {
    if (!user) return
    if (!confirm('Confirmer l’unfollow ?')) return
    setMessage('')
    try {
      await apiFetch(`/api/follows/${user.id}`, { method: 'DELETE' })
      await load()
    } catch (err: any) {
      setMessage(err.message)
    }
  }

  async function onUpdateProfile(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    const form = new FormData(e.currentTarget)
    setMessage('')
    try {
      const file = form.get('avatar') as File | null
      let avatarPath: string | undefined
      if (file && file.size > 0) {
        avatarPath = await uploadMedia(file)
      }
      const ageRaw = String(form.get('age') || '').trim()
      const ageValue = ageRaw === '' ? undefined : Number(ageRaw)
      if (ageValue !== undefined && Number.isNaN(ageValue)) {
        setMessage('Age invalide')
        return
      }
      await apiFetch('/api/users/me', {
        method: 'PATCH',
        body: JSON.stringify({
          is_public: form.get('is_public') === 'on',
          nickname: form.get('nickname') || undefined,
          about: form.get('about') || undefined,
          avatar: avatarPath,
          first_name: form.get('first_name') || undefined,
          last_name: form.get('last_name') || undefined,
          age: ageValue
        })
      })
      await load()
      await loadMe()
      setShowAvatarModal(false)
    } catch (err: any) {
      setMessage(err.message)
    }
  }

  

  return (
    <section className="profile-layout">
      <aside className="card profile-sidebar">
        <div className="profile-header">
          {isMe && (
            <a
              href="/settings"
              className="icon-button"
              title="Edit profile"
              aria-label="Edit profile"
            >
              ⚙
            </a>
          )}
          <div>
            <h1>
              {user
                ? user.nickname ||
                  (user.show_first_name === false && !isMe ? 'Profil' : user.first_name) ||
                  'Profil'
                : 'Profil'}
          {user && user.show_last_name && (
            <span className="profile-lastname">{user.last_name || '—'}</span>
          )}
            </h1>
          </div>
        </div>
        {message && <p>{message}</p>}
        {user && (
          <div>
            <div className="avatar-block">
              {user.avatar_path && !avatarError ? (
                <img
                  src={`${API_URL}/${user.avatar_path}`}
                  alt=""
                  aria-label="avatar"
                  className="avatar-img"
                  onError={() => setAvatarError(true)}
                />
              ) : (
                <div className="avatar-placeholder" aria-label="avatar placeholder">
                  <svg viewBox="0 0 120 120" role="img" aria-hidden="true">
                    <circle cx="60" cy="48" r="20" fill="#bdbdbd" />
                    <path d="M20 105c10-22 28-34 40-34s30 12 40 34" fill="#bdbdbd" />
                  </svg>
                </div>
              )}
            </div>
            {user.show_nickname && user.nickname && <p>@{user.nickname}</p>}
            {user.show_about && <p>Bio: {user.about || '—'}</p>}
            {user.show_age && <p>Âge: {user.age ?? '—'}</p>}
          </div>
        )}

        {user && me && !isMe && (
          <div>
            {!isFollowing && !isPending && <button type="button" onClick={onFollow}>Follow</button>}
            {isPending && <p>Demande envoyée</p>}
            {isFollowing && <button type="button" onClick={onUnfollow}>Unfollow</button>}
          </div>
        )}

        

        <h2>Followers <span className="count-badge">{followers.length}</span></h2>
        {followers.map((f) => (
          <div key={f.id}>{f.first_name} {f.last_name}</div>
        ))}

        <h2>Following <span className="count-badge">{following.length}</span></h2>
        {following.map((f) => (
          <div key={f.id}>{f.first_name} {f.last_name}</div>
        ))}
      </aside>

      <main className="profile-main">
        <h2>Posts</h2>
        {posts.map((post) => (
          <PostCard key={post.id} post={post} currentUserId={me?.id} onDeleted={load} />
        ))}
      </main>


    </section>
  )
}
