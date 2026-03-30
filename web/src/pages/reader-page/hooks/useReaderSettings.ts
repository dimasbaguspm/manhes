import { usePersistedState } from '../../../hooks/usePersistedState'

export interface ReaderSettings {
  stripWidth: string
  bgColor: string
  autoScroll: boolean
  autoScrollSpeed: number
  showProgress: boolean
  showPageIndicator: boolean
}

export const READER_SETTINGS_DEFAULT: ReaderSettings = {
  stripWidth: 'normal',
  bgColor: 'dark',
  autoScroll: false,
  autoScrollSpeed: 1,
  showProgress: true,
  showPageIndicator: true,
}

export function useReaderSettings() {
  const [settings, setSettings] = usePersistedState({
    key: 'manhes:reader-settings',
    fallback: READER_SETTINGS_DEFAULT,
  })

  function set<K extends keyof ReaderSettings>(key: K, value: ReaderSettings[K]) {
    setSettings(prev => ({ ...prev, [key]: value }))
  }

  const stripMaxWidthClass =
    settings.stripWidth === 'narrow' ? 'max-w-lg' :
    settings.stripWidth === 'wide'   ? 'max-w-5xl' :
    settings.stripWidth === 'full'   ? 'max-w-none' :
    'max-w-3xl'

  const bgClass =
    settings.bgColor === 'black' ? 'bg-black' :
    settings.bgColor === 'white' ? 'bg-white' :
    'bg-gray-950'

  return { settings, set, stripMaxWidthClass, bgClass }
}
