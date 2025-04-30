import { createApp, createSSRApp } from 'vue'
import App from './App.vue'
import { createAppRouter } from './router'
import { createPinia } from 'pinia'
import { createHead } from '@unhead/vue/client'
import './style.css'


export function makeApp() {
  const isServer = typeof window === 'undefined'
  const app = isServer ? createSSRApp(App) : createApp(App)
  const router = createAppRouter(isServer)
  const store = createPinia()
  const head = createHead()

  app.use(head)
  app.use(store)
  app.use(router)

  return {
    app,
    router,
    store,
    head,
  }
}
