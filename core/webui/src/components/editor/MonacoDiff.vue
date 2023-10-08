<script setup lang="ts">
import * as monaco from "monaco-editor";
import "./langs/toml";
import * as diff2html from "diff2html";
const props = defineProps<{
  value?: string;
  readOnly?: boolean;
  lang?: string;
  filename?: string;
}>();
let el = $ref(null as HTMLElement | null);

onMounted(() => {
  let parsed = diff2html.parse(props.value as string, {
    outputFormat: "side-by-side",
  });
  let left = parsed[0].blocks[0].lines
    .map((el) => {
      if (el.oldNumber) return el.content;
      else return "";
    })
    .filter((el) => el);

  let right = parsed[0].blocks[0].lines
    .map((el) => {
      if (el.newNumber) return el.content;
      else return "";
    })
    .filter((el) => el);

  let originalModel = monaco.editor.createModel(
    left.join("\n"),
    undefined,
    (props.filename && monaco.Uri.file(props.filename)) || undefined
  );

  let modifiedModel = monaco.editor.createModel(
    right.join("\n"),
    originalModel.getLanguageId()
  );
  let diffEditor = monaco.editor.createDiffEditor(el as HTMLElement, {
    theme: "vs-dark",
    smoothScrolling: true,
    readOnly: true,
  });
  diffEditor.setModel({
    original: originalModel,
    modified: modifiedModel,
  });
});

onUnmounted(() => {
  monaco.editor.getModels().forEach((model) => model.dispose());
});
</script>
<template>
  <div ref="el" style="height: 100%" />
</template>
