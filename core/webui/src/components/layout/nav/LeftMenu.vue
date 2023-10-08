<script setup lang="ts">
import { useI18n } from "vue-i18n";
import { useRoute, useRouter } from "vue-router";
import { useUserStore } from "@/store/userStore";

const { t } = useI18n();

const userStore = useUserStore();
const route = useRoute();
const router = useRouter();

// let userEmail = $ref("");
let logoutModal = $ref(false);

let activeKey = $computed(() => {
  if (route.name === "repository-files" || route.name === "commits")
    return "code";
  return route.path.split("/")[1];
});


let isLoginPage = $ref(window.location.pathname.startsWith("/login"));
watch(
  () => route.fullPath,
  () => {
    isLoginPage = route.fullPath.startsWith("/login")
    if (isLoginPage && route.fullPath.startsWith("/login/")) {
      router.push("/login");
    }
  }
);

const getUser = () => {
  if (!isLoginPage)
    api.get("/user-info").then((res) => {
      userStore.permissions = res.PermissionsGlobal;
      userStore.teams = res.Teams;
      userStore.userName = res.Name;
      userStore.email = res.Email;
      userStore.HideGitRepoLocal = res.HideGitRepoLocal;
    }).catch((e: any) => {
      if (
        e.message.split("received:").length &&
        (e.message.split("received:")[1] === '\n"`session` is invalid"' ||
          e.message.split("received:")[1] === '\n"`token` is invalid"' ||
          e.message.split("received:")[1] === '\n"`session` is expired"' ||
          e.message === "$c:31")
      ) {
        localStorage.removeItem('user-session');
        router.push("/login");
      }
    });
};

const logout = () => {
  // token.token = "";
  logoutModal = false;
  localStorage.removeItem("user-session");
  router.push("/login");
};
const menuOptions = computed(() => [
  {
    label: t("navbar.env"),
    key: "env",
    icon: mdiCloud,
    path: "/env",
    disabled: userStore.permissions && !userStore.havePermission("Env_View"),
  },
  {
    label: t("navbar.repo"),
    // label: t("navbar.repo"),
    key: "code",
    icon: mdiGit,
    path: "/code",
    disabled: userStore.permissions && !userStore.havePermission("Repo_View"),
  },
  {
    label: t("navbar.infra"),
    key: "infrastructure",
    path: "/infrastructure",
    icon: mdiServerNetwork,
    disabled: true,
  },
  {
    label: t("objects.member", 2),
    key: "users-list",
    disabled:
      userStore.permissions &&
      !userStore.havePermission("Glob_ManageGlobalMemebers"),
    path: "/users-list",
    icon: mdiAccount,
  },
  {
    label: t("fields.team", 2),
    key: "teams-list",
    disabled:
      userStore.permissions &&
      !userStore.havePermission("Glob_ManageGlobalMemebers"),
    path: "/teams-list",
    icon: mdiAccountGroup,
  },
  {
    label: "Platform",
    key: "admin",
    disabled:
      userStore.permissions &&
      !userStore.havePermission("Glob_AccessToAdminZone"),
    icon: mdiShieldCrown,
    path: "/admin/platform",
  },
  {
    label: userStore.userName,
    key: "user",
    icon: mdiAccount,
    path: "/user",
  },
]);
onMounted(() => {
  getUser();
});


const pushOnPath = (path: string) => {
  switch (path) {
    case "/env":
      if (userStore.permissions && !userStore.havePermission("Env_View"))
        return;
      else break;
    case "/users-list":
    case "/teams-list":
      if (
        userStore.permissions &&
        !userStore.havePermission("Glob_ManageGlobalMemebers")
      )
        return;
      else break;
    case "/code":
      if (userStore.permissions && !userStore.havePermission("Repo_View"))
        return;
      else break;
    case "/admin/platform":
      if (
        userStore.permissions &&
        !userStore.havePermission("Glob_AccessToAdminZone")
      )
        return;
      else break;
    case "/infrastructure":
      return;
  }
  router.push(path);
};
</script>

<template>
  <div class="navbar-left" v-if="!isLoginPage">
    <template v-for="option in menuOptions" :key="option.key">
      <button :class="
        option.key === activeKey
          ? 'link active-option'
          : option.disabled
            ? 'link disabled'
            : 'link'
      " @click="pushOnPath(option.path as string)">
        <div style="
              display: flex;
              flex-direction: column;
              align-items: center;
              width: 100%;
              padding: 0.5em 0;
            ">
          <n-icon size="20">
            <Mdi :path="option.icon || ''" />
          </n-icon>
          <p style="word-wrap: break-word; width: 90%">{{ option.label }}</p>
        </div>
      </button>
      <div style="flex-grow: 1" v-if="option.key === 'admin'"></div>
    </template>

    <button class="link" @click="logoutModal = true">
      <div style="width: 100%; padding: 0.5em 0">
        <n-icon size="20">
          <Mdi :path="mdiLogout" />
        </n-icon>
        <p>{{ t("home.logout") }}</p>
      </div>
    </button>
  </div>
  <Modal v-model:show="logoutModal" style="width: 20rem" :title="t('home.logout')" :touched="false" :showFooter="true"
    @positive-click="logout">
    <div>
      {{ t("questions.sure") }}
    </div>
  </Modal>
</template>
<style scoped>
.navbar-left {
  display: flex;
  flex-direction: column;
  min-width: 5rem;
  align-items: center;
  background: #fab900;
  margin-top: -15px;
  /* padding-top: 15px; */
  height: calc(100vh - 40px);
  width: 100px;
  position: relative;
  z-index: 2137;
}

.navbar-left .link {
  cursor: pointer;
  width: 100%;
  font-size: 12px;
  /* padding: 5px 0; */
  background: #fab900;
  border: none;
  text-align: center;
  color: black;
  text-decoration: none;
}

.navbar-left .link:hover {
  background-color: rgba(0, 0, 0, 0.103) !important;
  backdrop-filter: brightness(80%) contrast(140%);
}

.navbar-left .link.disabled {
  opacity: 0.5;
  cursor: default;
}

.active-option {
  position: relative;
  background-color: rgba(0, 0, 0, 0.103) !important;
  backdrop-filter: brightness(80%) contrast(140%);
}

.active-option::before {
  position: absolute;
  left: 0px;
  top: 0;
  width: 3px;
  height: 100%;
  background-color: black !important;
  content: "";
}
</style>
