<script setup lang="ts">
import { useSlots } from "vue";

const props = defineProps<{
  touched?: boolean;
  hide?: boolean;
  keyToHide?: number;
  locked?: boolean;
}>();
const slots = useSlots();

let show = $ref(false);

const close = () => {
  show = false;
};
watchEffect(() => {
  if (props.hide) show = false;
});
watch(
  () => props.keyToHide,
  () => {
    show = false;
  }
);
const options = [
  {
    type: "render",
    render: slots["content"] && (() => slots["content"]!({ close })),
  },
];

let shakeIt = $ref(false);
const shake = () => {
  shakeIt = true;
  setTimeout(() => {
    shakeIt = false;
  }, 1000);
};

const clickOutside = () => {
  if (props.locked) {
    return;
  }
  if (props.touched) {
    shake();
  } else {
    show = false;
  }
};
onKeyStroke("Escape", () => {
  show = false;
});
</script>

<template>
  <n-dropdown
    :show="show"
    @update:show="show = true"
    trigger="click"
    :options="options"
    :class="{ shakeIt }"
    placement="bottom-start"
    @clickoutside="clickOutside"
  >
    <slot name="trigger" />
  </n-dropdown>
</template>

<style scoped>
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
</style>

<style>
.v-binder-follower-container {
  z-index: 2138 !important;
}
</style>
