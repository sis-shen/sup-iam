import request from './request'

export function login(data) {
  return request.post('/auth/login', data)
}

export function register(data) {
  return request.post('/auth/register', data)
}

export function logout() {
  return request.post('/auth/logout')
}

export function getMe() {
  return request.get('/auth/me')
}

export function refreshToken(data) {
  return request.post('/auth/refresh', data)
}

export function changePassword(data) {
  return request.post('/auth/password/change', data)
}
