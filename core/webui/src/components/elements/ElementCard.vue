<script setup lang="ts">
type ElementStatus = "ok" | "error";
type StatusVariant = { color: string; icon: string };
type ElementCardProps = {
  elementId: string;
  status: ElementStatus;
  selected: boolean;
};

const props = defineProps<ElementCardProps>();

const statusVariant = computed<StatusVariant>(
  () =>
    ({
      ok: {
        color: "var(--successColor)",
        icon: mdiCheck,
      },
      error: {
        color: "var(--errorColor)",
        icon: mdiAlert,
      },
    }[props.status])
);
</script>

<template>
  <div :class="{ selected }" :style="{ '--color': statusVariant.color }">
    <n-card size="small">
      <div class="card-content">
        <n-icon-wrapper>
          <n-icon> <mdi :path="statusVariant.icon" /> </n-icon>
        </n-icon-wrapper>
        <div class="text">
          <span>postgresql</span>
        </div>
      </div>
      <template #footer>
        <n-button strong secondary type="primary" size="tiny">
          <template #icon>
            <mdi :path="mdiGit" />
          </template>
        </n-button>
      </template>
    </n-card>
  </div>
</template>

<style scoped>
.n-card {
  border-color: var(--color);
}

.selected .n-card {
  filter: hue-rotate(30deg);
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
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
