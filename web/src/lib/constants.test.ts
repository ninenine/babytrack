import { describe, it, expect } from 'vitest'
import {
  API_ENDPOINTS,
  FEEDING_TYPES,
  SLEEP_TYPES,
  MEDICATION_FREQUENCIES,
  MEDICATION_UNITS,
  APPOINTMENT_TYPES,
  DURATION_OPTIONS,
} from './constants'

describe('API_ENDPOINTS', () => {
  describe('FAMILIES', () => {
    it('builds BY_ID path correctly', () => {
      expect(API_ENDPOINTS.FAMILIES.BY_ID('abc123')).toBe('/api/families/abc123')
    })

    it('builds CHILDREN path correctly', () => {
      expect(API_ENDPOINTS.FAMILIES.CHILDREN('family-1')).toBe('/api/families/family-1/children')
    })
  })

  describe('FEEDINGS', () => {
    it('builds BY_ID path correctly', () => {
      expect(API_ENDPOINTS.FEEDINGS.BY_ID('feed-1')).toBe('/api/feeding/feed-1')
    })

    it('builds LAST path correctly', () => {
      expect(API_ENDPOINTS.FEEDINGS.LAST('child-1')).toBe('/api/feeding/last/child-1')
    })
  })

  describe('SLEEP', () => {
    it('builds END path correctly', () => {
      expect(API_ENDPOINTS.SLEEP.END('sleep-1')).toBe('/api/sleep/sleep-1/end')
    })

    it('builds ACTIVE path correctly', () => {
      expect(API_ENDPOINTS.SLEEP.ACTIVE('child-1')).toBe('/api/sleep/active/child-1')
    })
  })

  describe('MEDICATIONS', () => {
    it('builds LOGS path correctly', () => {
      expect(API_ENDPOINTS.MEDICATIONS.LOGS('med-1')).toBe('/api/medications/med-1/logs')
    })

    it('builds DEACTIVATE path correctly', () => {
      expect(API_ENDPOINTS.MEDICATIONS.DEACTIVATE('med-1')).toBe('/api/medications/med-1/deactivate')
    })
  })

  describe('VACCINATIONS', () => {
    it('builds GENERATE path correctly', () => {
      expect(API_ENDPOINTS.VACCINATIONS.GENERATE('child-1')).toBe('/api/vaccinations/generate/child-1')
    })

    it('builds RECORD path correctly', () => {
      expect(API_ENDPOINTS.VACCINATIONS.RECORD('vax-1')).toBe('/api/vaccinations/vax-1/record')
    })
  })

  describe('APPOINTMENTS', () => {
    it('builds COMPLETE path correctly', () => {
      expect(API_ENDPOINTS.APPOINTMENTS.COMPLETE('apt-1')).toBe('/api/appointments/apt-1/complete')
    })

    it('builds CANCEL path correctly', () => {
      expect(API_ENDPOINTS.APPOINTMENTS.CANCEL('apt-1')).toBe('/api/appointments/apt-1/cancel')
    })
  })
})

describe('FEEDING_TYPES', () => {
  it('contains expected feeding types', () => {
    const values = FEEDING_TYPES.map((t) => t.value)
    expect(values).toContain('breast')
    expect(values).toContain('bottle')
    expect(values).toContain('formula')
    expect(values).toContain('solid')
  })

  it('has labels and emojis for all types', () => {
    FEEDING_TYPES.forEach((type) => {
      expect(type.label).toBeTruthy()
      expect(type.emoji).toBeTruthy()
    })
  })
})

describe('SLEEP_TYPES', () => {
  it('contains nap and night types', () => {
    const values = SLEEP_TYPES.map((t) => t.value)
    expect(values).toEqual(['nap', 'night'])
  })
})

describe('MEDICATION_FREQUENCIES', () => {
  it('has at least 5 frequency options', () => {
    expect(MEDICATION_FREQUENCIES.length).toBeGreaterThanOrEqual(5)
  })

  it('includes common frequencies', () => {
    const values = MEDICATION_FREQUENCIES.map((f) => f.value)
    expect(values).toContain('once_daily')
    expect(values).toContain('twice_daily')
    expect(values).toContain('as_needed')
  })
})

describe('MEDICATION_UNITS', () => {
  it('contains common units', () => {
    expect(MEDICATION_UNITS).toContain('mg')
    expect(MEDICATION_UNITS).toContain('ml')
    expect(MEDICATION_UNITS).toContain('tablets')
  })
})

describe('APPOINTMENT_TYPES', () => {
  it('contains expected appointment types', () => {
    const values = APPOINTMENT_TYPES.map((t) => t.value)
    expect(values).toContain('well_visit')
    expect(values).toContain('sick_visit')
    expect(values).toContain('specialist')
  })

  it('has colors for all types', () => {
    APPOINTMENT_TYPES.forEach((type) => {
      expect(type.color).toMatch(/^bg-/)
    })
  })
})

describe('DURATION_OPTIONS', () => {
  it('has duration values in minutes', () => {
    const values = DURATION_OPTIONS.map((d) => d.value)
    expect(values).toContain(15)
    expect(values).toContain(30)
    expect(values).toContain(60)
  })

  it('has human-readable labels', () => {
    const hourOption = DURATION_OPTIONS.find((d) => d.value === 60)
    expect(hourOption?.label).toBe('1 hour')
  })
})
