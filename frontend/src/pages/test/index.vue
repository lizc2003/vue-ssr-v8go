<template>
  <div class="container">
    <h1 class="title">Test Page</h1>
    <h2 class='nav-button' @click="goToHome" style="cursor: pointer">GO HOME</h2>
    <p>Count: {{ counter.count }}</p>
    <div class="info">XMLHttpRequest Response:</div>
    <div v-if="counter.ip">
      <div v-for="(value, key) in counter.ip" :key="key" class="ip-info">
        <span class="ip-key">{{ key }}:</span>
        <span class="ip-value">{{ value }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue';
import { useRouter } from 'vue-router';
import { useCounterStore } from '@/stores/counter'
import axios from 'axios'
import APItest from '@/api/api/apiTest.js'
import { useSeoMeta } from '@unhead/vue'

useSeoMeta({
  title: 'Test',
  description: 'My Test page',
  ogDescription: 'Still about my about page',
  ogTitle: 'About',
  ogImage: 'https://example.com/image.png',
  twitterCard: 'summary_large_image',
})

const counter = useCounterStore()
const router = useRouter();
const ip = ref('')

const goToHome = () => {
  router.push('/');
};

const getIpInfo = async () => {
  try {
    const response = await APItest.test()
    counter.ip = response
  } catch (error) {
    console.error('xmlhttprequest fail:', error)
  }
}

if(!counter?.ip?.ip_addr){
  console.log('http request for ip now.')
  getIpInfo()
}

</script>

<style scoped>
.container {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  background-color: #1a1a1a;
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
.info {
  font-size: 16px;
  font-weight: bold;
  margin-bottom: 10px;
  text-align: left;
}

.ip-info{
  text-align: left;
  max-width: 1200px;
}
.nav-button:hover {
  background-color: #61dafb;
  color: #1a1a1a;
  transform: translateY(-2px);
}
</style>
