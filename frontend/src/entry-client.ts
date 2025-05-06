import { makeApp } from './app'

const { app, router, store} = makeApp()
if (window.__INITIAL_STATE__) {
    store.state.value = window.__INITIAL_STATE__
}

router.push(window.location.pathname)

app.mount('#app', true)
