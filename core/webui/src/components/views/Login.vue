<script setup lang="ts">
import { useRoute } from "vue-router";
import { useRouter } from "vue-router";
import { useMessage } from "naive-ui";
import { useUserStore } from "@/store/userStore";
import { nextTick } from "vue";
const userStore = useUserStore();
const message = useMessage();
const route = useRoute();
const router = useRouter();
let emailInput = $ref("");
let tokenInput = $ref("");
let step = $ref(1);
let msg = $ref("");
let email = $ref<HTMLInputElement>();
let token = $ref<HTMLInputElement>();

if (localStorage.getItem('user-session')!==null) {
  window.location.href = "/";
  // router.push("/");
}

const sendEmail = () => {

  nextTick(() => {
    email?.focus();
  });
  
  if (!emailInput) return;

  api
    .get("/user-login", {
      queries: {
        email: emailInput as string,
      },
    })
    .then((res) => {
      if (res.msg) {
        msg = res.msg;
        // qr = res.qr;
        step = 2;
        nextTick(() => {
          token?.focus();
        });
      }
      if (res.error && res.error !== "empty token") {
        message.error(res.error);
        return;
      }
      step = 2;
      nextTick(() => {
        token?.focus();
      });
    });
};





emailInput = route.query.email as string;
sendEmail();




const sendToken = () => {
  api
    .get("/user-login", {
      queries: {
        email: emailInput as string,
        token: tokenInput as string,
      },
    })
    .then((res) => {

      if (res.error) {
        message.error(res.error);
        token?.focus();
        return;
      }

      localStorage.setItem("user-session", res.session);
      api.get("/user-info").then((res) => {
        userStore.permissions = res.PermissionsGlobal;
        userStore.teams = res.Teams;
        userStore.email = res.Email;
        userStore.userName = res.Name;
        router.push("/env");
      });
    });
};
</script>
<template>
  <div
    style="
      display: flex;
      align-items: center;
      justify-content: center;
      height: 90vh;
    "
  >
    <n-card
      title=""
      style="max-width: 25rem"
      size="small"
      v-if="step === 1"
    >
      <n-input
        ref="email" 
        placeholder="Email"
        v-model:value="emailInput"
        @keydown.enter="sendEmail"
      />
      <n-button
        secondary
        type="primary"
        style="margin-top: 10px; float: right"
        @click="sendEmail"
      >
        Login
      </n-button>
    </n-card>
    <n-card
      title=""
      style="max-width: 25rem"
      size="small"
      v-if="step === 2"
    >
      <p style="margin-bottom: 1rem" v-if="msg" v-html="msg"></p>
      <n-input
        ref="token" 
        placeholder="Token"
        v-model:value="tokenInput"
        @keydown.enter="sendToken"
      />
      <n-button
        secondary
        type="primary"
        style="margin-top: 10px; float: right"
        @click="sendToken"
      >
        Login
      </n-button>
    </n-card>
  </div>
</template>
