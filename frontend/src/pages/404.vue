<template>
  <div class="not-found">
    <h1>404</h1>
    <h2>Page Not Found</h2>
    <button @click="goHome" class="home-button">GO HOME</button>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'

const router = useRouter()

const goHome = () => {
  router.push('/')
}
</script>

<style scoped>
.not-found {
  height: 100vh;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  background-color: #1a1a1a;
  color: #ffffff;
}

h1 {
  font-size: 8rem;
  margin: 0;
  color: #61dafb;
}

h2 {
  font-size: 2rem;
  margin: 1rem 0;
}

p {
  font-size: 1.2rem;
  margin-bottom: 2rem;
}

.home-button {
  padding: 12px 24px;
  font-size: 1.1rem;
  background-color: #2c2c2c;
  color: #ffffff;
  border: 2px solid #61dafb;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.home-button:hover {
  background-color: #61dafb;
  color: #1a1a1a;
}
</style>
