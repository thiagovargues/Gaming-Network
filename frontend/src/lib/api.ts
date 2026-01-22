export const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

export async function apiFetch(path: string, options: RequestInit = {}) {
  const res = await fetch(`${API_URL}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(options.headers || {})
    },
    credentials: 'include'
  })

  const contentType = res.headers.get('content-type') || ''
  const data = contentType.includes('application/json') ? await res.json() : null
  if (!res.ok) {
    const message = data?.error || `HTTP ${res.status}`
    throw new Error(message)
  }
  return data
}

export function wsURL() {
  const base = process.env.NEXT_PUBLIC_WS_URL
  if (base) return base
  return API_URL.replace('http', 'ws') + '/api/ws'
}
