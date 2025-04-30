import axios from 'axios'

const instance = axios.create({
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
    'Accept': 'application/json'
  }
})

instance.interceptors.request.use(
  async (config) => {
    return config
  },
  (error) => {
    console.error('axios req error:', error)
    return Promise.reject(error)
  }
)

instance.interceptors.response.use(
  (response) => {
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

const request = (url, data = {}, options = {}) => {
  const { method = 'get', headers = {}, ...rest } = options

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

export default request
