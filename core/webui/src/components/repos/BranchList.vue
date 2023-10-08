<script setup lang="ts">
const props = defineProps<{
  branch?: string;
  branchList?: string[] | undefined;
}>();

const emit = defineEmits<{
  (type: "update:branch", event: string): void;
  (type: "reload"): void;
}>();
let clicked = $ref(false);
const emitF = () => {
  emit("reload");
  clicked = true;
  setTimeout(() => {
    clicked = false;
  }, 500);
};
const { t } = useI18n();

let activeBranch = $(useVModel(props, "branch", emit));
</script>

<template>
  <n-card
    size="small"
    :title="t('fields.branch', 2)"
    style="height: calc(100vh - 5.1rem); position: relative"
  >
    <template #header-extra>
      <n-button quaternary type="primary" size="small" circle @click="emitF">
        <Mdi width="20" :path="mdiReload" :class="clicked ? 'clicked' : ''" />
      </n-button>
    </template>
    <Spinner :data="branchList?.length">
      <n-scrollbar style="height: 100%; position: absolute">
        <div style="display: flex; flex-flow: column; gap: 0.5rem">
          <n-button
            strong
            secondary
            v-for="branch in [...(props.branchList || [])].sort()"
            :key="branch"
            size="tiny"
            :type="activeBranch === branch ? 'primary' : undefined"
            @click="activeBranch = branch"
          >
            <n-tooltip
              trigger="hover"
              v-if="branch.length > 20"
              placement="right"
            >
              <template #trigger> {{ branch.substring(0, 17) }}... </template>
              {{ branch }}
            </n-tooltip>
            <span v-else>
              {{ branch }}
            </span>
          </n-button>
        </div>
      </n-scrollbar>
    </Spinner>
  </n-card>
</template>
