<script setup lang="ts">
// import { useEnv } from "@/store/envStore";
import { useRoute } from "vue-router";
import { ElementMapRespExtended } from "@/zodios/schemas/elements";
import { z } from "zod";
import { useEnvElements } from "@/store/envStore";
import { useUserStore } from "@/store/userStore";
import { useMessage } from "naive-ui";

const userStore = useUserStore();
// import { ElementInputs } from "../elements/ElementsContainer.vue";
// import { Input } from "../../zodios/schemas/inputs";
// import { EnvModRequest } from "../../zodios/schemas/envMod";
// import { useMessage } from "naive-ui";
// import { z } from "zod";
// import Spinner from "../Spinner.vue";
// import { useI18n } from "vue-i18n";

// interface MultiElementInputs {
//   [k: string]: ElementInputs;
// }

// // @ts-ignore
// type ElementInput = z.infer<typeof Input>;
// type EnvModReq = z.infer<typeof EnvModRequest>;

const { t } = useI18n();
const route = useRoute();
const message = useMessage();

let variablesLoading = $ref(false);

let changeModal = $ref(false);
let currentElement = $ref<z.infer<typeof ElementMapRespExtended>>();
let selectedVariable = $ref("");

const env = useEnvElements(computed(() => route.params.id as string));

const elements = computed(() => {
  return Object.values(env.value?.EnvElements || {});
});

let data = $ref<any>();

const columns = $computed(() => [
  {
    title: "",
    template: "status",
    width: "2%",
  },
  {
    title: t("fields.name"),
    key: "Name",
    width: "25%",
  },
  {
    title: t("scratch.rawValue"),
    key: "CurrentValue",
    width: "23%",
  },
  {
    title: t("fields.value"),
    key: "ResolvedValue",
    width: "23%",
    template: "value",
  },
  // {
  //   title: t("fields.description"),
  //   key: "Description",
  //   width: "19%",
  // },
  {
    title: "",
    template: "change",
    width: "10%",
  },
]);

const fetchVariables = () => {
  variablesLoading = true;
  api
    .get("/env-variables", {
      queries: {
        env: route.params.id as string,
      },
    })
    .then((res) => {
      data = Object.entries(res)
        .map((el) => {
          return {
            Name: el[0],
            ...(el[1] as Object),
          };
        })
        .sort((a, b) =>
          a.Name.split(".")[0].localeCompare(b.Name.split(".")[0])
        );
      variablesLoading = false;
    });
};

const openModal = (elName: string, variable: string) => {
  changeModal = true;
  selectedVariable = variable;
  currentElement = elements.value.find((el) => el.Info.Name === elName)!;
};

const variableErrorMessages = (code: number) => {
  switch (code) {
    case 10:
      return t("errorMessages.elementNotFound");
    case 11:
      return t("errorMessages.variableNotFound");
    case 12:
      return t("errorMessages.invalidReference");
    case 20:
      return t("errorMessages.emptyValue");
    case 30:
      return t("errorMessages.invalidValidator");
    case 31:
      return t("errorMessages.invalidValidatorArgs");
    case 40:
      return t("errorMessages.invalidName");
    case 41:
      return t("errorMessages.validationFailed");
    default:
      return t("errorMessages.unknownError");
  }
};

const variableCopy = (row: any) => {
  if (row.Secret) {
    const [element, variable] = row.Name.split(".");
    api
      .get("/env-variable-get-secret", {
        queries: {
          env: route.params.id as string,
          element,
          variable,
        },
      })
      .then((res) => {
        if (res === "permission denied")
          message.error(t("messages.permissionDenied"));
        else {
          navigator.clipboard.writeText(res);
          message.success(t("messages.copied"));
        }
      });
  } else {
    navigator.clipboard.writeText(row.ResolvedValue);
    message.success(t("messages.copied"));
  }
};

onBeforeMount(() => {
  fetchVariables();
});
</script>

