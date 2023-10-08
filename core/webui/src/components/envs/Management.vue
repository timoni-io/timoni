<script lang="ts" setup>
import { useRoute } from "vue-router";
import { useEnv } from "@/store/envStore";
import { useMessage } from "naive-ui";
import z from "zod";
import { GitOpsMap } from "@/zodios/schemas/repos";
const message = useMessage();
const route = useRoute();
const { t } = useI18n();
defineProps<{
  manage: boolean;
}>();
const env = useEnv(computed(() => route.params.id as string));
let management = $ref(0);
let selectedRepo = $ref("");
let repoList = $ref<{ label: string; value: string }[]>([]);
let branchList = $ref<{ label: string; value: string }[]>([]);
let filePathList = $ref<{ label: string; value: string }[]>([]);
let selectedBranch = $ref("");
let globalGitOps: z.infer<typeof GitOpsMap>;
// let elementList = $ref<{ label: string; value: string }[]>([]);
let selectedElement = $ref("");
let assign = true;
let showPopover = $ref(false);
const openShowManagementModal = () => {
  selectedRepo = env.value.EnvInfo?.Env.GitOps.GitRepoName ?? "";
  selectedBranch = env.value.EnvInfo?.Env.GitOps.BranchName ?? "";
  selectedElement = env.value.EnvInfo?.Env.GitOps.FilePath ?? "";
  management = env.value.EnvInfo?.Env.GitOps.Enabled ? 1 : 0;
  api.get("/gitops-repo-map").then((res) => {
    globalGitOps = res;
    repoList = Object.keys(res).map((el) => ({ label: el, value: el }));
  });
};

// const fetchFilePaths = () => {
//   api
//     .get("/git-repo-env-map", {
//       queries: {
//         "git-repo-name": selectedRepo as string,
//         branch: selectedBranch as string,
//       },
//     })
//     .then((res) => {
//       filePathList = Object.values(res || {}).map((el) => {
//         return { label: el.Source.FilePath, value: el.Source.FilePath };
//       });
//     });
// };

watch(
  () => env.value,
  () => {
    if (assign) {
      assign = false;
      selectedRepo = env.value.EnvInfo?.Env.GitOps.GitRepoName ?? "";
      selectedBranch = env.value.EnvInfo?.Env.GitOps.BranchName ?? "";
      selectedElement = env.value.EnvInfo?.Env.GitOps.FilePath ?? "";
      management = env.value.EnvInfo?.Env.GitOps.Enabled ? 1 : 0;
    }
  }
);
// watch($$(management), () => {
//   if (management === 1) {
//     api.get("/git-repo-map").then((res) => {
//       repoList = Object.keys(res).map((el) => {
//         return { label: el, value: el };
//       });
//     });
//   }
// });
const uptdateBranchList = () => {
  branchList = Object.keys(globalGitOps[selectedRepo]).map((el) => ({
    label: el,
    value: el,
  }));
};
watch($$(selectedRepo), () => {
  if (selectedRepo && globalGitOps) {
    uptdateBranchList();
  }
  // api
  //   .get("/git-repo-branch-list", {
  //     queries: {
  //       name: selectedRepo as string,
  //       context: "App",
  //       level: 2,
  //     },
  //   })
  //   .then((res) => {
  //     branchList = res.map((el) => ({ label: el, value: el }));
  //   });
});
watch($$(selectedBranch), () => {
  if (selectedBranch && globalGitOps) {
    filePathList = globalGitOps[selectedRepo][selectedBranch].map((el) => ({
      label: el,
      value: el,
    }));
  }
});
let showManagement = $ref(false);
let sureApplyModal = $ref(false);

const applyManagement = (auto: boolean) => {
  sureApplyModal = false;
  api
    .post(
      "/env-gitops-set",
      {
        Enabled: auto,
        GitRepoName: auto ? selectedRepo : "",
        BranchName: auto ? selectedBranch : "",
        FilePath: auto ? selectedElement : "",
      },
      {
        queries: {
          env: route.params.id as string,
        },
      }
    )
    .then((res) => {
      // if (res === "permission denied") {
      //   message.error(t("messages.permissionDenied"));
      //   return;
      // }
      if (res === "ok") {
        message.success(t("messages.managmentChanged"));
        if (!auto) {
          selectedRepo = "";
          selectedBranch = "";
          selectedElement = "";
        }
        showManagement = !showManagement;
      } else {
        message.error(res);
      }
    });
};

