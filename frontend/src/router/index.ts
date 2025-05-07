import { createRouter as _createRouter, createWebHistory, createMemoryHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import Home from '@/pages/home/index.vue'
import Test from '@/pages/test/index.vue'

const routes: RouteRecordRaw[] = [
  { path: '/', component: Home },
  { path: '/test', component: Test },
]

export function createRouter() {
  return _createRouter({
    history: import.meta.env.SSR ? createMemoryHistory() : createWebHistory(),
    routes
  })
}
