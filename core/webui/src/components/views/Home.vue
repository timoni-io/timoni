<script setup lang="ts">
import { useDashboard } from "@/store/envStore";
import { useI18n } from "vue-i18n";
import { useMessage } from "naive-ui";
import { useRouter } from "vue-router";
// import { EnvMap } from "@/zodios/schemas/dashboard";
// import { z } from "zod";
import { useUserStore } from "@/store/userStore";
import { ComputedRef } from "vue";
import { Column } from "@/components/utils/DataTable.vue";

const userStore = useUserStore();
const router = useRouter();

const { t } = useI18n();
const message = useMessage();
const envRef = $ref<HTMLInputElement | null>(null);
let createEnvModal = $ref(false);
// use env store
let envNameError = $ref("");
// const envStore = useEnvStore();
const dashboard = useDashboard();

// let showEditGroup = $ref<{ [key: string]: boolean }>({});
// let groupName = $ref("");
let envName = $ref("");
let newEnvTag = $ref("");
let envTags = $ref<string[]>([]);
let selectedTag = $ref<string | null>(null);
let dynamicEnvironmentModal = $ref(false);
let showManagement = $ref(false);
let showConfirmationDeleteModal = $ref(false);
let repoList = $ref<{ label: string; value: string }[]>([]);
let dynamicEnvironmentsDetails = $ref<any>([]);
let dynamicEnvironmentModalDetails = $ref(false);
const getRepoList = () => {
  selectedRepo = "";
  selectedBranch = "";
  dirPath = "";
  api.get("/git-repo-map").then((res) => {
    repoList = Object.keys(res).map((el) => {
      return { label: el, value: el };
    });
  });
};

const columnsDetails: ComputedRef<Column[]> = computed(() => [
  {
    title: "File path",
    key: "filepath",
    width: "50%",
  },
  {
    title: "Environment name",
    template: "action",
    width: "50%",
  },
]);

const columns: ComputedRef<Column[]> = computed(() => [
  {
    title: t("navbar.repo"),
    key: "RepoName",
    width: "15%",
  },
  {
    title: t("fields.branch"),
    key: "BranchName",
    width: "15%",
  },
  {
    title: "Dir path",
    key: "DirPath",
    width: "15%",
  },
  {
    title: t("fields.found"),
    // key: "EnvironmentFound",
    template: "found",
    width: "15%",
  },
  {
    title: "",
    template: "action",
    width: "2%",
  },
]);
watch(
  () => envRef,
  () => {
    if (envRef) envRef.focus();
  }
);

const addEnv = () => {
  envName = "";
  envTags = [];
}

watch(
  () => selectedTag,
  () => {
    selectedTag = null;
  }
);

const createEnvironment = async () => {
  api
    .post("/env-create", {
      name: envName.trim(),
    })
    .then((res) => {
      if (res.substring(0, 3) === "env") {
        message.success(t("messages.createdEnvironment"));
        setTimeout(() => {
          router.push(`/env/${res}`);
        });
        return;
      }

      message.error(res);
    });
};

const createEnvF = () => {
  if (!envName.trim()) {
    envNameError = t("messages.nameRequired");
  }

  if (envName.trim()) {
    if (envName.trim().length >= 31) {
      return;
    }
    createEnvironment();
  }
};

watch(
  () => envName,
  () => {
    if (envName.trim().length >= 31) envNameError = t("messages.tooLongName");
    else envNameError = "";
  }
);

const envsToDelete = $computed(() =>
  Object.values((dashboard.envMap || {}) as Object)
    ?.filter((env) => env.ToDelete)
    .map((env) => env.Name)
);

watch($$(envsToDelete), (newv, old) => {
  if (newv.length < old.length)
    old
      .filter((el) => !newv.includes(el))
      .forEach(() => {
        message.success(t("messages.envDeleted"));
      });
});
let branchList = $ref<{ label: string; value: string }[]>([]);
let selectedRepo = $ref("");
let selectedBranch = $ref("");
let dirPath = $ref("");
let dynamicEnvironments = $ref([]);
watch($$(selectedRepo), () => {
  if (selectedRepo)
    api
      .get("/git-repo-branch-list", {
        queries: {
          name: selectedRepo as string,
          level: 2,
        },
      })
      .then((res) => {
        branchList = res.map((el) => ({ label: el, value: el }));
      });
});

