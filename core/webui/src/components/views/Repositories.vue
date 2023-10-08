<script setup lang="ts">
import { useDashboard } from "@/store/envStore";
import { useI18n } from "vue-i18n";
import { useMessage } from "naive-ui";
import { useRouter } from "vue-router";
import { useUserStore } from "@/store/userStore";

const userStore = useUserStore();

const message = useMessage();

const { t } = useI18n();

const dashboard = useDashboard();
let localRepoName = $ref("");
const router = useRouter();
let addLocalRepo = $ref(true);
const openLocalRepoPopModal = () => {
  localRepoName = "";
  localErrorMsg = "";
};
const createLocal = () => {
  // if (!localRepoName) {
  //   message.error("Name is required");
  //   return;
  // }
  api
    .post("/git-repo-create", {
      Name: localRepoName,
    })
    .then((res) => {
      // if (res && res)
      if (res === localRepoName) {
        message.success(t("messages.localRepoCreated"));
        addLocalRepo = false;
        router.push("/code/" + res);
      } else message.error(res);
    });
};

const envRef = $ref(null as HTMLInputElement | null);
watch(
  () => envRef,
  () => {
    envRef?.focus();
  }
);

// validation
let localErrorMsg = $ref("");

watch(
  () => localRepoName,
  () => {
    if (!/^[a-zA-Z0-9_.-]*$/.test(localRepoName))
      localErrorMsg = t("messages.invalidName");
    else localErrorMsg = "";
  }
);
</script>

<template>
  <div>
    <PageLayout>
      <div v-if="userStore.havePermission('Repo_View')">
        <n-card size="small" :title="t('objects.localRepo', 2)" v-if="userStore.HideGitRepoLocal===false">
          <template #header-extra>
            <PopModal
              :title="
                t('actions.add') + ' ' + t('objects.localRepo').toLowerCase()
              "
              @negative-click="() => {}"
              :width="'30rem'"
              :show-footer="{
                positiveText: t('actions.confirm'),
                negativeText: t('actions.cancel'),
              }"
              @positive-click="createLocal"
              :touched="localRepoName.length > 0"
              :show="addLocalRepo"
            >
              <template #trigger>
                <n-button
                  :disabled="
                    !userStore.havePermission('Glob_CreateAndDeleteGitRepos')
                  "
                  @click="openLocalRepoPopModal"
                  strong
                  secondary
                  type="primary"
                  size="tiny"
                >
                  {{ t("objects.localRepo") }}
                  <template #icon>
                    <n-icon><mdi :path="mdiPlus" /></n-icon>
                  </template>
                </n-button>
              </template>
              <template #content>
                <div
                  style="
                    display: flex;
                    justify-content: space-between;
                    gap: 0.5rem;
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
                    >{{ t("fields.repoName") }}</span
                  >
                  <!-- <n-input
                ref="envRef"
                type="text"
                placeholder="Repo name"
                v-model:value="localRepoName"
                @keyup.enter="createLocal"
              /> -->
                  <Input
                    style="width: 70%"
                    :placeholder="t('fields.repoName')"
                    :errorMessage="localErrorMsg"
                    :focus="true"
                    :removeWhiteSpace="true"
                    @keyup.enter="createLocal"
                    @update:value="
                  (v: string) => {
                    localRepoName = v;
                  }
                "
                  />
                </div>
                <div></div>
              </template>
            </PopModal>
          </template>
          <div class="repo-grid">
            <RepoCard
              v-for="repo in Object.values(dashboard.repoMap || {}).filter(
                (el) => el.Local
              )"
              :key="repo.Name"
              :repo="repo"
              class="listing-card"
            />
          </div>
          <div
            class="base-alert"
            v-if="
              dashboard.repoMap &&
              !Object.values(dashboard.repoMap).filter((el) => el.Local).length
            "
          >
            {{ t("messages.noReposToShowLocal") }}
          </div>
        </n-card>
        <n-card
          size="small"
          :title="t('objects.remoteRepo', 2)"
          style="margin-top: 1rem"
        >
          <template #header-extra>
            <AddRemoteRepo
              :disabled="
                !userStore.havePermission('Glob_CreateAndDeleteGitRepos')
              "
              :btnSize="'tiny'"
              :noPush="false"
            />
          </template>
          <div
            class="base-alert"
            v-if="
              dashboard.repoMap &&
              !Object.values(dashboard.repoMap).filter((el) => !el.Local).length
            "
          >
            {{ t("messages.noReposToShowRemote") }}
          </div>
          <div class="repo-grid" v-else>
            <RepoCard
              v-for="repo in Object.values(dashboard.repoMap || {}).filter(
                (el) => !el.Local
              )"
              :key="repo.Name"
              :repo="repo"
              class="listing-card"
            />
          </div>
        </n-card>
      </div>
      <ErrorDialog v-else :msg="$t('messages.permissionDenied')" />
    </PageLayout>
  </div>
</template>

<style scoped>
.repo-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(15rem, 1fr));
  gap: 1rem;
}
.form {
  display: flex;
  justify-content: space-between;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
}
.form-label {
  height: 34px;
  display: flex;
  align-items: center;
  width: 10rem;
}
</style>
<style>
.listing-card:hover {
  background: rgb(124, 124, 124);
}
</style>
