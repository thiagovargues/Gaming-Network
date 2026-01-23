'use client'

import { useEffect, useState } from 'react'
import { apiFetch, uploadMedia } from '../../src/lib/api'

type User = {
  id: number
  first_name: string
  last_name: string
  about?: string | null
  age?: number | null
  show_first_name?: boolean
  show_last_name?: boolean
  show_age?: boolean
  show_about?: boolean
  is_public: boolean
}

export default function SettingsPage() {
  const [user, setUser] = useState<User | null>(null)
  const [message, setMessage] = useState('')
  const [isPublic, setIsPublic] = useState<boolean>(true)

  async function load() {
    try {
      const data = await apiFetch('/api/me')
      setUser(data)
      setIsPublic(Boolean(data.is_public))
    } catch (err: any) {
      setMessage(err.message)
    }
  }

  useEffect(() => {
    load()
  }, [])

  async function onSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    const form = new FormData(e.currentTarget)
    setMessage('')
    try {
      const file = form.get('avatar') as File | null
      let avatarPath: string | undefined
      if (file && file.size > 0) {
        avatarPath = await uploadMedia(file)
      }
      const ageRaw = String(form.get('age') || '').trim()
      const ageValue = ageRaw === '' ? undefined : Number(ageRaw)
      if (ageValue !== undefined && Number.isNaN(ageValue)) {
        setMessage('Age invalide')
        return
      }
      await apiFetch('/api/users/me', {
        method: 'PATCH',
        body: JSON.stringify({
          is_public: isPublic,
          first_name: form.get('first_name') || undefined,
          last_name: form.get('last_name') || undefined,
          about: form.get('about') || undefined,
          age: ageValue,
          avatar: avatarPath,
          show_last_name: form.get('show_last_name') === 'on',
          show_age: form.get('show_age') === 'on',
          show_about: form.get('show_about') === 'on'
        })
      })
      setMessage('Profil mis à jour')
      load()
    } catch (err: any) {
      setMessage(err.message)
    }
  }

  return (
    <section className="card settings-card">
      <h1>Modifier le profil</h1>
      {message && <p>{message}</p>}
      {user && (
        <form onSubmit={onSubmit} className="settings-form">
          <div className="settings-row">
            <div className="field-with-label">
              <span className="field-label">Compte :</span>
              <span className="field-value">{isPublic ? 'public' : 'privé'}</span>
            </div>
            <label className="switch-field">
              <input
                name="is_public"
                type="checkbox"
                checked={isPublic}
                onChange={(e) => setIsPublic(e.target.checked)}
              />
              <span className="switch-slider" />
            </label>
          </div>
          <div className="settings-row">
            <div className="field-with-label">
              <span className="field-label">Prénom :</span>
              <input name="first_name" placeholder="Prénom" defaultValue={user.first_name || ''} maxLength={30} />
            </div>
            <span className="switch-spacer" />
          </div>
          <div className="settings-row">
            <div className="field-with-label">
              <span className="field-label">Nom :</span>
              <input name="last_name" placeholder="Nom" defaultValue={user.last_name || ''} maxLength={30} />
            </div>
            <label className="switch-field">
              <input name="show_last_name" type="checkbox" defaultChecked={user.show_last_name ?? true} />
              <span className="switch-slider" />
            </label>
          </div>
          <div className="settings-row">
            <div className="field-with-label">
              <span className="field-label">Âge :</span>
              <input name="age" placeholder="Âge" defaultValue={user.age ?? ''} maxLength={3} inputMode="numeric" />
            </div>
            <label className="switch-field">
              <input name="show_age" type="checkbox" defaultChecked={user.show_age ?? true} />
              <span className="switch-slider" />
            </label>
          </div>
          
          <div className="settings-row">
            <div className="field-with-label">
              <span className="field-label">Bio :</span>
              <textarea name="about" placeholder="Bio" defaultValue={user.about || ''} maxLength={300} />
            </div>
            <label className="switch-field">
              <input name="show_about" type="checkbox" defaultChecked={user.show_about ?? true} />
              <span className="switch-slider" />
            </label>
          </div>
          <div className="settings-row">
            <div className="field-with-label">
              <span className="field-label">Photo :</span>
              <label className="file-trigger">
                Uploader une image de profil
                <input name="avatar" type="file" accept="image/png,image/jpeg,image/gif" />
              </label>
            </div>
          </div>
          <div className="file-row">
            <button type="submit">Enregistrer</button>
          </div>
        </form>
      )}
    </section>
  )
}
