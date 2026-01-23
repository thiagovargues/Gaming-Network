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

  async function markAllRead() {
    try {
      await apiFetch('/api/notifications/read-all', { method: 'POST' })
      load()
    } catch (err: any) {
      setMessage(err.message)
    }
  }

  async function markRead(id: number) {
    try {
      await apiFetch(`/api/notifications/${id}/read`, { method: 'POST' })
      load()
    } catch (err: any) {
      setMessage(err.message)
    }
  }

  async function acceptFollow(requestId: number) {
    try {
      await apiFetch(`/api/follows/request/${requestId}/accept`, { method: 'POST' })
      load()
    } catch (err: any) {
      setMessage(err.message)
    }
  }

  async function refuseFollow(requestId: number) {
    try {
      await apiFetch(`/api/follows/request/${requestId}/refuse`, { method: 'POST' })
      load()
    } catch (err: any) {
      setMessage(err.message)
    }
  }

  async function acceptInvite(inviteId: number) {
    try {
      await apiFetch(`/api/groups/invites/${inviteId}/accept`, { method: 'POST' })
      load()
    } catch (err: any) {
      setMessage(err.message)
    }
  }

  async function refuseInvite(inviteId: number) {
    try {
      await apiFetch(`/api/groups/invites/${inviteId}/refuse`, { method: 'POST' })
      load()
    } catch (err: any) {
      setMessage(err.message)
    }
  }

  async function acceptJoinRequest(requestId: number) {
    try {
      await apiFetch(`/api/groups/join-requests/${requestId}/accept`, { method: 'POST' })
      load()
    } catch (err: any) {
      setMessage(err.message)
    }
  }

  async function refuseJoinRequest(requestId: number) {
    try {
      await apiFetch(`/api/groups/join-requests/${requestId}/refuse`, { method: 'POST' })
      load()
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
      <button type="button" onClick={markAllRead}>Tout marquer comme lu</button>
      {message && <p>{message}</p>}
      {items.map((item) => {
        let payload: any = {}
        try {
          payload = JSON.parse(item.payload_json || item.payload || item.Payload || '{}')
        } catch {
          payload = {}
        }
        const id = item.id ?? item.ID
        const isRead = item.is_read ?? item.IsRead ?? false
        const createdAt = item.created_at || item.CreatedAt || ''
        return (
        <div key={id} className="card">
          <p>{item.type || item.Type} {isRead ? '(lu)' : '(non lu)'}</p>
          <small>{createdAt}</small>
          <div>
            {payload.request_id && (
              <>
                <button type="button" onClick={() => acceptFollow(Number(payload.request_id))}>
                  Accepter follow
                </button>
                <button type="button" onClick={() => refuseFollow(Number(payload.request_id))}>
                  Refuser follow
                </button>
              </>
            )}
            {payload.invite_id && (
              <>
                <button type="button" onClick={() => acceptInvite(Number(payload.invite_id))}>
                  Accepter invite
                </button>
                <button type="button" onClick={() => refuseInvite(Number(payload.invite_id))}>
                  Refuser invite
                </button>
              </>
            )}
            {payload.join_request_id && (
              <>
                <button type="button" onClick={() => acceptJoinRequest(Number(payload.join_request_id))}>
                  Accepter demande
                </button>
                <button type="button" onClick={() => refuseJoinRequest(Number(payload.join_request_id))}>
                  Refuser demande
                </button>
              </>
            )}
          </div>
          <button type="button" onClick={() => markRead(Number(id))}>Marquer comme lu</button>
        </div>
      )})}
    </section>
  )
}
