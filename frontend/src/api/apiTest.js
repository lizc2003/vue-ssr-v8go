export default {
  test(fetchFn) {
    if (import.meta.env.SSR) {
      return fetchFn('https://ifconfig.me/all.json', {}, { method: 'get' })
    } else {
      return fetchFn('/api/ifconfig/all.json', {}, { method: 'get' })
    }
  }
}
