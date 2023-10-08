<script lang="ts" setup>
import { useI18n } from "vue-i18n";
import { GitEnvS } from "@/zodios/schemas/elements";
import { z } from "zod";
import { useRoute } from "vue-router";
import { computed, ComputedRef } from "vue";
import { Column } from "@/components/utils/DataTable.vue";
import { useRouteParam } from "@/utils/router";
import { useRepoBranch } from "@/store/repoStore";
import { useUserStore } from "@/store/userStore";
import { useMessage } from "naive-ui";
import { useRouter } from "vue-router";

const userStore = useUserStore();
const message = useMessage();
const repoBranch = useRepoBranch();
const { t } = useI18n();
const route = useRoute();
const router = useRouter();

let errorsModal = $ref(false);
let elementsLoading = $ref(false);

let addEnv = $ref(false);
let envName = $ref("");
let envNameError = $ref("");
let errorTeamMessage = $ref("");
let teamList = $ref<{ value: string; label: string }[]>();
let selectedTeam = $ref<string[]>([]);
let addEnvFilePath = $ref("");

let fileModal = $ref(false);
let fileContent = $ref("");
let fileName = $ref("");
let filePath = $ref("");

let branchList = $ref<string[]>([]);
let selectedRepo = $ref("");
let selectedBranch = $(useRouteParam("branch"));
watch($$(selectedBranch), () => {
  repoBranch.selectedBranch = selectedBranch;
  reloadF();
});

watch(
  () => envName,
  () => {
    if (envName.trim().length >= 31) envNameError = t("messages.tooLongName");
    else envNameError = "";
  }
);

const file = (env: z.infer<typeof GitEnvS>) => {
  fileModal = true;
  fileContent = window.atob(env?.FileContent);
  fileName = env?.Name as string;
  filePath = env?.Source?.FilePath as string;
};

const applyManagement = (envID: string) => {
  api
    .post(
      "/env-gitops-set",
      {
        Enabled: true,
        GitRepoName: selectedRepo,
        BranchName: selectedBranch,
        FilePath: addEnvFilePath,
      },
      {
        queries: {
          env: envID as string,
        },
      }
    )
    .then((res) => {
      if (res === "ok") {
        message.success(t("messages.createdEnvironment"));
        addEnv = false;
        setTimeout(() => {
          router.push(`/env/${envID}`);
        }, 500);
      } else {
        message.error(res);
      }
    });
};

const createEnvironment = async () => {
  api
    .post("/env-create", {
      name: envName.trim(),
      teams: selectedTeam as string[],
    })
    .then((res) => {
      if (res.substring(0, 3) === "env") {
        applyManagement(res);
        return;
      }

      message.error(res);
    });
};

const createEnvF = () => {
  if (!envName.trim()) {
    envNameError = t("messages.nameRequired");
  }
  if (!selectedTeam.length) {
    errorTeamMessage = t("messages.requiredTeam");
  }
  if (envName.trim() && selectedTeam && selectedTeam!.length) {
    if (envName.trim().length >= 31) {
      return;
    }
    createEnvironment();
  }
};

const columns: ComputedRef<Column[]> = computed(() => [
  {
    title: t("fields.name"),
    key: "Name",
    width: "20%",
  },
  {
    title: t("fields.description"),
    key: "Description",
    width: "20%",
  },
  {
    title: t("fields.filePath"),
    key: "Source.FilePath",
    width: "20%",
  },
  {
    title: t("objects.element", 2),
    template: "elements",
    width: "20%",
  },
  {
    title: "",
    template: "action",
    width: "10%",
  },
]);

let envs = $ref<Record<string, z.infer<typeof GitEnvS>>>();

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
const getRepoEnvList = () => {
  // envs = [];
  elementsLoading = true;
  api
    .get("/git-repo-env-map", {
      queries: {
        "git-repo-name": route.params.name as string,
        branch: selectedBranch as string,
      },
    })
    .then((res) => {
      envs = res || {};
      elementsLoading = false;
    });
};

const getTeams = () => {
  if (userStore && userStore.teams) {
    api.get("/team-list").then((res) => {
      teamList = res.map((el) => {
        return {
          value: el.ID,
          label: el.Name,
        };
      });
    });
  }
};
let clicked = $ref(false);

