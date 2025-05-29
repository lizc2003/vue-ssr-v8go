import { makeApp } from './app'
import { createAxiosInstance } from '@/utils/axios.ts'

const { app, router, store} = makeApp()

if (window.__INITIAL_STATE__) {
  store.state.value = window.__INITIAL_STATE__
}
app.config.globalProperties.$fetchFn = createAxiosInstance()

router.push(window.location.pathname)

app.mount('#app', true)
