<script lang="ts" setup>
import { useRoute } from "vue-router";
import { useMessage } from "naive-ui";

const message = useMessage();
const { t } = useI18n();

const route = useRoute();

const props = defineProps<{ show: boolean }>();
const emit = defineEmits<{
  (type: "hide"): void;
}>();
let envMap = $ref<any>();
let globalRun = $ref(true);
let localShow = $ref(false);

const getData = () => {
  api
    .get("/env-element-map", {
      queries: {
        env: route.params.id as string,
      },
    })
    .then((res) => {
      let tmp: Record<string, boolean> = {};
      Object.values(res).forEach((el) => {
        if (el.Info.Type !== "config")
          tmp[el.Info.Name as string] = !el.Info.Stopped;
        if (el.Info.Stopped) globalRun = false;
      });
      envMap = tmp;
    });
};

watch(
  () => props.show,
  () => {
    localShow = props.show;
  }
);
watch($$(localShow), () => {
  if (localShow) getData();
  if (!localShow) emit("hide");
});
const changeGlobal = () => {
  for (let el in envMap) {
    envMap[el] = globalRun;
  }
};
const updateGlobal = (state: boolean) => {
  if (!state) {
    globalRun = state;
  } else if (!Object.values(envMap).includes(false)) {
    globalRun = state;
  }
};
const updateRunningElements = () => {
  if (globalRun) {
    api
      .post("/env-element-run-control", {
        EnvID: route.params.id as string,
        Control: 1,
      })
      .then((res) => {
        if (res === "ok") message.success(t("messages.elementsUpdated"));
      });
  } else {
    let errorOccured = false;
    for (let el in envMap) {
      api
        .post("/env-element-run-control", {
          EnvID: route.params.id as string,
          Element: el,
          Control: envMap[el] ? 1 : 2,
        })
        .then((res) => {
          // if (res === "ok") message.success(t("messages.elementStarted"));
          if (res !== "ok") errorOccured = true;
        });
    }
    if (errorOccured) {
      message.error(t("messages.errorOccured"));
    } else {
      message.success(t("messages.elementsUpdated"));
    }
  }
  localShow = false;
};
</script>
<template>
  <Modal
    v-model:show="localShow"
    :title="t('fields.toggleElements')"
    :show-icon="false"
    :showFooter="true"
    @positive-click="updateRunningElements"
    @negative-click="() => {}"
    style="width: 20rem"
  >
    <div v-for="(_, key) in envMap" :key="key">
      <div
        style="
          display: flex;
          justify-content: space-between;
          width: 100%;
          margin-bottom: 5px;
        "
      >
        {{ key }}
        <n-switch
          v-model:value="envMap[key]"
          :round="false"
          @update:value="
            (e:boolean) => {
              updateGlobal(e);
            }
          "
        >
          <template #checked> Running </template>
          <template #unchecked> Stopped </template>
        </n-switch>
      </div>
    </div>
    <div
      style="
        display: flex;
        justify-content: space-between;
        width: 100%;
        margin-bottom: 5px;
        border-top: 1px solid grey;
        padding-top: 5px;
      "
    >
      {{ t("fields.allElements") }}
      <n-switch
        v-model:value="globalRun"
        :round="false"
        @update:value="changeGlobal"
      >
        <template #checked> Running </template>
        <template #unchecked> Stopped </template>
      </n-switch>
    </div>
  </Modal>
</template>
