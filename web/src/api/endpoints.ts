export const API = {
  // Auth
  AUTH: {
    GOOGLE: '/api/auth/google',
    REFRESH: '/api/auth/refresh',
    ME: '/api/auth/me',
  },

  // Family
  FAMILIES: {
    LIST: '/api/families',
    CREATE: '/api/families',
    GET: (id: string) => `/api/families/${id}`,
    CHILDREN: (familyId: string) => `/api/families/${familyId}/children`,
    CHILD: (familyId: string, childId: string) =>
      `/api/families/${familyId}/children/${childId}`,
    INVITE: (familyId: string) => `/api/families/${familyId}/invite`,
  },

  // Feeding
  FEEDING: {
    LIST: '/api/feeding',
    CREATE: '/api/feeding',
    GET: (id: string) => `/api/feeding/${id}`,
    UPDATE: (id: string) => `/api/feeding/${id}`,
    DELETE: (id: string) => `/api/feeding/${id}`,
    LAST: (childId: string) => `/api/feeding/last/${childId}`,
  },

  // Sleep
  SLEEP: {
    LIST: '/api/sleep',
    CREATE: '/api/sleep',
    GET: (id: string) => `/api/sleep/${id}`,
    UPDATE: (id: string) => `/api/sleep/${id}`,
    DELETE: (id: string) => `/api/sleep/${id}`,
    START: '/api/sleep/start',
    END: (id: string) => `/api/sleep/${id}/end`,
    ACTIVE: (childId: string) => `/api/sleep/active/${childId}`,
  },

  // Medications
  MEDICATIONS: {
    LIST: '/api/medications',
    CREATE: '/api/medications',
    GET: (id: string) => `/api/medications/${id}`,
    UPDATE: (id: string) => `/api/medications/${id}`,
    DELETE: (id: string) => `/api/medications/${id}`,
    LOG: '/api/medications/log',
    LOGS: (id: string) => `/api/medications/${id}/logs`,
    LAST_LOG: (id: string) => `/api/medications/${id}/logs/last`,
  },

  // Vaccinations
  VACCINATIONS: {
    LIST: '/api/vaccinations',
    CREATE: '/api/vaccinations',
    GET: (id: string) => `/api/vaccinations/${id}`,
    RECORD: (id: string) => `/api/vaccinations/${id}/record`,
    UPCOMING: (childId: string) => `/api/vaccinations/upcoming/${childId}`,
    SCHEDULE: '/api/vaccinations/schedule',
  },

  // Appointments
  APPOINTMENTS: {
    LIST: '/api/appointments',
    CREATE: '/api/appointments',
    GET: (id: string) => `/api/appointments/${id}`,
    UPDATE: (id: string) => `/api/appointments/${id}`,
    DELETE: (id: string) => `/api/appointments/${id}`,
    UPCOMING: (childId: string) => `/api/appointments/upcoming/${childId}`,
  },

  // Notes
  NOTES: {
    LIST: '/api/notes',
    CREATE: '/api/notes',
    GET: (id: string) => `/api/notes/${id}`,
    UPDATE: (id: string) => `/api/notes/${id}`,
    DELETE: (id: string) => `/api/notes/${id}`,
    SEARCH: '/api/notes/search',
  },

  // Sync
  SYNC: {
    PUSH: '/api/sync/push',
    PULL: '/api/sync/pull',
    STATUS: '/api/sync/status',
  },
} as const
