<script lang="ts" setup>
import { useI18n } from "vue-i18n";
import { ElementRespExtended } from "@/zodios/schemas/elements";
import { z } from "zod";
import { useRoute, useRouter } from "vue-router";
import { computed, ComputedRef } from "vue";
import { Column } from "@/components/utils/DataTable.vue";
import { useRouteParam } from "@/utils/router";
import { useRepoBranch } from "@/store/repoStore";
import { useUserStore } from "@/store/userStore";

const userStore = useUserStore();
interface Usage {
  Branch: string;
  ElementName: string;
  EnvID: string;
  EnvName: string;
}

const repoBranch = useRepoBranch();
const { t } = useI18n();
const route = useRoute();
const router = useRouter();

let fileModal = $ref(false);
let errorsModal = $ref(false);
let usageModal = $ref(false);
let elementsLoading = $ref(false);

let usageContent = $ref<Array<Usage>>([]);
let fileContent = $ref("");
let fileName = $ref("");
let filePath = $ref("");

let branchList = $ref<string[]>([]);
let selectedBranch = $(useRouteParam("branch"));
watch($$(selectedBranch), () => {
  repoBranch.selectedBranch = selectedBranch;
  reloadF();
});

const file = (element: z.infer<typeof ElementRespExtended>) => {
  fileModal = true;
  fileContent = window.atob(element.Element.FileContent);
  fileName = element.Element.Name;
  filePath = element.Element.Source.FilePath;
};

const usageColumns: ComputedRef<Column[]> = computed(() => [
  {
    title: t("fields.elementName"),
    key: "ElementName",
    width: "50%",
  },
  {
    title: t("fields.envName"),
    template: "envName",
    width: "50%",
  },
]);

const columns: ComputedRef<Column[]> = computed(() => [
  {
    title: t("fields.name"),
    key: "Element.Name",
    width: "20%",
  },
  {
    title: t("fields.type"),
    key: "Element.Type",
    width: "10%",
  },
  {
    title: t("fields.filePath"),
    key: "Element.Source.FilePath",
    width: "25%",
  },
  {
    title: t("fields.description"),
    key: "Element.Description",
    width: "20%",
  },
  // {
  //   title: t("objects.error"),
  //   key: "Error",
  //   template: "error",
  //   width: "25%",
  // },
  {
    title: "",
    template: "action",
    width: "10%",
  },
]);

let elements = $ref<z.infer<typeof ElementRespExtended>[]>([]);

// let elementsList = computed(() => {
//   return elements.filter((el) => el.Element.Source.BranchName === selectedBranch);
// });
const getBranchList = () => {
  branchList = [];

  api
    .get("/git-repo-branch-list", {
      queries: {
        name: route.params.name as string,
        level: 1,
      },
    })
    .then((res) => {
      if (typeof res === "object")
        branchList = res.sort((a: string, b: string) => {
          return a > b ? 1 : -1;
        }) as string[];
      selectedBranch = route.params.branch as string;
    });
};
const getRepoElementList = () => {
  elements = [];
  elementsLoading = true;
  api
    .get("/git-repo-element-list", {
      queries: {
        "git-repo-name": route.params.name as string,
        branch: selectedBranch as string,
      },
    })
    .then((res) => {
      elements = res;
      elementsLoading = false;
    });
};
let clicked = $ref(false);
const reloadF = () => {
  getRepoElementList();
  clicked = true;
  setTimeout(() => {
    clicked = false;
  }, 500);
};

// window width
let screenWidth = $ref<number>(0);
onBeforeMount(() => {
  screenWidth = window.innerWidth;
});

onMounted(() => {
  getBranchList();
  getRepoElementList();

  window.addEventListener("resize", () => {
    screenWidth = window.innerWidth;
  });
});
</script>

