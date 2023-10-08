<script setup lang="ts">
import { useRepo } from "@/store/repoStore";
import { useRoute, useRouter } from "vue-router";
// const { t } = useI18n();
import { useMessage } from "naive-ui";
import { useUserStore } from "@/store/userStore";

const userStore = useUserStore();
const message = useMessage();
const { t } = useI18n();

let defaultBranch = $ref("");
const route = useRoute();
const router = useRouter();
const repo = $(useRepo(computed(() => route.params.name as string)));
let elementTypes = $ref<{ label: string; value: string }[]>([]);
let newRepoUrl = $ref("");
let changeURLmodal = $ref(true);
let showDeleteModal = $ref(false);
watch($$(repo), () => {
  defaultBranch = repo?.DefaultBranch as string;
  newRepoUrl = repo?.RemoteURL as string;
  if (repo?.Name)
    api
      .get("/git-repo-branch-list", {
        queries: {
          name: repo?.Name,
          level: 2,
        },
      })
      .then((res) => {
        elementTypes = res.map((el) => {
          return {
            label: el,
            value: el,
          };
        });
      });
});
watch($$(defaultBranch), (newBranch, oldBranch) => {
  if (oldBranch && repo?.Name && newBranch) {
    api
      .get("/git-repo-update", {
        queries: {
          "git-repo": repo?.Name as string,
          default: newBranch,
        },
      })
      .then((res) => {
        if (res === "ok") {
          message.success(t("messages.defaultBranchChanges"));
        } else message.error(res);
      });
  }
});
const changeSourceUrl = () => {
  api
    .get("/git-repo-remote-access-update", {
      queries: {
        "git-repo": repo?.Name as string,
        url: newRepoUrl as string,
        login: "",
        password: "",
      },
    })
    .then((res) => {
      if (res === "ok") {
        message.success(t("messages.remoteUrlChanged"));
        changeURLmodal = !changeURLmodal;
      } else message.error(res);
    });
};

const copyCloneUrl = () => {
  navigator.clipboard.writeText(repo?.CloneURL || "");
  message.success(t("messages.copied"));
};

const copyGitAccess = () => {
  navigator.clipboard.writeText(repo?.AccessToCode || "");
  message.success(t("messages.copied"));
};

const deleteRepo = () => {
  api
    .get("/git-repo-delete", {
      queries: { name: route.params.name as string },
    })
    .then((res) => {
      if (res === "ok") {
        router.push(`/code`);
        message.success(t("messages.repoDeleted"));
      } else {
        message.error(res as string);
      }
    });
};
</script>
<template>
  <div>
    <RepoTabs />
    <PageLayout>
      <n-card
        v-if="userStore.havePermission('Repo_View')"
        :title="$t('objects.settings')"
        size="small"
        style="height: calc(100vh - 5.1rem)"
      >
        <template #header-extra>
          <n-button
            secondary
            size="tiny"
            type="error"
            :disabled="
              !userStore.havePermission('Glob_CreateAndDeleteGitRepos')
            "
            @click="showDeleteModal = true"
            >{{ $t("actions.deleteRepo") }}
            <template #icon>
              <n-icon>
                <Mdi :path="mdiTrashCan" />
              </n-icon>
            </template>
          </n-button>
        </template>
        <div style="display: flex; gap: 1rem; flex-direction: column">
          <div v-if="repo?.RemoteURL" class="form">
            <p class="form-label" style="white-space: nowrap">
              {{ t("fields.sourceURL") }}
            </p>
            <PopModal
              :title="'change source url'"
              @positive-click="changeSourceUrl"
              @negative-click="() => {}"
              :show="changeURLmodal"
              :width="'27rem'"
              :show-footer="true"
            >
              <template #trigger>
                <n-button
                  secondary
                  :disabled="!userStore.havePermission('Repo_SettingsManage')"
                  type="primary"
                  style="width: 20%; justify-content: flex-start"
                >
                  <n-icon style="margin-right: 0.5em">
                    <Mdi :path="mdiPen" />
                  </n-icon>
                  <div style="inline-size: 90%; overflow: hidden">
                    {{ repo?.RemoteURL }}&nbsp;&nbsp;
                  </div>
                </n-button>
              </template>
              <template #content>
                <div style="display: flex; align-items: center; gap: 0.5rem">
                  <p>URL:</p>
                  <Input
                    style="width: 70%"
                    :placeholder="'https://google.com'"
                    :focus="true"
                    :removeWhiteSpace="true"
                    :valueFromParent="repo?.RemoteURL"
                    @keyup.enter="changeSourceUrl"
                    @update:value="
                  (v: string) => {
                    newRepoUrl = v;
                  }
                "
                  />
                </div>
              </template>
            </PopModal>
          </div>
          <div class="form" style="min-width: 300px">
            <p class="form-label" style="white-space: nowrap">
              {{ t("fields.defaultBranch") }}
            </p>
            <n-select
              v-model:value="defaultBranch"
              :options="elementTypes"
              style="width: 20%"
            />
          </div>
          <div class="form" style="min-width: 300px">
            <p class="form-label" style="white-space: nowrap">
              {{ t("objects.cloneUrl") }}
            </p>
            <n-button
              secondary
              type="primary"
              icon-placement="right"
              class="clone"
              style="width: 20%; justify-content: space-between"
              @click="copyCloneUrl"
            >
              <div style="inline-size: 90%; overflow: hidden">
                {{ repo?.CloneURL }}&nbsp;&nbsp;
              </div>
              <n-icon>
                <Mdi :path="mdiContentCopy" />
              </n-icon>
            </n-button>
          </div>
          <div class="form" style="min-width: 300px">
            <p class="form-label" style="white-space: nowrap">
              {{ t("actions.addAccess") }}
            </p>
            <n-button
              secondary
              type="primary"
              icon-placement="right"
              class="clone"
              style="width: 20%; justify-content: space-between"
              @click="copyGitAccess"
            >
              <div style="inline-size: 90%; overflow: hidden">
                {{ repo?.AccessToCode }}&nbsp;&nbsp;
              </div>
              <n-icon>
                <Mdi :path="mdiContentCopy" />
              </n-icon>
            </n-button>
          </div>
        </div>
      </n-card>
      <ErrorDialog v-else :msg="$t('messages.permissionDenied')" />
    </PageLayout>
  </div>
  <Modal
    v-model:show="showDeleteModal"
    show-footer
    :title="$t('actions.deleteRepo')"
    style="max-width: 20rem"
    @positive-click="deleteRepo()"
  >
    {{ $t("questions.sure") }}
  </Modal>
</template>
<style scoped lang="scss">
.form {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.form-label {
  height: 34px;
  display: flex;
  align-items: center;
  width: 10rem;
}

.clone {
  & :deep(.n-button__content) {
    justify-content: space-between;
    width: 100%;
  }
}
</style>
