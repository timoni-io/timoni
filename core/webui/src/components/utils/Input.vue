<script setup lang="ts">
let props = defineProps<{
  placeholder: string;
  errorMessage?: string;
  valueFromParent?: string;
  focus: boolean;
  removeWhiteSpace?: boolean;
  replaceInvalidCharacter?: RegExp;
  name?: string;
}>();

const emit = defineEmits<{
  (type: "update:value", value: string): void;
  (type: "keyup.enter", event: string): void;
}>();

let value = $ref(props.name ? props.name : "");

watchEffect(() => {
  if (props.valueFromParent) value = props.valueFromParent;
});

watch(
  () => value,
  () => {
    if (props.removeWhiteSpace) value = value.trim();
    if (props.replaceInvalidCharacter)
      value = value.replace(props.replaceInvalidCharacter, "-");
    emit("update:value", value);
  }
);

// watch(
//   () => props.valueFromParent,
//   () => {
//     if (props.valueFromParent) value = props.valueFromParent;
//   }
// );

const keyupEnter = (event: Event) => {
  // @ts-ignore
  emit("keyup.enter", event.target?.value);
};

const inputRef = $ref(null as HTMLInputElement | null);
watch(
  () => inputRef,
  () => {
    if (props.focus) inputRef?.focus();
  }
);
</script>

<template>
  <label>
    <n-input
      ref="inputRef"
      type="text"
      :placeholder="placeholder"
      v-model:value="value"
      @keyup.enter="keyupEnter"
    />
    <div
      style="
        font-size: 0.8rem;
        color: tomato;
        transition: all 0.2s ease-in-out;
        transform: translateY(-1rem);
        opacity: 0;
        height: 0;
      "
      :style="errorMessage ? { transform: 'translateY(0rem)', opacity: 1 } : {}"
    >
      {{ errorMessage }}
    </div>
  </label>
</template>
