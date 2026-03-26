export enum DateFormat {
  /** "Mar 26, 2026" */
  ShortDate,
  /** "Mar 26, 2026, 08:42" */
  ShortDateTime,
  /** "08:42" */
  TimeOnly,
}

const FORMAT_OPTIONS: Record<DateFormat, Intl.DateTimeFormatOptions> = {
  [DateFormat.ShortDate]:     { dateStyle: 'medium' },
  [DateFormat.ShortDateTime]: { dateStyle: 'medium', timeStyle: 'short' },
  [DateFormat.TimeOnly]:      { timeStyle: 'short' },
}

export function formatDate(
  value: string | Date | null | undefined,
  fmt: DateFormat = DateFormat.ShortDate,
): string {
  if (!value) return ''
  try {
    const date = value instanceof Date ? value : new Date(value)
    return new Intl.DateTimeFormat(undefined, FORMAT_OPTIONS[fmt]).format(date)
  } catch {
    return ''
  }
}
