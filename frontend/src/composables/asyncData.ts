import { getCurrentInstance, onServerPrefetch } from 'vue'

export function useAsyncData<T>(
  store: any,
  key: string,
  asyncFn: () => Promise<T>,
  options: {
    serverOnly?: boolean
    lazy?: boolean
  } = {}
) {
  const instance = getCurrentInstance()

  // 服务端预取
  if (import.meta.env.SSR) {
    onServerPrefetch(async () => {
      const promise = asyncFn()
      store.addPromise(key, promise)
      const data = await promise
      store.setData(key, data)
    })
  }

  // 客户端处理
  if (!import.meta.env.SSR) {
    const initialData = store.getData<T>(key)

    if (initialData) {
      // 有服务端数据直接使用
      return { data: initialData, refresh }
    }

    if (!options.lazy) {
      // 立即执行客户端请求
      return refresh()
    }
  }

  async function refresh() {
    if (!options.serverOnly || import.meta.env.SSR) {
      const data = await asyncFn()
      store.setData(key, data)
      return data
    }
    return null
  }

  return { data: undefined, refresh }
}
