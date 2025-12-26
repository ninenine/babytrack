import { describe, it, expect } from 'vitest'
import { cn } from './utils'

describe('cn', () => {
  it('merges class names', () => {
    expect(cn('foo', 'bar')).toBe('foo bar')
  })

  it('handles single class name', () => {
    expect(cn('foo')).toBe('foo')
  })

  it('handles empty input', () => {
    expect(cn()).toBe('')
  })

  it('handles undefined and null values', () => {
    expect(cn('foo', undefined, 'bar', null)).toBe('foo bar')
  })

  it('handles conditional classes', () => {
    const isActive = true
    const isDisabled = false
    expect(cn('base', isActive && 'active', isDisabled && 'disabled')).toBe('base active')
  })

  it('merges Tailwind conflicting classes correctly', () => {
    // tailwind-merge should keep the last conflicting class
    expect(cn('px-2', 'px-4')).toBe('px-4')
    expect(cn('text-red-500', 'text-blue-500')).toBe('text-blue-500')
  })

  it('handles object syntax from clsx', () => {
    expect(cn({ foo: true, bar: false, baz: true })).toBe('foo baz')
  })

  it('handles array syntax from clsx', () => {
    expect(cn(['foo', 'bar'])).toBe('foo bar')
  })

  it('handles complex mixed input', () => {
    expect(
      cn(
        'base-class',
        ['array-class'],
        { 'object-class': true, 'excluded-class': false },
        undefined,
        'final-class'
      )
    ).toBe('base-class array-class object-class final-class')
  })
})
