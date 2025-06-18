# Vue SSR powered by v8go

A High-performance Server-side Rendering (SSR) Framework for Vue Apps, Powered by V8 and Golang.

## How to run

First, compile and launch the server:
```bash
make
./vue-ssr-v8go -config conf-dev.toml
```

Then, visit the test page using curl or a browser:
```bash
curl http://localhost:9090/test
```

## Used Packages

- [v8go](https://github.com/tommie/v8go) - Go bindings for V8 JavaScript engine
- [Vue](https://vuejs.org/) - The Progressive JavaScript Framework | Vue.js
- [Vite](https://vite.dev/) - Next Generation Frontend Tooling ｜ Vite

## Pros and Cons

Pros:
- No need to use Node.js, powered by V8 and Golang.
- Thread-safe V8 scheduler with dynamic scaling (1-50+ isolates)
- Prevents memory leaks via lifetime-controlled isolation (0-3600s+).
- An enhanced and optimized XMLHttpRequest is implemented, with setTimeout and setInterval removed, thus bettering SSR.
- Automatically fall back to client-side rendering when server-side rendering encounters an error.

Cons:
- You can't use Vite features like hot module replacement (HMR), but you can run `npm run watch` to build on-the-fly JavaScript scripts.


## Keynotes

### Separate config for Vite Build

The default config is configured in [`vite.config.ts`](frontend/vite.config.ts) and the additional configuration for the SSR build is located in [`vite.config.server.ts`](frontend/vite.config.server.ts).
SSR config builds the frontend application with target `webworker` as it is required for the V8 engine.

### Entry point for SSR and Client are different

See:
- Client-side entry: [`frontend/src/entry-client.ts`](frontend/src/entry-client.ts)
- Server-side entry: [`frontend/src/entry-server.ts`](frontend/src/entry-server.ts)

