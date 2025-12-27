// API endpoints
export const API_ENDPOINTS = {
  // Auth
  AUTH: {
    GOOGLE: '/api/auth/google',
    CALLBACK: '/api/auth/google/callback',
    REFRESH: '/api/auth/refresh',
    ME: '/api/auth/me',
  },
  // Families
  FAMILIES: {
    BASE: '/api/families',
    BY_ID: (id: string) => `/api/families/${id}`,
    JOIN: (id: string) => `/api/families/${id}/join`,
    CHILDREN: (familyId: string) => `/api/families/${familyId}/children`,
  },
  // Feedings
  FEEDINGS: {
    BASE: '/api/feeding',
    BY_ID: (id: string) => `/api/feeding/${id}`,
    LAST: (childId: string) => `/api/feeding/last/${childId}`,
  },
  // Sleep
  SLEEP: {
    BASE: '/api/sleep',
    BY_ID: (id: string) => `/api/sleep/${id}`,
    START: '/api/sleep/start',
    END: (id: string) => `/api/sleep/${id}/end`,
    ACTIVE: (childId: string) => `/api/sleep/active/${childId}`,
  },
  // Medications
  MEDICATIONS: {
    BASE: '/api/medications',
    BY_ID: (id: string) => `/api/medications/${id}`,
    LOG: '/api/medications/log',
    LOGS: (id: string) => `/api/medications/${id}/logs`,
    DEACTIVATE: (id: string) => `/api/medications/${id}/deactivate`,
  },
  // Vaccinations
  VACCINATIONS: {
    BASE: '/api/vaccinations',
    BY_ID: (id: string) => `/api/vaccinations/${id}`,
    SCHEDULE: '/api/vaccinations/schedule',
    GENERATE: (childId: string) => `/api/vaccinations/generate/${childId}`,
    RECORD: (id: string) => `/api/vaccinations/${id}/record`,
    UPCOMING: (childId: string) => `/api/vaccinations/upcoming/${childId}`,
  },
  // Appointments
  APPOINTMENTS: {
    BASE: '/api/appointments',
    BY_ID: (id: string) => `/api/appointments/${id}`,
    COMPLETE: (id: string) => `/api/appointments/${id}/complete`,
    CANCEL: (id: string) => `/api/appointments/${id}/cancel`,
    UPCOMING: (childId: string) => `/api/appointments/upcoming/${childId}`,
  },
  // Notes
  NOTES: {
    BASE: '/api/notes',
    BY_ID: (id: string) => `/api/notes/${id}`,
    PIN: (id: string) => `/api/notes/${id}/pin`,
    SEARCH: '/api/notes/search',
  },
  // Sync
  SYNC: {
    PUSH: '/api/sync/push',
    PULL: '/api/sync/pull',
    STATUS: '/api/sync/status',
  },
} as const

// Feeding types
export const FEEDING_TYPES = [
  { value: 'breast', label: 'Breast', emoji: '🤱' },
  { value: 'bottle', label: 'Bottle', emoji: '🍼' },
  { value: 'formula', label: 'Formula', emoji: '🥛' },
  { value: 'solid', label: 'Solid Food', emoji: '🥣' },
] as const

// Sleep types
export const SLEEP_TYPES = [
  { value: 'nap', label: 'Nap', emoji: '😴' },
  { value: 'night', label: 'Night Sleep', emoji: '🌙' },
] as const

// Medication frequencies
export const MEDICATION_FREQUENCIES = [
  { value: 'once_daily', label: 'Once daily' },
  { value: 'twice_daily', label: 'Twice daily' },
  { value: 'three_times_daily', label: 'Three times daily' },
  { value: 'four_times_daily', label: 'Four times daily' },
  { value: 'every_4_hours', label: 'Every 4 hours' },
  { value: 'every_6_hours', label: 'Every 6 hours' },
  { value: 'every_8_hours', label: 'Every 8 hours' },
  { value: 'as_needed', label: 'As needed' },
] as const

// Medication units
export const MEDICATION_UNITS = [
  'mg', 'ml', 'g', 'drops', 'tablets', 'capsules', 'tsp', 'tbsp'
] as const

// Appointment types
export const APPOINTMENT_TYPES = [
  { value: 'well_visit', label: 'Well Visit', color: 'bg-green-500' },
  { value: 'sick_visit', label: 'Sick Visit', color: 'bg-orange-500' },
  { value: 'specialist', label: 'Specialist', color: 'bg-blue-500' },
  { value: 'dental', label: 'Dental', color: 'bg-purple-500' },
  { value: 'other', label: 'Other', color: 'bg-gray-500' },
] as const

// Duration options for appointments
export const DURATION_OPTIONS = [
  { value: 15, label: '15 minutes' },
  { value: 30, label: '30 minutes' },
  { value: 45, label: '45 minutes' },
  { value: 60, label: '1 hour' },
  { value: 90, label: '1.5 hours' },
  { value: 120, label: '2 hours' },
] as const
