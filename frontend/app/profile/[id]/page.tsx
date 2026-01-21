'use client'

import { useEffect, useState } from 'react'
import { apiFetch } from '../../../src/lib/api'

export default function ProfilePage({ params }: { params: { id: string } }) {
  const [user, setUser] = useState<any>(null)
  const [posts, setPosts] = useState<any[]>([])
  const [message, setMessage] = useState('')

  useEffect(() => {
    async function load() {
      try {
        const data = await apiFetch(`/api/users/${params.id}`)
        setUser(data)
        const postsData = await apiFetch(`/api/users/${params.id}/posts`)
        setPosts(postsData.posts || [])
      } catch (err: any) {
        setMessage(err.message)
      }
    }
    load()
  }, [params.id])

  return (
    <section className="card">
      <h1>Profile</h1>
      {message && <p>{message}</p>}
      {user && (
        <div>
          <p>{user.first_name} {user.last_name}</p>
          <p>{user.email}</p>
          <p>Public: {String(user.is_public)}</p>
        </div>
      )}
      <h2>Posts</h2>
      {posts.map((post) => (
        <div key={post.id} className="card">
          <p>{post.text}</p>
        </div>
      ))}
    </section>
  )
}
