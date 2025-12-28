import { format, parseISO } from 'date-fns'
import { formatInTimeZone } from 'date-fns-tz'

/**
 * Get the user's local timezone
 */
export function getLocalTimezone(): string {
  return Intl.DateTimeFormat().resolvedOptions().timeZone
}

/**
 * Get short timezone abbreviation (e.g., "EAT", "PST", "UTC")
 */
export function getTimezoneAbbr(date: Date = new Date()): string {
  return date.toLocaleTimeString('en-US', { timeZoneName: 'short' }).split(' ').pop() || ''
}

/**
 * Format a date for display in local timezone
 * @param date - Date object or ISO string
 * @param formatStr - date-fns format string
 */
export function formatLocal(date: Date | string, formatStr: string): string {
  const d = typeof date === 'string' ? parseISO(date) : date
  return format(d, formatStr)
}

/**
 * Format a date with timezone indicator
 * @param date - Date object or ISO string
 * @param formatStr - date-fns format string
 */
export function formatWithTimezone(date: Date | string, formatStr: string): string {
  const d = typeof date === 'string' ? parseISO(date) : date
  const tz = getLocalTimezone()
  return formatInTimeZone(d, tz, formatStr)
}

/**
 * Convert a Date to ISO string for API submission (UTC)
 * Used for datetime fields (feeding times, sleep times, etc.)
 */
export function toAPIDateTime(date: Date): string {
  return date.toISOString()
}

/**
 * Convert a Date to date-only string for API submission
 * Used for date-only fields (date_of_birth, medication dates, etc.)
 * Format: YYYY-MM-DD (no timezone conversion to prevent date shifting)
 */
export function toAPIDate(date: Date): string {
  return format(date, 'yyyy-MM-dd')
}

/**
 * Parse an ISO datetime string from API to local Date
 */
export function fromAPIDateTime(isoString: string): Date {
  return parseISO(isoString)
}

/**
 * Parse a date-only string from API to Date (at midnight local time)
 * Handles both "YYYY-MM-DD" and full ISO strings
 */
export function fromAPIDate(dateString: string): Date {
  // If it's just a date (no T), parse as local date
  if (!dateString.includes('T')) {
    const parts = dateString.split('-').map(Number)
    const year = parts[0] ?? 0
    const month = parts[1] ?? 1
    const day = parts[2] ?? 1
    return new Date(year, month - 1, day)
  }
  // Otherwise parse as ISO and extract just the date part in local timezone
  const parsed = parseISO(dateString)
  return new Date(parsed.getFullYear(), parsed.getMonth(), parsed.getDate())
}

/**
 * Create a Date at a specific local time today
 */
export function todayAt(hours: number, minutes: number = 0): Date {
  const now = new Date()
  now.setHours(hours, minutes, 0, 0)
  return now
}

/**
 * Check if two dates are the same calendar day (in local timezone)
 */
export function isSameDay(date1: Date, date2: Date): boolean {
  return (
    date1.getFullYear() === date2.getFullYear() &&
    date1.getMonth() === date2.getMonth() &&
    date1.getDate() === date2.getDate()
  )
}
