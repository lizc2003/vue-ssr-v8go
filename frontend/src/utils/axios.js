import axios from 'axios'

function createAxiosInstance(ssrCtx) {
  const _headers = {
    'Content-Type': 'application/json',
    'Accept': 'application/json'
  }
  if (ssrCtx) {
    _headers['SSR-Headers'] = JSON.stringify(ssrCtx.ssrHeaders)
  }

  const instance = axios.create({
    headers: _headers,
    timeout: 10000,
  })

  instance.interceptors.request.use((config) => {
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

  return async function (url, data = {}, options = {}) {
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

export default createAxiosInstance
