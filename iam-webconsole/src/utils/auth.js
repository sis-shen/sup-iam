import { STORAGE_KEYS } from './constants'

export function getAccessToken() {
  return localStorage.getItem(STORAGE_KEYS.ACCESS_TOKEN)
}

export function getRefreshToken() {
  return localStorage.getItem(STORAGE_KEYS.REFRESH_TOKEN)
}

export function setTokens(accessToken, refreshToken, expiresIn) {
  localStorage.setItem(STORAGE_KEYS.ACCESS_TOKEN, accessToken)
  if (refreshToken) {
    localStorage.setItem(STORAGE_KEYS.REFRESH_TOKEN, refreshToken)
  }
  if (expiresIn) {
    localStorage.setItem(STORAGE_KEYS.EXPIRES_IN, String(expiresIn))
  }
}

export function clearTokens() {
  localStorage.removeItem(STORAGE_KEYS.ACCESS_TOKEN)
  localStorage.removeItem(STORAGE_KEYS.REFRESH_TOKEN)
  localStorage.removeItem(STORAGE_KEYS.EXPIRES_IN)
  localStorage.removeItem(STORAGE_KEYS.USER_INFO)
}

export function isAuthenticated() {
  return !!getAccessToken()
}
