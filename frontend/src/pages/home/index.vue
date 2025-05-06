<template>
  <div class="container">
    <h1 class="title">Hello SSR</h1>
    <button class="nav-button" @click="goToTest">GO TEST</button>
    <p>Count: {{ counter.count }}</p>
    <div class='flex button-box'>
      <button @click="counter.increment()">+</button>
      <button @click="counter.decrement()">-</button>
    </div>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router';
import { useCounterStore } from '@/stores/counter'
import { useSeoMeta } from '@unhead/vue'

useSeoMeta({
  title: 'Home',
  description: 'My Home page',
  ogDescription: 'Still about my about page',
  ogTitle: 'About',
  ogImage: 'https://example.com/image.png',
  twitterCard: 'summary_large_image',
})

const counter = useCounterStore()
const router = useRouter();

const goToTest = () => {
  router.push('/test');
};
</script>

<style lang='scss' scoped>
.container {
  width: 100%;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  color: #ffffff;
}

.title {
  font-size: 2.5rem;
  margin-bottom: 2rem;
  color: #61dafb;
}

.nav-button {
  padding: 12px 24px;
  font-size: 1.1rem;
  background-color: #2c2c2c;
  color: #ffffff;
  border: 2px solid #61dafb;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.nav-button:hover {
  background-color: #61dafb;
  color: #1a1a1a;
  transform: translateY(-2px);
}

.button-box{
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 10px;
  button{
    background-color: #aaa;
  }
}

</style>
