'use client'

import { useEffect, useState } from 'react'
import { apiFetch } from '../../src/lib/api'

interface Post {
  id: number
  user_id: number
  text: string
  visibility: string
  media_path?: string
  created_at: string
}

export default function FeedPage() {
  const [posts, setPosts] = useState<Post[]>([])
  const [message, setMessage] = useState('')

  async function load() {
    try {
      const data = await apiFetch('/api/feed')
      setPosts(data.posts || [])
    } catch (err: any) {
      setMessage(err.message)
    }
  }

  async function onCreate(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    const form = new FormData(e.currentTarget)
    try {
      await apiFetch('/api/posts', {
        method: 'POST',
        body: JSON.stringify({
          text: form.get('text'),
          visibility: form.get('visibility')
        })
      })
      e.currentTarget.reset()
      load()
    } catch (err: any) {
      setMessage(err.message)
    }
  }

  useEffect(() => {
    load()
  }, [])

  return (
    <section>
      <div className="card">
        <h1>Feed</h1>
        <form onSubmit={onCreate}>
          <textarea name="text" placeholder="Votre post" required />
          <select name="visibility" defaultValue="public">
            <option value="public">Public</option>
            <option value="followers">Followers</option>
            <option value="private">Private</option>
          </select>
          <button type="submit">Publier</button>
        </form>
        {message && <p>{message}</p>}
      </div>

      {posts.map((post) => (
        <div key={post.id} className="card">
          <p>#{post.id} par {post.user_id}</p>
          <p>{post.text}</p>
          <small>{post.visibility} - {post.created_at}</small>
        </div>
      ))}
    </section>
  )
}
