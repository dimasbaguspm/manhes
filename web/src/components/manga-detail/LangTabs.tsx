interface LangInfo {
  lang: string
  totalChapters: number
  uploadedChapters: number
}

interface LangTabsProps {
  langs: LangInfo[]
  activeLang: string | null
  onSelect: (lang: string) => void
}

export function LangTabs({ langs, activeLang, onSelect }: LangTabsProps) {
  return (
    <div className="mb-4 flex items-end gap-1 border-b border-gray-800">
      {langs.map(({ lang, totalChapters, uploadedChapters }) => {
        const isActive = lang === activeLang
        return (
          <button
            key={lang}
            onClick={() => onSelect(lang)}
            className={`flex items-center gap-2 rounded-t-lg border border-b-0 px-4 py-2 text-sm font-medium transition ${
              isActive
                ? 'border-gray-700 bg-gray-900 text-white'
                : 'border-transparent text-gray-500 hover:text-gray-300'
            }`}
          >
            <span className="uppercase">{lang}</span>
            {totalChapters > 0 && (
              <span className={`text-xs ${isActive ? 'text-gray-400' : 'text-gray-600'}`}>
                {uploadedChapters}
              </span>
            )}
          </button>
        )
      })}
    </div>
  )
}
