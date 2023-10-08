<script setup lang="ts">
import { getMaterialFileIcon } from "file-extension-icon-js";

const props = defineProps<{ file?: string; fileList?: string[] }>();

const emit = defineEmits<{
  (type: "update:file", event: string): void;
}>();

let activeFile = $(useVModel(props, "file", emit));

let trimFilePath = (path: string, max: number): string => {
  if (path.length <= max) return path;

  const [_head, ...tail] = path.split("/");

  if (tail.length === 0) return path;

  return trimFilePath(tail.join("/"), max);
};
</script>

<template>
  <div style="display: flex; flex-flow: column" class="file-item">
    <n-button
      v-for="file in [...(props.fileList || [])].sort()"
      strong
      quaternary
      :key="file"
      size="small"
      :type="activeFile === file ? 'primary' : undefined"
      @click="activeFile = file"
    >
      <img
        :src="getMaterialFileIcon(file)"
        style="
          width: 1rem;
          height: 1rem;
          transform: scale(1.1);
          margin-right: 0.5rem;
        "
      />
      <template v-if="trimFilePath(file, 25).length < file.length">
        <n-tooltip placement="right">
          <template #trigger>
            <span> ...{{ trimFilePath(file, 25) }} </span>
          </template>
          {{ file }}
        </n-tooltip>
      </template>
      <template v-else>
        {{ file }}
      </template>
    </n-button>
  </div>
</template>

<style>
.file-item .n-button {
  justify-content: start;
}
</style>
