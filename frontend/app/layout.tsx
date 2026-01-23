import './globals.css'
import type { ReactNode } from 'react'
import Nav from './components/Nav'
import ChatDock from './components/ChatDock'

export const metadata = {
  title: 'Gaming Network',
  description: 'Social network MVP'
}

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="fr">
      <body>
        <Nav />
        <main>{children}</main>
        <ChatDock />
      </body>
    </html>
  )
}
