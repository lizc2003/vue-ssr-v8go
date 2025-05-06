export default {
  test(fetcher) {
    if (import.meta.env.SSR) {
      return fetcher.fetch('https://ifconfig.me/all.json', {}, { method: 'get' })
    } else {
      return fetcher.fetch('/api/ifconfig/all.json', {}, { method: 'get' })
    }
  }
}
