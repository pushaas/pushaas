import baseClient from 'clients/baseClient'

const getResources = () => baseClient.get('/resources')

export default {
  getResources,
}
