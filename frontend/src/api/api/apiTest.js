import request from "@/api/axios.js"

export default {
  test() {
    return request('/api/ifconfig/all.json', {}, { method: 'get' })
  }
}