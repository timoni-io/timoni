<script setup lang="ts">
import { darkTheme, lightTheme } from "naive-ui";
import ErrorDialog from "@/components/layout/error/ErrorDialog.vue";
import ErrorDialogEnv from "@/components/layout/error/ErrorDialogEnv.vue";
import { useThemeCssVars } from "@/utils/styles";
import useDefaultFontSize from "@/utils/defaultFontSize";
import useErrorMsg from "@/utils/errorMsg";
import useErrorMsgEnv from "@/utils/errorMsgEnv";
import { onBeforeMount } from "vue";
import { useI18n } from "vue-i18n";
import { useUserSettings } from "@/store/userSettings";
import chroma from "chroma-js";
import { useRoute } from "vue-router";
import { useEnv } from "@/store/envStore";
import { useRepo } from "@/store/repoStore";
// import { useMessage } from "naive-ui";

// const message = useMessage();
const { locale, t } = useI18n();
const userSettings = useUserSettings();
const route = useRoute();
watchEffect(() => {
  locale.value = userSettings.lang;
});

const env = useEnv(computed(() => route.params.id as string));
const repo = $(useRepo(computed(() => route.params.name as string)));

let left = $ref(userSettings.color!.left);
let right = $ref(userSettings.color!.right);
let opacity = $ref(userSettings.opacity / 100);
let errorMsg = useErrorMsg().errorMsg;
let errorMsgEnv = useErrorMsgEnv().errorMsgEnv;

let collapsed = $ref(true);
watch(
  () => [
    userSettings.color!.left,
    userSettings.color!.right,
    userSettings.opacity,
  ],
  () => {
    left = userSettings.color!.left;
    right = userSettings.color!.right;
    opacity = userSettings.opacity / 100;
  }
);
watch(
  () => collapsed,
  () => {
    localStorage.setItem("collapsed", collapsed.toString());
  }
);

onBeforeMount(() => {
  let localCollapsed = localStorage.getItem("collapsed");
  if (localCollapsed === "false") {
    collapsed = false;
  }
});

// const primaryColor = chroma(`#d01bfd`);
const primaryColor = chroma(`#1ba3fd`);
const accentColor = chroma("#fab900");
const backgroundColor = chroma("#101014");
// const cardColor = chroma("#232e3c");
const cardColor = backgroundColor.brighten(0.5);
// const primaryColor = chroma(`#ffffff`);
// const primaryColor = chroma(`#ffffff`);

useThemeCssVars({
  primaryColor: primaryColor.hex(),
  primaryColorHover: primaryColor.brighten().hex(),
  primaryColorPressed: primaryColor.darken().hex(),
  primaryColorSuppl: primaryColor.darken(2).hex(),
  itemTextColorHover: primaryColor.darken(2).hex(),
  accentColor: accentColor.hex(),
  accentColorDarken: accentColor.darken().hex(),
  cardColor: cardColor.alpha(opacity).hex(),
  cardColorOpacity: cardColor.alpha(0.93).hex(),
  scrollThumbColor: backgroundColor.alpha(0.5).hex(),
});
const { fontSize } = useDefaultFontSize();

let showFilters = $ref(false);
let showAdminPanel = $ref(false);

provide("showFilters", $$(showFilters));
provide("showAdminPanel", $$(showAdminPanel));

