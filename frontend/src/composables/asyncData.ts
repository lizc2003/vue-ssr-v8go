import { getCurrentInstance, onServerPrefetch } from 'vue'

export async function useAsyncData<T>(
  store: any,
  key: string,
  asyncFn: (fetchFn : any) => Promise<T>,
)  {
  const instance = getCurrentInstance()
  const fetchFn = instance?.appContext.config.globalProperties.$fetchFn

  if (import.meta.env.SSR) {
    onServerPrefetch(async () => {
      const data = await asyncFn(fetchFn)
      store.setData(key, data)
    })
  } else {
    const initialData = store.getData(key)
    if (!initialData) {
      const data = await asyncFn(fetchFn)
      store.setData(key, data)
    }
  }
}