<template>
  <div>
    <RepoTabs />
    <PageLayout>
      <Spinner :data="elements" v-if="userStore.havePermission('Repo_View')">
        <div style="display: grid; gap: 1rem; grid-template-columns: 12rem 1fr">
          <BranchList
            v-model:branch="selectedBranch"
            :branch-list="branchList"
            @reload="getBranchList"
          />
          <n-card
            :title="t('objects.element', 2)"
            size="small"
            style="height: calc(100vh - 5.1rem); overflow-y: hidden"
          >
            <template #header-extra>
              <n-button quaternary type="primary" size="small" circle>
                <Mdi
                  width="20"
                  :path="mdiReload"
                  :class="clicked ? 'clicked' : ''"
                  @click="reloadF"
                />
              </n-button>
            </template>
            <n-spin
              v-if="elementsLoading"
              :size="60"
              stroke="#1ba3fd"
              :stroke-width="10"
              style="height: 100%; max-height: 80vh"
            />
            <data-table
              v-else
              :columns="columns"
              :data="
                elements.sort((a, b) =>
                  a.Element.Name.localeCompare(b.Element.Name)
                )
              "
              :max-height="'77vh'"
            >
              <template #error="element">
                <n-tooltip
                  trigger="hover"
                  v-if="element.Element.Error.length"
                  style="max-width: 600px"
                >
                  <template #trigger>
                    <n-tag type="error" :bordered="false" :size="'tiny'">
                      {{ element.Element.Error.substring(0, 50) }}
                      <span v-if="element.Element.Error.length > 50">...</span>
                    </n-tag>
                  </template>
                  {{ element.Element.Error }}
                </n-tooltip>
              </template>
              <template #action="element">
                <div
                  style="
                    width: 100%;
                    height: 100%;
                    display: flex;
                    justify-content: flex-end;
                  "
                >
                  <n-tooltip
                    v-if="element.Element.Error"
                    trigger="hover"
                    :disabled="screenWidth > 1300"
                  >
                    <template #trigger>
                      <n-button
                        size="tiny"
                        secondary
                        type="error"
                        style="margin-right: 0.5em"
                        :style="
                          screenWidth > 1300
                            ? 'min-width: 5rem'
                            : 'min-width: 2rem'
                        "
                        @click="
                          () => {
                            errorsModal = true;
                            fileContent = element.Element.Error;
                            fileName = element.Element.Name;
                          }
                        "
                      >
                        <div
                          style="
                            border-radius: 50%;
                            background: var(--errorColor);
                            height: 16px;
                            width: 16px;
                          "
                        >
                          <Mdi
                            :path="mdiExclamationThick"
                            width="14"
                            style="color: black"
                          />
                        </div>
                        <span
                          v-if="screenWidth > 1300"
                          style="margin-left: 10px"
                          >{{ t("objects.error", 2) }}</span
                        >
                      </n-button>
                    </template>
                    {{ t("objects.error", 2) }}
                  </n-tooltip>

                  <n-tooltip
                    v-if="element.Usage"
                    trigger="hover"
                    :disabled="screenWidth > 1300"
                  >
                    <template #trigger>
                      <n-button
                        size="tiny"
                        secondary
                        type="primary"
                        style="margin-right: 0.5em"
                        :style="
                          screenWidth > 1300 ? 'width: 5rem' : 'min-width: 2rem'
                        "
                        @click="
                          () => {
                            usageModal = true;
                            usageContent = element.Usage;
                            fileName = element.Element.Name;
                          }
                        "
                      >
                        <span
                          v-if="screenWidth > 1300"
                          style="margin-right: 2px"
                          >{{ t("fields.usage") }} </span
                        >({{ element.Usage.length }})
                      </n-button>
                    </template>
                    {{ t("fields.usage") }} ({{ element.Usage.length }})
                  </n-tooltip>

                  <n-tooltip trigger="hover" :disabled="screenWidth > 1300">
                    <template #trigger>
                      <n-button
                        size="tiny"
                        secondary
                        type="primary"
                        :style="
                          screenWidth > 1300 ? 'width: 3rem' : 'min-width: 2rem'
                        "
                        @click="file(element)"
                      >
                        <n-icon class="icon-color" style="cursor: pointer">
                          <mdi :path="mdiFile" width="25" />
                        </n-icon>
                        <span
                          v-if="screenWidth > 1300"
                          style="margin-left: 0.5em"
                        >
                          {{ t("objects.file") }}</span
                        >
                      </n-button>
                    </template>
                    {{ t("objects.file") }}
                  </n-tooltip>
                </div>
              </template>
            </data-table>
          </n-card>
        </div>
      </Spinner>
      <ErrorDialog v-else :msg="$t('messages.permissionDenied')" />
    </PageLayout>
    <Modal
      v-model:show="fileModal"
      :title="t('fields.filePath') + ': ' + filePath"
      style="width: 60%"
    >
      <div v-if="fileContent" style="height: 30rem">
        <Monaco :value="fileContent" lang="toml" read-only />
      </div>
      <p v-else>
        {{ t("fields.noContent") }}
      </p>
    </Modal>
    <Modal
      v-model:show="errorsModal"
      :title="t('fields.errorContent') + ': ' + fileName"
      style="width: 60%"
    >
      <div v-if="fileContent" style="height: 30rem">
        <Monaco :value="fileContent" lang="toml" read-only />
      </div>
      <p v-else>
        {{ t("fields.noContent") }}
      </p>
    </Modal>
    <Modal
      v-model:show="usageModal"
      :title="t('fields.usage') + ': ' + fileName"
      style="width: 40%"
    >
      <data-table
        :columns="usageColumns"
        :data="usageContent"
        :max-height="'40vh'"
      >
        <template #envName="usage">
          <n-button
            secondary
            type="primary"
            size="small"
            @click="router.push(`/env/elements/` + usage.EnvID)"
          >
            {{ usage.EnvName }}
          </n-button>
        </template>
      </data-table>
    </Modal>
  </div>
</template>

<style scoped>
.icon-color {
  color: var(--primaryColor);
}
</style>
