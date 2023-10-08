<script setup lang="ts">
import { useDashboard } from "@/store/envStore";
import { useI18n } from "vue-i18n";
import { useMessage } from "naive-ui";
import { RepoMap } from "@/zodios/schemas/dashboard";
import { z } from "zod";

const message = useMessage();

const { t } = useI18n();
const dashboard = useDashboard();
let localRepoName = $ref("");
let openAddRemoteValidationModal = $ref(false);
let projectURL = $ref("");
let formName = $ref("");
// let formBranch = ref("");
let formLogin = $ref("");
let formPassword = $ref("");
let errorProjects = $ref("");
let modalLoader = $ref(false);
let modalValidationLoader = $ref(false);
const validProjectURL = $computed(() => {
  return /^(?:(?:(?:https?|ftp):)?\/\/)(?:\S+(?::\S*)?@)?(?:(?!(?:10|127)(?:\.\d{1,3}){3})(?!(?:169\.254|192\.168)(?:\.\d{1,3}){2})(?!172\.(?:1[6-9]|2\d|3[0-1])(?:\.\d{1,3}){2})(?:[1-9]\d?|1\d\d|2[01]\d|22[0-3])(?:\.(?:1?\d{1,2}|2[0-4]\d|25[0-5])){2}(?:\.(?:[1-9]\d?|1\d\d|2[0-4]\d|25[0-4]))|(?:(?:[a-z\u00a1-\uffff0-9]-*)*[a-z\u00a1-\uffff0-9]+)(?:\.(?:[a-z\u00a1-\uffff0-9]-*)*[a-z\u00a1-\uffff0-9]+)*(?:\.(?:[a-z\u00a1-\uffff]{2,})))(?::\d{2,5})?(?:[/?#]\S*)?$/i.test(
    projectURL
  );
});
watch(
  () => projectURL,
  () => {
    if (validProjectURL) {
      let splitted = projectURL;
      let splittedList: string[];
      if (projectURL.slice(0, 8) === "https://")
        splitted = splitted.split("https://")[1];
      else if (projectURL.slice(0, 7) === "http://")
        splitted = splitted.split("http://")[1];
      splittedList = splitted.split("/");
      if (splittedList.length > 1) {
        // formName = splittedList.slice(1).join("/").replace(".git", "");
        formName = splittedList[splittedList.length - 1].replace(".git", "");
      }
    }
  }
);
let addLocalRepo = $ref(true);
const openLocalRepo = () => {
  localRepoName = "";
  localErrorMsg = "";
};
const createLocal = () => {
  if (!localRepoName) {
    message.error("Name is required");
  }
  api
    .post("/git-repo-create", {
      Name: localRepoName,
    })
    .then((res) => {
      if (res === localRepoName) {
        message.success("Local repo created");
        addLocalRepo = false;
      } else message.error(res);
    })
    .then(() => {
      addLocalRepo = true;
    });
};

let openAddRemoteModal = $ref(false);
const createRemote = () => {
  modalLoader = true;
  api
    .post("/git-repo-create-remote", {
      URL: projectURL,
      Name: formName,
    })
    .then((res) => {
      modalLoader = false;
      if (typeof res === "string") {
        message.error(res);
        if (res === "login and password are required") {
          // console.log("validation ppopup");
          openAddRemoteValidationModal = true;
          return;
        }
        // errorProjects = res;
      } else {
        message.success("successfully added remote repository");
        openAddRemoteModal = false;
      }
    });
};
const createRemoteWithValidation = () => {
  modalValidationLoader = true;
  api
    .post("/git-repo-create-remote", {
      URL: projectURL,
      Name: formName,
      Login: formLogin,
      Password: formPassword,
    })
    .then((res) => {
      modalValidationLoader = false;
      if (typeof res === "string") {
        message.error(res);
        // errorProjects = res;
      } else {
        message.success("Pomyślnie dodano projekt zdalny");
        openAddRemoteModal = false;
      }
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
let projectURLErrorMsg = $ref("");
let formNameErrorMsg = $ref("");
watch(
  () => localRepoName,
  () => {
    if (!/^[a-zA-Z0-9_.-]*$/.test(localRepoName))
      localErrorMsg = t("messages.invalidName");
    else localErrorMsg = "";
  }
);
watch(
  () => projectURL,
  () => {
    if (
      (!validProjectURL ||
        !/^[a-zA-Z0-9_.-/]*$/.test(
          projectURL.split("/")[projectURL.split("/").length - 1]
        )) &&
      projectURL.length > 0
    )
      projectURLErrorMsg = t("messages.invalidURL");
    else projectURLErrorMsg = "";
  }
);
watch(
  () => formName,
  () => {
    if (!/^[a-zA-Z0-9_.-]*$/.test(formName) && formName.length > 0)
      formNameErrorMsg = t("messages.invalidName");
    else formNameErrorMsg = "";
  }
);
</script>

<template>
  <PageLayout>
    <n-card size="small" :title="'Local Repos'">
      <template #header-extra>
        <PopModal
          title="Add local repo"
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
              @click="openLocalRepo"
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
              style="display: flex; justify-content: space-between; gap: 0.5rem"
            >
              <span
                style="
                  white-space: nowrap;
                  width: 30%;
                  height: 34px;
                  display: flex;
                  align-items: center;
                "
                >Repo name</span
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
                :placeholder="'Repo name'"
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
        <template
          v-for="repo, repoName in dashboard.repoMap as z.infer<typeof RepoMap> || {}"
        >
          <RepoCard
            v-if="repo.Local"
            :key="repoName"
            :repo="repo"
            class="listing-card"
          >
          </RepoCard>
        </template>
      </div>
      <div
        class="base-alert"
        v-if="
          dashboard.repoMap &&
          !Object.values(dashboard.repoMap).filter((el) => el.Local).length
        "
      >
        No repos to show, create new local repo
      </div>
    </n-card>
    <n-card size="small" :title="'Remote Repos'" style="margin-top: 1rem">
      <template #header-extra>
        <n-button
          strong
          secondary
          type="primary"
          size="tiny"
          @click="
            () => {
              openAddRemoteModal = true;
              projectURLErrorMsg = '';
              formName = '';
              projectURL = '';
            }
          "
        >
          {{ t("objects.remoteRepo") }}
          <template #icon>
            <n-icon><mdi :path="mdiPlus" /></n-icon>
          </template>
        </n-button>
      </template>
      <div
        class="base-alert"
        v-if="
          dashboard.repoMap &&
          !Object.values(dashboard.repoMap).filter((el) => !el.Local).length
        "
      >
        No repos to show, create new remote repo
      </div>
      <div class="repo-grid" v-else>
        <template
          v-for="repo, repoName in dashboard.repoMap as z.infer<typeof RepoMap> || {}"
        >
          <RepoCard
            v-if="!repo.Local"
            :key="repoName"
            :repo="repo"
            class="listing-card"
          >
          </RepoCard>
        </template>
      </div>
    </n-card>
    <Modal
      v-model:show="openAddRemoteModal"
      title="Add remote repos"
      style="width: 1000px"
      :showFooter="true"
      :loading="modalLoader"
      @positive-click="createRemote"
      @negative-click="openAddRemoteModal = false"
      :touched="projectURL.length > 0"
    >
      <p style="color: red">{{ errorProjects }}</p>
      <div class="form">
        <p class="form-label">URL</p>
        <!-- <n-input
          v-model:value="projectURL"
          ref="lol"
          placeholder="np. https://google.com"
        /> -->
        <Input
          style="width: 70%"
          :placeholder="'np. https://google.com'"
          :errorMessage="projectURLErrorMsg"
          :focus="true"
          :removeWhiteSpace="true"
          @keyup.enter="createRemote"
          @update:value="
                  (v: string) => {
                    projectURL = v;
                  }
                "
        />
      </div>
      <div class="form" v-if="validProjectURL">
        <p class="form-label">Nazwa</p>
        <Input
          style="width: 70%"
          :placeholder="t('fields.name')"
          :errorMessage="formNameErrorMsg"
          :valueFromParent="formName"
          :focus="false"
          :removeWhiteSpace="true"
          :replaceInvalidCharacter="/[^a-zA-Z0-9_.-]/gi"
          @keyup.enter="createRemote"
          @update:value="
                  (v: string) => {
                    formName = v;
                  }
                "
        />
        <!-- <n-input
          class="form-input"
          v-model:value="formName"
          type="text"
          placeholder="Nazwa"
        /> -->
      </div>
      <!-- <div class="form">
        <p class="form-label">Branch</p>
        <n-input
          class="form-input"
          v-model:value="formBranch"
          type="text"
          placeholder="Branch"
        />
      </div> -->
      <!-- <div class="form">
        <p class="form-label">Login</p>
        <n-input
          class="form-input"
          v-model:value="formLogin"
          type="text"
          placeholder="Login"
        />
      </div>
      <div class="form">
        <p class="form-label">Hasło</p>
        <n-input
          class="form-input"
          v-model:value="formPassword"
          type="password"
          placeholder="Hasło"
        />
      </div> -->
    </Modal>
    <Modal
      v-model:show="openAddRemoteValidationModal"
      title="Validation Login "
      style="width: 600px"
      :showFooter="true"
      :loading="modalValidationLoader"
      @positive-click="createRemoteWithValidation"
      @negative-click="openAddRemoteValidationModal = false"
    >
      <p style="color: red">{{ errorProjects }}</p>

      <div class="form">
        <p class="form-label">Login</p>
        <n-input
          class="form-input"
          v-model:value="formLogin"
          type="text"
          placeholder="Login"
        />
      </div>
      <div class="form">
        <p class="form-label">Hasło</p>
        <n-input
          class="form-input"
          v-model:value="formPassword"
          type="password"
          placeholder="Hasło"
        />
      </div>
    </Modal>
  </PageLayout>
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
