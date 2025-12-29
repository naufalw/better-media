import { createRoot } from 'react-dom/client'
import { Player } from './components/Player'

const config = (window as any).__VIDEO_CONFIG__ || {}

createRoot(document.getElementById('player')!).render(
  <Player
    src={config.src}
    poster={config.poster}
    autoPlay={config.autoPlay}
    theme={config.theme}
  />
)
