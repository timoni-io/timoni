<script setup lang="ts">
import { useEnv } from "@/store/envStore";
import { useRoute } from "vue-router";
import { useI18n } from "vue-i18n";
import { useUserStore } from "@/store/userStore";

const userStore = useUserStore();
const route = useRoute();
const env = useEnv(computed(() => route.params.id as string));

const { t } = useI18n();

let timeChangeTrigger = $ref(0);

let screenWidth = $ref<number>(0);

onBeforeMount(() => {
  screenWidth = window.innerWidth;
});

onMounted(() => {
  window.addEventListener("resize", () => {
    screenWidth = window.innerWidth;
  });
});
</script>

<template>
  <div>
    <EnvTab />
    <PageLayout>
      <n-card
        v-if="userStore.havePermission('Env_View')"
        :title="t('elements.charts.mostChangedElements')"
        style="margin-bottom: 1em; height: calc(100vh - 5rem)"
        size="small"
      >
        <template #header-extra>
          <div style="padding-left: 0.5rem">
            <TimeSetBtn
              @timeSettingChanged="timeChangeTrigger += 1"
              :id="'-most-changed-elements'"
              :size="'tiny'"
              :timeUnitTo="'mo'"
              :responsive-version="screenWidth <= 1400"
              :daysMonths="true"
              :defaultTimeProps="{ time: 30, timeUnit: 'd' }"
            />
          </div>
        </template>
        <MostActiveElements
          :timeChangeTrigger="timeChangeTrigger"
          :id="'-most-changed-elements'"
          :MostChangedElements="env?.EnvInfo?.MostChangedElements || []"
        />
      </n-card>
      <ErrorDialog v-else :msg="$t('messages.permissionDenied')" />
    </PageLayout>
  </div>
</template>

<style scoped>
.dropdown-menu {
  display: flex;
  align-items: center;
  justify-content: right;
  cursor: pointer;
  opacity: 0.6;
  transition: 0.2s ease-in-out;
}
.dropdown-menu:hover {
  opacity: 1;
}
</style>
