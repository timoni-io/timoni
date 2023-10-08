<script setup lang="ts">
import { useRepo } from "@/store/repoStore";
import { invoke } from "@vueuse/core";
import { NIcon, TreeOption } from "naive-ui";
import { useRoute } from "vue-router";
import MdiVue from "../Mdi.vue";
import {
  getMaterialFileIcon,
  getMaterialFolderIcon,
} from "file-extension-icon-js";
import { useRouteParam, useRouteParamArray } from "@/utils/router";
import { useI18n } from "vue-i18n";
import { useRepoBranch } from "@/store/repoStore";
import { useUserStore } from "@/store/userStore";

const userStore = useUserStore();
const props = defineProps<{
  repo?: string;
  branch?: string;
  dockerfile?: string[];
}>();
const emit = defineEmits(["exportDocker"]);

const route = useRoute();
const { t } = useI18n();
const repoBranch = useRepoBranch();
const repo = $(
  useRepo(computed(() => props.repo || (route.params.name as string)))
);

let selectedBranch = $(
  props.branch ? ref(props.branch) : useRouteParam("branch")
);

watch($$(selectedBranch), () => {
  repoBranch.selectedBranch = selectedBranch;
});
let selectedFileName = $(
  props.repo ? ref(props.dockerfile || []) : useRouteParamArray("filepath")
);
let selectedFileContent = $ref("");
let fileLoader = $ref(true);
let openFileLoader = $ref(true);
let isMarkdown = $computed(
  () => selectedFileName.at(-1)?.endsWith(".md") || false
);

type FileList = ResType<"/git-repo-file-list">;

const fileListToTree = (fileList: FileList, options?: { dir?: string }) =>
  fileList.map((item) => ({
    key: options?.dir ? `${options.dir}/${item.name}` : item.name,
    label: item.name,
    isLeaf: item.isFile === "file",
    prefix: getPrefix(item.name, item.isFile === "file"),
  }));

// let branchList = useAsyncState<string[] | null>(
//   api.get("/git-repo-branch-list", {
//     queries: {
//       name: route.params.name as string,
//       context: "Code",
//       level: 1,
//     },
//   }) as Promise<string[]>,
//   null
// );
let branchList = $ref<string[]>([]);

const getBranchList = async () => {
  branchList = [];
  branchList = (await api.get("/git-repo-branch-list", {
    queries: {
      name: props.repo || (route.params.name as string),
      level: 1,
    },
  })) as string[];
};
getBranchList();
let tree = $ref([] as TreeOption[]);

const getPrefix = (filename: string, isFile: boolean) => {
  const getIcon = isFile ? getMaterialFileIcon : getMaterialFolderIcon;

  return () =>
    h("img", {
      src: getIcon(filename),
      style: "width: 1rem; height: 1rem; transform: scale(1.1)",
    });
};
const getFileContent = async () => {
  openFileLoader = true;
  selectedFileContent = await api.get("/git-repo-file-open", {
    queries: {
      name: repo!.Name as string,
      branch: selectedBranch as string,
      path: selectedFileName.join("/"),
    },
  });
  openFileLoader = false;
};
invoke(async () => {
  await until($$(repo)).toBeTruthy();

  const deepTree = await getDeepTree("", selectedFileName.join("/"));
  if (deepTree) {
    tree = deepTree;
  }
  getFileContent();
});
const getRepoFileList = async () => {
  fileLoader = true;
  const data = await api.get("/git-repo-file-list", {
    queries: {
      name: repo!.Name as string,
      branch: selectedBranch! as string,
    },
  });

  tree = fileListToTree(data);
  fileLoader = false;
};

watch($$(selectedBranch), async () => {
  getRepoFileList();
});

const loadDir = async (dir: TreeOption) => {
  const data = await api.get("/git-repo-file-list", {
    queries: {
      name: repo!.Name,
      branch: selectedBranch! as string,
      directory: dir.key as string,
    },
  });

  dir.children = fileListToTree(data, { dir: dir.key as string });
};

const renderSwitcherIcon = () =>
  h(NIcon, null, {
    default: () =>
      h(MdiVue, { path: mdiChevronRight, style: "transform: scale(1.4)" }),
  });

const select = async (keys: string[], items: (TreeOption | null)[]) => {
  if (items[0]?.isLeaf) {
    openFileLoader = true;
    selectedFileContent = await api.get("/git-repo-file-open", {
      queries: {
        name: repo!.Name,
        branch: selectedBranch as string,
        path: keys[0],
      },
    });
    openFileLoader = false;
    selectedFileName = keys[0].split("/");
    emit("exportDocker", keys[0]);
  }
};

const getDeepTree = async (dir: string, path: string) => {
  const [next, ...rest] = path.split("/");
  const nextDir = dir ? [dir, next].join("/") : next;
  const data = await api.get("/git-repo-file-list", {
    queries: {
      name: repo!.Name,
      branch: selectedBranch! as string,
      directory: dir,
    },
  });

  const tree = fileListToTree(data, { dir }) as TreeOption[];

  const dirNode = tree.find((node) => node.key === nextDir);
  if (dirNode && !dirNode.isLeaf) {
    dirNode.children = await getDeepTree(nextDir, rest.join("/"));
  }
  fileLoader = false;
  return tree;
};

