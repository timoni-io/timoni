<script setup lang="ts">
import { ComputedRef } from "vue";
import { Column } from "@/components/utils/DataTable.vue";
import { useI18n } from "vue-i18n";
import { useMessage } from "naive-ui";
import { ElementResp } from "@/zodios/schemas/elements";
import { useRoute } from "vue-router";
import { z } from "zod";
import { useUserStore } from "@/store/userStore";

const userStore = useUserStore();

const { t } = useI18n();
const emit = defineEmits(["creatingDone", "isTouched"]);
const route = useRoute();
const message = useMessage();
const filterRef = $ref(null as HTMLInputElement | null);

const columns: ComputedRef<Column[]> = computed(() => [
  {
    title: t("fields.name"),
    template: "select",
    width: "25%",
  },
  {
    title: t("fields.source"),
    template: "source",
    key: "Source.RepoName",
    width: "15%",
    sorter: (row1, row2) =>
      // @ts-ignore
      (("" + row1.Source.RepoName) as string).localeCompare(
        // @ts-ignore
        row2.Source.RepoName as string
      ),
  },
  {
    title: t("fields.branch"),
    key: "Source.BranchName",
    width: "10%",
    template: "branch",
    sorter: (row1, row2) =>
      // @ts-ignore
      ("" + row1.Source.BranchName).localeCompare(
        // @ts-ignore
        row2.Source.BranchName as string
      ),
  },
  {
    title: t("fields.type"),
    key: "Type",
    width: "10%",
    sorter: (row1, row2) => ("" + row1.Type).localeCompare(row2.Type as string),
  },
  {
    title: t("fields.filePath"),
    width: "15%",
    template: "filePath",
  },
  {
    title: t("fields.description"),
    width: "25%",
    template: "description",
  },
]);

const elementTypes = $ref([
  {
    label: t("fields.any").toLowerCase(),
    value: "",
  },
  {
    label: t("scratch.types.domain"),
    value: "domain",
  },
  {
    label: t("scratch.types.pod"),
    value: "pod",
  },
  {
    label: t("scratch.types.config"),
    value: "config",
  },
  {
    label: "MongoDB",
    value: "mongodb",
  },
  {
    label: "Elasticsearch",
    value: "elasticsearch",
  },
]);

let elements = $ref<z.infer<typeof ElementResp>[]>();
let selectedElement = $ref<z.infer<typeof ElementResp>>(
  {} as z.infer<typeof ElementResp>
);
let selectedName = $ref("");
let selectedAddStopped = $ref(false);
let selectedEditAfterAdd = $ref(false);
let selectModal = $ref(false);
let elementsLoader = $ref(true);
let filterRepoBranchValue = $ref("");
let filterValue = $ref("");
let filterRepoValue = $ref("");
let filterBuffer = $ref<Array<number>>([]);
let filterType = $ref("");

let projectURL = ref("");
let formName = ref("");
let formLogin = ref("");
let formPassword = ref("");
let errorProjects = ref("");

let openAddRemoteModal = $ref(false);

watch($$(filterValue), () => {
  filterBuffer.push(1);
  setTimeout(() => {
    filterBuffer.pop();
    if (!filterBuffer.length) {
      fetchElementsList(filterType as string);
    }
  }, 200);
});

const filteredElements = $computed(() => {
  if (filterValue) {
    emit("isTouched");
  }
  return elements
    ?.filter((el) => {
      return (
        el.Source.RepoName.includes(filterRepoValue) &&
        el.Source.BranchName.includes(filterRepoBranchValue)
      );
    })
    .sort((a, b) => ("" + a.Name).localeCompare(b.Name));
});

const filteredElementsLength = computed(() => {
  return filteredElements?.length;
});

const testElementName = computed(() => {
  return /^[a-z0-9\-]+$/.test(selectedName);
});

const createRemote = () => {
  api
    .post("/git-repo-create-remote", {
      URL: projectURL.value,
      Name: formName.value,
      Login: formLogin.value,
      Password: formPassword.value,
    })
    .then((res) => {
      if (typeof res === "string") {
        message.error(res);
        errorProjects.value = res;
      } else {
        message.success(t("messages.remoteRepoCreated"));
        openAddRemoteModal = false;
      }
    });
};

const selectElement = (element: z.infer<typeof ElementResp>) => {
  selectedAddStopped = false;
  selectedEditAfterAdd = false;
  selectedName = element.Name;
  selectedElement = element;
  selectModal = true;
};

const createElement = () => {
  api
    .post("/env-element-create-from-git", {
      EnvID: route.params.id as string,
      GitID: selectedElement?.ID as string,
      Name: selectedName,
      DontStart: selectedAddStopped,
    })
    .then((res) => {
      if (res === "ok") {
        selectModal = false;
        emit("creatingDone", selectedEditAfterAdd ? selectedName : "");
      } else message.error(t("messages." + res));
    });
};

