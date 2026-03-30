import { createContext, useContext, type ReactNode, type HTMLAttributes } from 'react'
import { cn } from '@/lib/cn'

interface TabsContextValue {
  activeTab: string
  selectTab: (id: string) => void
}

const TabsContext = createContext<TabsContextValue | null>(null)

function useTabsContext() {
  const ctx = useContext(TabsContext)
  if (!ctx) throw new Error('Tab sub-components must be used within <Tab>')
  return ctx
}

interface TabProps {
  value: string
  onChange: (value: string) => void
  children: ReactNode
  className?: string
}

export function Tab({ value, onChange, children, className }: TabProps) {
  return (
    <TabsContext.Provider value={{ activeTab: value, selectTab: onChange }}>
      <div className={className} data-tabs>
        {children}
      </div>
    </TabsContext.Provider>
  )
}

interface TabItemProps extends Omit<HTMLAttributes<HTMLButtonElement>, 'onClick'> {
  name: string
  children: ReactNode
  className?: string
}

export function TabItem({ name, children, className, ...props }: TabItemProps) {
  const { activeTab, selectTab } = useTabsContext()
  const isActive = activeTab === name

  return (
    <button
      role="tab"
      aria-selected={isActive}
      onClick={() => selectTab(name)}
      className={cn(
        'flex items-center gap-2 rounded-t-lg border border-b-0 px-4 py-2 text-sm font-medium transition',
        isActive
          ? 'border-gray-700 bg-gray-900 text-white'
          : 'border-transparent text-gray-500 hover:text-gray-300',
        className
      )}
      data-tab-item={name}
      data-tab-active={isActive ? 'true' : undefined}
      {...props}
    >
      {children}
    </button>
  )
}