let expandedKeys = $ref(
  selectedFileName.reduce((acc, p, idx) => {
    if (idx === 0) {
      return [p];
    } else {
      acc.push(`${acc[idx - 1]}/${p}`);
      return acc;
    }
  }, [] as string[])
);

let showMdCode = $ref(false);
let clicked = $ref(false);
let clickedFile = $ref(false);
const reloadF = (isFile: boolean) => {
  if (isFile) {
    clickedFile = true;
    getFileContent();
  } else {
    clicked = true;
    getRepoFileList();
  }
  setTimeout(() => {
    clicked = false;
    clickedFile = false;
  }, 500);
};
</script>

<template>
  <div>
    <RepoTabs />
    <PageLayout>
      <!-- <Spinner
        :data="branchList.state.value"
        style="height: calc(100vh - 5.1rem)"
      > -->
      <div
        v-if="userStore.havePermission('Repo_View')"
        :class="props.repo !== undefined ? 'files-for-docker' : 'files-grid'"
        :style="props.repo !== undefined ? 'height: 50em' : 'height: 100%'"
      >
        <BranchList
          v-if="props.repo === undefined"
          v-model:branch="selectedBranch"
          :branch-list="branchList || []"
          @reload="getBranchList"
        />
        <n-card
          size="small"
          :title="t('fields.filetree')"
          style="position: relative"
          :style="
            props.repo !== undefined
              ? 'height: inherit'
              : 'calc(100vh - 5.1rem)'
          "
        >
          <template #header-extra>
            <n-button
              quaternary
              type="primary"
              size="small"
              circle
              @click="reloadF(false)"
            >
              <Mdi
                width="20"
                :path="mdiReload"
                :class="clicked ? 'clicked' : ''"
              />
            </n-button>
          </template>
          <Spinner
            :data="!fileLoader"
            style="position: absolute; width: 95%"
            :style="
              props.repo !== undefined
                ? 'height: 40rem'
                : 'height: calc(100vh - 9.1rem)'
            "
          >
            <n-scrollbar
              :style="
                props.repo !== undefined
                  ? 'height: 40rem'
                  : 'height: calc(100vh - 9.1rem)'
              "
            >
              <n-tree
                :data="tree"
                :on-load="loadDir"
                :render-switcher-icon="renderSwitcherIcon"
                selectable
                @update:selected-keys="select"
                :selected-keys="[selectedFileName.join('/')]"
                v-model:expanded-keys="expandedKeys"
                expand-on-click
                accordion
              />
            </n-scrollbar>
          </Spinner>
        </n-card>
        <n-card size="small" class="monaco-card">
          <Spinner
            :data="!openFileLoader"
            style="position: relative"
            :style="
              props.repo !== undefined
                ? 'height: 40rem'
                : 'calc(100vh - 5.1rem)'
            "
          >
            <div
              style="
                display: flex;
                align-items: center;
                justify-content: space-between;
              "
            >
              <p
                style="
                  height: 2rem;
                  display: flex;
                  align-items: center;
                  padding: 0.5rem;
                  gap: 0.7rem;
                "
              >
                <img
                  :src="getMaterialFileIcon(selectedFileName.join('/'))"
                  style="width: 1rem; height: 1rem; transform: scale(1.1)"
                />
                {{ selectedFileName.join("/") }}
              </p>
              <div style="display: flex; align-items: center">
                <n-button
                  quaternary
                  type="primary"
                  size="small"
                  circle
                  style="margin-right: 0.5rem"
                  @click="reloadF(true)"
                >
                  <Mdi
                    width="20"
                    :path="mdiReload"
                    :class="clickedFile ? 'clicked' : ''"
                  />
                </n-button>
                <n-button
                  v-if="isMarkdown"
                  quaternary
                  type="primary"
                  circle
                  size="small"
                  @click="showMdCode = !showMdCode"
                  style="margin-right: 0.5rem"
                >
                  <template #icon>
                    <Mdi :path="showMdCode ? mdiScriptText : mdiFileCode" />
                  </template>
                </n-button>
              </div>
            </div>
            <div
              style="position: absolute; width: 100%"
              :style="
                props.repo !== undefined
                  ? 'height: 40rem'
                  : 'height: calc(100% - 2rem)'
              "
            >
              <Markdown
                v-if="isMarkdown && !showMdCode"
                :absolute="true"
                :text="selectedFileContent"
              />
              <Monaco
                v-else
                :key="selectedFileContent"
                :value="selectedFileContent"
                read-only
                :filename="selectedFileName.join('/')"
              />
            </div>
          </Spinner>
        </n-card>
      </div>
      <ErrorDialog v-else :msg="$t('messages.permissionDenied')" />

      <!-- </Spinner> -->
    </PageLayout>
  </div>
</template>

<style scoped lang="scss">
.files-grid {
  display: grid;
  grid-template-columns: 12rem 17rem 1fr;
  gap: 1rem;
}

.files-for-docker {
  display: flex;
  gap: 10px;
}

.file-tree-card {
  height: calc(100vh - 5rem);
  max-height: calc(100vh - 5rem);
  background-color: #3d3d3d48;
  & :deep(.n-card__content) {
    padding: 0px;
    --px: 2px;
    // padding-left: var(--px);
    // padding-right: var(--px);
  }
}

.monaco-card :deep(.n-card__content) {
  padding: 0px;
}
</style>
