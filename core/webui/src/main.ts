import { createApp } from "vue";
import "./style.css";
import App from "./App.vue";
import router from "./router";
import { createPinia } from "pinia";
import persistentPinia from "pinia-plugin-persistedstate";
import messages from "@/locales.json";
import { createI18n } from "vue-i18n";

const pinia = createPinia().use(persistentPinia); //🍍

const i18n = createI18n({
  legacy: false,
  locale: "pl",
  messages,
});

createApp(App).use(i18n).use(router).use(pinia).mount("#app");
