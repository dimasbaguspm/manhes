import { useInterval } from '@/hooks/use-interval'
import { useState } from 'react'

export function useClock(): string {
  const [time, setTime] = useState(() => new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }))

  useInterval(() => {
    setTime(new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }))
  }, 60_000)

  return time
}
