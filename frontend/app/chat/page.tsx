'use client'

import { useEffect, useRef, useState } from 'react'
import { wsURL } from '../../src/lib/api'

interface Message {
  type: string
  text?: string
  from_user_id?: number
  group_id?: number
}

export default function ChatPage() {
  const [messages, setMessages] = useState<Message[]>([])
  const [status, setStatus] = useState('')
  const wsRef = useRef<WebSocket | null>(null)

  useEffect(() => {
    const ws = new WebSocket(wsURL())
    wsRef.current = ws
    ws.onopen = () => setStatus('connected')
    ws.onclose = () => setStatus('disconnected')
    ws.onerror = () => setStatus('error')
    ws.onmessage = (evt) => {
      const data = JSON.parse(evt.data)
      setMessages((prev) => [data, ...prev])
    }
    return () => {
      ws.close()
    }
  }, [])

  function sendDM(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    const form = new FormData(e.currentTarget)
    const payload = {
      type: 'dm_send',
      to_user_id: Number(form.get('to_user_id')),
      text: String(form.get('text'))
    }
    wsRef.current?.send(JSON.stringify(payload))
    e.currentTarget.reset()
  }

  function sendGroup(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    const form = new FormData(e.currentTarget)
    const payload = {
      type: 'group_send',
      group_id: Number(form.get('group_id')),
      text: String(form.get('text'))
    }
    wsRef.current?.send(JSON.stringify(payload))
    e.currentTarget.reset()
  }

  return (
    <section>
      <div className="card">
        <h1>Chat</h1>
        <p>Status: {status}</p>
        <form onSubmit={sendDM}>
          <input name="to_user_id" placeholder="To user id" required />
          <input name="text" placeholder="Message" required />
          <button type="submit">Send DM</button>
        </form>
        <form onSubmit={sendGroup}>
          <input name="group_id" placeholder="Group id" required />
          <input name="text" placeholder="Message" required />
          <button type="submit">Send Group</button>
        </form>
      </div>

      {messages.map((msg, idx) => (
        <div key={idx} className="card">
          <p>{msg.type}</p>
          <p>{msg.text}</p>
        </div>
      ))}
    </section>
  )
}
