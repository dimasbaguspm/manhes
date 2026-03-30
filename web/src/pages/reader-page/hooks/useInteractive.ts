import { useContext } from 'react'
import { Ctx, type InteractiveCtx } from '../components/InteractiveProvider'

export function useInteractive(): InteractiveCtx {
  const ctx = useContext(Ctx)
  if (!ctx) throw new Error('useInteractive must be used within InteractiveProvider')
  return ctx
}
