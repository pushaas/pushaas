import axios from 'axios'
import { toast } from 'react-toastify'

import credentialsService from 'services/credentialsService'

const baseClient = axios.create({
  baseURL: '/api/v1',
})

baseClient.interceptors.request.use((config) => ({
    ...config,
    auth: config.auth || credentialsService.getCredentials(),
  }), (error) => Promise.reject(error))

baseClient.interceptors.response.use(({ data }) => data, (error) => {
    let message = 'Unknown error'
    if (error.response.data && error.response.data.message) {
      message = `${error.response.data.message} (error code ${error.response.data.code})`
    }
    console.error(message)
    toast.error(message)
    return Promise.reject(error)
  })

export default baseClient
