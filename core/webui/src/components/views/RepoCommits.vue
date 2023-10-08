<script lang="ts" setup>
import { useRoute, useRouter } from "vue-router";
import { CommitList } from "@/zodios/schemas/commitList";
import { z } from "zod";
import { useI18n } from "vue-i18n";
import moment from "moment";
import stc from "string-to-color";
import { useRouteParam } from "@/utils/router";
import { useRepoBranch } from "@/store/repoStore";
import { useUserStore } from "@/store/userStore";

const userStore = useUserStore();
const { t } = useI18n();
type commitListType = z.infer<typeof CommitList>;
const repoBranch = useRepoBranch();

const route = useRoute();
const router = useRouter();
let commitList = $ref<commitListType>([]);
let displayCommitCompare = $ref(false);
// let html = $ref("");
let selectedFile = $ref<string>("");
let files = $ref<string[]>([]);
let selectedSHA = $ref("");
let data = $ref("");
let branchList = $ref<string[] | undefined>(undefined);
let selectedBranch = $(useRouteParam("branch"));

watch($$(selectedBranch), () => {
  repoBranch.selectedBranch = selectedBranch;
});
let currentPage = $ref<number>(1);
let lastPage = $ref(false);
let commitSpinner = $ref(true);

const getFile = (path: string) => {
  api
    .get("/git-repo-commit-info-file-diff", {
      queries: {
        name: route.params.name as string,
        branch: route.params.branch as string,
        commit: selectedSHA as string,
        path: path,
      },
    })
    .then((res) => {
      data = res;
    });
};
const getCommitList = () => {
  commitSpinner = true;
  api
    .get("/git-repo-commit-list", {
      queries: {
        name: route.params.name as string,
        branch: selectedBranch as string,
        page: currentPage as number,
      },
    })
    .then((res) => {
      commitSpinner = false;
      commitList = res;
      lastPage = !res[res.length - 1].Files ? true : false;
    });
};
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

const options = [
  {
    title: t("fields.time"),
    template: "fromTime",
    width: "8rem",
  },
  {
    title: t("navbar.user"),
    template: "user",
    width: "6rem",
  },
  {
    title: "Hash",
    template: "sha",
    width: "7rem",
  },
  {
    title: t("fields.message"),
    key: "Message",
    template: "message",
    // width: "1fr",
  },
  {
    title: "",
    template: "show",
    width: "6.5rem",
  },
];
const openCompare = (sha: string) => {
  selectedSHA = sha;
  api
    .get("/git-repo-commit-info-file-list", {
      queries: {
        name: route.params.name as string,
        branch: route.params.branch as string,
        commit: sha,
      },
    })
    .then((res) => {
      displayCommitCompare = true;
      files = res;
      selectedFile = res[0];
      getFile(selectedFile);
    });
};

watch($$(selectedFile), () => {
  getFile(selectedFile);
});

const changePage = () => {
  getCommitList();
  router.push({
    query: {
      page: currentPage,
    },
  });
};

onMounted(() => {
  getBranchList();
  if (route.query.page) currentPage = parseInt(route.query.page as string);
  getCommitList();
});

