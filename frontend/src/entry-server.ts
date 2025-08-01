import { renderToString } from '@vue/server-renderer'
import { makeApp } from './app'
import { createAxiosInstance } from '@/utils/axios.ts'
import { renderSSRHead } from '@unhead/ssr'

declare function dumpObject(obj: any): string;

(globalThis as any).v8goRenderToString = async function (ctx: any) {
  const { app, router, store, head } = makeApp()

  await router.push(ctx.url)
  await router.isReady();

  if (router.currentRoute.value.matched.length === 0) {
    throw new Error("404 NotFound");
  }

  if (router.currentRoute.value.meta?.ssrOff) {
    throw new Error("ssr-off");
  }

  app.config.globalProperties.$fetchFn = createAxiosInstance(ctx)

  const html = await renderToString(app, ctx)
  const {headTags} = await renderSSRHead(head)

  ctx.htmlState = store.state.value
  ctx.htmlMeta = headTags
  if (ctx.modules && ctx.modules.size > 0) {
    ctx.htmlModules = JSON.stringify([...ctx.modules])
    ctx.modules = null
  }

  store.state.value = {}

  return html
}
