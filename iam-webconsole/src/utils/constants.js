// API base path - used by axios, proxied by nginx to backend
export const API_BASE_URL = '/api/v1'

// Local storage keys
export const STORAGE_KEYS = {
  ACCESS_TOKEN: 'access_token',
  REFRESH_TOKEN: 'refresh_token',
  EXPIRES_IN: 'expires_in',
  USER_INFO: 'user_info',
}

// Pagination defaults
export const DEFAULT_PAGE_SIZE = 20
export const MAX_PAGE_SIZE = 100

// Route meta constants
export const ROUTE_META = {
  REQUIRES_AUTH: 'requiresAuth',
  REQUIRES_ADMIN: 'requiresAdmin',
}
