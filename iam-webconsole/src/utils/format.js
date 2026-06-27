/**
 * Format a date string or timestamp to a readable format.
 * @param {string|number|Date} date
 * @returns {string}
 */
export function formatDateTime(date) {
  if (!date) return '-'
  const d = new Date(date)
  if (isNaN(d.getTime())) return '-'
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  const hours = String(d.getHours()).padStart(2, '0')
  const minutes = String(d.getMinutes()).padStart(2, '0')
  const seconds = String(d.getSeconds()).padStart(2, '0')
  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
}

/**
 * Format a date to a short date string (YYYY-MM-DD).
 */
export function formatDate(date) {
  if (!date) return '-'
  const d = new Date(date)
  if (isNaN(d.getTime())) return '-'
  return d.toISOString().split('T')[0]
}

/**
 * Truncate a string with ellipsis.
 */
export function truncate(str, maxLen = 20) {
  if (!str) return ''
  return str.length > maxLen ? str.slice(0, maxLen) + '...' : str
}

/**
 * Mask a secret key for display (show first 4 and last 4 chars).
 */
export function maskKey(key) {
  if (!key) return ''
  if (key.length <= 8) return key
  return key.slice(0, 4) + '****' + key.slice(-4)
}
