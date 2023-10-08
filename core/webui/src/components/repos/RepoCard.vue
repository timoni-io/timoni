<script setup lang="ts">
// import { useRepo } from "@/store/repoStore";
import { Repo } from "@/zodios/schemas/dashboard";
import { envIcon } from "@/utils/iconFactory";
import { z } from "zod";
import { useRouter, useRoute } from "vue-router";
import { useSpinner } from "@/store/spinner";
const router = useRouter();
const route = useRoute();
const spinner = useSpinner();
type StatusVariant = { color: string; icon: string; background: string };

const props = defineProps<{ repo: z.infer<typeof Repo> }>();
const statusVariant = computed<StatusVariant>(() => envIcon(props.repo.Status));

// const repo = useRepo(computed(() => props.repo.Name));
const push = (path: string) => {
  if (route.path !== path) spinner.spinner = true;
  router.push(path);
};
</script>

<template>
  <div
    :style="{
      '--color': statusVariant.color,
      cursor: props.repo.Status === 4 ? 'auto' : 'pointer',
    }"
  >
    <n-tooltip trigger="hover" v-if="props.repo?.Name.length > 20">
      <template #trigger>
        <n-card
          @click="push(`/code/${props.repo.Name}`)"
          size="small"
          :style="{
            'background-color': statusVariant.background,
          }"
        >
          {{ props.repo?.Name.substring(0, 20) + "..." }}
        </n-card>
      </template>
      {{ props.repo?.Name }}
    </n-tooltip>
    <span v-else>
      <n-card
        @click="push(`/code/${props.repo.Name}`)"
        size="small"
        :style="{
          'background-color': statusVariant.background,
        }"
      >
        {{ props.repo?.Name }}
      </n-card>
    </span>
  </div>
</template>

<style scoped>
.n-card {
  border-color: var(--color);
  color: black;
  cursor: pointer;
}
</style>
