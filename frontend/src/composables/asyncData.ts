import { getCurrentInstance, onServerPrefetch } from 'vue'

export async function useAsyncData<T>(
  store: any,
  key: string,
  asyncFn: (fetcher : any) => Promise<T>,
)  {
  const instance = getCurrentInstance()
  const fetcher = instance?.appContext.config.globalProperties.$fetcher

  if (import.meta.env.SSR) {
    onServerPrefetch(async () => {
      const data = await asyncFn(fetcher)
      store.setData(key, data)
    })
  } else {
    const initialData = store.getData(key)
    if (!initialData) {
      const data = await asyncFn(fetcher)
      store.setData(key, data)
    }
  }
}
