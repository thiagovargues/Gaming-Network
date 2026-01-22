'use client'

import { useEffect, useState } from 'react'
import { apiFetch } from '../../src/lib/api'

export default function NotificationsPage() {
  const [items, setItems] = useState<any[]>([])
  const [message, setMessage] = useState('')

  async function load() {
    try {
      const data = await apiFetch('/api/notifications')
      setItems(data.notifications || [])
    } catch (err: any) {
      setMessage(err.message)
    }
  }

  useEffect(() => {
    load()
  }, [])

  return (
    <section className="card">
      <h1>Notifications</h1>
      {message && <p>{message}</p>}
      {items.map((item) => (
        <div key={item.id} className="card">
          <p>{item.type}</p>
          <small>{item.created_at}</small>
        </div>
      ))}
    </section>
  )
}