const fetchElementsList = (filterType: string) => {
  elementsLoader = true;
  api
    .get("/git-elements-list", {
      queries: {
        type: filterType,
        filter: filterValue as string,
      },
    })
    .then((res) => {
      elements = res;
      elementsLoader = false;
    });
};

onMounted(() => {
  fetchElementsList("");

  setTimeout(() => {
    filterRef?.focus();
  }, 100);
});

watch($$(filterType), () => {
  fetchElementsList(filterType as string);
});
</script>

<template>
  <div style="display: flex; flex-direction: column">
    <div
      style="
        display: flex;
        justify-content: space-between;
        align-content: center;
        margin-bottom: 2rem;
      "
    >
      <div style="display: flex; align-items: center; width: 25rem">
        <p
          style="
            display: inline-block;
            text-align: left;
            margin: 0 0.6em;
            white-space: nowrap;
          "
        >
          {{ t("fields.name") }}
        </p>
        <n-input
          v-model:value="filterValue"
          type="text"
          :placeholder="t('fields.name')"
          style="width: 68%"
          ref="filterRef"
        />
      </div>
      <div style="display: flex; align-items: center; width: 25rem">
        <p
          style="
            display: inline-block;
            text-align: right;
            margin: 0 0.6em;
            white-space: nowrap;
          "
        >
          {{ t("fields.source") }}
        </p>
        <n-input
          v-model:value="filterRepoValue"
          type="text"
          :placeholder="t('fields.source')"
          style="width: 68%"
        />
      </div>
      <div
        style="
          display: flex;
          align-items: center;
          width: 25rem;
          white-space: nowrap;
        "
      >
        <p style="display: inline-block; text-align: right; margin: 0 0.6em">
          {{ t("fields.branch") }}
        </p>
        <n-input
          v-model:value="filterRepoBranchValue"
          type="text"
          :placeholder="t('fields.branch')"
          style="width: 68%"
        />
      </div>
      <div style="display: flex; align-items: center; width: 25rem">
        <p style="display: inline-block; text-align: right; margin: 0 0.6em">
          {{ t("fields.type") }}
        </p>
        <n-select
          v-model:value="filterType"
          :options="elementTypes"
          style="width: 68%"
        />
      </div>
      <div style="display: flex; justify-content: space-between">
        <AddRemoteRepo
          :btnSize="'medium'"
          :noPush="true"
          :disabled="!userStore.havePermission('Glob_CreateAndDeleteGitRepos')"
          style="padding-left: 0.5rem"
          @repoAdded="fetchElementsList(filterType as string)"
        />
      </div>
    </div>
    <div style="height: 80%">
      <div
        v-if="elementsLoader"
        style="
          display: flex;
          justify-content: center;
          align-content: center;
          min-height: 33em;
          height: 1px;
        "
      >
        <n-spin
          :size="60"
          stroke="#1ba3fd"
          :stroke-width="10"
          style="height: 100%"
        />
      </div>
      <data-table
        v-else
        :bordered="false"
        :single-line="false"
        :columns="columns"
        :data="filteredElements"
        :min-height="filteredElements?.length ? '23em' : '14.5rem'"
        :max-height="filteredElements?.length ? '30em' : '14.5rem'"
        :pagination="{
          pageSize: 10,
          prefix() {
            return `${t('fields.results')}: ${filteredElementsLength}`;
          },
        }"
      >
        <template #select="element">
          <n-button
            strong
            secondary
            size="medium"
            style="width: 100%; justify-content: left"
            @click="selectElement(element)"
            >+ {{ element.Name }}
          </n-button>
        </template>
        <template #filePath="element">
          {{
            element.Source.FilePath.split("/")
              .slice(0, -1)
              .join("/")
              .substring(1)
          }}
        </template>
        <template #source="element">
          <!-- <div
            style="
              display: flex;
              justify-content: flex-start;
              align-items: center;
              gap: 0.5rem;
            "
          >
            <n-tooltip trigger="hover">
              <template #trigger>
                <n-button size="tiny" secondary type="primary" circle>
                  <n-icon :size="20">
                    <Mdi :path="mdiInformation" />
                  </n-icon>
                </n-button>
              </template>
              <div>
                <p>
                  <span style="font-weight: bold">Repo: </span>
                  {{ element.Source.RepoName }}
                </p>
                <p>
                  <span style="font-weight: bold"
                    >{{ t("fields.branch") }}:
                  </span>
                  {{ element.Source.BranchName }}
                </p>
                <p>
                  <span style="font-weight: bold"
                    >{{ t("objects.commit") }}:
                  </span>
                  {{ element.Source.CommitHash.substring(0, 8) }}
                </p>
                <p>
                  <span style="font-weight: bold"
                    >{{ t("scratch.path") }}:
                  </span>
                  {{ element.Source.FilePath }}
                </p>
              </div>
            </n-tooltip> -->
          {{ element.Source.RepoName }}
          <!-- </div> -->
        </template>
        <template #branch="element">
          <div
            style="
              display: flex;
              justify-content: flex-start;
              align-items: center;
              gap: 0.5rem;
            "
          >
            <n-tooltip trigger="hover">
              <template #trigger>
                <n-button size="tiny" secondary type="primary" circle>
                  <n-icon :size="20">
                    <Mdi :path="mdiInformation" />
                  </n-icon>
                </n-button>
              </template>
              <div>
                <p>
                  <span style="font-weight: bold">Repo: </span>
                  {{ element.Source.RepoName }}
                </p>
                <p>
                  <span style="font-weight: bold"
                    >{{ t("fields.branch") }}:
                  </span>
                  {{ element.Source.BranchName }}
                </p>
                <p>
                  <span style="font-weight: bold"
                    >{{ t("objects.commit") }}:
                  </span>
                  {{ element.Source.CommitHash.substring(0, 8) }}
                </p>
                <p>
                  <span style="font-weight: bold"
                    >{{ t("scratch.path") }}:
                  </span>
                  {{ element.Source.FilePath }}
                </p>
              </div>
            </n-tooltip>
            {{ element.Source.BranchName }}
          </div>
        </template>
        <template #description="{ Description }">
          <n-tooltip
            trigger="hover"
            v-if="Description.length > 150"
            style="max-width: 600px"
          >
            <template #trigger>
              {{ Description.substring(0, 150) }}
            </template>
            {{ Description }}
          </n-tooltip>
          <div v-else>
            {{ Description }}
          </div>
        </template>
      </data-table>
    </div>
  </div>
  <Modal
    v-model:show="openAddRemoteModal"
    :title="t('actions.add') + ' ' + t('objects.remoteRepo').toLowerCase()"
    style="width: 1200px"
    :showFooter="true"
    @positive-click="createRemote"
    @negative-click="openAddRemoteModal = false"
    :touched="!!projectURL || !!formName || !!formLogin || !!formPassword"
  >
    <p style="color: red">{{ errorProjects }}</p>
    <div class="form">
      <p class="form-label">URL</p>
      <n-input
        v-model:value="projectURL"
        autofocus
        ref="lol"
        placeholder="np. https://google.com"
      />
    </div>
    <div class="form">
      <p class="form-label">{{ t("fields.name") }}</p>
      <n-input
        class="form-input"
        v-model:value="formName"
        type="text"
        :placeholder="t('fields.name')"
      />
    </div>
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
      <p class="form-label">{{ t("fields.password") }}</p>
      <n-input
        class="form-input"
        v-model:value="formPassword"
        type="password"
        :placeholder="t('fields.password')"
      />
    </div>
  </Modal>
  <Modal
    v-model:show="selectModal"
    :title="
      $t('actions.add') +
      ' ' +
      selectedElement?.Name +
      ' (' +
      selectedElement?.Source?.RepoName +
      ' / ' +
      selectedElement?.Source?.BranchName +
      ' / ' +
      selectedElement?.Source?.CommitHash.substring(0, 8) +
      ')'
    "
    style="width: 40rem"
    :showFooter="true"
    :disabled="!testElementName"
    @positive-click="createElement"
  >
    <div class="element-adding-modal-row">
      <p style="width: 50%">{{ t("fields.elementName") }}</p>
      <n-input
        v-model:value="selectedName"
        type="text"
        :placeholder="t('fields.elementName')"
        style="width: 50%"
        :status="testElementName ? undefined : 'error'"
        @keyup.enter="createElement"
      />
    </div>
    <div
      v-if="['pod', 'elasticsearch', 'mongodb'].includes(selectedElement.Type)"
      class="element-adding-modal-row"
    >
      <span style="width: 70%">{{ t("elements.options.addStopped") }}</span>
      <n-checkbox v-model:checked="selectedAddStopped" />
    </div>
    <div class="element-adding-modal-row">
      <p style="width: 70%">{{ t("elements.options.editAfterAdd") }}</p>
      <n-checkbox v-model:checked="selectedEditAfterAdd" />
    </div>
  </Modal>
</template>

<style scoped>
.project-add {
  margin-left: 5px;
}

.form {
  display: flex;
  gap: 1rem;
  width: 100%;
}

.form-label {
  width: 15%;
  align-self: center;
}

.form-input {
  margin-top: 5px;
}

.n-pagination {
  justify-content: space-between !important;
}

.element-adding-modal-row {
  display: flex;
  align-items: center;
  margin-bottom: 0.5em;
}
</style>
