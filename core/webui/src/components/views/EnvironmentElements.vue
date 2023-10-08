<script setup lang="ts">
import { Ref } from "vue";
import "splitpanes/dist/splitpanes.css";
import { Splitpanes, Pane } from "splitpanes";
import { useEnv } from "@/store/envStore";
import { useRoute } from "vue-router";
import { useUserSettings } from "@/store/userSettings";
import { z } from "zod";
import { ElementMapRespExtended } from "@/zodios/schemas/elements";
import chroma from "chroma-js";
import { useUserStore } from "@/store/userStore";

const userStore = useUserStore();
const route = useRoute();
const userSettings = useUserSettings();
// const env = useEnv(computed(() => z.string().parse(route.params.id)));
const env = useEnv(computed(() => route.params.id as string));

const { t } = useI18n();
let showStoppedEnvModal = $ref(false);
let showExportModal = $ref(false);
let exportToTOMLtext = $ref("");

let logsSize = $(useLocalStorage("logsSize", 50));
let showModal = $ref(false);
let touched = $ref(false);
let openAddFromScratchModal = $ref(false);
let fromScratchTouched = $ref(false);
let currentElement = $ref<z.infer<typeof ElementMapRespExtended>>(
  {} as z.infer<typeof ElementMapRespExtended>
);
let editShow = $ref(false);
let elementToEdit = $ref("");

const showAdding = () => {
  showModal = !showModal;
  touched = false;
};
const isTouched = () => {
  touched = true;
};
const openWeb = (url: string) => {
  window.open(url, "_blank");
};

const creatingDone = (data: string) => {
  showModal = false;
  fromScratchTouched = false;
  if (data) elementToEdit = data;
};
let showFilters = $(inject("showFilters") as Ref<boolean>);
let showAdminPanel = $(inject("showAdminPanel") as Ref<boolean>);
const panelsHeight = $computed(() =>
  showFilters || showAdminPanel
    ? "calc(100vh - (var(--navbar-h) * 2))"
    : "calc(100vh - (var(--navbar-h))) "
);

// card bg color
let backgroundColor = $ref("");
let renderKey = $ref(0);
const getCardBGColor = () => {
  backgroundColor = chroma("#101014")
    .brighten(0.5)
    .alpha(userSettings.opacity / 100)
    .hex();
};
const fromScratchCreated = () => {
  openAddFromScratchModal = false;
  env.value.refetch();
  // emit("creatingDone");
};
const fromScratchEdited = () => {
  editShow = false;
  env.value.refetch();
}
watch(
  () => userSettings.opacity,
  () => {
    getCardBGColor();
    renderKey += 1;
  }
);
watch(
  () => env.value.EnvElements,
  () => {
    if (
      elementToEdit &&
      env.value.EnvElements &&
      Object.keys(env.value.EnvElements).includes(elementToEdit)
    ) {
      currentElement = env.value.EnvElements[elementToEdit];
      editShow = true;
      elementToEdit = "";
    }
  }
);
onBeforeMount(() => {
  getCardBGColor();
});

const exportToTOML = () => {
  api
    .get("/env-export-toml", {
      queries: {
        env: route.params.id as string,
      },
    })
    .then((res) => {

      exportToTOMLtext = atob(res);
      showExportModal = true;

    });
};

