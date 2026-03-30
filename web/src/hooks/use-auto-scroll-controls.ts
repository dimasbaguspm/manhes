import { useCallback } from 'react'
import { useReaderSettings } from '@/pages/reader-page/components/use-reader-settings'

export function useAutoScrollControls() {
  const { settings, set } = useReaderSettings()

  const toggle = useCallback(() => {
    set('autoScroll', !settings.autoScroll)
  }, [settings.autoScroll, set])

  const cycleSpeed = useCallback(() => {
    const next = settings.autoScrollSpeed >= 5 ? 1 : settings.autoScrollSpeed + 1
    set('autoScrollSpeed', next)
  }, [settings.autoScrollSpeed, set])

  return {
    isActive: settings.autoScroll,
    speed: settings.autoScrollSpeed,
    toggle,
    cycleSpeed,
  }
}
