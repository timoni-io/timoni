<script setup lang="ts">
import { IInput } from "./ElementAddRemote.vue";

interface IInputResp {
  Text: string;
  Bool: boolean;
  Number: number;
}

const props = defineProps<{
  name: string;
  input: IInput;
  edit: boolean;
  element?: string;
}>();
const emit = defineEmits(["updateInput"]);

const inputResp = ref<IInputResp>({
  Bool: props.edit
    ? props.input.Value === "true"
    : props.input.Default === "true",
  Number: props.edit
    ? isNaN(parseFloat(props.input.Value))
      ? props.input.Min
      : parseFloat(props.input.Value)
    : isNaN(parseFloat(props.input.Default))
    ? props.input.Min
    : parseFloat(props.input.Default),
  Text: props.edit ? props.input.Value : props.input.Default,
});

const regexpCheck = (value: string) => {
  return new RegExp(props.input.MatchRegEx).test(value);
};

watch(inputResp.value, () => {
  switch (props.input.Type) {
    case "text":
    case "secret":
    case "one-of":
      emit("updateInput", [props.name, inputResp.value.Text, props.element]);
      break;
    case "bool":
      emit("updateInput", [props.name, inputResp.value.Text, props.element]);
      break;
    case "integer":
    case "float":
      emit("updateInput", [props.name, inputResp.value.Number, props.element]);
      break;
  }
});
</script>

<template>
  <div style="display: flex">
    <p
      style="
        width: 15%;
        display: flex;
        justify-content: flex-end;
        padding-right: 2em;
      "
    >
      {{ props.name }}
    </p>
    <div style="width: 85%">
      <n-input
        v-if="props.input.Type === 'text'"
        :placeholder="name"
        v-model:value="inputResp.Text"
        :allow-input="regexpCheck"
      />
      <n-input
        v-else-if="props.input.Type === 'secret'"
        type="password"
        show-password-on="mousedown"
        :placeholder="name"
        v-model:value="inputResp.Text"
        :allow-input="regexpCheck"
      />
      <n-input-number
        v-else-if="props.input.Type === 'integer'"
        :placeholder="name"
        v-model:value="inputResp.Number"
        :validator="
          props.input.Min === props.input.Max
            ? is(z.number().int())
            : is(z.number().int().min(props.input.Min).max(props.input.Max))
        "
      />
      <n-input-number
        v-else-if="props.input.Type === 'float'"
        :placeholder="name"
        v-model:value="inputResp.Number"
        :validator="
          props.input.Min === props.input.Max
            ? is(z.number())
            : is(z.number().min(props.input.Min).max(props.input.Max))
        "
      />
      <div
        v-else-if="props.input.Type === 'bool'"
        style="display: flex; gap: 5px"
      >
        <n-radio-group
          v-for="option in ['true', 'false']"
          :key="option"
          v-model:value="inputResp.Text"
        >
          <n-radio-button :value="option">
            {{ option }}
          </n-radio-button>
        </n-radio-group>
      </div>
      <div
        v-else-if="props.input.Type === 'one-of'"
        style="display: flex; gap: 5px"
      >
        <n-radio-group
          v-for="option in props.input.Options"
          :key="option"
          v-model:value="inputResp.Text"
        >
          <n-radio-button :value="option">
            {{ option }}
          </n-radio-button>
        </n-radio-group>
      </div>
    </div>
  </div>
</template>

<style scoped>
.n-card__content {
  padding-right: 20em;
  padding-left: 20em;
}
</style>
