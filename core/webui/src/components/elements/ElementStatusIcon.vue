<script setup lang="ts">
import { Icon } from "@/utils/iconFactory";

const props = defineProps<{
  status: number;
  toDelete: boolean;
  iconFunc: (status: number) => Icon;
  cron?: boolean;
}>();

const iconF = computed(() => props.iconFunc(props.status));
</script>

<template>
  <div
    style="
      /* width: 100%; */
      display: flex;
      position: relative;
      align-items: center;
      justify-content: flex-end;
      right: -10px;
      z-index: 9;
    "
  >
    <!-- <n-tooltip placement="top-start" trigger="hover">
        <template #trigger> -->
    <div style="position: relative; height: 17px; width: 17px; z-index: 9">
      <div
        style="position: absolute; right: -9px; top: -9px; z-index: 9"
        v-if="cron"
      >
        <Mdi :path="mdiClockOutline" width="15px" />
      </div>

      <n-icon-wrapper
        :size="17"
        :border-radius="24"
        :color="iconF.color"
        style="margin-right: 4px"
      >
        <n-icon v-if="props.toDelete">
          <Mdi :path="mdiDelete" />
          <n-spin :size="26" :stroke="iconF.color" />
        </n-icon>
        <n-icon
          v-else
          :size="iconF.iconSize"
          :class="iconF.icon === mdiAutorenew ? 'rotate' : ''"
        >
          <Mdi :path="iconF.icon" />
        </n-icon>
      </n-icon-wrapper>
      <!-- <div
              v-if="status.StatusMessage"
              style="
                position: absolute;
                bottom: -1px;
                right: -1px;
                height: 6px;
                width: 6px;
                border-radius: 50%;
                background-color: red;
                border: 1px solid var(--n-th-color);
              "
            ></div>
          </div>
        </template> -->
      <!-- <div>
          {{
            statusDictionary[
              status.Status as 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7
            ]
          }}
          {{ status.StatusMessage ? ` - ${status.StatusMessage}` : "" }} -->
    </div>
    <!-- </n-tooltip> -->
  </div>
</template>

<style scoped>
@keyframes rotation {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(359deg);
  }
}
.rotate {
  animation: rotation 2s infinite linear;
}
</style>
