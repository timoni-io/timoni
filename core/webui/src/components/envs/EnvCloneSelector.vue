<script setup lang="ts">
import { useDashboard } from "@/store/envStore";
import { useMessage } from "naive-ui";
import { useI18n } from "vue-i18n";
import { useRouter } from "vue-router";

defineProps<{
  manage: boolean;
}>();
const dashboard = useDashboard();
const message = useMessage();
const { t } = useI18n();
const router = useRouter();

let cloneEnvPopmodal = $ref(false);
let selectedEnv = $ref(undefined);
let envNameError = $ref("");
let selectedEnvError = $ref("");
let clonedEnv = $ref<string>("");

let envList = computed<{ label: string; value: string }[]>(() => {
  return Object.values(dashboard.envMap || {}).map((env) => {
    return { label: env.Name, value: env.ID };
  });
});

const cloneEnv = () => {
  api
    .get("/env-clone", {
      queries: {
        env: selectedEnv || "",
        targetName: clonedEnv as string,
      },
    })
    .then((res) => {
      if (res.substring(0, 3) === "env") {
        message.info(t("messages.createdEnvironment"));
        setTimeout(() => {
          router.push(`/env/${res}`);
        });
      } else {
        message.error(res);
      }
    });
};

const createEnvF = () => {
  if (!clonedEnv.trim().length) envNameError = t("messages.nameRequired");
  if (!selectedEnv) selectedEnvError = t("messages.envRequired");

  if (clonedEnv.trim().length && selectedEnv) {
    if (clonedEnv.trim().length >= 31) {
      return;
    }
    cloneEnv();
  }
};

watch(
  () => clonedEnv,
  () => {
    if (clonedEnv.trim().length >= 31) envNameError = t("messages.tooLongName");
    else envNameError = "";
  }
);

watch(
  () => selectedEnv,
  () => {
    if (selectedEnv) selectedEnvError = "";
  }
);
</script>

<template>
  <PopModal
    :title="t('actions.cloneEnv')"
    :show="cloneEnvPopmodal"
    :touched="selectedEnv !== undefined || clonedEnv.length > 0"
    :width="'30rem'"
  >
    <template #trigger>
      <n-button
        size="tiny"
        strong
        secondary
        type="primary"
        :disabled="!manage"
        @click="
          () => {
            selectedEnv = undefined;
            clonedEnv = '';
            envNameError = '';
          }
        "
      >
        <template #icon>
          <n-icon>
            <Mdi :path="mdiCheckboxMultipleBlankOutline" />
          </n-icon>
        </template>
        {{ t("actions.cloneEnv") }}
      </n-button>
    </template>
    <template #content>
      <div style="display: flex; gap: 0.5rem">
        <div>
          <div
            style="
              display: grid;
              grid-template-columns: 6rem 1fr;
              align-items: center;
              margin-bottom: 0.5rem;
            "
            :style="envNameError ? 'padding-bottom: 0.65rem' : ''"
          >
            <p>{{ t("fields.name") }}</p>
            <Input
              :placeholder="t('fields.envName')"
              :errorMessage="envNameError"
              :focus="false"
              :removeWhiteSpace="true"
              :name="clonedEnv"
              @keyup.enter="createEnvF"
              @update:value="
                          (v: string) => {
                            clonedEnv = v;
                          }
                        "
            />
          </div>
          <div
            style="
              display: grid;
              grid-template-columns: 6rem 1fr;
              align-items: center;
            "
          >
            <p>{{ t("objects.environment") }}</p>
            <div>
              <n-select
                v-model:value="selectedEnv"
                :options="envList"
                :placeholder="t('objects.environment')"
              />
              <div
                style="
                  font-size: 0.8rem;
                  color: tomato;
                  transition: all 0.2s ease-in-out;
                  transform: translateY(-1rem);
                  opacity: 0;
                  height: 0;
                "
                :style="
                  selectedEnvError
                    ? { transform: 'translateY(0rem)', opacity: 1 }
                    : {}
                "
              >
                {{ selectedEnvError }}
              </div>
            </div>
          </div>
        </div>
        <n-button
          :style="
            envNameError && selectedEnvError
              ? 'height: 96px'
              : envNameError || selectedEnvError
              ? 'height: 87px'
              : 'height: 78px'
          "
          secondary
          type="primary"
          @click="createEnvF"
          >{{ t("actions.confirm") }}</n-button
        >
      </div>
    </template>
  </PopModal>
</template>