const reloadF = () => {
  getRepoEnvList();
  clicked = true;
  setTimeout(() => {
    clicked = false;
  }, 500);
};
let screenWidth = $ref<number>(0);
onBeforeMount(() => {
  screenWidth = window.innerWidth;
});
onMounted(() => {
  getBranchList();
  getRepoEnvList();
  getTeams();

  window.addEventListener("resize", () => {
    screenWidth = window.innerWidth;
  });
});
</script>

<template>
  <div>
    <RepoTabs />
    <PageLayout>
      <Spinner :data="envs" v-if="userStore.havePermission('Repo_View')">
        <div style="display: grid; gap: 1rem; grid-template-columns: 12rem 1fr">
          <BranchList
            v-model:branch="selectedBranch"
            :branch-list="branchList"
            @reload="getBranchList"
          />
          <n-card :title="t('objects.environment', 2)" size="small">
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
                Object.values(envs || {}).sort((a, b) =>
                  a.Name.localeCompare(b.Name)
                )
              "
              :max-height="'80vh'"
            >
              <template #elements="env">
                {{ Object.keys(env.Element || {}).length }}
              </template>
              <template #action="env">
                <div
                  style="
                    width: 100%;
                    height: 100%;
                    display: flex;
                    gap: 0.5em;
                    justify-content: flex-end;
                  "
                >
                  <n-tooltip trigger="hover" :disabled="screenWidth > 1300">
                    <template #trigger>
                      <n-button
                        size="tiny"
                        secondary
                        type="primary"
                        :style="
                          screenWidth > 1300
                            ? 'min-width: 3rem'
                            : 'min-width: 2rem'
                        "
                        @click="
                          () => {
                            addEnv = true;
                            envName = env.Name;
                            addEnvFilePath = env.Source.FilePath;
                            selectedRepo = env.Source.RepoName;
                          }
                        "
                      >
                        +
                        <span
                          v-if="screenWidth > 1300"
                          style="margin-left: 10px"
                          >{{ t("actions.create") }}</span
                        >
                      </n-button>
                    </template>
                    {{ t("actions.create") }}
                  </n-tooltip>

                  <n-tooltip
                    v-if="env.Error"
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
                            fileContent = env.Error;
                            fileName = env.Name;
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

                  <n-tooltip trigger="hover" :disabled="screenWidth > 1300">
                    <template #trigger>
                      <n-button
                        size="tiny"
                        secondary
                        type="primary"
                        :style="
                          screenWidth > 1300 ? 'width: 3rem' : 'min-width: 2rem'
                        "
                        @click="file(env)"
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
      v-model:show="addEnv"
      :title="t('actions.add') + ' ' + t('objects.environment').toLowerCase()"
      style="width: 25%"
    >
      <div style="display: flex; gap: 0.5rem; justify-content: space-between">
        <div style="display: flex; flex-flow: column; gap: 1.2rem">
          <div style="display: flex; gap: 0.5rem; align-items: center">
            <span style="width: 30%">{{ t("fields.name") }}</span>
            <Input
              style="width: 100%"
              :placeholder="t('fields.envName')"
              :errorMessage="envNameError"
              :focus="true"
              :valueFromParent="envName"
              :removeWhiteSpace="true"
              @update:value="
                            (v: string) => {
                              envName = v;
                              envNameError = ''
                            }
                          "
            />
          </div>
          <div style="display: flex; gap: 0.5rem; align-items: center">
            <span style="width: 30%">{{ t("fields.team") }}</span>
            <div style="width: 100%">
              <n-select
                multiple
                v-model:value="selectedTeam"
                :options="teamList"
                :placeholder="t('fields.selectTeam')"
                @update:value="errorTeamMessage = ''"
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
                :style="
                  errorTeamMessage
                    ? { transform: 'translateY(0rem)', opacity: 1 }
                    : {}
                "
              >
                {{ errorTeamMessage }}
              </div>
            </div>
          </div>
        </div>
        <n-button
          style="height: 87px"
          secondary
          type="primary"
          @click="createEnvF"
          >{{ t("actions.confirm") }}</n-button
        >
      </div>
    </Modal>
  </div>
</template>

<style scoped>
.icon-color {
  color: var(--primaryColor);
}
</style>
