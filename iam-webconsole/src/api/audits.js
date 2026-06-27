import request from './request'

export function getPolicyAudits(params) {
  return request.get('/audits/policies', { params })
}

export function getPolicyAudit(id) {
  return request.get(`/audits/policies/${id}`)
}

export function getBindingAudits(params) {
  return request.get('/audits/bindings', { params })
}

export function getBindingAudit(id) {
  return request.get(`/audits/bindings/${id}`)
}
