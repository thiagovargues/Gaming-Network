'use client'

import { useEffect, useRef, useState } from 'react'
import { API_URL, apiFetch, uploadMedia } from '../../src/lib/api'

type Post = {
  id: number
  user_id: number
  text: string
  visibility: string
  media_path?: string | null
  created_at: string
}

type Comment = {
  id: number
  post_id: number
  user_id: number
  text: string
  media_path?: string | null
  created_at: string
}

export default function PostCard({ post, currentUserId, onDeleted }: { post: Post; currentUserId?: number | null; onDeleted?: () => void }) {
  const [comments, setComments] = useState<Comment[] | null>(null)
  const [message, setMessage] = useState('')
  const [loading, setLoading] = useState(false)
  const [fileName, setFileName] = useState('')
  const [showComposer, setShowComposer] = useState(false)
  const [liked, setLiked] = useState(false)
  const [author, setAuthor] = useState<{ avatar_path?: string | null; first_name?: string | null; last_name?: string | null } | null>(null)
  const fileRef = useRef<HTMLInputElement | null>(null)
  const isMine = currentUserId != null && post.user_id === currentUserId

  useEffect(() => {
    let cancelled = false
    async function loadAuthor() {
      try {
        const data = await apiFetch(`/api/users/${post.user_id}`)
        if (!cancelled) setAuthor({ avatar_path: data.avatar_path, first_name: data.first_name, last_name: data.last_name })
      } catch {
        if (!cancelled) setAuthor(null)
      }
    }
    loadAuthor()
    return () => {
      cancelled = true
    }
  }, [post.user_id])

  async function toggleComments() {
    if (comments) {
      setComments(null)
      return
    }
    setLoading(true)
    setMessage('')
    try {
      const data = await apiFetch(`/api/posts/${post.id}/comments`)
      setComments(data.comments || [])
    } catch (err: any) {
      setMessage(err.message)
    } finally {
      setLoading(false)
    }
  }

  async function onCommentSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    setMessage('')
    const form = new FormData(e.currentTarget)
    const text = String(form.get('text') || '').trim()
    const file = form.get('file') as File | null
    try {
      let mediaPath: string | undefined
      if (file && file.size > 0) {
        mediaPath = await uploadMedia(file)
      }
      await apiFetch(`/api/posts/${post.id}/comments`, {
        method: 'POST',
        body: JSON.stringify({ text, media_path: mediaPath })
      })
      e.currentTarget.reset()
      setFileName('')
      if (fileRef.current) fileRef.current.value = ''
      await toggleComments()
    } catch (err: any) {
      setMessage(err.message)
    }
  }

  return (
    <div className="card">
      <div className="post-head">
        <div className="post-avatar">
          {author?.avatar_path ? (
            <img src={`${API_URL}/${author.avatar_path}`} alt="" />
          ) : (
            <div className="post-avatar-placeholder" />
          )}
        </div>
        <div>
          <div className="post-meta">
            <span className="post-author">
              {author?.first_name || 'Profil'}{author?.last_name ? ` ${author.last_name}` : ''}
            </span>
            <span className="post-date">{post.created_at}</span>
            <span className="post-visibility">{post.visibility}</span>
          </div>
        </div>
      </div>
      <p className="post-text">{post.text}</p>
      {post.media_path && (
        <img src={`${API_URL}/${post.media_path}`} alt="media" style={{ maxWidth: '100%' }} />
      )}

      <div style={{ marginTop: 12 }}>
        <div className="post-actions">
          <button type="button" onClick={() => setLiked((v) => !v)} aria-label="Like" title="Like">
            <svg className="post-action-icon" viewBox="0 0 24 24" aria-hidden="true">
              <path
                d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 6 4 4 6.5 4c1.7 0 3.4.9 4.3 2.3C11.7 4.9 13.4 4 15.5 4 18 4 20 6 20 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z"
                fill={liked ? "currentColor" : "none"}
                stroke="currentColor"
                strokeWidth="1.2"
              />
            </svg>
          </button>
          <button type="button" onClick={() => setShowComposer((v) => !v)} aria-label="Commenter" title="Commenter">
            <svg className="post-action-icon" viewBox="0 0 24 24" aria-hidden="true">
              <path d="M3 17.25V21h3.75L19.81 7.94l-3.75-3.75L3 17.25z" fill={showComposer ? "currentColor" : "none"} stroke="currentColor" strokeWidth="1.2" />
              <path d="M14.06 4.19l3.75 3.75" fill="none" stroke="currentColor" strokeWidth="1.2" />
            </svg>
          </button>
          <button type="button" onClick={toggleComments} disabled={loading} aria-label="Afficher commentaires" title="Afficher commentaires">
            <svg className="post-action-icon" viewBox="0 0 24 24" aria-hidden="true">
              <path d="M4 4h16v11H7l-3 3V4z" fill={comments ? "currentColor" : "none"} stroke="currentColor" strokeWidth="1.2" strokeLinejoin="round" />
            </svg>
          </button>
          {isMine ? (
            <button
              type="button"
              onClick={async () => {
                if (!confirm('Supprimer ce post ?')) return
                try {
                  await apiFetch(`/api/posts/${post.id}`, { method: 'DELETE' })
                  onDeleted?.()
                } catch (err: any) {
                  setMessage(err.message)
                }
              }}
              aria-label="Supprimer"
              title="Supprimer"
            >
              <svg className="post-action-icon post-action-icon--delete" viewBox="0 0 24 24" aria-hidden="true">
                <path d="M6 7h12M9 7V5h6v2M8 7l1 12h6l1-12" fill="none" stroke="currentColor" strokeWidth="1.2" strokeLinecap="round" />
              </svg>
            </button>
          ) : (
            <button type="button" onClick={() => setMessage('Fonction "Transférer" à brancher')} aria-label="Transférer" title="Transférer">
              <svg className="post-action-icon post-action-icon--transfer" viewBox="0 0 24 24" aria-hidden="true">
                <path d="M14 7l5 5-5 5" fill="none" stroke="currentColor" strokeWidth="1.2" strokeLinecap="round" strokeLinejoin="round" />
                <path d="M5 12h13" fill="none" stroke="currentColor" strokeWidth="1.2" strokeLinecap="round" />
              </svg>
            </button>
          )}
        </div>
        {!showComposer ? (
          <></>
        ) : (
          <form onSubmit={onCommentSubmit}>
            <input name="text" placeholder="Écrire un commentaire" required autoFocus />
            <div className="file-row">
              <label className="file-trigger">
                Ajouter une photo
                <input
                  ref={fileRef}
                  name="file"
                  type="file"
                  accept="image/png,image/jpeg,image/gif"
                  onChange={(e) => setFileName(e.target.files?.[0]?.name || '')}
                />
              </label>
              {fileName && (
                <span className="file-chip">
                  {fileName}
                  <button
                    type="button"
                    onClick={() => {
                      setFileName('')
                      if (fileRef.current) fileRef.current.value = ''
                    }}
                  >
                    ×
                  </button>
                </span>
              )}
            </div>
            <div className="file-row">
              <button type="submit">Envoyer</button>
              <button type="button" className="secondary" onClick={() => setShowComposer(false)}>
                Annuler
              </button>
            </div>
          </form>
        )}
        {message && <p>{message}</p>}
        <div style={{ marginTop: 10 }} />
        {comments && comments.length === 0 && <p>Aucun commentaire</p>}
        {comments && comments.map((c) => (
          <div key={c.id} className="card">
            <p>#{c.id} par {c.user_id}</p>
            <p>{c.text}</p>
            {c.media_path && (
              <img src={`${API_URL}/${c.media_path}`} alt="comment media" style={{ maxWidth: '100%' }} />
            )}
            <small>{c.created_at}</small>
          </div>
        ))}
      </div>
    </div>
  )
}
