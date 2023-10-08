<script setup lang="ts">
import { useI18n } from "vue-i18n";
import { Annotation, Path } from "../ElementAddFromScratch.vue";
import { ElementMapRespExtended, DomainType } from "@/zodios/schemas/elements";
import { dynamicInputsHandle } from "@/utils/dynamicInputsHandle";
import { findDuplicates } from "@/utils/findDuplicates";
import { useRoute } from "vue-router";
import { z } from "zod";

const emit = defineEmits(["tipMode", "confirm", "delete"]);
const props = defineProps<{
  element?: z.infer<typeof ElementMapRespExtended>;
  testElementName: boolean;
  allowApplyResp: string;
  selectedTab: string;
}>();

const { t } = useI18n();
const route = useRoute();
let domainName = $ref(
  props.element && props.element.Info && "Domain" in props.element.Info
    ? props.element.Info.Domain
    : ""
);
let annotations = $ref<Annotation[]>([
  {
    key: "",
    value: "",
  },
]);
let paths = $ref<Array<Path>>([]);
let domainTargets = $ref<Array<string>>();
let httpOnly = $ref<boolean>(
  props.element && props.element.Info && "Domain" in props.element.Info
    ? props.element.Info.HttpOnly
    : false
);
let wwwRedirect = $ref<boolean>(
  props.element && props.element.Info && "Domain" in props.element.Info
    ? props.element.Info.WWWredirect
    : false
);
  
const onAnnotationCreate = () => {
  return {
    key: "",
    value: "",
  };
};

const onPathCreate = () => {
  return {
    path: "",
    target: "",
    label: "",
    prefix: ""
  };
};

let advancedPathShow = $ref(false);
let advancedPath = $ref<Path>(onPathCreate());
let advancedPathIndex = $ref<number>(-1);
let advancedTouched = $ref<boolean>(false);
  
const annotationsDuplicates = $computed<Array<string>>(() =>
  findDuplicates<string>(annotations.map((annotation) => annotation.key))
);

const emptyPaths = $computed<boolean>(() => {
  return paths.filter((el) => el.path !== "" && el.target === "").length > 0;
});

const pathsDuplicates = $computed<Array<string>>(() =>
  findDuplicates<string>(paths.map((path) => path.path))
);

const allowApply = $computed<boolean>(
  () =>
    annotationsDuplicates.length === 0 &&
    pathsDuplicates.length === 0 &&
    !props.allowApplyResp &&
    !emptyPaths
);

const applyTooltipMessage = (): string => {
  let sections = [];
  let resp = [];
  if (props.allowApplyResp.includes(t("scratch.removeDuplicates")))
    sections.push(t("scratch.tabs.variables"));
  if (annotationsDuplicates.length)
    sections.push(t("scratch.tabs.annotations"));
  if (pathsDuplicates.length) sections.push(t("scratch.path", 2));
  if (sections.length)
    resp.push(
      t("scratch.removeDuplicates") +
        (t("scratch.removeDuplicates").includes("from") && sections.length > 1
          ? "s: "
          : ": ") +
        sections.join(", ")
    );
  if (emptyPaths) resp.push(t("scratch.emptyPaths"));
  // if (props.allowApplyResp.includes(t("scratch.emptySecret")))
  //   resp.push(t("scratch.emptySecret"));
  return resp.join(" ");
};

watch(
  $$(annotations),
  () => {
    dynamicInputsHandle<Annotation>(
      annotations,
      ["key", "value"],
      [""],
      "annotations-dynamic",
      onAnnotationCreate
    );
  },
  { deep: true }
);

watch(
  $$(paths),
  () => {
    dynamicInputsHandle<Path>(
      paths,
      ["path", "target", "prefix", "label"],
      [""],
      "paths-dynamic",
      onPathCreate
    );
  },
  { deep: true }
);

onBeforeMount(() => {
  const elInfo = props.element?.Info as z.infer<typeof DomainType>;
  if (elInfo) {
    Object.keys(elInfo.Annotations || {}).forEach((annotation) => {
      if (elInfo.Annotations) {
        annotations.push({
          key: annotation,
          value: elInfo.Annotations[annotation],
        } as Annotation);
      }
    });
    Object.keys(elInfo.Paths || {}).forEach((path) => {
      if (elInfo.Paths) {
        paths.push({
          path: path,
          target:
            elInfo.Paths[path].ElementName +
            ":" +
            elInfo.Paths[path].Port,
          label: elInfo.Paths[path].Label,
          prefix: elInfo.Paths[path].Prefix
        } as Path);
      }
    });
  }
  annotations.push(onAnnotationCreate());
  if (paths.length === 0)
    paths.push({
      path: "/",
      target: "",
      label: "",
      prefix: ""
    });
  paths.push(onPathCreate());
});

