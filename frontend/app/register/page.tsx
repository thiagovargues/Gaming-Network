'use client'

import { useState } from 'react'
import { apiFetch, API_URL } from '../../src/lib/api'

export default function RegisterPage() {
  const [message, setMessage] = useState('')

  async function onSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    const form = new FormData(e.currentTarget)
    try {
      await apiFetch('/api/auth/register', {
        method: 'POST',
        body: JSON.stringify({
          email: form.get('email'),
          password: form.get('password'),
          first_name: form.get('first_name'),
          last_name: form.get('last_name'),
          dob: form.get('dob')
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
        <input name="dob" placeholder="YYYY-MM-DD" required />
        <button type="submit">Créer</button>
      </form>
      {message && <p>{message}</p>}
    </section>
  )
}
