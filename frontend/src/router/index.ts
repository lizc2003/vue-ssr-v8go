import { createRouter, createWebHistory, createMemoryHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import Home from '../pages/home/index.vue'
import Test from '../pages/test/index.vue'

const routes: RouteRecordRaw[] = [
  { path: '/', component: Home },
  { path: '/test', component: Test },
  {
    path: '/:pathMatch(.*)*',
    component: () => import('../pages/404.vue')
  }
]

export function createAppRouter(isServer: boolean) {
  return createRouter({
    history: isServer ? createMemoryHistory() : createWebHistory(),
    routes
  })
}
