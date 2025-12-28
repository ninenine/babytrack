import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import {
  getLocalTimezone,
  getTimezoneAbbr,
  formatLocal,
  formatWithTimezone,
  toAPIDateTime,
  toAPIDate,
  fromAPIDateTime,
  fromAPIDate,
  todayAt,
  isSameDay,
} from './dates'

describe('dates utilities', () => {
  describe('getLocalTimezone', () => {
    it('should return a timezone string', () => {
      const tz = getLocalTimezone()
      expect(typeof tz).toBe('string')
      expect(tz.length).toBeGreaterThan(0)
    })
  })

  describe('getTimezoneAbbr', () => {
    it('should return a timezone abbreviation', () => {
      const abbr = getTimezoneAbbr()
      expect(typeof abbr).toBe('string')
    })

    it('should accept a date parameter', () => {
      const date = new Date('2024-06-15T12:00:00Z')
      const abbr = getTimezoneAbbr(date)
      expect(typeof abbr).toBe('string')
    })
  })

  describe('formatLocal', () => {
    it('should format a Date object', () => {
      const date = new Date('2024-03-15T10:30:00')
      const formatted = formatLocal(date, 'yyyy-MM-dd')
      expect(formatted).toBe('2024-03-15')
    })

    it('should format an ISO string', () => {
      const formatted = formatLocal('2024-03-15T10:30:00', 'yyyy-MM-dd')
      expect(formatted).toBe('2024-03-15')
    })

    it('should format with time', () => {
      const date = new Date('2024-03-15T10:30:00')
      const formatted = formatLocal(date, 'HH:mm')
      expect(formatted).toBe('10:30')
    })
  })

  describe('formatWithTimezone', () => {
    it('should format a Date object with timezone', () => {
      const date = new Date('2024-03-15T10:30:00')
      const formatted = formatWithTimezone(date, 'yyyy-MM-dd')
      expect(formatted).toBe('2024-03-15')
    })

    it('should format an ISO string with timezone', () => {
      const formatted = formatWithTimezone('2024-03-15T10:30:00', 'yyyy-MM-dd')
      expect(formatted).toBe('2024-03-15')
    })
  })

  describe('toAPIDateTime', () => {
    it('should convert Date to ISO string', () => {
      const date = new Date('2024-03-15T10:30:00Z')
      const result = toAPIDateTime(date)
      expect(result).toBe('2024-03-15T10:30:00.000Z')
    })
  })

  describe('toAPIDate', () => {
    it('should convert Date to date-only string', () => {
      const date = new Date(2024, 2, 15) // March 15, 2024
      const result = toAPIDate(date)
      expect(result).toBe('2024-03-15')
    })
  })

  describe('fromAPIDateTime', () => {
    it('should parse ISO datetime string to Date', () => {
      const result = fromAPIDateTime('2024-03-15T10:30:00.000Z')
      expect(result instanceof Date).toBe(true)
      expect(result.toISOString()).toBe('2024-03-15T10:30:00.000Z')
    })
  })

  describe('fromAPIDate', () => {
    it('should parse date-only string to local Date', () => {
      const result = fromAPIDate('2024-03-15')
      expect(result instanceof Date).toBe(true)
      expect(result.getFullYear()).toBe(2024)
      expect(result.getMonth()).toBe(2) // March (0-indexed)
      expect(result.getDate()).toBe(15)
    })

    it('should parse ISO string and extract date in local timezone', () => {
      const result = fromAPIDate('2024-03-15T10:30:00.000Z')
      expect(result instanceof Date).toBe(true)
      expect(result.getFullYear()).toBe(2024)
    })

    it('should handle incomplete date string with defaults', () => {
      // Edge case: date string with missing parts
      const result = fromAPIDate('2024')
      expect(result instanceof Date).toBe(true)
      expect(result.getFullYear()).toBe(2024)
    })

    it('should handle date string with only year and month', () => {
      const result = fromAPIDate('2024-06')
      expect(result instanceof Date).toBe(true)
      expect(result.getFullYear()).toBe(2024)
      expect(result.getMonth()).toBe(5) // June
    })
  })

  describe('todayAt', () => {
    beforeEach(() => {
      vi.useFakeTimers()
      vi.setSystemTime(new Date('2024-03-15T08:00:00'))
    })

    afterEach(() => {
      vi.useRealTimers()
    })

    it('should create Date at specific hour', () => {
      const result = todayAt(14)
      expect(result.getHours()).toBe(14)
      expect(result.getMinutes()).toBe(0)
      expect(result.getSeconds()).toBe(0)
    })

    it('should create Date at specific hour and minutes', () => {
      const result = todayAt(14, 30)
      expect(result.getHours()).toBe(14)
      expect(result.getMinutes()).toBe(30)
    })
  })

  describe('isSameDay', () => {
    it('should return true for same day', () => {
      const date1 = new Date('2024-03-15T08:00:00')
      const date2 = new Date('2024-03-15T20:00:00')
      expect(isSameDay(date1, date2)).toBe(true)
    })

    it('should return false for different days', () => {
      const date1 = new Date('2024-03-15T08:00:00')
      const date2 = new Date('2024-03-16T08:00:00')
      expect(isSameDay(date1, date2)).toBe(false)
    })

    it('should return false for different months', () => {
      const date1 = new Date('2024-03-15T08:00:00')
      const date2 = new Date('2024-04-15T08:00:00')
      expect(isSameDay(date1, date2)).toBe(false)
    })

    it('should return false for different years', () => {
      const date1 = new Date('2024-03-15T08:00:00')
      const date2 = new Date('2025-03-15T08:00:00')
      expect(isSameDay(date1, date2)).toBe(false)
    })
  })
})
