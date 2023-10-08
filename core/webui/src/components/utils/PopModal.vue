<script setup lang="ts">
type Footer = {
  positiveText?: string;
  negativeText?: string;
};
let props = defineProps<{
  show?: boolean;
  title?: string;
  touched?: boolean;
  showFooter?: Footer | boolean;
  width?: string;
  locked?: boolean;
}>();
const emit = defineEmits<{
  (type: "update:show", event: boolean): void;
  (type: "negative-click"): void;
  (type: "positive-click"): void;
}>();
watch(
  () => props.show,
  () => {
    keyToHide++;
  }
);
// let show = $(useVModel(props, "show", emit));

// let shakeIt = $ref(false);

// const shake = () => {
//   shakeIt = true;
//   setTimeout(() => {
//     shakeIt = false;
//   }, 1000);
// };

// const maskClick = () => {
//   if (props.touched) {
//     shake();
//   } else {
//     // show = false;
//   }
// };

let keyToHide = $ref(0);
</script>

<template>
  <Dropdown :touched="touched" :keyToHide="keyToHide" :locked="locked">
    <template #trigger>
      <slot name="trigger" />
    </template>
    <template #content="props">
      <div style="padding: 0.5rem 1rem" :style="width ? `width: ${width}` : ''">
        <div style="display: flex; justify-content: space-between">
          <div
            v-if="title"
            style="font-size: var(--n-title-font-size); padding-bottom: 0.3rem"
          >
            {{ title }}
          </div>
          <div
            @click="
              () => {
                emit('negative-click');
                keyToHide++;
              }
            "
            size="tiny"
            class="close-button"
          >
            <n-icon size="18">
              <mdi :path="mdiClose" />
            </n-icon>
          </div>
        </div>

        <slot name="content" v-bind="props" />
        <div class="modal-options">
          <!-- <n-button
            v-if="showFooter === true || 'negativeText' in (showFooter || {})"
            strong
            secondary
            @click="
              () => {
                emit('negative-click');
                keyToHide++;
              }
            "
            size="tiny"
          >
            {{
              typeof showFooter === "boolean"
                ? "Cancel"
                : showFooter?.negativeText
            }}
          </n-button> -->
          <n-button
            v-if="showFooter"
            type="primary"
            secondary
            strong
            @click="emit('positive-click')"
            size="tiny"
            style="width: 24%; height: 1.5rem"
          >
            {{
              typeof showFooter === "boolean"
                ? "Confirm"
                : showFooter.positiveText
            }}
          </n-button>
        </div>
      </div>
    </template>
  </Dropdown>
</template>

<style lang="scss">
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
