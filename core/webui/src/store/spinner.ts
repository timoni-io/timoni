import { defineStore } from "pinia";

export const useSpinner = defineStore("spinner", {
  state: () => ({ spinner: false }),
});