watch($$(selectedBranch), () => {
  getCommitList();
  currentPage = 1;
});
let clicked = $ref(false);
const reloadF = () => {
  getCommitList();
  clicked = true;
  setTimeout(() => {
    clicked = false;
  }, 500);
};
</script>
<template>
  <div>
    <RepoTabs />
    <PageLayout>
      <div
        v-if="userStore.havePermission('Repo_View')"
        style="
          display: grid;
          gap: 1rem;
          grid-template-columns: 12rem 1fr;
          align-items: stretch;
          height: 100%;
        "
      >
        <BranchList
          v-model:branch="selectedBranch"
          :branch-list="branchList"
          @reload="getBranchList"
        />
        <n-card
          size="small"
          title="Commits"
          style="min-height: 10rem; position: relative"
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
          <Spinner :data="!commitSpinner" :showSpinnerProp="commitSpinner">
            <div
              style="
                display: flex;
                flex-direction: column;
                position: absolute;
                justify-content: space-between;
                height: 100%;
              "
            >
              <div style="display: flex">
                <data-table
                  style="flex: 1"
                  :columns="options"
                  :data="commitList"
                  :max-height="'75vh'"
                >
                  <template #fromTime="commit">
                    {{ moment(commit.TimeStamp * 1000).fromNow() }}
                  </template>
                  <template #sha="{ SHA }">
                    {{ SHA.substring(0, 8) }}
                  </template>
                  <template #user="row">
                    <n-tooltip trigger="hover">
                      <template #trigger>
                        <n-tag
                          :bordered="false"
                          size="tiny"
                          :color="{
                            color: stc(row.AuthorEmail) + '4D',
                            textColor: stc(row.AuthorEmail),
                          }"
                        >
                          <div style="display: flex; align-items: center">
                            <img
                              :src="`https://robohash.org/${row.AuthorName}.png?set=set4`"
                              alt="avatar"
                              style="height: 1.4rem; width: 1.4rem"
                            />
                            <p
                              style="
                                width: 2rem;
                                display: flex;
                                justify-content: center;
                              "
                            >
                              {{ row.AuthorInitials }}
                            </p>
                          </div>
                        </n-tag>
                      </template>
                      {{ row.AuthorEmail }}
                    </n-tooltip>
                  </template>

                  <template #show="{ SHA }">
                    <div
                      style="
                        display: flex;
                        justify-content: center;
                        cursor: pointer;
                        margin-right: 0.5em;
                      "
                    >
                      <n-button
                        v-if="SHA"
                        secondary
                        type="primary"
                        size="small"
                        class="change-version"
                        @click="openCompare(SHA)"
                      >
                        <n-icon size="18" class="icon-color">
                          <Mdi :path="mdiFileCompare" />
                        </n-icon>
                        {{ t("fields.diff") }}
                      </n-button>
                    </div>
                  </template>
                  <template #message="{ Message }">
                    <p v-if="Message.length < 100">{{ Message }}</p>
                    <n-tooltip v-else placement="top" style="max-width: 1000px">
                      <template #trigger>
                        {{ Message.substring(0, 100) }}...
                      </template>
                      {{ Message }}
                    </n-tooltip>
                  </template>
                </data-table>
              </div>
              <div
                style="
                  display: flex;
                  gap: 1rem;
                  float: right;
                  margin-top: 10px;
                  align-items: center;
                  justify-content: flex-end;
                "
                v-if="!(currentPage === 1 && lastPage)"
              >
                <n-button
                  secondary
                  type="primary"
                  size="small"
                  :disabled="currentPage < 2"
                  @click="() => currentPage-- && changePage()"
                  >Prev</n-button
                >
                <p>{{ currentPage }}</p>
                <n-button
                  secondary
                  type="primary"
                  size="small"
                  :disabled="lastPage"
                  @click="currentPage++ && changePage()"
                  >Next</n-button
                >
              </div>
            </div>
          </Spinner>
        </n-card>
      </div>
      <ErrorDialog v-else :msg="$t('messages.permissionDenied')" />
    </PageLayout>

    <Modal
      v-model:show="displayCommitCompare"
      :title="t('fields.fileComparison')"
      style="width: 80rem"
    >
      <div style="display: flex; gap: 1rem">
        <div style="height: 600px">
          <n-scrollbar>
            <FileList v-model:file="selectedFile" :file-list="files" />
          </n-scrollbar>
        </div>
        <MonacoDiff
          :value="data"
          v-if="data"
          style="height: 600px; flex: 1"
          :key="data"
          :filename="selectedFile"
        />
      </div>
    </Modal>
  </div>
</template>

<style scoped>
.icon-color {
  color: var(--primaryColor);
}
</style>