<template>
  <div>
    <EnvTab />
    <PageLayout>
      <n-card
        v-if="userStore.havePermission('Env_View')"
        :title="t('envTabs.variables', 2)"
        :size="'small'"
        style="height: calc(100vh - 5.1rem); position: relative"
      >
        <Spinner :data="true">
          <n-spin
            v-if="variablesLoading"
            :size="60"
            stroke="#1ba3fd"
            :stroke-width="10"
            style="height: 100%; max-height: 80vh"
          />
          <data-table
            v-else
            :columns="columns"
            :data="data"
            :max-height="'calc(100vh - 11rem)'"
          >
            <template #value="row">
              <div>
                <div style="display: flex; justify-content: space-between">
                  <n-button
                    v-if="row.ResolvedValue"
                    secondary
                    size="tiny"
                    class="copy"
                    icon-placement="right"
                    style="margin-right: 10px"
                    @click="variableCopy(row)"
                  >
                    <div class="wrap" v-if="row.Secret">*****</div>
                    <div class="wrap" v-else>
                      {{ row.ResolvedValue }}
                    </div>
                    <template #icon>
                      <n-icon>
                        <mdi :path="mdiContentCopy" />
                      </n-icon>
                    </template>
                  </n-button>
                  <!-- {{ showPassword.find((el) => el.rowName === row.Name) }} -->
                </div>
              </div>
            </template>

            <template #status="row">
              <n-tooltip
                trigger="hover"
                v-if="row.Errors && Object.keys(row.Errors).length"
              >
                <template #trigger>
                  <button class="success-error-btn">
                    <n-icon color="#d03050">
                      <mdi :path="mdiInformation" />
                    </n-icon>
                  </button>
                </template>
                <span
                  v-for="err in (Object.values(row.Errors || {}) as Array<number>)"
                  :key="err"
                >
                  {{ variableErrorMessages(err) }}
                </span>
              </n-tooltip>

              <div
                v-if="
                  row.ValueInDB === row.ValueInGit &&
                  !Object.keys(row.Errors).length
                "
                class="success-error-btn"
              >
                <n-icon color="#18a058">
                  <mdi :path="mdiCheckCircle" />
                </n-icon>
              </div>
            </template>

            <template #change="row">
              <div
                style="
                  display: flex;
                  gap: 0.5rem;
                  align-items: center;
                  justify-content: flex-end;
                "
              >
                <n-button
                  v-if="!row.System"
                  size="tiny"
                  strong
                  secondary
                  type="primary"
                  @click="
                    openModal(row.Name.split('.')[0], row.Name.split('.')[1])
                  "
                  :disabled="!userStore.havePermission('Env_ElementFullManage')"
                >
                  <template #icon>
                    <n-icon class="icon">
                      <mdi :path="mdiPencil" />
                    </n-icon>
                  </template>
                  {{ t("actions.edit") }}
                </n-button>
              </div>
            </template>
          </data-table>
        </Spinner>
      </n-card>
      <ErrorDialog v-else :msg="$t('messages.permissionDenied')" />
    </PageLayout>
    <Modal
      v-model:show="changeModal"
      :title="
        t('actions.edit') +
        ': ' +
        currentElement?.Info?.Name +
        ' (' +
        currentElement?.Info?.Type +
        ')'
      "
      style="width: 80rem"
    >
      <ElementAddFromScratch
        :element="currentElement"
        @fromScratchCreated="
          async () => {
            await fetchVariables();
            changeModal = !changeModal;
          }
        "
        :selectedTab="'variables'"
        :selectedVariable="selectedVariable"
      />
    </Modal>
  </div>
</template>

<style scoped lang="scss">
.element-input {
  display: flex;
  margin: 10px;
  justify-content: space-between;

  .input {
    width: 400px;
  }
}

.success-error-btn {
  display: flex;
  background-color: transparent;
  border: none;
}

.copy {
  width: 100%;
  & :deep(.n-button__content) {
    justify-content: space-between;
    width: 100%;
  }
}

.wrap {
  inline-size: 90%;
  overflow: hidden;
  text-align: left;
}
</style>