</script>
<template>
  <div>
  <EnvTab />
  <PageLayout>
    <div v-if="userStore.havePermission('Env_View')">
      <div class="home-page">
        <div class="items-section">
          <Splitpanes horizontal @resize="logsSize = $event[1].size">
              <Pane :size="100 - logsSize" style="padding-bottom: 0.5rem">
                <div class="n-card" style="overflow: hidden; transition: opacity 0.2s" :style="`background-color: ${backgroundColor} !important; opacity: ${logsSize > 95 ? 0 : 1
                  }`" :key="renderKey">
                  <PanelHeader :title="t('objects.element', 2)">
                    <template #trailing>
                      <div style="display: flex; align-items: center">
                        <Cron style="margin-right: 0.5em"
                          :custom="Object.values(env.EnvElements || {}).find((el: any) => (el.Info.Type === 'pod' || el.Info.Type === 'domain') && el.Info.Unschedulable === true) ? true : false"
                          :inactive="Object.values(env.EnvElements || {}).find((el: any) => (el.Info.Type === 'pod' || el.Info.Type === 'domain') && el.Info.Unschedulable === false) ? false : true"
                          :manage="
                            userStore.havePermission('Env_ManageSchedule')
                          " />

                        <n-button strong secondary type="primary" size="tiny" @click="exportToTOML()"
                          style="margin-right: 0.5em"
                          :disabled="!Object.values(env.EnvElements || {}).find((el: any) => el.Info.Type === 'pod' || el.Info.Type === 'domain')">
                          <template #icon>
                            <n-icon>
                              <Mdi :path="mdiApplicationExport" />
                            </n-icon>
                          </template>
                          Export
                        </n-button>

                        <n-button
                          strong
                          secondary
                          type="primary"
                          size="tiny"
                          @click="showStoppedEnvModal = true"
                          style="margin-right: 0.5em"
                          :disabled="!Object.values(env.EnvElements || {}).find((el: any) => el.Info.Type === 'pod' || el.Info.Type === 'domain')"
                        >
                          <template #icon>
                            <n-icon>
                              <Mdi :path="mdiAutorenew" />
                            </n-icon>
                          </template>
                          {{ t("fields.toggleElements") }}
                        </n-button>

                        <n-button
                          strong
                          secondary
                          type="primary"
                          size="tiny"
                          @click="showAdding"
                          style="margin-right: 0.5em" 
                        >
                          <template #icon>
                            <n-icon>
                              <mdi :path="mdiPlus" />
                            </n-icon>
                          </template>
                          {{ t("objects.elementFromGit") }}
                        </n-button>

                        <n-button 
                          strong
                          secondary
                          type="primary"
                          size="tiny"
                          @click="openAddFromScratchModal = true"
                        >
                          {{ t("objects.visualCreator") }}
                        </n-button>
                      </div>
                    </template>
                  </PanelHeader>
                  <div
                    v-if="Object.keys(env.URLs).length > 0"
                    style="display: flex; gap: 5px;"
                  >
                    <n-button
                      v-for="key in Object.keys(env.URLs)"
                      :key="key"
                      secondary
                      type="primary"
                      size="tiny"
                      style="margin-right: 0.5em"
                      @click="openWeb(env.URLs[key])"
                    >
                      <n-icon style="cursor: pointer">
                        <mdi :path="mdiDomain" width="15" />
                      </n-icon>

                      <span style="margin-left: 5px">
                      {{ key.slice(0, 25) }}{{ key.length > 25 ? "..." : "" }}
                      </span>
                    </n-button>
                  </div>
                  <ElementsContainer :logsSize="logsSize" />
                </div>
              </Pane>
              <Pane :size="logsSize" style="padding-top: 0.5rem; transition: opacity 0.2s"
                :style="{ opacity: logsSize < 7 ? 0 : 1 }">
                <Logs :fullscreen="logsSize > 95" :backgroundColor="backgroundColor"
                  @update:fullscreen="logsSize = $event ? 100 : 50" @minimize="logsSize = 0" v-if="
                    userStore.havePermission('Env_ViewLogsBuild') ||
                    userStore.havePermission('Env_ViewLogsEvents') ||
                    userStore.havePermission('Env_ViewLogsRuntime')
                  " />
              </Pane>
            </Splitpanes>
          </div>
        </div>
        <Teleport to="#layout-content" v-if="
          userStore.havePermission('Env_ViewLogsBuild') ||
          userStore.havePermission('Env_ViewLogsEvents') ||
          userStore.havePermission('Env_ViewLogsRuntime')
        ">
          <div class="show-logs" :class="{ show: logsSize < 5 }">
            <button class="show-logs-btn" @click="logsSize = 50" style="">
              {{ $t("messages.showLogs") }}
              <n-icon :size="18">
                <Mdi :path="mdiArrowUpDropCircle" />
              </n-icon>
            </button>
          </div>
        </Teleport>
      </div>
      <ErrorDialog v-else :msg="$t('messages.permissionDenied')" />
    </PageLayout>

    <Modal v-model:show="showModal" :title="`${t('actions.add')} ${t('fields.new').toLowerCase()} ${t(
      'objects.element'
    ).toLowerCase()}`" :show-icon="false" :touched="touched" style="width: 80%; min-height: 45em; height: 1px">
      <ElementAdd @isTouched="isTouched" @creatingDone="creatingDone" />
    </Modal>
    <Modal v-model:show="openAddFromScratchModal" :title="t('fields.elementCreator')"
      :style="fromScratchTouched ? 'width: 100em' : 'width: 60em'" :touched="fromScratchTouched">
      <ElementAddFromScratch @fromScratchCreated="fromScratchCreated" @isTouched="fromScratchTouched = true"
        @detouch="fromScratchTouched = false" />
    </Modal>
    <Modal v-model:show="editShow" :title="
      t('actions.edit') +
      ': ' +
      currentElement?.Info?.Name +
      ' (' +
      currentElement?.Info?.Type +
      ')'
    " :touched="false" style="width: 80rem" :show-footer="false">
      <ElementAddFromScratch @fromScratchCreated="fromScratchEdited" @isTouched="fromScratchTouched = true"
        @detouch="fromScratchTouched = false" :element="currentElement" />
    </Modal>
    <StartStopElements :show="showStoppedEnvModal" @hide="showStoppedEnvModal = false" />

    <Modal v-model:show="showExportModal" title="Export" :style="fromScratchTouched ? 'width: 100em' : 'width: 60em'"
      :touched="fromScratchTouched">
      <div style="height: 30rem">
        <Monaco :value="exportToTOMLtext" lang="toml" read-only />
      </div>
    </Modal>

  </div>
</template>
<style scoped>
h1 {
  padding: 0;
  margin: 0;
}

.n-card {
  height: 100%;
  background: var(--cardColor);
  padding: 15px;
  border-radius: 5px;
}

.home-page {
  display: grid;
  grid-template-columns: 1fr;
  gap: 1rem;
  overflow-y: hidden;
  position: relative;
}

.items-section {
  width: 100%;
  min-width: 0;
}

.show-logs {
  position: absolute;
  display: grid;
  place-items: center;
  bottom: 0;
  left: 0;
  right: 0;
  transform: translateY(4rem);
  transition: all 0.2s;
  z-index: 100;
}

.show-logs.show {
  transform: translateY(0);
}

.show-logs-btn {
  background-color: var(--primaryColor);
  color: black;
  padding: 0.5rem 1rem;
  border-top-right-radius: 10px;
  border-top-left-radius: 10px;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
}

.n-layout {
  padding: 0.5rem;
}

.items-grid {
  --colums: 4;
  display: grid;
  grid-template-columns: repeat(var(--colums), 1fr);
  gap: 1rem;
}

@media screen and (max-width: 1000px) {
  .items-grid {
    --colums: 2;
  }
}

@media screen and (max-width: 600px) {
  .items-grid {
    --colums: 1;
  }
}

.header__extra {
  display: flex;
  gap: 1rem;
}

.element-menu {
  overflow: hidden;
}

.logs-panel,
.elements-panel {
  margin-top: 30px;
}

.splitpanes {
  max-height: v-bind(panelsHeight);
  height: v-bind(panelsHeight);
  transition: height 0.25s ease-in-out;
}
</style>

<style lang="scss">
.splitpanes__splitter {
  position: relative;
  cursor: move !important;
}

.splitpanes__splitter:before {
  content: "•••";
  position: absolute;
  display: grid;
  place-content: center;
  left: 0;
  right: 0;
  top: 0;
  transition: background-color 0.4s, color 0.4s;
  background-color: rgba(121, 121, 121, 0);
  border-radius: 10px;
  z-index: 1;
}

.splitpanes__splitter:hover:before {
  opacity: 1;
  background-color: rgba(121, 121, 121, 0.603);
  color: var(--primaryColor);
}

.splitpanes--vertical>.splitpanes__splitter:before {
  left: -0.5rem;
  right: -0.5rem;
  height: 100%;
}

.splitpanes--horizontal>.splitpanes__splitter:before {
  top: -0.5rem;
  bottom: -0.5rem;
  width: 100%;
}

.home-page,
.splitpanes__pane {
  // box-shadow: 0px 0px 14px 0px #00000098;
  overflow: visible !important;
}

.splitpanes--horizontal .splitpanes__pane {
  transition: none;
}
</style>
