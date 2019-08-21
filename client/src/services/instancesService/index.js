import baseClient from 'clients/baseClient'

const getInstances = () => baseClient.get('/resources/instances')

export default {
  getInstances,
}
