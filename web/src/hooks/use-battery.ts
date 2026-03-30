import { useEffect, useState } from 'react'

interface BatteryState {
  level: number
  charging: boolean
}

export function useBattery(): BatteryState {
  const [state, setState] = useState<BatteryState>({ level: 1, charging: false })

  useEffect(() => {
    const nav = navigator as Navigator & {
      getBattery?: () => Promise<{
        level: number
        charging: boolean
        addEventListener: (event: string, handler: () => void) => void
        removeEventListener: (event: string, handler: () => void) => void
      }>
    }

    if (!nav.getBattery) return

    nav.getBattery().then(battery => {
      const update = () => setState({ level: battery.level, charging: battery.charging })
      update()
      battery.addEventListener('levelchange', update)
      battery.addEventListener('chargingchange', update)
      return () => {
        battery.removeEventListener('levelchange', update)
        battery.removeEventListener('chargingchange', update)
      }
    })
  }, [])

  return state
}
