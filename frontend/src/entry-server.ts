import { renderToString } from '@vue/server-renderer'
import { makeApp } from './app'
import { createAxiosInstance } from '@/utils/axios.ts'

declare function dumpObject(obj: any): string;

(globalThis as any).v8goRenderToString = async function (ctx: any) {
  //const { app, router, store, head } = makeApp()
  const { app, router, store } = makeApp()
  app.config.globalProperties.$fetchFn = createAxiosInstance(ctx)

  await router.push(ctx.url)

  const html = await renderToString(app, ctx)
  ctx.htmlState = store.state.value
  //console.log("head:", dumpObject(head))
  return html
}
