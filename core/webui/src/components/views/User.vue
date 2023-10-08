<script setup lang="ts">
// import { useToken } from "@/store/token";
import { useI18n } from "vue-i18n";
import { useUserSettings } from "@/store/userSettings";
const { locale, t } = useI18n();
const userSettings = useUserSettings();
// const token = useToken();

const languages = $ref<"pl" | "en">(locale.value as "pl" | "en");
const changeLocale = () => {
  userSettings.lang = languages;
};

let themeDark = $ref<boolean>(userSettings.themeDark || true);
const changeLocaleThemeDark = () => {
  userSettings.themeDark = themeDark;
};

watchEffect(() => {
  if (!themeDark) {
    setTimeout(() => {
      themeDark = true;
      changeLocaleThemeDark();
    }, 100);
  }
});
const model = $ref({
  left: userSettings.color?.left,
  right: userSettings.color?.right,
});
const opacity = $ref(userSettings.opacity);
const updateColor = (val: string, side: "left" | "right") => {
  userSettings.color![side] = val;
};
const updateOpacity = (val: number) => {
  userSettings.opacity = val;
};
</script>

<template>
  <div>
    <NavbarTabs />
    <PageLayout>
      <n-card :title="t('navbar.user')">
        <!-- <template #header-extra>
          <n-button secondary type="error" @click="logout" size="small">
            {{ t("home.logout") }}
            <template #icon>
              <n-icon><mdi :path="mdiLogoutVariant" /></n-icon>
            </template>
          </n-button>
        </template> -->
        <div
          style="
            display: flex;
            align-items: center;
            gap: 1rem;
            margin-top: -15px;
          "
        >
          <div style="border-right: 1px solid grey; height: 60px">
            <p>{{ t("home.language") }}</p>
            <n-radio-group
              v-model:value="languages"
              @update:value="changeLocale"
            >
              <n-space item-style="display: flex;">
                <n-radio value="pl" label="polish" />
                <n-radio value="en" label="english" />
              </n-space>
            </n-radio-group>
          </div>
          <div
            style="
              padding-right: 15px;
              border-right: 1px solid grey;
              height: 60px;
            "
          >
            <p>{{ t("theme.title") }}</p>
            <n-switch
              v-model:value="themeDark"
              @update:value="changeLocaleThemeDark"
            >
              <template #checked>{{ t("theme.dark") }}</template>
              <template #unchecked>{{ t("theme.light") }}</template>
              <template #checked-icon>
                <n-icon>
                  <mdi :path="mdiWeatherNight" />
                </n-icon>
              </template>
              <template #unchecked-icon>
                <n-icon>
                  <mdi :path="mdiWhiteBalanceSunny" />
                </n-icon>
              </template>
            </n-switch>
          </div>
          <div
            style="
              height: 60px;
              border-right: 1px solid grey;
              padding-right: 15px;
            "
          >
            <p>
              <!-- {{ t("theme.title") }} -->
              Gradient
            </p>
            <n-color-picker
              v-model:value="model.left"
              :show-alpha="false"
              style="width: 100px; margin-right: 5px"
              @update:value="(color: string) => updateColor(color, 'left')"
            />
            <n-color-picker
              v-model:value="model.right"
              :show-alpha="false"
              style="width: 100px"
              @update:value="(color: string) => updateColor(color, 'right')"
            />
          </div>
          <div style="height: 60px">
            <p>
              {{ t("theme.opacity") }}
            </p>
            <n-slider
              v-model:value="opacity"
              @update:value="updateOpacity"
              :step="10"
              style="width: 10rem; margin-top: 10px"
            />
          </div>
        </div>
      </n-card>
    </PageLayout>
  </div>
</template>
<style scoped>
.first {
  border-right: 1px solid grey;
}
</style>
