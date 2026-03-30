import { Tab, TabItem } from '@/components/ui'
import type { DomainMangaLangResponse } from '@/types'

interface LangTabsProps {
  langs: DomainMangaLangResponse[]
  activeLang: string | null
  onSelect: (lang: string) => void
}

export function LangTabs({ langs, activeLang, onSelect }: LangTabsProps) {
  return (
    <Tab value={activeLang ?? ''} onChange={onSelect} className="mb-4 flex items-end gap-1 border-b border-gray-800">
      {langs.map((langInfo, i) => (
        <TabItem
          key={langInfo.lang ?? i}
          name={langInfo.lang ?? ''}
          className="uppercase"
        >
          <span>{langInfo.lang}</span>
          {(langInfo.total_chapters ?? 0) > 0 && (
            <span className="text-xs text-gray-400">
              {langInfo.available_chapters ?? 0}
            </span>
          )}
        </TabItem>
      ))}
    </Tab>
  )
}
