<script setup lang="ts">
import { Env } from "@/zodios/schemas/dashboard";
// import { envIcon } from "@/utils/iconFactory";
import { z } from "zod";
import { useRouter, useRoute } from "vue-router";
import { useSpinner } from "@/store/spinner";

const router = useRouter();
const route = useRoute();
const spinner = useSpinner();
// type EnvStatus = "ok" | "error";
// type StatusVariant = { color: string; icon: string; background: string };
type EnvCardProps = {
  envId: string;
  // status: EnvStatus;
  env: z.infer<typeof Env>;
  elementStatuses: Record<string, number>;
  toDelete: boolean;
};

const props = defineProps<EnvCardProps>();
// const statusVariant = computed<StatusVariant>(() => envIcon(props.env.Status));

const push = (path: string) => {
  if (route.path !== path) spinner.spinner = true;
  router.push(path);
};

const statusCount = $computed(() => ({
  success: props.elementStatuses[3] || 0,
  warn:
    (props.elementStatuses[0] || 0) +
    (props.elementStatuses[1] || 0) +
    (props.elementStatuses[5] || 0),
  error: (props.elementStatuses[2] || 0) + (props.elementStatuses[4] || 0),
  disabled: props.elementStatuses[6] || 0,
}));
</script>

<template>
  <div
    :style="{
      // '--color': statusVariant.color,
      cursor: 'pointer',
    }"
  >
    <n-card
      size="small"
      @click="push(`env/${envId}`)"
      :style="{
        // 'background-color': statusVariant.background,
        border: '1px solid black',
      }"
      class="listing-element"
    >
      <div
        class="card-content"
        style="
          display: flex;
          align-items: center;
          justify-content: space-between;
        "
      >
        <!-- <n-icon-wrapper>
          <n-icon> <mdi :path="statusVariant.icon" /> </n-icon>
        </n-icon-wrapper> -->
        <div class="text">
          <span>{{ env.Name }}</span>
        </div>
        <div
          v-if="toDelete"
          style="
            display: flex;
            place-items: center;
            height: 1.7rem;
            justify-content: end;
          "
        >
          <n-icon :size="18">
            <Mdi :path="mdiTrashCan" />
            <n-spin size="small" />
          </n-icon>
        </div>
        <div
          v-else
          style="
            display: grid;
            gap: 0.5rem;
            grid-auto-flow: column;
            grid-auto-columns: 1fr;
            height: 1.7rem;
          "
        >
          <n-tag v-if="statusCount.success" :bordered="false" type="success">
            {{ statusCount.success }}
          </n-tag>
          <n-tag v-if="statusCount.warn" :bordered="false" type="warning">
            {{ statusCount.warn }}
          </n-tag>
          <n-tag v-if="statusCount.error" :bordered="false" type="error">
            {{ statusCount.error }}
          </n-tag>
          <n-tag v-if="statusCount.disabled" :bordered="false">
            {{ statusCount.disabled }}
          </n-tag>
        </div>
      </div>
      <div class="hover"></div>
    </n-card>
  </div>
</template>

<style scoped>
.n-tag {
  justify-content: center;
  padding: 0.3rem;
  border-radius: 4px;
}
.n-card {
  border-color: var(--color);
}

.n-icon-wrapper {
  background-color: var(--color);
}
.card-content {
  display: flex;
  gap: 1rem;
}

.text {
  line-height: 1.15;
}
</style>
<style>
.listing-element {
  position: relative;
}
.listing-element:hover .hover {
  position: absolute;
  left: 0;
  top: 0;
  width: 100%;
  height: 100%;
  border-radius: 1px;
  background: rgba(255, 255, 255, 0.2);
}
</style>
