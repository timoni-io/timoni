<script setup lang="ts">
import { useI18n } from "vue-i18n";
import { useMessage } from "naive-ui";
import { useRouter } from "vue-router";
const router = useRouter();
const { t } = useI18n();
const message = useMessage();

const props = defineProps<{
  btnSize: "tiny" | "small" | "medium" | "large";
  noPush: boolean;
  disabled: boolean;
}>();
const emit = defineEmits(["repoAdded"]);
let redirectConfirmationModal = $ref(false);
let createdRepoName = $ref("");
let openAddRemoteModal = $ref(false);
let openAddRemoteValidationModal = $ref(false);
let modalLoader = $ref(false);
let modalValidationLoader = $ref(false);
let projectURL = $ref("");
let errorProjects = $ref("");
let formName = $ref("");
let formLogin = $ref("");
let formPassword = $ref("");

const validProjectURL = $computed(() => {
  return /[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=-]*)/i.test(
    projectURL
  );
});
// validation
let projectURLErrorMsg = $ref("");
let formNameErrorMsg = $ref("");
watch(
  () => projectURL,
  () => {
    if (
      (!validProjectURL ||
        !/^[a-zA-Z0-9-/]*\.?[a-zA-Z0-9/]*$/.test(
          projectURL.split("/")[projectURL.split("/").length - 1]
        )) &&
      projectURL.length > 0
    ) {
      projectURLErrorMsg = t("messages.invalidURL");
    } else projectURLErrorMsg = "";
  }
);
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
        formName = splittedList[splittedList.length - 1]
          .replace(".git", "")
          .replaceAll(".", "");
      }
    }
  }
);

const createRemote = () => {
  modalLoader = true;
  if (!projectURL) {
    message.error(t("messages.urlRequired"));
    modalLoader = false;
    return;
  }
  if (projectURLErrorMsg) {
    message.error(projectURLErrorMsg);
    modalLoader = false;
    return;
  }

  api
    .post("/git-repo-create-remote", {
      URL: projectURL.toLowerCase(),
      Name: formName.toLowerCase(),
    })
    .then((res) => {
      modalLoader = false;
      if (typeof res === "string") {
        if (res !== "login and password are required") {
          message.error(res);
        }
        if (res === "login and password are required") {
          openAddRemoteValidationModal = true;
          return;
        }
      } else {
        if (res.IsTaken) {
          createdRepoName = res.GitRepoName;
          redirectConfirmationModal = true;
          return;
        }
        if (
          !/^[:a-z0-9_.-/-]*$/.test(projectURL) ||
          !/^[a-z0-9_.-/-]*$/.test(formName)
        ) {
          message.success(
            "Successfully added remote repository with conversion to lower case"
          );
        } else message.success("Successfully added remote repository");
        if (res.GitRepoName && !props.noPush) {
          router.push("/code/" + res.GitRepoName);
        }
        openAddRemoteModal = false;
        emit("repoAdded");
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
  <div>
    <n-button
      :disabled="props.disabled"
      strong
      secondary
      type="primary"
      :size="btnSize"
      @click="
        () => {
          openAddRemoteModal = true;
          projectURL = '';
          formName = '';
          formLogin = '';
          formPassword = '';
        }
      "
    >
      {{ t("objects.remoteRepo") }}
      <template #icon>
        <n-icon><mdi :path="mdiPlus" /></n-icon>
      </template>
    </n-button>
    <Modal
      v-model:show="openAddRemoteModal"
      :title="t('actions.add') + ' ' + t('objects.remoteRepo').toLowerCase()"
      style="width: 1000px"
      :showFooter="true"
      :loading="modalLoader"
      @positive-click="createRemote"
      @negative-click="openAddRemoteModal = false"
      :touched="projectURL.length > 0"
    >
      <p style="color: red">{{ errorProjects }}</p>
      <div class="form" style="margin-bottom: 1rem">
        <p class="form-label">URL</p>
        <Input
          style="width: 70%"
          :placeholder="'https://google.com'"
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
        <p class="form-label">{{ t("fields.name") }}</p>
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
      </div>
    </Modal>
    <Modal
      v-model:show="openAddRemoteValidationModal"
      title="Validation Login "
      style="width: 600px"
      :showFooter="true"
      :loading="modalValidationLoader"
      @positive-click="createRemoteWithValidation"
      @negative-click="openAddRemoteModal = false"
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
        <p class="form-label">{{ t("scratch.password") }}</p>
        <n-input
          class="form-input"
          v-model:value="formPassword"
          type="password"
          :placeholder="t('scratch.password')"
        />
      </div>
    </Modal>
    <Modal
      v-model:show="redirectConfirmationModal"
      title=""
      style="width: 600px"
      :showFooter="true"
      @positive-click="router.push('/code/' + createdRepoName)"
      @negative-click="redirectConfirmationModal = false"
    >
      <p>{{ t("messages.nameExistWillRedirect") }}</p>
    </Modal>
  </div>
</template>
