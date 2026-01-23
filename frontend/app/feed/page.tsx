'use client'

import { useEffect, useMemo, useState } from 'react'
import { apiFetch, uploadMedia } from '../../src/lib/api'
import PostCard from '../components/PostCard'

interface Post {
  id: number
  user_id: number
  text: string
  visibility: string
  media_path?: string | null
  created_at: string
}

type User = {
  id: number
  email: string
  first_name: string
  last_name: string
}

export default function FeedPage() {
  const [posts, setPosts] = useState<Post[]>([])
  const [message, setMessage] = useState('')
  const [me, setMe] = useState<User | null>(null)
  const [followers, setFollowers] = useState<User[]>([])
  const [isPublic, setIsPublic] = useState(true)

  async function load() {
    try {
      const data = await apiFetch('/api/feed')
      setPosts(data.posts || [])
    } catch (err: any) {
      setMessage(err.message)
    }
  }

  async function loadMe() {
    try {
      const user = await apiFetch('/api/me')
      setMe(user)
      const followersRes = await apiFetch(`/api/users/${user.id}/followers`)
      setFollowers(followersRes.users || [])
    } catch (err) {
      setMe(null)
      setFollowers([])
    }
  }

  async function onCreate(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    const form = new FormData(e.currentTarget)
    setMessage('')
    try {
      const file = form.get('file') as File | null
      let mediaPath: string | undefined
      if (file && file.size > 0) {
        mediaPath = await uploadMedia(file)
      }
      const allowedFollowerIDs = form.getAll('allowed_follower_ids').map((v) => Number(v)).filter((v) => v > 0)
      await apiFetch('/api/posts', {
        method: 'POST',
        body: JSON.stringify({
          text: form.get('text'),
          visibility: isPublic ? 'public' : 'private',
          allowed_follower_ids: allowedFollowerIDs,
          media_path: mediaPath
        })
      })
      e.currentTarget.reset()
      setIsPublic(true)
      load()
    } catch (err: any) {
      setMessage(err.message)
    }
  }

  useEffect(() => {
    load()
    loadMe()
  }, [])

  const followersOptions = useMemo(() => followers.map((f) => (
    <label key={f.id} style={{ display: 'block' }}>
      <input type="checkbox" name="allowed_follower_ids" value={String(f.id)} />
      {f.first_name} {f.last_name} ({f.email})
    </label>
  )), [followers])

  return (
    <section className="feed-layout">
      <div className="card">
        <div className="feed-header">
          <h1>Feed</h1>
          <div className="privacy-row">
            <span>{isPublic ? 'Public' : 'Privé'}</span>
            <button
              type="button"
              className={`switch ${isPublic ? 'on' : 'off'}`}
              onClick={() => setIsPublic((v) => !v)}
              aria-pressed={isPublic}
            >
              <span className="switch-thumb" />
            </button>
          </div>
        </div>
        <form onSubmit={onCreate}>
          <textarea name="text" placeholder="Votre post" required className="feed-composer" />
          <div className="feed-actions">
            <label className="file-trigger">
              Ajouter une image
              <input name="file" type="file" accept="image/png,image/jpeg,image/gif" />
            </label>
            <button type="submit">Publier</button>
          </div>
          {!isPublic && (
            <div>
              <p>Choisir les followers autorisés :</p>
              {followersOptions.length > 0 ? followersOptions : <p>Aucun follower</p>}
            </div>
          )}
        </form>
        {message && <p>{message}</p>}
      </div>

      {posts.map((post) => (
        <PostCard key={post.id} post={post} currentUserId={me?.id} onDeleted={load} />
      ))}
    </section>
  )
}
