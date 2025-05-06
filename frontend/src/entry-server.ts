import { renderToString } from '@vue/server-renderer'
import { makeApp } from './app'
import createAxiosInstance from '@/utils/axios.js'

declare function dumpObject(obj: any): string;

async function doRenderToString(ctx: any) {
  //const { app, router, store, head } = makeApp()
  const { app, router, store } = makeApp()
  app.config.globalProperties.$fetchFn = createAxiosInstance(JSON.stringify(ctx.ssrHeaders))

  await router.push(ctx.url)

  const html = await renderToString(app, ctx)
  ctx.htmlState = store.state.value
  //console.log("head:", dumpObject(head))
  return html
}

(globalThis as any).v8goRenderToString = doRenderToString
