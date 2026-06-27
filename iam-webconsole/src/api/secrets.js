import request from './request'

export function getSecrets(params) {
  return request.get('/secrets', { params })
}

export function getSecret(id) {
  return request.get(`/secrets/${id}`)
}

export function createSecret(data) {
  return request.post('/secrets', data)
}

export function updateSecret(id, data) {
  return request.put(`/secrets/${id}`, data)
}

export function deleteSecret(id) {
  return request.delete(`/secrets/${id}`)
}

export function rotateSecret(id) {
  return request.put(`/secrets/${id}/rotate`)
}

export function getSecretPolicies(id, params) {
  return request.get(`/secrets/${id}/policies`, { params })
}
