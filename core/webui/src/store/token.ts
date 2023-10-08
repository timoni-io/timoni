import { defineStore } from "pinia";

export const useToken = defineStore("token", {
  state: () => ({ token: "" }),
  persist: true,
});
