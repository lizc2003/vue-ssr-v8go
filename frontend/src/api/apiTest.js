export default {
  test(fetcher) {
    if (import.meta.env.SSR) {
      return fetcher('https://ifconfig.me/all.json', {}, { method: 'get' })
    } else {
      return fetcher('/api/ifconfig/all.json', {}, { method: 'get' })
    }
  }
}
