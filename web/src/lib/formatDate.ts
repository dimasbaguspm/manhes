import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import updateLocale from 'dayjs/plugin/updateLocale'

dayjs.extend(relativeTime)
dayjs.extend(updateLocale)

export enum DateFormat {
  /** "Mar 26, 2026" */
  ShortDate,
  /** "Mar 26, 2026, 08:42" */
  ShortDateTime,
  /** "08:42" */
  TimeOnly,
}

const FORMAT_MAP: Record<DateFormat, string> = {
  [DateFormat.ShortDate]:     'MMM D, YYYY',
  [DateFormat.ShortDateTime]: 'MMM D, YYYY, HH:mm',
  [DateFormat.TimeOnly]:      'HH:mm',
}

export function formatDate(
  value: string | Date | null | undefined,
  fmt: DateFormat = DateFormat.ShortDate,
): string {
  if (!value) return ''
  try {
    return dayjs(value).format(FORMAT_MAP[fmt])
  } catch {
    return ''
  }
}
