import { defineStore } from "pinia";

interface CounterState {
  count: any;
  ip: any;
  [key: string]: any;
}

export const useCounterStore = defineStore("counter", {
  state: (): CounterState => ({
    count: 5,
    ip: null,
  }),
  actions: {
    setData(key: string, data: any) {
      this[key] = data
    },
    getData(this: CounterState, key: string) {
      return this[key]
    },
    increment(this: CounterState) {
      this.count++;
    },
    decrement(this: CounterState): any {
      this.count--;
    },
  },
  getters: {
    doubleCount: (state: any) => state.count * 2,
  },
});