watch(
  [() => route.fullPath, () => env.value?.EnvInfo?.Env?.Name, () => repo?.Name],
  () => {
    let tabName: string;
    let routePath = route.fullPath.split("/");
    switch (true) {
      case routePath[1] === "env" && routePath.length < 3:
        tabName = `${t("objects.env")} | Timoni`;
        break;
      case routePath[1] === "env":
        tabName = `${t("objects.env")} > ${
          env.value?.EnvInfo?.Env?.Name || ""
        } | Timoni`;
        break;
      case routePath[1] === "code" && routePath.length < 3:
        tabName = "Git Repo | Timoni";
        break;
      case routePath[1] === "code":
        tabName = `Git Repo > ${repo?.Name || ""} | Timoni`;
        break;
      case routePath[1] === "users-list":
        tabName = `${t("objects.member", 2)} | Timoni`;
        break;
      case routePath[1] === "teams-list":
        tabName = `${t("fields.team", 2)} | Timoni`;
        break;
      case routePath[1] === "user":
        tabName = `SuperUser | Timoni`;
        break;
      case route.fullPath === "/admin/platform":
        tabName = `${t("adminPanel.platform")} | Timoni`;
        break;
      case routePath[1] === "admin":
        tabName = "Admin | Timoni";
        break;
      default:
        tabName = "Timoni";
    }
    window.document.title = tabName;
  }
);
</script>

<template>
  <n-config-provider
    :theme="userSettings.themeDark ? darkTheme : lightTheme"
    :theme-overrides="{
      common: {
        fontSize: `${fontSize}px`,
        fontSizeMini: `${fontSize - 2}px`,
        fontSizeTiny: `${fontSize - 2}px`,
        fontSizeSmall: `${fontSize}px`,
        fontSizeMedium: `${fontSize}px`,
        // fontSizeLarger: `${fontSize + 1}px`,
        fontSizeHuge: `${fontSize + 2}px`,
        primaryColor: primaryColor.hex(),
        primaryColorHover: primaryColor.brighten().hex(),
        primaryColorPressed: primaryColor.darken().hex(),
        primaryColorSuppl: primaryColor.darken(2).hex(),
        // itemTextColorHover: primaryColor.darken(2).hex(),
        fontFamily: `'Poppins', sans-serif`,
      },
      DataTable: {
        tdColor: chroma('black').alpha(0.45).hex(),
      },

      Card: {
        titleFontSizeSmall: `${fontSize + 4}px`,
        titleFontSizeMedium: `${fontSize + 4}px`,
        titleFontSizeLarge: `${fontSize + 4}px`,
        titleFontSizeHuge: `${fontSize + 4}px`,
        borderColor: `transparent`,
        color: cardColor.alpha(opacity).hex(),
      },
      Tabs: {
        tabFontSizeSmall: `${fontSize}px`,
      },
      Table: {
        fontSizeSmall: `${fontSize - 2}px`,
      },
      Menu: {
        itemTextColor: backgroundColor.hex(),
        itemIconColor: backgroundColor.hex(),
        itemIconColorHover: backgroundColor.hex(),
        itemIconColorCollapsed: backgroundColor.hex(),
        itemIconColorActive: accentColor.hex(),
        // itemIconColorActive: backgroundColor.hex(),
        itemColorActive: backgroundColor.hex(),
        itemTextColorActive: accentColor.hex(),
        itemTextColorHover: backgroundColor.hex(),
        itemColorHover: accentColor.darken().hex(),
        itemColorActiveHover: backgroundColor.hex(),
        itemColorActiveCollapsed: backgroundColor.hex(),
        // itemColorActiveCollapsedHover: backgroundColor.brighten().hex(),
        itemTextColorActiveHover: accentColor.hex(),
        itemIconColorActiveHover: accentColor.hex(),
      },
      Layout: {
        siderBorderColor: `transparent`,
        // backgroundColor: `transparent`,
      },
      // Input: {
      //   borderFocus: `1px solid ${primaryColor.brighten()}`,
      //   borderHover: `1px solid ${primaryColor.brighten()}`,
      // },
    }"
  >
    <n-message-provider placement="bottom" :duration="7000">
      <NetworkStatus />
      <n-layout has-sider style="height: 100vh" class="container">
        <!-- <n-layout-sider
          v-if="route.path.split('/')[1] !== 'login'"
          bordered
          show-trigger
          :collapsed="collapsed"
          collapse-mode="width"
          :collapsed-width="64"
          :width="240"
          :native-scrollbar="false"
          style="height: 100vh"
          @collapse="collapsed = true"
          @expand="collapsed = false"
        >
          <NavbarLeft />
        </n-layout-sider> -->

        <n-layout-content
          content-style="padding: 0px; height: 100vh"
          id="layout-content"
        >
          <div style="height: 100vh; max-height: 100vh; overflow-y: hidden">
            <div class="app" style="padding-top: 55px">
              <Navbar />
              <div style="display: flex">
                <LeftMenu />
                <div style="width: calc(100% - 100px)">
                  <router-view v-slot="{ Component, route }">
                    <GlobalSpinner :data="Component">
                      <transition name="fade">
                        <component :is="Component" :key="route.path" />
                      </transition>
                    </GlobalSpinner>
                  </router-view>
                </div>
              </div>
            </div>
          </div>
        </n-layout-content>
      </n-layout>
      <!-- <Navbar class="fixed-navbar" /> -->
      <ErrorDialog v-if="errorMsg" />
      <ErrorDialogEnv v-if="errorMsgEnv" />
    </n-message-provider>
    <SmallSize />
    <n-global-style />
  </n-config-provider>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: all 0.2s ease;
  position: absolute;
  left: 0;
  right: 0;
}
.container {
  background: linear-gradient(30deg, v-bind(left) 0%, v-bind(right) 100%);
}
.fade-leave-to {
  opacity: 0;
}
</style>

