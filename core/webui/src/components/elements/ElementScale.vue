<script setup lang="ts">
import { onBeforeMount } from "vue";
import { useI18n } from "vue-i18n";
import { useRoute } from "vue-router";
import { useMessage } from "naive-ui";

const props = defineProps<{
  scale: any;
  elementName: string;
  disable: boolean;
}>();

const { t } = useI18n();
const route = useRoute();
const message = useMessage();

// scale variants
const scaleVariants = [
  // { value: "off", name: t("actions.scaling.scaleToZero") },
  { value: "static", name: t("actions.scaling.static") },
  { value: "hpa", name: t("actions.scaling.dynamic") },
];
// dynamic scale value
let nrOfPodsMin = $ref(1);
let nrOfPodsMax = $ref(1);
let CPUTargetProc = $ref(props.scale.CPUTargetProc || 0);
let CPUTarget = $ref(props.scale.CPUTargetProc || 0);

// set current scale variant & static scale value
let currentScaleVariant: "off" | "static" | "hpa" | undefined = $ref(undefined);
let staticScalingValue = $ref(1);
onBeforeMount(() => {
  if (props.disable) currentScaleVariant = "off";
  else if (props.scale.NrOfPodsMin === props.scale.NrOfPodsMax) {
    currentScaleVariant = "static";
    staticScalingValue =
      props.scale.NrOfPodsMin > 0 ? props.scale.NrOfPodsMin : 1;
  } else {
    currentScaleVariant = "hpa";
    nrOfPodsMin = props.scale.NrOfPodsMin;
    nrOfPodsMax = props.scale.NrOfPodsMax;
    CPUTargetProc =
      props.scale.CPUTargetProc === 0
        ? props.scale.CPUTargetMinCores / 100
        : props.scale.CPUTargetProc;
  }
});

// change scale type
const changeScaleType = (e: "off" | "static" | "hpa") => {
  currentScaleVariant = e;
  if (e === "off") staticScalingValue = 0;
  if (e === "static")
    staticScalingValue =
      props.scale.NrOfPodsMin > 0 ? props.scale.NrOfPodsMin : 1;
};

// emit close dialog
const emit = defineEmits<{
  (e: "closeDialog"): void;
}>();

// handle confirm
const handleConfirm = () => {
  if (currentScaleVariant === "off" || currentScaleVariant === "static") {
    if (currentScaleVariant === "static") {
      // validation
      if (staticScalingValue <= 0 || !Number.isInteger(staticScalingValue)) {
        message.error(t("messages.positiveInteger"));
        return;
      }
    }
    api
      .post(
        "/env-element-static-scale",
        {
          [props.elementName]: staticScalingValue,
        },
        {
          queries: {
            env: route.params.id as string,
          },
        }
      )
      .then((res) => {
        if (res === "ok") {
          message.success(t("messages.scalingChanged"));
          emit("closeDialog");
        } else {
          message.error(res);
        }
      })
      .catch((error) => {
        message.error(error);
      });
  } else if (currentScaleVariant === "hpa") {
    api
      .post("/env-element-scale", {
        CPUTargetProc: CPUTarget,
        Element: props.elementName,
        EnvID: route.params.id as string,
        NrOfPodsMax: nrOfPodsMax,
        NrOfPodsMin: nrOfPodsMin,
      })
      .then((res) => {
        if (res === "ok") {
          message.success(t("messages.scalingChanged"));
          emit("closeDialog");
        } else {
          message.error(res);
        }
      })
      .catch((error) => {
        message.error(error);
      });
  }
};
</script>

