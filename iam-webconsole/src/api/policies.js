import request from './request'

export function getPolicies(params) {
  return request.get('/policies', { params })
}

export function getPolicy(id) {
  return request.get(`/policies/${id}`)
}

export function createPolicy(data) {
  return request.post('/policies', data)
}

export function updatePolicy(id, data) {
  return request.put(`/policies/${id}`, data)
}

export function deletePolicy(id) {
  return request.delete(`/policies/${id}`)
}

export function getPolicySecrets(id, params) {
  return request.get(`/policies/${id}/secrets`, { params })
}
