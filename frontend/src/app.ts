import { createApp, createSSRApp } from 'vue'
import App from './App.vue'
import { createRouter } from './router'
import { createPinia } from 'pinia'
import { createHead } from '@unhead/vue/client'
import './style.css'


export function makeApp() {
  const app = import.meta.env.SSR ? createSSRApp(App) : createApp(App)
  const router = createRouter()
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
