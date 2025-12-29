interface PlayerProps {
  src: string
  poster?: string
  autoPlay?: boolean
  muted?: boolean
  controls?: boolean

  theme?: 'light' | 'dark'
  accentColor?: string
  borderRadius?: number
  controlsStyle?: 'minimal' | 'full'

  onPlay?: () => void
  onPause?: () => void
  onEnded?: () => void
  onTimeUpdate?: (time: number) => void
  onError?: (error: Error) => void
}