<style lang="scss">
* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
  // color-scheme: dark;
  --navbar-h: 4.5rem;

  &::-webkit-scrollbar {
    width: 0.7rem;
  }
  &::-webkit-scrollbar-track {
    box-shadow: inset 0 0 6px rgba(0, 0, 0, 0.3);
  }
  &::-webkit-scrollbar-thumb {
    background-color: var(--scrollThumbColor);
    // outline: 1px solid slategrey;
  }
}
.clicked {
  transform: rotate(360deg);
  transition: transform 0.5s ease-in-out;
}
.n-button.n-button--primary-type {
  --n-color-hover: rgba(27, 163, 253, 0.5) !important;
}
.n-button.n-button--error-type {
  --n-color-hover: rgba(232, 128, 128, 0.5) !important;
}
.monaco-editor,
.monaco-editor-background,
.monaco-editor .margin {
  background-color: #0000002c !important;
}

.monaco-editor .view-overlays .current-line {
  background-color: #0000002c !important;
  border-color: transparent !important;
  border-radius: 4px;
}
// body {
//   background-image: linear-gradient(300deg, rgb(24, 24, 24), rgb(0, 24, 36));
// }

.n-layout {
  background-color: transparent !important;
}

.fixed-navbar {
  position: sticky;
  top: 0;
  width: 100%;
  z-index: 2;
  // height: var(--navbar-h);
}

.n-layout-toggle-button {
  top: calc(100% - 2rem) !important;
}

.n-card {
  -webkit-box-shadow: 0px 0px 14px 0px #00000098;
  -moz-box-shadow: 0px 0px 14px 0px #00000098;
  box-shadow: 0px 0px 14px 0px #00000098;
}

.admin-cards {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  // margin: 0 20px 2rem;

  & .n-card-header__main {
    font-size: 1rem;
  }
}
.base-alert {
  color: #777;
  padding: 0.25rem 0;
}
.n-modal-mask {
  backdrop-filter: blur(1px);
}

.n-card > .n-card-header .n-card-header__main {
  font-size: 0.8rem;
  color: #777;
}

.panel-header {
  font-size: 0.8rem;
  color: #777;
}
.listing-card .n-card {
  box-shadow: none;
  border: none;
}

.dots-button {
  margin-right: 10px;
  padding: 0 5px;
}
.dots-button span {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 13px;
}

.dynamic-table {
  overflow-y: auto;
}
.dynamic-table thead th {
  position: sticky;
  background-color: var(--n-color-modal) !important;
  z-index: 1000;
  top: 0;
}
</style>
