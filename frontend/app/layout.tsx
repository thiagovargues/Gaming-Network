import './globals.css'
import Link from 'next/link'
import type { ReactNode } from 'react'

export const metadata = {
  title: 'Gaming Network',
  description: 'Social network MVP'
}

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="fr">
      <body>
        <nav className="nav">
          <Link href="/">Home</Link>
          <Link href="/register">Register</Link>
          <Link href="/login">Login</Link>
          <Link href="/feed">Feed</Link>
          <Link href="/groups">Groups</Link>
          <Link href="/chat">Chat</Link>
          <Link href="/notifications">Notifications</Link>
        </nav>
        <main>{children}</main>
      </body>
    </html>
  )
}
