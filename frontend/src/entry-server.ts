import { renderToString } from '@vue/server-renderer'
import { makeApp } from './app'

async function render(ctx: any) {
    const { app, router, store, head } = makeApp()
    await router.push(ctx.url)

    const html = await renderToString(app, ctx)
    console.log(dumpObject(store))
    console.log(dumpObject(head))
    return html
}

(globalThis as any).ssrRender = render
