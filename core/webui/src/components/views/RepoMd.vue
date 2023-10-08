<script setup lang="ts">
import { useRoute } from "vue-router";
import { useRepo } from "@/store/repoStore";
import "github-markdown-css/github-markdown-dark.css";
import { FileList } from "@/zodios/schemas/repos";
import { z } from "zod";

const route = useRoute();
const repo = $(useRepo(computed(() => route.params.name as string)));

let fileName = $ref<z.infer<typeof FileList>[number]>(
  {} as z.infer<typeof FileList>[number]
);
let fileContent = $ref<string>("");
let showReadme = $ref(false);
watch(
  () => repo,
  () => {
    if (repo && repo.Name)
      api
        .get("/git-repo-file-list", {
          queries: {
            name: repo!.Name,
            branch: repo!.DefaultBranch,
          },
        })
        .then((res) => {
          fileName = res.filter((file) =>
            file.name.toLowerCase().includes("readme.md")
          )[0];
          if (!fileName) {
            showReadme = true;
          }
        });
  }
);
watch(
  () => fileName,
  () => {
    if (fileName)
      api
        .get("/git-repo-file-open", {
          queries: {
            name: repo!.Name,
            branch: repo!.DefaultBranch,
            path: fileName.name,
          },
        })
        .then((res) => {
          showReadme = true;
          fileContent = res;
        });
  }
);
</script>
<template>
  <n-card
    :title="fileName ? fileName.name : 'README.md'"
    size="small"
    style="height: calc(100vh - 5.1rem)"
  >
    <Spinner :data="showReadme">
      <n-scrollbar
        style="
          height: 100%;
          position: absolute;
          width: 980px;
          border-radius: 0.5rem;
          left: calc(48vw - 490px);
          display: block;
        "
      >
        <Markdown
          class="markdown-body"
          :text="fileContent"
          v-if="fileContent"
        />
        <div
          class="base-alert"
          v-else
          style="text-align: center; margin-top: 50px"
        >
          {{ $t("messages.noReadme") }}
        </div>
      </n-scrollbar>
    </Spinner>
  </n-card>
</template>

<style scoped>
.markdown-body {
  box-sizing: border-box;
  min-width: 200px;
  max-width: 980px;
  border-radius: 0.5rem;
  margin: 0 auto;
  padding: 2rem;
}

@media (max-width: 767px) {
  .markdown-body {
    padding: 15px;
  }
}
</style>
<style>
pre code.hljs {
  display: block;
  overflow-x: auto;
  padding: 1em;
}
code.hljs {
  padding: 3px 5px;
}
.hljs {
  color: #c9d1d9;
  background: #0d1117;
}
.hljs-doctag,
.hljs-keyword,
.hljs-meta .hljs-keyword,
.hljs-template-tag,
.hljs-template-variable,
.hljs-type,
.hljs-variable.language_ {
  color: #ff7b72;
}
.hljs-title,
.hljs-title.class_,
.hljs-title.class_.inherited__,
.hljs-title.function_ {
  color: #d2a8ff;
}
.hljs-attr,
.hljs-attribute,
.hljs-literal,
.hljs-meta,
.hljs-number,
.hljs-operator,
.hljs-selector-attr,
.hljs-selector-class,
.hljs-selector-id,
.hljs-variable {
  color: #79c0ff;
}
.hljs-meta .hljs-string,
.hljs-regexp,
.hljs-string {
  color: #a5d6ff;
}
.hljs-built_in,
.hljs-symbol {
  color: #ffa657;
}
.hljs-code,
.hljs-comment,
.hljs-formula {
  color: #8b949e;
}
.hljs-name,
.hljs-quote,
.hljs-selector-pseudo,
.hljs-selector-tag {
  color: #7ee787;
}
.hljs-subst {
  color: #c9d1d9;
}
.hljs-section {
  color: #1f6feb;
  font-weight: 700;
}
.hljs-bullet {
  color: #f2cc60;
}
.hljs-emphasis {
  color: #c9d1d9;
  font-style: italic;
}
.hljs-strong {
  color: #c9d1d9;
  font-weight: 700;
}
.hljs-addition {
  color: #aff5b4;
  background-color: #033a16;
}
.hljs-deletion {
  color: #ffdcd7;
  background-color: #67060c;
}
</style>
