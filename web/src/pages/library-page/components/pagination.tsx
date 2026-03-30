import { ChevronLeft, ChevronRight } from 'lucide-react'
import { Icon } from '@/components/icon'
import { Button } from '@/components/ui/button/button'
import { Text } from '@/components/ui/text/text'

interface Props {
  page: number
  total: number
  onChange: (page: number) => void
}

export default function Pagination({ page, total, onChange }: Props) {
  if (total <= 1) return null

  return (
    <div className="flex items-center justify-center gap-3">
      <Button
        variant="outline"
        size="sm"
        onClick={() => onChange(page - 1)}
        disabled={page <= 1}
      >
        <Icon as={ChevronLeft} size="small" className="mr-1 inline" /> Prev
      </Button>
      <Text color="muted">
        {page} / {total}
      </Text>
      <Button
        variant="outline"
        size="sm"
        onClick={() => onChange(page + 1)}
        disabled={page >= total}
      >
        Next <Icon as={ChevronRight} size="small" className="ml-1 inline" />
      </Button>
    </div>
  )
}