<template>
  <div>
    <n-tabs
      :default-value="currentScaleVariant"
      type="line"
      animated
      @update:value="changeScaleType($event)"
    >
      <n-tab-pane
        v-for="variant in scaleVariants"
        :key="variant.value"
        :name="variant.value"
        :tab="variant.name"
        style="min-height: 5em"
      >
        <div
          v-if="currentScaleVariant === 'off'"
          style="display: flex; align-items: center"
        >
          <p>{{ t("fields.scaleToZero") }}</p>
        </div>
        <div
          v-if="currentScaleVariant === 'static'"
          style="display: flex; align-items: center"
        >
          <p style="width: 25%">{{ t("fields.nrOfPods") }}</p>
          <n-input-number
            v-model:value="staticScalingValue"
            style="width: 75%"
            :validator="(x: number) => {
              if (scale.MaxOnePod) return (x === 0 || x === 1);
              else return x > 0;
            }"
          />
        </div>
        <div
          v-else-if="currentScaleVariant === 'hpa'"
          style="
            display: flex;
            flex-direction: column;
            gap: 0.5rem;
            padding-bottom: 0.75rem;
          "
        >
          <p style="font-weight: bold; font-size: 1.1rem">
            {{ t("scaling.numberOfContainers") }}:
          </p>
          <div style="display: flex; gap: 0.5rem; align-items: center">
            <p style="width: 6rem">{{ t("scaling.minimal", 1) }}:</p>
            <n-input-number
              :min="0"
              :max="nrOfPodsMax"
              v-model:value="nrOfPodsMin"
            />
          </div>
          <div style="display: flex; gap: 0.5rem; align-items: center">
            <p style="width: 6rem">{{ t("scaling.current", 1) }}:</p>
            <span>
              {{
                scale.NrOfPodsMin == 0 && scale.NrOfPodsMax == 0
                  ? 0
                  : scale.NrOfPodsCurrent
              }}
            </span>
          </div>
          <div style="display: flex; gap: 0.5rem; align-items: center">
            <p style="width: 6rem">{{ t("scaling.maximal", 1) }}:</p>
            <n-input-number :min="nrOfPodsMin" v-model:value="nrOfPodsMax" />
          </div>
          <p style="font-weight: bold; font-size: 1.1rem">
            {{ t("scaling.processorLoad") }}:
          </p>
          <n-table :single-line="false" class="data-table" :bordered="false">
            <thead>
              <tr>
                <th></th>
                <th>{{ t("scaling.minimal", 2) }}</th>
                <th>{{ t("scaling.current", 2) }}</th>
                <th>{{ t("scaling.target") }}</th>
                <th>{{ t("scaling.maximal", 2) }}</th>
                <th>{{ t("scaling.requested") }}</th>
              </tr>
            </thead>
            <tbody>
              <!-- <tr>
                <td>{{ t("scaling.numberOfCore") }}</td>
                <td>{{ scale.CPUTargetMinCores / 100 }}</td>
                <td></td>
                <td>
                  <n-input-number
                    v-model:value="CPUTargetProc"
                    @update:value="assignTargetCPU(true)"
                    style="max-width: 9rem"
                    :min="scale.CPUTargetMinCores / 100"
                    :max="scale.CPUTargetMaxCores / 100"
                    :step="0.1"
                  />
                </td>
                <td>{{ scale.CPUTargetMaxCores / 100 }}</td>
                <td>{{ scale.CPURequestedCores / 100 }}</td>
              </tr> -->
              <tr>
                <td>%</td>
                <td>0</td>
                <td>{{ CPUTargetProc }}</td>
                <td>
                  <n-input-number
                    v-model:value="CPUTarget"
                    style="max-width: 9rem"
                    :min="0"
                    :max="80"
                    :step="5"
                  />
                </td>
                <td>80</td>
                <td>100</td>
              </tr>
            </tbody>
          </n-table>
        </div>
      </n-tab-pane>
    </n-tabs>
    <div
      style="
        width: 100%;
        display: flex;
        justify-content: right;
        margin-top: 0.7em;
        flex-grow: 1;
        align-items: flex-end;
      "
    >
      <n-button @click="handleConfirm" type="primary" secondary strong>{{
        t("actions.confirm")
      }}</n-button>
    </div>
  </div>
</template>

<style>
.n-modal .n-table td {
  padding-top: 4px !important;
  padding-bottom: 4px !important;
}

.n-modal .n-table tr,
.n-modal .n-table th,
.n-modal .n-table thead {
  background: none;
}

.n-modal .n-table th {
  padding-top: 4px !important;
  padding-bottom: 4px !important;
  font-size: var(--fontSizeMini);
}
</style>
