'use client'

import { useState } from 'react'
import { apiFetch } from '../../src/lib/api'

export default function LoginPage() {
  const [message, setMessage] = useState('')

  async function onSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    const form = new FormData(e.currentTarget)
    try {
      await apiFetch('/api/auth/login', {
        method: 'POST',
        body: JSON.stringify({
          email: form.get('email'),
          password: form.get('password')
        })
      })
      setMessage('Connecté')
    } catch (err: any) {
      setMessage(err.message)
    }
  }

  return (
    <section className="card">
      <h1>Login</h1>
      <form onSubmit={onSubmit}>
        <input name="email" placeholder="Email" required />
        <input name="password" type="password" placeholder="Password" required />
        <button type="submit">Login</button>
      </form>
      {message && <p>{message}</p>}
    </section>
  )
}
