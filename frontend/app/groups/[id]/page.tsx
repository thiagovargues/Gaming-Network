'use client'

import { useEffect, useState } from 'react'
import { apiFetch, uploadMedia } from '../../../src/lib/api'
import PostCard from '../../components/PostCard'

type Event = {
  id: number
  title: string
  description: string
  datetime: string
}

export default function GroupPage({ params }: { params: { id: string } }) {
  const [group, setGroup] = useState<any>(null)
  const [posts, setPosts] = useState<any[]>([])
  const [members, setMembers] = useState<any[]>([])
  const [events, setEvents] = useState<Event[]>([])
  const [isMember, setIsMember] = useState(false)
  const [me, setMe] = useState<any>(null)
  const [message, setMessage] = useState('')

  async function load() {
    try {
      const data = await apiFetch(`/api/groups/${params.id}`)
      setGroup(data)
    } catch (err: any) {
      setMessage(err.message)
      return
    }

    try {
      const membersRes = await apiFetch(`/api/groups/${params.id}/members`)
      setMembers(membersRes.users || [])
      setIsMember(true)
      const postsData = await apiFetch(`/api/groups/${params.id}/posts`)
      setPosts(postsData.posts || [])
      const eventsData = await apiFetch(`/api/groups/${params.id}/events`)
      setEvents(eventsData.events || [])
    } catch (err: any) {
      setIsMember(false)
      setMembers([])
      setPosts([])
      setEvents([])
    }
    try {
      const meRes = await apiFetch('/api/me')
      setMe(meRes)
    } catch {
      setMe(null)
    }
  }

  useEffect(() => {
    load()
  }, [params.id])

  async function onJoinRequest() {
    setMessage('')
    try {
      await apiFetch(`/api/groups/${params.id}/join-request`, { method: 'POST' })
      setMessage('Demande envoyée')
    } catch (err: any) {
      setMessage(err.message)
    }
  }

  async function onInvite(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    const form = new FormData(e.currentTarget)
    setMessage('')
    try {
      await apiFetch(`/api/groups/${params.id}/invite`, {
        method: 'POST',
        body: JSON.stringify({ user_id: Number(form.get('user_id')) })
      })
      e.currentTarget.reset()
      setMessage('Invitation envoyée')
    } catch (err: any) {
      setMessage(err.message)
    }
  }

  async function onCreatePost(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    const form = new FormData(e.currentTarget)
    setMessage('')
    try {
      const file = form.get('file') as File | null
      let mediaPath: string | undefined
      if (file && file.size > 0) {
        mediaPath = await uploadMedia(file)
      }
      await apiFetch(`/api/groups/${params.id}/posts`, {
        method: 'POST',
        body: JSON.stringify({ text: form.get('text'), media_path: mediaPath })
      })
      e.currentTarget.reset()
      load()
    } catch (err: any) {
      setMessage(err.message)
    }
  }

  async function onCreateEvent(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    const form = new FormData(e.currentTarget)
    setMessage('')
    try {
      await apiFetch(`/api/groups/${params.id}/events`, {
        method: 'POST',
        body: JSON.stringify({
          title: form.get('title'),
          description: form.get('description'),
          datetime: form.get('datetime')
        })
      })
      e.currentTarget.reset()
      load()
    } catch (err: any) {
      setMessage(err.message)
    }
  }

  async function onRespondEvent(eventID: number, status: 'going' | 'not_going') {
    setMessage('')
    try {
      await apiFetch(`/api/events/${eventID}/respond`, {
        method: 'POST',
        body: JSON.stringify({ status })
      })
      setMessage(`Réponse enregistrée: ${status}`)
    } catch (err: any) {
      setMessage(err.message)
    }
  }

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

      {!isMember && (
        <button type="button" onClick={onJoinRequest}>Demander à rejoindre</button>
      )}

      {isMember && (
        <div className="card">
          <h2>Inviter un utilisateur</h2>
          <form onSubmit={onInvite}>
            <input name="user_id" placeholder="User ID" required />
            <button type="submit">Inviter</button>
          </form>
        </div>
      )}

      {isMember && (
        <div className="card">
          <h2>Créer un post</h2>
          <form onSubmit={onCreatePost}>
            <textarea name="text" placeholder="Votre post" required />
            <input name="file" type="file" accept="image/png,image/jpeg,image/gif" />
            <button type="submit">Publier</button>
          </form>
        </div>
      )}

      <h2>Membres</h2>
      {members.map((m) => (
        <div key={m.id}>{m.first_name} {m.last_name}</div>
      ))}

      {isMember && (
        <>
          <h2>Posts</h2>
          {posts.map((post) => (
            <PostCard key={post.id} post={post} currentUserId={me?.id} onDeleted={load} />
          ))}

          <h2>Événements</h2>
          <div className="card">
            <form onSubmit={onCreateEvent}>
              <input name="title" placeholder="Titre" required />
              <input name="description" placeholder="Description" />
              <input name="datetime" type="datetime-local" required />
              <button type="submit">Créer un événement</button>
            </form>
          </div>
          {events.map((ev) => (
            <div key={ev.id} className="card">
              <p>{ev.title}</p>
              <p>{ev.description}</p>
              <small>{ev.datetime}</small>
              <div>
                <button type="button" onClick={() => onRespondEvent(ev.id, 'going')}>Going</button>
                <button type="button" onClick={() => onRespondEvent(ev.id, 'not_going')}>Not going</button>
              </div>
            </div>
          ))}
        </>
      )}
    </section>
  )
}
