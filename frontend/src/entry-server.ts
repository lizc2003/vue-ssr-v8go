import { renderToString } from '@vue/server-renderer'
import { makeApp } from './app'

declare function dumpObject(obj: any): string;

async function doRenderToString(ctx: any) {
    //const { app, router, store, head } = makeApp()
    const { app, router, store } = makeApp()
    await router.push(ctx.url)

    const html = await renderToString(app, ctx)
    console.log("store:", dumpObject(store.state.value))
    //console.log("head:", dumpObject(head))
    return html
}

(globalThis as any).v8goRenderToString = doRenderToString
