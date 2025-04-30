import { defineStore } from "pinia"

export const useCounterStore = defineStore("counter", {
  state: () => ({
    count: 5,
    ip: {}
  }),
  actions: {
    increment() {
      this.count++
    },
    decrement() {
      this.count--
    },
  },
  getters: {
    doubleCount: (state) => state.count * 2,
  },
})
