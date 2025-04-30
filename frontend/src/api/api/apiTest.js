import request from "@/api/axios.js"

export default {
  test() {
    if (import.meta.env.SSR) {
      return request('https://ifconfig.me/all.json', {}, { method: 'get' })
    } else {
      return request('/api/ifconfig/all.json', {}, { method: 'get' })
    }
  }
}
