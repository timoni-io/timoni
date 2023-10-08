<script setup lang="ts">
import * as monaco from "monaco-editor";
import "./langs/toml";
import "./langs/vue";

type LineNumbersType =
  | "on"
  | "off"
  | "relative"
  | "interval"
  | ((lineNumber: number) => string);

let el = $ref(null as HTMLElement | null);

const props = defineProps<{
  value?: string;
  readOnly?: boolean;
  lang?: string;
  filename?: string;
  lineNumbers?: LineNumbersType;
}>();

const emit = defineEmits<{
  (type: "update:value", event: string): void;
}>();

let editor: monaco.editor.IStandaloneCodeEditor;

onMounted(() => {
  editor = monaco.editor.create(el!, {
    model:
      (props.filename &&
        monaco.editor.createModel(
          props.value || "",
          undefined,
          monaco.Uri.file(props.filename)
        )) ||
      undefined,
    value: props.value,
    language: props.lang,
    theme: "vs-dark",
    readOnly: props.readOnly,
    fontLigatures: true,
    minimap: {
      autohide: true,
    },
    smoothScrolling: true,
    automaticLayout: true,
    lineNumbers: props.lineNumbers,
  });

  editor.onDidChangeModelContent(() => emit("update:value", editor.getValue()));
});

onUnmounted(() => {
  // monaco.editor.getModels().forEach((model) => model.dispose());
  editor.getModel()?.dispose();
});
</script>

<template>
  <div ref="wrapperEl" style="position: relative; height: 100%; width: 100%">
    <div ref="el" style="position: absolute; inset: 0; height: 100%" />
  </div>
</template>
