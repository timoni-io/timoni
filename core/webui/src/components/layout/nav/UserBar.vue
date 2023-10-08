<script setup lang="ts">
import { useI18n } from "vue-i18n";
import { useUserSettings } from "@/store/userSettings";
import useDefaultFontSize from "@/utils/defaultFontSize";

const { t } = useI18n();
const settings = useUserSettings();

const { fontSize, setFontSize } = useDefaultFontSize();

// change font size
let defaultFontSize = $ref(fontSize);
const formatTooltip = (value: number) => `${value}px`;
const updateFontSize = (size: number) => {
  setFontSize(size);
};

onMounted(() => {
  let fontSize = localStorage.getItem("defaultFontSize");
  if (fontSize) {
    defaultFontSize = parseInt(fontSize);
  }
});
</script>

<template>
  <Dropdown>
    <template #trigger>
      <n-avatar class="user-avatar" :size="22"> A </n-avatar>
    </template>
    <template #content>
      <div class="drop-down">
        <n-form-item-row :label="t('home.language')">
          <n-radio-group v-model:value="settings.$state.lang" name="radiogroup">
            <n-space>
              <n-radio key="pl" value="pl" label="polski" />
              <n-radio key="en" value="en" label="english" />
            </n-space>
          </n-radio-group>
        </n-form-item-row>
        <n-form-item-row :label="t('home.fontSize')">
          <n-slider
            :default-value="defaultFontSize"
            :step="1"
            :format-tooltip="formatTooltip"
            :min="12"
            :max="16"
            :on-update:value="updateFontSize"
          />
        </n-form-item-row>
        <n-button size="tiny">{{ t("home.logout") }}</n-button>
      </div>
    </template>
  </Dropdown>
</template>

<style lang="scss">
.user-tab .user-avatar {
  filter: brightness(1);
  transition: 0.2s ease;
}

.user-tab:hover {
  filter: brightness(1.15);
}

.n-form-item-feedback-wrapper {
  min-height: 0.5rem !important;
}
</style>