const touched = computed(() => {
  let tempManagement = env.value.EnvInfo?.Env.GitOps.Enabled ? 1 : 0;
  if (tempManagement === 0 && management === 0) return false;

  let tempSelectedRepo = env.value.EnvInfo?.Env.GitOps.GitRepoName ?? "";
  let tempSelectedBranch = env.value.EnvInfo?.Env.GitOps.BranchName ?? "";
  let tempSelectedElement = env.value.EnvInfo?.Env.GitOps.FilePath ?? "";

  if (
    tempSelectedRepo !== selectedRepo ||
    tempSelectedBranch !== selectedBranch ||
    tempSelectedElement !== selectedElement ||
    tempManagement !== management
  )
    return true;

  return false;
});
const toltipHover = () => {
  if (env.value.EnvInfo?.Env.GitOps.Enabled) showPopover = true;
};

const gitopsName = $computed(() => {
  if (env.value.EnvInfo?.Env.GitOps.GitRepoName) {
    return (
      env.value.EnvInfo?.Env.GitOps.GitRepoName.slice(0, 20) +
      (env.value.EnvInfo?.Env.GitOps.GitRepoName.length > 20
        ? "... > "
        : " > ") +
      env.value.EnvInfo?.Env.GitOps.BranchName.slice(0, 20) +
      (env.value.EnvInfo?.Env.GitOps.BranchName.length > 20
        ? "... > "
        : " > ") +
      env.value.EnvInfo?.Env.GitOps.FilePath.slice(0, 20) +
      (env.value.EnvInfo?.Env.GitOps.FilePath.length > 20 ? "..." : "")
    );
  } else return "";
});
</script>

<template>
  <PopModal
    title="GitOps"
    style="width: 30rem"
    :show="showManagement"
    :touched="touched"
  >
    <template #trigger>
      <n-button
        :disabled="!manage"
        size="tiny"
        strong
        secondary
        type="primary"
        @click="openShowManagementModal"
        @mouseover="toltipHover"
        @mouseleave="showPopover = false"
      >
        {{
          env.EnvInfo?.Env.GitOps.Enabled ? gitopsName : t("actions.disabled")
        }}
        <template #icon>
          <n-icon class="icon">
            <mdi :path="mdiPencil" />
          </n-icon>
        </template>
      </n-button>
    </template>
    <template #content>
      <div
        v-if="!repoList.length"
        style="margin-top: 0.5rem; text-align: center"
      >
        {{ t("messages.noGitops") }}
      </div>
      <div v-else>
        <n-radio-group
          v-model:value="management"
          style="width: 100%; text-align: center"
          @update:value="(val:number) => (management = val)"
        >
          <n-radio-button :value="0">{{
            t("actions.disabled")
          }}</n-radio-button>
          <n-radio-button :value="1">{{ t("actions.enabled") }}</n-radio-button>
        </n-radio-group>
        <n-button
          v-if="management === 0 && env.EnvInfo?.Env.GitOps.Enabled"
          secondary
          type="primary"
          style="float: right; margin-top: 0.5rem"
          @click="applyManagement(false)"
          >{{ t("actions.apply") }}</n-button
        >
        <div style="clear: both"></div>
        <div v-if="management">
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
                  uptdateBranchList();
                  selectedBranch = '';
                  selectedElement = '';
                }
              "
            />
            <!-- 
            <Input
              style="width: 70%"
              :placeholder="'Repo name'"
              :focus="true"
              :removeWhiteSpace="true"
            /> -->
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
              @update:value="selectedElement = ''"
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
              >{{ t("fields.filePath") }}</span
            >

            <n-select
              v-model:value="selectedElement"
              filterable
              placeholder="Please select a branch"
              :options="filePathList"
            />
          </div>
          <n-button
            v-if="selectedElement"
            secondary
            type="primary"
            style="float: right; margin-top: 0.5rem"
            @click="sureApplyModal = true"
            >{{ t("actions.apply") }}</n-button
          >
          <div style="clear: both"></div>
        </div>
      </div>
    </template>
  </PopModal>
  <Modal
    v-model:show="sureApplyModal"
    title="GitOps"
    :show-icon="false"
    :showFooter="true"
    @positive-click="applyManagement(true)"
    style="width: 20rem"
  >
    {{ t("questions.gitOps") }}
  </Modal>
</template>
