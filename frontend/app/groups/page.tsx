'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import { apiFetch } from '../../src/lib/api'

export default function GroupsPage() {
  const [groups, setGroups] = useState<any[]>([])
  const [message, setMessage] = useState('')

  useEffect(() => {
    load()
  }, [])

  async function load() {
    try {
      const data = await apiFetch('/api/groups')
      setGroups(data.groups || [])
    } catch (err: any) {
      setMessage(err.message)
    }
  }

  async function onCreate(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    const form = new FormData(e.currentTarget)
    setMessage('')
    try {
      await apiFetch('/api/groups', {
        method: 'POST',
        body: JSON.stringify({
          title: form.get('title'),
          description: form.get('description')
        })
      })
      e.currentTarget.reset()
      load()
    } catch (err: any) {
      setMessage(err.message)
    }
  }

  return (
    <section className="card">
      <h1>Groups</h1>
      <form onSubmit={onCreate}>
        <input name="title" placeholder="Titre" required />
        <input name="description" placeholder="Description" />
        <button type="submit">Créer un groupe</button>
      </form>
      {message && <p>{message}</p>}
      {groups.map((group) => (
        <div key={group.id}>
          <Link href={`/groups/${group.id}`}>{group.title}</Link>
        </div>
      ))}
    </section>
  )
}
