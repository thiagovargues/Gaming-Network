'use client'

import { useEffect, useState } from 'react'
import { apiFetch } from '../../../src/lib/api'

export default function GroupPage({ params }: { params: { id: string } }) {
  const [group, setGroup] = useState<any>(null)
  const [posts, setPosts] = useState<any[]>([])
  const [message, setMessage] = useState('')

  useEffect(() => {
    async function load() {
      try {
        const data = await apiFetch(`/api/groups/${params.id}`)
        setGroup(data)
        const postsData = await apiFetch(`/api/groups/${params.id}/posts`)
        setPosts(postsData.posts || [])
      } catch (err: any) {
        setMessage(err.message)
      }
    }
    load()
  }, [params.id])

  return (
    <section className="card">
      <h1>Group</h1>
      {message && <p>{message}</p>}
      {group && (
        <div>
          <p>{group.title}</p>
          <p>{group.description}</p>
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
