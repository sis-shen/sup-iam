import request from './request'

export function getBindings(params) {
  return request.get('/bindings', { params })
}

export function getBinding(id) {
  return request.get(`/bindings/${id}`)
}

export function createBinding(data) {
  return request.post('/bindings', data)
}

export function deleteBinding(id) {
  return request.delete(`/bindings/${id}`)
}
