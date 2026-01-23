'use client'

import { useState } from 'react'
import { apiFetch, API_URL, uploadMedia } from '../../src/lib/api'

export default function RegisterPage() {
  const [message, setMessage] = useState('')

  async function onSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    const form = new FormData(e.currentTarget)
    try {
      const file = form.get('avatar') as File | null
      let avatarPath: string | undefined
      if (file && file.size > 0) {
        avatarPath = await uploadMedia(file)
      }
      await apiFetch('/api/auth/register', {
        method: 'POST',
        body: JSON.stringify({
          email: form.get('email'),
          password: form.get('password'),
          first_name: form.get('first_name'),
          last_name: form.get('last_name'),
          dob: form.get('dob'),
          avatar: avatarPath,
          nickname: form.get('nickname') || undefined,
          about: form.get('about') || undefined
        })
      })
      setMessage('Inscription OK. Tu peux login.')
    } catch (err: any) {
      setMessage(err.message)
    }
  }

  return (
    <section className="card">
      <h1>Register</h1>
      <button type="button" onClick={() => (window.location.href = `${API_URL}/api/auth/google/start`)}>
        Continuer avec Google
      </button>
      <form onSubmit={onSubmit}>
        <input name="email" placeholder="Email" required />
        <input name="password" type="password" placeholder="Password" required />
        <input name="first_name" placeholder="Prénom" required />
        <input name="last_name" placeholder="Nom" required />
        <input name="dob" type="date" required />
        <input name="nickname" placeholder="Pseudo (optionnel)" />
        <input name="about" placeholder="À propos (optionnel)" />
        <input name="avatar" type="file" accept="image/png,image/jpeg,image/gif" />
        <button type="submit">Créer</button>
      </form>
      {message && <p>{message}</p>}
    </section>
  )
}
