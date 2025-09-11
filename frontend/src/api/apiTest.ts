export function apiTest(fetchFn:any) {
  return fetchFn('/all.json', {}, { method: 'get' })
}
