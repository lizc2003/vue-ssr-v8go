import axios from 'axios'

function createAxiosInstance(ssrHeaders) {
  const instance = axios.create({
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json'
    },
    timeout: 10000,
  })

  instance.interceptors.request.use((config) => {
    if (ssrHeaders) {
      config.headers['SSR-Headers'] = ssrHeaders
    }
    return config
  })

  instance.interceptors.response.use((response) => {
    const res = response?.data ?? {}
      if (res.code === 0) {
        res.code = 200
      }
      return res
    },
    (error) => {
      console.error('axios response error:', error)
      return Promise.reject(error)
    }
  )

  return {
    async fetch(url, data = {}, options = {}) {
      const {method = 'get', headers = {}, ...rest} = options

      const config = {
        url,
        method,
        headers,
        ...rest
      }

      if (method.toLowerCase() === 'get') {
        config.params = data
      } else {
        config.data = data
      }

      return instance(config)
    }
  }
}

export default createAxiosInstance
