import { defineStore } from "pinia";

interface CounterState {
  count: any;
  ip: Record<string, any>;
}

export const useCounterStore = defineStore("counter", {
  state: (): CounterState => ({
    count: 5,
    ip: {},
  }),
  actions: {
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