const createDynamicEnv = () => {
  api
    .post("/env-dynamic-sources-add", {
      RepoName: selectedRepo,
      BranchName: selectedBranch,
      DirPath: dirPath,
    })
    .then(() => {
      showManagement = !showManagement;
      api.get("/env-dynamic-sources-map").then((res) => {
        dynamicEnvironments = Object.values(res || {});
      });
    });
};
onMounted(() => {
  api.get("/env-dynamic-sources-map").then((res) => {
    dynamicEnvironments = Object.values(res || {});
  });
});
let envToDelete = $ref("");
const removeDynamicEnvironment = () => {
  api
    .get("/env-dynamic-sources-delete", {
      queries: {
        envSourceID: envToDelete as string,
      },
    })
    .then((res) => {
      if (res === "ok") {
        showConfirmationDeleteModal = false;
        api.get("/env-dynamic-sources-map").then((res) => {
          dynamicEnvironments = Object.values(res || {});
        });
      }
    });
};
const createDetails = (row: any) => {
  dynamicEnvironmentsDetails = Object.entries(row.Environments).map((el) => {
    return {
      filepath: el[0],
      // @ts-ignore
      name: el[1].Name,
      // @ts-ignore
      id: el[1].ID,
    };
  });
};
</script>

<template>
  <!-- <NavbarTabs /> -->
  <PageLayout>
    <div class="home-page" v-if="userStore.havePermission('Env_View')">
      <!-- <div class="home-page" v-if="true"> -->
      <div class="items-section">
        <n-card size="small">
          <n-space
            class="card-header"
            style="gap: none; justify-content: space-between"
          >
            <div class="panel-header">
              <p>{{ t("objects.environment", 2) }}</p>
            </div>
            <div style="display: flex; align-items: center; gap: 0.5rem">
              <PopModal
                :title="t('actions.addEnv')"
                :show="createEnvModal"
                :touched="
                  envName !== '' ||
                  envTags.length > 0 ||
                  newEnvTag.length > 0
                "
                :width="'25rem'"
              >
                <template #trigger>
                  <n-button
                    strong
                    secondary
                    type="primary"
                    size="tiny"
                    :disabled="
                      !userStore.havePermission('Glob_CreateAndDeleteEnvs')
                    "
                    @click="addEnv"
                  >
                    <template #icon>
                      <n-icon><mdi :path="mdiPlus" /></n-icon> </template
                    >{{ t("objects.environment") }}
                  </n-button>
                </template>
                <template #content>
                  <div style="display: flex; gap: 0.5rem">
                    <div style="display: flex; flex-flow: column; gap: 1.2rem">
                      <div
                        style="display: flex; gap: 0.5rem; align-items: center"
                      >
                        <span>{{ t("fields.name") }}</span>
                        <Input
                          style="width: 100%"
                          :placeholder="t('fields.envName')"
                          :errorMessage="envNameError"
                          :focus="true"
                          :removeWhiteSpace="true"
                          @update:value="
                            (v: string) => {
                              envName = v;
                              envNameError = ''
                            }
                          "
                        />
                      </div>
                    </div>
                    <n-button
                      secondary
                      type="primary"
                      @click="createEnvF"
                      >{{ t("actions.confirm") }}</n-button
                    >
                  </div>
                </template>
              </PopModal>
              <EnvCloneSelector
                v-if="Object.keys(dashboard.envMap || {}).length"
                :manage="userStore.havePermission('Glob_CreateAndDeleteEnvs')"
              />

              <n-button
                size="tiny"
                secondary
                type="primary"
                @click="dynamicEnvironmentModal = true"
              >
                <Mdi :path="mdiCog" width="15" />
                <!-- {{ t("objects.dynamicEnvironment") }} -->
              </n-button>
            </div>
          </n-space>
          <div
            class="items-grid"
            v-if="Object.keys(dashboard.envMap || {})?.length"
          >
            <EnvCard
              v-for="(env, envID) in dashboard.envMap"
              :key="envID"
              :env-id="envID"
              :env="env"
              :element-statuses="env.Status"
              :to-delete="env.ToDelete"
              class="listing-card"
            >
            </EnvCard>
          </div>
          <div v-else>
            <div class="base-alert">
              {{ t("messages.noEnvironments") }}
            </div>
          </div>
        </n-card>
      </div>
      <Modal
        v-model:show="dynamicEnvironmentModal"
        :title="t('objects.dynamicEnvironment')"
        style="width: 1200px"
      >
        <PopModal title="source" style="width: 30rem" :show="showManagement">
          <template #trigger>
            <n-button
              secondary
              type="primary"
              style="width: 5rem; margin-bottom: 1rem"
              size="tiny"
              @click="getRepoList"
            >
              <Mdi :path="mdiPlus" width="15" style="margin-right: 3px" />
              Source
            </n-button>
          </template>
          <template #content>
            <div>
              <div>
                <div
                  style="
                    display: flex;
                    justify-content: space-between;
                    gap: 0.5rem;
                    margin-top: 1rem;
                  "
                >
                  <span
                    style="
                      white-space: nowrap;
                      width: 30%;
                      height: 34px;
                      display: flex;
                      align-items: center;
                    "
                    >Repo</span
                  >
                  <n-select
                    v-model:value="selectedRepo"
                    filterable
                    placeholder="Please select a repo"
                    :options="repoList"
                    @update:value="
                      () => {
                        selectedBranch = '';
                      }
                    "
                  />
                </div>
                <div
                  v-if="selectedRepo"
                  style="
                    display: flex;
                    justify-content: space-between;
                    gap: 0.5rem;
                    margin-top: 0.5rem;
                  "
                >
                  <span
                    style="
                      white-space: nowrap;
                      width: 30%;
                      height: 34px;
                      display: flex;
                      align-items: center;
                    "
                    >{{ t("fields.branch") }}</span
                  >

                  <n-select
                    v-model:value="selectedBranch"
                    filterable
                    placeholder="Please select a branch"
                    :options="branchList"
                  />
                </div>
                <div
                  v-if="selectedBranch"
                  style="
                    display: flex;
                    justify-content: space-between;
                    gap: 0.5rem;
                    margin-top: 0.5rem;
                  "
                >
                  <span
                    style="
                      white-space: nowrap;
                      width: 30%;
                      height: 34px;
                      display: flex;
                      align-items: center;
                    "
                    >Dir path</span
                  >

                  <n-input v-model:value="dirPath" placeholder="Dir path" />
                </div>
                <n-button
                  v-if="dirPath"
                  secondary
                  type="primary"
                  style="float: right; margin-top: 0.5rem"
                  @click="createDynamicEnv"
                  >{{ t("actions.apply") }}</n-button
                >
                <div style="clear: both"></div>
              </div>
            </div>
          </template>
        </PopModal>

        <data-table :columns="columns" :data="dynamicEnvironments">
          <template #found="row">
            <n-button
              size="tiny"
              secondary
              type="primary"
              @click="
                () => {
                  createDetails(row);
                  dynamicEnvironmentModalDetails = true;
                }
              "
            >
              {{ Object.values(row.Environments || {}).length }}
            </n-button>
          </template>
          <template #action="row">
            <n-button
              size="tiny"
              secondary
              type="error"
              style="float: right"
              @click="
                () => {
                  envToDelete = row.ID;
                  showConfirmationDeleteModal = true;
                }
              "
              >{{ t("actions.remove") }}</n-button
            >
          </template>
        </data-table>
      </Modal>
      <Modal
        v-model:show="dynamicEnvironmentModalDetails"
        :title="t('objects.dynamicEnvironment')"
        style="width: 600px"
      >
        <data-table
          :columns="columnsDetails"
          :data="dynamicEnvironmentsDetails"
        >
          <template #action="row">
            <n-button
              size="tiny"
              secondary
              type="primary"
              @click="$router.push(`/env/${row.id}`)"
            >
              {{ row.name }}
            </n-button>
          </template>
        </data-table>
      </Modal>
      <Modal
        :showFooter="true"
        @positive-click="removeDynamicEnvironment"
        v-model:show="showConfirmationDeleteModal"
        :title="t('actions.deleteEnv')"
        style="width: 30rem"
      >
        <p>{{ t("questions.sure") }}</p>
      </Modal>
    </div>
    <ErrorDialog v-else :msg="$t('messages.permissionDenied')" />
  </PageLayout>
</template>

<style scoped>
.home-page {
  display: grid;
  grid-template-columns: 1fr;
  gap: 1rem;
}
.items-section {
  width: 100%;
  min-width: 0;
}

.items-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(15rem, 1fr));
  gap: 1rem;
  padding: 1rem;
}

.header__extra {
  display: flex;
  gap: 1rem;
}
</style>
