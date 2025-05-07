import { renderToString } from '@vue/server-renderer'
import { makeApp } from './app'
import { createAxiosInstance } from '@/utils/axios.ts'
import { renderSSRHead } from '@unhead/ssr'

declare function dumpObject(obj: any): string;

(globalThis as any).v8goRenderToString = async function (ctx: any) {
  const { app, router, store, head } = makeApp()
  app.config.globalProperties.$fetchFn = createAxiosInstance(ctx)

  await router.push(ctx.url)

  const html = await renderToString(app, ctx)
  const {headTags} = await renderSSRHead(head)

  ctx.htmlState = store.state.value
  ctx.htmlMeta = headTags
  ctx.preloadLinks = renderPreloadLinks(ctx.modules, ctx.manifest);
  return html
}

function renderPreloadLinks(modules: any, manifest: any): string {
  if (!manifest) {
    return '';
  }
  let links = '';
  const seen = new Set();
  modules.forEach((id: string) => {
    const files = manifest[id];
    if (files) {
      files.forEach((file: string) => {
        if (!seen.has(file)) {
          seen.add(file);
          const filename = _basename(file);
          if (manifest[filename]) {
            for (const depFile of manifest[filename]) {
              links += renderPreloadLink(depFile);
              seen.add(depFile);
            }
          }
          links += renderPreloadLink(file);
        }
      });
    }
  });
  return links;
}

function renderPreloadLink(file: string): string {
  if (file.endsWith('.js')) {
    return `<link rel="modulepreload" crossorigin href="${file}">`;
  } else if (file.endsWith('.css')) {
    return `<link rel="stylesheet" href="${file}">`;
  } else if (file.endsWith('.woff')) {
    return ` <link rel="preload" href="${file}" as="font" type="font/woff" crossorigin>`;
  } else if (file.endsWith('.woff2')) {
    return ` <link rel="preload" href="${file}" as="font" type="font/woff2" crossorigin>`;
  } else if (file.endsWith('.gif')) {
    return ` <link rel="preload" href="${file}" as="image" type="image/gif">`;
  } else if (file.endsWith('.jpg') || file.endsWith('.jpeg')) {
    return ` <link rel="preload" href="${file}" as="image" type="image/jpeg">`;
  } else if (file.endsWith('.png')) {
    return ` <link rel="preload" href="${file}" as="image" type="image/png">`;
  } else {
    return '';
  }
}

function _basename(str: string): string {
  return str.substring(str.lastIndexOf('/') + 1);
}
