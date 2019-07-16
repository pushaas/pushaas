const KEY = 'push-api.credentials'

const clearCredentials = () => {
  localStorage.removeItem(KEY)
}

const getCredentials = () => {
  return JSON.parse(localStorage.getItem(KEY))
}

const setCredentials = (credentials) => {
  localStorage.setItem(KEY, JSON.stringify(credentials))
}

export default {
  clearCredentials,
  getCredentials,
  setCredentials,
}
