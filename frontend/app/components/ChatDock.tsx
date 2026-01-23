'use client'

import { useEffect, useMemo, useRef, useState } from 'react'
import { apiFetch, wsURL } from '../../src/lib/api'

type User = {
  id: number
  first_name: string
  last_name: string
}

type DMMessage = {
  type: string
  text?: string
  from_user_id?: number
  to_user_id?: number
  created_at?: string
}

export default function ChatDock() {
  const [me, setMe] = useState<User | null>(null)
  const [people, setPeople] = useState<User[]>([])
  const [openConvos, setOpenConvos] = useState<number[]>([])
  const [messages, setMessages] = useState<Record<number, DMMessage[]>>({})
  const [dockCollapsed, setDockCollapsed] = useState(false)
  const wsRef = useRef<WebSocket | null>(null)

  useEffect(() => {
    async function load() {
      try {
        const meRes = await apiFetch('/api/me')
        setMe(meRes)
        const followersRes = await apiFetch(`/api/users/${meRes.id}/followers`)
        const followingRes = await apiFetch(`/api/users/${meRes.id}/following`)
        const list: User[] = [...(followersRes.users || []), ...(followingRes.users || [])]
        const uniq = new Map<number, User>()
        list.forEach((u) => uniq.set(u.id, u))
        setPeople(Array.from(uniq.values()))
      } catch {
        setMe(null)
        setPeople([])
      }
    }
    load()
  }, [])

  useEffect(() => {
    if (!me) return
    const ws = new WebSocket(wsURL())
    wsRef.current = ws
    ws.onmessage = (evt) => {
      const data = JSON.parse(evt.data)
      if (data.type !== 'dm_new') return
      const from = data.from_user_id
      const to = data.to_user_id
      const other = from === me.id ? to : from
      if (!other) return
      setMessages((prev) => {
        const current = prev[other] || []
        return { ...prev, [other]: [data, ...current] }
      })
      setOpenConvos((prev) => (prev.includes(other) ? prev : [other, ...prev]))
    }
    return () => {
      ws.close()
    }
  }, [me])

  const peopleById = useMemo(() => {
    const map = new Map<number, User>()
    people.forEach((p) => map.set(p.id, p))
    return map
  }, [people])

  function openConversation(id: number) {
    setOpenConvos((prev) => (prev.includes(id) ? prev : [id, ...prev]))
  }

  function closeConversation(id: number) {
    setOpenConvos((prev) => prev.filter((x) => x !== id))
  }

  function sendMessage(toUserId: number, text: string) {
    wsRef.current?.send(JSON.stringify({ type: 'dm_send', to_user_id: toUserId, text }))
  }

  if (!me) return null

  return (
    <div className="chat-dock">
      <div className={`chat-dock__panel ${dockCollapsed ? 'chat-dock__panel--collapsed' : ''}`}>
        <button
          type="button"
          className="chat-dock__header"
          onClick={() => setDockCollapsed((v) => !v)}
        >
          <span>Messagerie</span>
          <span>{dockCollapsed ? '▴' : '▾'}</span>
        </button>
        {!dockCollapsed && (
          <div className="chat-dock__list">
            {people.map((u) => (
              <button key={u.id} type="button" onClick={() => openConversation(u.id)} className="chat-dock__item">
                {u.first_name} {u.last_name}
              </button>
            ))}
            {people.length === 0 && <div className="chat-dock__empty">Aucun contact</div>}
          </div>
        )}
      </div>

      {openConvos.map((id) => {
        const user = peopleById.get(id)
        return (
          <ChatWindow
            key={id}
            user={user}
            messages={messages[id] || []}
            onClose={() => closeConversation(id)}
            onSend={(text) => sendMessage(id, text)}
          />
        )
      })}
    </div>
  )
}

function ChatWindow({
  user,
  messages,
  onClose,
  onSend
}: {
  user?: User
  messages: DMMessage[]
  onClose: () => void
  onSend: (text: string) => void
}) {
  const [text, setText] = useState('')

  function submit(e: React.FormEvent) {
    e.preventDefault()
    const value = text.trim()
    if (!value) return
    onSend(value)
    setText('')
  }

  return (
    <div className="chat-dock__panel chat-dock__panel--window">
      <div className="chat-dock__header">
        <span>{user ? `${user.first_name} ${user.last_name}` : 'Conversation'}</span>
        <button type="button" onClick={onClose} className="chat-dock__close">×</button>
      </div>
      <div className="chat-dock__messages">
        {messages.length === 0 && <div className="chat-dock__empty">Aucun message</div>}
        {messages.map((m, idx) => (
          <div key={idx} className="chat-dock__message">{m.text}</div>
        ))}
      </div>
      <form onSubmit={submit} className="chat-dock__input">
        <input value={text} onChange={(e) => setText(e.target.value)} placeholder="Écrire..." />
      </form>
    </div>
  )
}
