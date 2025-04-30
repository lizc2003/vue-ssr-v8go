import { makeApp } from './app'

const { app, router} = makeApp()
router.push(window.location.pathname)

app.mount('#app', true)
