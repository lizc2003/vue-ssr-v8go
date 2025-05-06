import { onServerPrefetch } from 'vue'

export async function useAsyncData<T>(
  store: any,
  key: string,
  asyncFn: () => Promise<T>,
)  {
  if (import.meta.env.SSR) {
    onServerPrefetch(async () => {
      const data = await asyncFn()
      store.setData(key, data)
    })
  } else {
    const initialData = store.getData(key)
    if (!initialData) {
      const data = await asyncFn()
      store.setData(key, data)
    }
  }
}
