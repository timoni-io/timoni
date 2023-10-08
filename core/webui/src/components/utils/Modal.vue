<script setup lang="ts">
// ale git
import { useI18n } from "vue-i18n";

type Footer = {
  positiveText?: string;
  negativeText?: string;
};
const props = defineProps<{
  show: boolean;
  title: string;
  touched?: boolean;
  showFooter?: Footer | boolean;
  width?: string;
  errorModal?: boolean;
  unclose?: boolean;
  loading?: boolean;
  disabled?: boolean;
}>();
const emit = defineEmits<{
  (type: "update:show", event: boolean): void;
  (type: "negative-click"): void;
  (type: "positive-click"): void;
}>();

let show = $(useVModel(props, "show", emit));
const { t } = useI18n();

let shakeIt = $ref(false);

const shake = () => {
  shakeIt = true;
  setTimeout(() => {
    shakeIt = false;
  }, 1000);
};

const maskClick = () => {
  if (props.unclose) return;
  if (props.touched) {
    shake();
  } else {
    show = false;
  }
};
</script>

<template>
  <n-modal
    v-model:show="show"
    :title="title"
    :class="{ shakeIt, errorModal }"
    :mask-closable="false"
    :closable="!unclose"
    @mask-click="maskClick"
    style="max-width: 80vw"
  >
    <n-card size="small">
      <template #header>
        <div class="error-title" v-if="errorModal">
          <n-icon size="32"><mdi :path="mdiAlertCircleOutline" /></n-icon>
          {{ title }}
        </div>
        <div v-else>
          {{ title }}
        </div>
      </template>
      <template #header-extra>
        <div
          v-if="!unclose"
          size="tiny"
          class="close-button"
          @click="show = false"
        >
          <n-icon size="18"> <mdi :path="mdiClose" /> </n-icon>
        </div>
      </template>
      <div style="display: flex; flex-direction: column; height: 100%">
        <slot />
        <div class="modal-options">
          <n-button
            v-if="showFooter"
            type="primary"
            secondary
            strong
            :disabled="disabled || false"
            :loading="loading"
            @click="emit('positive-click')"
          >
            {{
              typeof showFooter === "boolean"
                ? t("actions.confirm")
                : showFooter.positiveText
            }}
          </n-button>
        </div>
      </div>
    </n-card>
  </n-modal>
</template>
<style scoped lang="scss">
.close-button {
  cursor: pointer;

  & .n-icon {
    color: var(--primaryColor);
    opacity: 0.9;
    transition: all 0.2s ease-in-out;
  }

  &:hover .n-icon {
    opacity: 1;
    filter: brightness(1.2);
  }
}
</style>
<style>
.modal-options {
  display: flex;
  justify-content: flex-end;
  align-items: flex-end;
  flex-grow: 1;
  gap: 7px;
  margin-top: 0.7rem;
}

.shakeIt {
  animation: shake 0.82s cubic-bezier(0.36, 0.07, 0.19, 0.97) both;
  transform: translate3d(0, 0, 0);
  backface-visibility: hidden;
  perspective: 1000px;
}

@keyframes shake {
  10%,
  90% {
    transform: translate3d(-1px, 0, 0);
  }

  20%,
  80% {
    transform: translate3d(2px, 0, 0);
  }

  30%,
  50%,
  70% {
    transform: translate3d(-4px, 0, 0);
  }

  40%,
  60% {
    transform: translate3d(4px, 0, 0);
  }
}

.error-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: tomato;
  font-size: 1.75rem;
}

.errorModal {
  border: 1px tomato solid;
}
</style>