onMounted(() => {
  api
    .get("/env-domain-targets", {
      queries: {
        env: route.params.id as string,
      },
    })
    .then((res) => {
      domainTargets = res;
    })
    .catch(() => {
      domainTargets = [];
    });
});
</script>

<template>
  <div>
    <div v-if="selectedTab === 'general'" class="element-from-scratch-row">
      <ElementAddFromScratchRowHeader
        :name="t('scratch.domainName')"
        :tip="t('scratch.docs.domainName')"
      />
      <div style="width: 80%; padding-right: 5em">
        <n-input
          v-model:value="domainName"
          type="text"
          :placeholder="t('scratch.domainName')"
        />
      </div>
    </div>
    <div v-if="selectedTab === 'general'" class="element-from-scratch-row">
      <ElementAddFromScratchRowHeader
        :name="t('scratch.httpOnly')"
        :tip="t('scratch.docs.httpOnly')"
      />
      <div style="width: 80%">
        <n-checkbox v-model:checked="httpOnly" />
      </div>
    </div>
    <div v-if="selectedTab === 'general'" class="element-from-scratch-row">
      <ElementAddFromScratchRowHeader
        :name="t('scratch.wwwRedirect')"
        :tip="t('scratch.docs.wwwRedirect')"
      />
      <div style="width: 80%">
        <n-checkbox v-model:checked="wwwRedirect" />
      </div>
    </div>

    <div v-if="selectedTab === 'annotations'">
      <!-- <p class="element-from-scratch-row-header">
        {{ t("scratch.annotation", 2) }}
      </p> -->
      <div
        id="annotations-dynamic"
        style="
          width: 100%;
          padding: 0 5em 0 5em;
          overflow-y: scroll;
          max-height: 27em;
        "
      >
        <n-table class="dynamic-table" :bordered="false" :single-line="true">
          <thead>
            <tr>
              <th style="width: 46%">
                <ElementAddFromScratchTableHeader
                  :name="t('scratch.annotation')"
                  :tip="t('scratch.docs.annotations.self')"
                />
              </th>
              <th style="width: 48%">
                <ElementAddFromScratchTableHeader
                  :name="t('scratch.value')"
                  :tip="t('scratch.docs.annotations.value')"
                />
              </th>
              <th style="width: 6%"></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(annotation, index) in annotations" :key="index">
              <td>
                <n-input
                  v-model:value="annotation.key"
                  :placeholder="t('scratch.annotation')"
                  :status="
                    annotationsDuplicates.includes(annotation.key)
                      ? 'warning'
                      : undefined
                  "
                />
              </td>
              <td>
                <n-input
                  v-model:value="annotation.value"
                  :placeholder="t('scratch.value')"
                />
              </td>
              <td>
                <n-button
                  v-if="
                    annotation.key + annotation.value ||
                    index < annotations.length - 1
                  "
                  type="primary"
                  quaternary
                  circle
                  style="margin-left: 5px"
                  @click="
                    () => {
                      annotations.splice(index, 1);
                    }
                  "
                >
                  <template #icon>
                    <n-icon>
                      <mdi :path="mdiWindowClose" />
                    </n-icon>
                  </template>
                </n-button>
              </td>
            </tr>
          </tbody>
        </n-table>
      </div>
    </div>
    <div v-if="selectedTab === 'paths'">
      <div
        id="paths-dynamic"
        style="
          width: 100%;
          padding: 0 5em 0 5em;
          overflow-y: scroll;
          max-height: 27em;
        "
      >
        <n-table class="dynamic-table" :bordered="false" :single-line="true">
          <thead>
            <tr>
              <th style="width: 29%">
                <ElementAddFromScratchTableHeader
                  :name="t('scratch.path')"
                  :tip="t('scratch.docs.paths.self')"
                />
              </th>
              <th style="width: 31%">
                <ElementAddFromScratchTableHeader
                  :name="t('scratch.target')"
                  :tip="t('scratch.docs.paths.target')"
                />
              </th>
              <th style="width: 34%">
                  <ElementAddFromScratchTableHeader
                    :name="t('scratch.advanced')"
                    :tip="t('scratch.docs.variables.advanced')"
                  />
                </th>
              <th style="width: 6%"></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(path, index) in paths" :key="index">
              <td>
                <n-input
                  v-model:value="path.path"
                  :placeholder="t('scratch.path')"
                  :status="
                    pathsDuplicates.includes(path.path) ? 'warning' : undefined
                  "
                />
              </td>
              <td>
                <n-auto-complete
                  v-model:value="path.target"
                  :input-props="{
                    autocomplete: 'disabled',
                  }"
                  :options="
                    domainTargets?.filter((target) =>
                      target.includes(path.target)
                    )
                  "
                  :placeholder="t('scratch.target')"
                  :get-show="() => true"
                />
              </td>
              <td>
                <n-button
                  dashed
                  style="width: 100%"
                  @click="
                    () => {
                      advancedPath = JSON.parse(JSON.stringify(path));
                      advancedPathIndex = index;
                      advancedPathShow = true;
                    }
                  "
                >
                  {{ t("scratch.advanced") }}
                </n-button>
              </td>
              <td>
                <n-button
                  v-if="
                    path.path + path.target === '/' || index < paths.length - 1
                  "
                  type="primary"
                  quaternary
                  circle
                  style="margin-left: 5px"
                  @click="
                    () => {
                      paths.splice(index, 1);
                    }
                  "
                >
                  <template #icon>
                    <n-icon>
                      <mdi :path="mdiWindowClose" />
                    </n-icon>
                  </template>
                </n-button>
              </td>
            </tr>
          </tbody>
        </n-table>
      </div>
    </div>
  </div>
  <div
    style="display: flex"
    :style="
      props.element
        ? 'justify-content: space-between'
        : 'justify-content: flex-end'
    "
  >
    <n-button
      v-if="props.element"
      secondary
      type="error"
      @click="emit('delete')"
    >
      {{ t("actions.remove") }}
    </n-button>
    <n-button
      v-if="allowApply"
      type="primary"
      secondary
      @click="
        emit('confirm', {
          domain: domainName,
          annotations: annotations.filter((el) => el.key),
          paths: paths.filter((el) => el.path),
          httpOnly,
          wwwRedirect,
        })
      "
    >
      {{ t("actions.apply") }}
    </n-button>
    <n-tooltip v-else>
      <template #trigger>
        <n-button type="primary" secondary disabled tag="div">
          {{ t("actions.apply") }}
        </n-button>
      </template>
      {{ applyTooltipMessage() }}
    </n-tooltip>
  </div>
  <Modal
    v-model:show="advancedPathShow"
    :title="t('scratch.advanced')"
    style="width: 50rem"
    :showFooter="true"
    :touched="advancedTouched"
    @positive-click="
      () => {
        paths[advancedPathIndex] = advancedPath;
        advancedPathShow = false;
      }
    "
  >
    <div class="element-from-scratch-row">
      <ElementAddFromScratchRowHeader
        name="Prefix"
        :tip="t('scratch.docs.paths.prefix')"
      />

      <div style="width: 80%; height: 100%; padding-right: 5em">
        <n-input
          v-model:value="advancedPath.prefix"
          @input="advancedTouched = true"
          placeholder="Prefix"
        />
      </div>
    </div>
    <div class="element-from-scratch-row">
      <ElementAddFromScratchRowHeader
        :name="t('fields.label')"
        :tip="t('scratch.docs.paths.label')"
      />

      <div style="width: 80%; height: 100%; padding-right: 5em">
        <n-input
          v-model:value="advancedPath.label"
          @input="advancedTouched = true"
          :placeholder="t('fields.label')"
        />
      </div>
    </div>
  </Modal>
</template>

<style scoped lang="scss">
.element-from-scratch-row {
  display: flex;
  align-items: center;
  margin-bottom: 5px;
  &:hover {
    background-color: rgba(gray, 0.05);
  }
}

.element-from-scratch-row-header {
  width: 20%;
  display: flex;
  justify-content: flex-end;
  padding-right: 2em;
  text-align: right;
}
</style>
