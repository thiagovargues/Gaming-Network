'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import { apiFetch } from '../../src/lib/api'

export default function GroupsPage() {
  const [groups, setGroups] = useState<any[]>([])
  const [message, setMessage] = useState('')

  useEffect(() => {
    apiFetch('/api/groups')
      .then((data) => setGroups(data.groups || []))
      .catch((err) => setMessage(err.message))
  }, [])

  return (
    <section className="card">
      <h1>Groups</h1>
      {message && <p>{message}</p>}
      {groups.map((group) => (
        <div key={group.id}>
          <Link href={`/groups/${group.id}`}>{group.title}</Link>
        </div>
      ))}
    </section>
  )
}
