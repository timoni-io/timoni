<script setup lang="ts">
import { useI18n } from "vue-i18n";
import {
  ElementMapRespExtended,
  ElementMapStatus,
  PodType,
} from "@/zodios/schemas/elements";
import { dynamicInputsHandle } from "@/utils/dynamicInputsHandle";
import { z } from "zod";
import TOML from "@ltd/j-toml";
import { useRoute } from "vue-router";
import { useMessage } from "naive-ui";
import { findDuplicates } from "@/utils/findDuplicates";

export interface Annotation {
  key: string;
  value: string;
}

export interface Action {
  key: string;
  command: Array<string>;
}

export interface Path {
  path: string;
  target: string;
  label: string;
  prefix: string;
}

export interface Probe {
  disable: boolean;
  path: string;
  "initial-delay-seconds": number;
  "period-seconds": number;
  "timeout-seconds": number;
  "success-threshold": number;
  "failure-threshold": number;
  "restart-on-fail": boolean;
}

export interface Exposition {
  port: number;
  type: string;
  metricsPath: string;
  name: string;
  probe: Probe;
}

export interface Storage {
  path: string;
  type: string;
  maxSizeMB: number;
  readOnly: boolean;
  class: string;
  name: string;
  login: string;
  password: string;
  remoteHost: string;
  remotePath: string;
  options: string;
}

type KebabCase<S> = S extends `${infer H}${infer T}`
  ? T extends Uncapitalize<T>
    ? `${Uncapitalize<H>}${KebabCase<T>}`
    : `${Uncapitalize<H>}-${KebabCase<T>}`
  : S;

interface Validation {
  type: string;
  min: number;
  max: number;
  allowSpaces: boolean;
  allowSpecial: boolean;
  oneOfValues: Array<string>;
  regexString: string;
  minLen: number;
  minLetters: number;
  upAndLow: boolean;
  minDigits: number;
  minSpecial: number;
}

interface Variable {
  name: string;
  value: string;
  secret: boolean;
  system: boolean;
  validation: Validation;
  description: string;
  errors: Record<string, number>;
  validationResult?: {
    success: boolean;
    message: string;
  };
}

interface Target {
  element: string;
  port: number;
  prefix: string;
  label: string;
}

interface ScratchMode {
  label: string;
  value: string;
  description: string;
}

interface BaseManifest {
  type: string;
  description: string;
  variables: Record<string, any>;
}

interface ConfigManifest extends BaseManifest {}
interface DomainManifest<T, B> {
  annotations: T;
  domain: string;
  paths: B;
  "http-only": boolean;
  "www-redirect": boolean;
}
interface DomainManifestProps {
  annotations: Array<Annotation>;
  domain: string;
  paths: Array<Path>;
  httpOnly: boolean;
  wwwRedirect: boolean;
}
interface PodManifest {
  build: {
    type: string;
    script?: string;
    dockerfile?: string;
    image?: string;
  };
  "run-cmd": string[];
  "run-as-user"?: Array<number>;
  "run-writable-file-system": boolean;
  actions: Record<string, Array<string>>;
  expose: Record<string, KebabCase<Exposition>>;
  storage: Record<string, KebabCase<Storage>>;
}
interface PodManifestProps extends Pick<PodManifest, "build"> {
  runCmd: Array<string>;
  runAsUser: number | null;
  actions: Action[];
  expositions: Record<string, Exposition>;
  storages: Record<string, Storage>;
  writableFileSystem: boolean;
  image: string;
  dockerfilePath: string;
  buildType: string;
  script: string;
}

type Pod = {
  Info: z.infer<typeof PodType>;
  Status: z.infer<typeof ElementMapStatus>;
};

const { t } = useI18n();
const emit = defineEmits<{
  (e: "isTouched"): void;
  (e: "detouch"): void;
  (e: "fromScratchCreated"): void;
}>();
const route = useRoute();
const props = defineProps<{
  element?: z.infer<typeof ElementMapRespExtended>;
  selectedTab?: string;
  selectedVariable?: string;
}>();
const message = useMessage();

let inputRef = $ref(null as HTMLInputElement | null);

let selectedType = $ref(props.element?.Info.Type || "");
let description = $ref(props.element?.Info.Description || "");
let build = $ref("");
let buildType = $ref("");
let image = $ref("");
let dockerfilePath = $ref("");
let runCmd = $ref<Array<string>>([]);
let runAsUser = $ref<number | null>(null);
let domainName = $ref("");
let httpOnly = $ref<boolean>(false);
let wwwRedirect = $ref<boolean>(false);
let elementName = $ref(props.element?.Info.Name || "");
let variables = $ref<Variable[]>([]);
let annotations = $ref<Annotation[]>([]);
let paths = $ref<Array<Path>>([]);
let actions = $ref<Action[]>([]);
let expositions = $ref<Record<string, Exposition>>({});
let storages = $ref<Record<string, Storage>>({});
let writableFileSystem = $ref(false);
let openErrorModal = $ref(false);
let nameModal = $ref(false);
let deleteShow = $ref(false);
let anotherTry = $ref(false);
let addStopped = $ref(false);
let errorMessage = $ref("");
let selectedTab = $ref(props.selectedTab || "general");

const onVariableCreate = (): Variable => {
  return {
    name: "",
    value: "",
    secret: false,
    system: false,
    validation: {
      type: "noValidation",
      min: 0,
      max: 0,
      allowSpaces: true,
      allowSpecial: true,
      oneOfValues: [],
      regexString: "",
      minLen: 0,
      minLetters: 0,
      upAndLow: false,
      minDigits: 0,
      minSpecial: 0,
    },
    description: "",
    errors: {},
  };
};

let advancedVariableShow = $ref(false);
let advancedVariable = $ref<Variable>(onVariableCreate());
let advancedVariableIndex = $ref<number>(-1);
let advancedTouched = $ref<boolean>(false);

const boolValues = [
  "true",
  "false",
  "0",
  "1",
  "yes",
  "no",
  "on",
  "off",
  "y",
  "n",
  "t",
  "f",
];
const validationFunctions = $ref([
  {
    label: t("scratch.noValidation"),
    value: "noValidation",
  },
  {
    label: "int",
    value: "int",
  },
  {
    label: "bool",
    value: "bool",
  },
  {
    label: "text",
    value: "text",
  },
  {
    label: "oneof",
    value: "oneof",
  },
  {
    label: "regex",
    value: "regex",
  },
  {
    label: "password",
    value: "password",
  },
]);

const validationDecoder = (validationString: string): Validation => {
  let validation: Validation = {
    type: "noValidation",
    min: 0,
    max: 0,
    allowSpaces: true,
    allowSpecial: true,
    oneOfValues: [],
    regexString: "",
    minLen: 0,
    minLetters: 0,
    upAndLow: false,
    minDigits: 0,
    minSpecial: 0,
  };

  if (validationString.slice(0, 4) === "int(") {
    const minmax = validationString
      .replace("int(", "")
      .replace(")", "")
      .split(",")
      .map((el) => el.trim());
    validation.type = "int";
    validation.min = parseInt(minmax[0]);
    validation.max = parseInt(minmax[1]);
  } else if (validationString === "bool") {
    validation.type = "bool";
  } else if (validationString.slice(0, 5) === "text(") {
    const allow = validationString
      .replace("text(", "")
      .replace(")", "")
      .split(",");

    validation.type = "text";
    validation.allowSpaces = allow[0].trim() !== "false";
    validation.allowSpecial = allow[1].trim() !== "false";
  } else if (validationString.slice(0, 6) === "oneof(") {
    validation.type = "oneof";
    validation.oneOfValues = validationString
      .replace("oneof(", "")
      .replace(")", "")
      .split(",")
      .map((el) => el.trim());
  } else if (validationString.slice(0, 6) === "regex(") {
    validation.type = "regex";
    validation.regexString = validationString
      .replace("regex(", "")
      .replace(")", "")
      .replaceAll('"', "");
  } else if (validationString.slice(0, 9) === "password(") {
    const constraints = validationString
      .replace("password(", "")
      .replace(")", "")
      .split(",")
      .map((el) => el.trim());

    validation.type = "password";
    validation.minLen = parseInt(constraints[0]);
    validation.minLetters = parseInt(constraints[1]);
    validation.upAndLow = constraints[2] !== "false";
    validation.minDigits = parseInt(constraints[3]);
    validation.minSpecial = parseInt(constraints[4]);
  }
  return validation;
};

const validationEncoder = (validation: Validation): string => {
  switch (validation.type) {
    case "int":
      return "int(" + validation.min + ", " + validation.max + ")";
    case "bool":
      return "bool";
    case "text":
      return (
        "text(" + validation.allowSpaces + ", " + validation.allowSpecial + ")"
      );
    case "oneof":
      return "oneof(" + validation.oneOfValues.join(", ") + ")";
    case "regex":
      return 'regex("' + validation.regexString + '")';
    case "password":
      return (
        "password(" +
        validation.minLen +
        ", " +
        validation.minLetters +
        ", " +
        validation.upAndLow +
        ", " +
        validation.minDigits +
        ", " +
        validation.minSpecial +
        ")"
      );
    default:
      return "";
  }
};

let secretModal = $ref<boolean>(false);
// let secretIndex = $ref<number>(-1);
let secretOverwriteModal = $ref<boolean>(false);
let secretOverwriteValue = $ref("");
let secretOverwriteIndex = $ref<number>(-1);

const scratchModes = $ref<ScratchMode[]>([
  {
    label: t("scratch.types.domain"),
    value: "domain",
    description: t("scratch.descriptions.domain"),
  },
  {
    label: t("scratch.types.pod"),
    value: "pod",
    description: t("scratch.descriptions.pod"),
  },
  {
    label: t("scratch.types.config"),
    value: "config",
    description: t("scratch.descriptions.config"),
  },
  {
    label: "Elasticsearch",
    value: "elasticSearch",
    description: "Coming soon",
  },
  {
    label: "MongoDB",
    value: "mongo",
    description: "Coming soon",
  },
]);

// const elementTypes = $ref<Record<string, string>>({
//   domain: t("scratch.types.domain"),
//   pod: t("scratch.types.pod"),
//   config: t("scratch.types.config"),
//   mongoDB: "Mongo DB",
//   elasticSearch: "Elasticsearch",
// });

const emptyVariables = [""];

const tabs = [
  {
    name: "variables",
    value: t("scratch.tabs.variables"),
  },
];

const domainTabs = [
  {
    name: "annotations",
    value: t("scratch.tabs.annotations"),
  },
  {
    name: "paths",
    value: t("scratch.path", 2),
  },
];

const podTabs = [
  {
    name: "build",
    value: t("scratch.tabs.build"),
  },
  {
    name: "actions",
    value: t("scratch.tabs.actions"),
  },
  {
    name: "network",
    value: t("scratch.tabs.network"),
  },
  {
    name: "storage",
    value: t("scratch.tabs.storage"),
  },
  // {
  //   name: "scaling",
  //   value: t("scratch.tabs.scaling"),
  // },
];

const tabsFactory = (
  type: string
  // create: boolean
): Array<Pick<Variable, "name" | "value">> => {
  return [
    ...(selectedType === "config"
      ? []
      : [
          {
            name: "general",
            value: t("scratch.tabs.general"),
          },
        ]),
    ...((type: string) => {
      switch (type) {
        case "pod":
          return [...tabs, ...podTabs];
        case "domain":
          return [...tabs, ...domainTabs];
        case "config":
          return tabs;
        default:
          return tabs;
      }
    })(type),
    ...[
      {
        name: "description",
        value: t("scratch.tabs.description"),
      },
    ],
  ];
};

const testElementName = computed(() => {
  return /^[a-z0-9\-]+$/.test(elementName);
});

const reset = () => {
  description = "";
  build = "";
  runCmd = [];
  elementName = "";
  variables = [];
  annotations = [];
  paths = [];
  actions = [];
  expositions = {};
  storages = {};
  writableFileSystem = false;
};

const kebabCase = (str: string): string => {
  return str
    .replace(/([a-z])([A-Z])/g, "$1-$2")
    .replace(/[\s_]+/g, "-")
    .toLowerCase();
};

type ChangeFields<T, R> = Omit<T, keyof R> & R;

const createElementFromScratch = () => {
  if (elementName && (!anotherTry || props.element)) {
    let toml = "";
    let baseManifest: BaseManifest = {
      type: selectedType,
      description: description.trim(),
      variables: TOML.Section({}),
    };
    for (const vari of variables) {
      if (vari.name) {
        baseManifest["variables"][vari.name] = {
          value: ["\{\{ secret \}\}", "\{\{ random-string \}\}"].includes(
            vari.value
          )
            ? ""
            : vari.value,
          secret: vari.secret,
          description: vari.description,
          validation: vari.secret ? "" : validationEncoder(vari.validation),
        };
      }
    }

    switch (selectedType) {
      case "config":
        let configRes: ConfigManifest = baseManifest;
        toml = TOML.stringify(configRes as Readonly<ConfigManifest>);
        break;

      case "domain":
        let domainRes: KebabCase<
          DomainManifest<Record<string, string>, Record<string, Target>> &
            BaseManifest
        > = {
          ...baseManifest,
          domain: domainName,
          annotations: TOML.Section({}),
          paths: TOML.Section({}),
          "http-only": httpOnly,
          "www-redirect": wwwRedirect,
        };
        for (const ann in annotations) {
          domainRes["annotations"][annotations[ann].key] =
            annotations[ann].value;
        }
        for (const path of paths) {
          const splitedTarget = path.target.split(":");
          domainRes["paths"][path.path] = {
            element: splitedTarget[0],
            port: splitedTarget.length === 2 ? parseInt(splitedTarget[1]) : 0,
            prefix: path.prefix,
            label: path.label
          };
        }
        toml = TOML.stringify(
          // @ts-ignore
          domainRes,
          {
            integer: 50000,
          }
        );
        break;

      case "pod":
        let podRes: KebabCase<
          ChangeFields<
            PodManifest,
            {
              expose: Omit<PodManifest["expose"], "port">;
              storage: Omit<PodManifest["storage"], "path">;
            }
          > &
            BaseManifest
        > = {
          ...baseManifest,
          build: TOML.Section({
            type: buildType,
          }),
          "run-cmd": runCmd,
          "run-writable-file-system": writableFileSystem,
          expose: TOML.Section({}),
          actions: TOML.Section({}), //tu dokończyć
          storage: TOML.Section({}),
          // scale: TOML.Section({
          //   min: parseInt(scale.min + ""),
          //   max: parseInt(scale.max + ""),
          //   targetCPU: parseInt(scale.targetCPU + ""),
          //   disable: scale.disable,
          // }) as Scale,
        };
        if (buildType === "dockerfile")
          podRes["build"]["dockerfile"] = dockerfilePath;
        if (buildType === "script") podRes["build"]["script"] = build;
        if (buildType === "image") podRes["build"]["image"] = image;
        if (runAsUser !== null) podRes["run-as-user"] = [runAsUser];

        for (const exp in expositions) {
          podRes["expose"][exp] = <KebabCase<Exposition>>{};
          Object.keys(expositions[exp]).forEach((key: string) => {
            if (key !== "port") {
              // @ts-ignore
              podRes["expose"][exp][
                kebabCase(key) as keyof KebabCase<Exposition>
              ] = expositions[exp][key as keyof Exposition];
            }
          });
        }
        for (const storage in storages) {
          podRes["storage"][storage] = <KebabCase<Storage>>{};
          Object.keys(storages[storage]).forEach((key: string) => {
            if (key !== "path") {
              // @ts-ignore
              podRes["storage"][storage][
                kebabCase(key) as keyof KebabCase<Storage>
              ] = storages[storage][key as keyof Storage];
            }
          });
        }
        for (const action in actions) {
          if (actions[action].key)
            podRes["actions"][actions[action].key] = actions[
              action
            ].command.filter((comm) => comm);
        }
        // @ts-ignore
        toml = TOML.stringify(podRes, {
          integer: 50000,
        });
        break;

      default:
        break;
    }
    api
      .post(
        props.element
          ? "/env-element-update-from-toml"
          : "/env-element-create-from-toml",
        // @ts-ignore
        toml.join("\n"),
        {
          queries: {
            element: elementName as string,
            env: route.params.id as string,
            "dont-start": addStopped as boolean,
          },
        }
      )
      .then((resp) => {
        if (resp === "ok") {
          emit("fromScratchCreated");
          props.element === undefined
            ? message.success(t("scratch.created"))
            : message.success(t("scratch.edited"));
        } else {
          errorMessage = resp as string;
          anotherTry = true;
          openErrorModal = true;
        }
      });
  } else nameModal = true;
};

const variablesDuplicates = $computed<Array<string>>(() =>
  findDuplicates<string>(variables.map((vari) => vari.name))
);

const allowApplyBase = $computed<boolean>(
  () => variablesDuplicates.length === 0
  // && !variables.filter((el) => el.secret && !el.value).length
);

const applyTooltipMessage = computed<string>((): string => {
  let resp = [];
  if (variablesDuplicates.length)
    resp.push(
      t("scratch.removeDuplicates") + " " + t("scratch.variables") + "."
    );
  // if (variables.filter((el) => el.secret && !el.value).length)
  //   resp.push(t("scratch.emptySecret") + ".");

  return resp.join(" ");
});

const createClick = () => {
  nameModal = false;
  anotherTry = false;
  createElementFromScratch();
};

const deleteElement = () => {
  api
    .get("/env-element-delete", {
      queries: {
        env: route.params.id as string,
        element: props.element?.Info.Name as string,
      },
    })
    .then((res) => {
      deleteShow = false;
      if (res === "ok") {
        message.success(t("messages.willBeDeleted"));
        emit("fromScratchCreated");
      } else message.error(res);
    })
    .catch(() => {
      deleteShow = false;
    });
};

const validationFunctionFactory = (variable: Variable) => {
  if (
    !(
      variable.name !== undefined &&
      (/^[a-zA-Z_]+\w*$/.test(variable.name) || variable.name === "")
    )
  )
    return {
      success: false,
      message: t("errorMessages.invalidName"),
    };
  if (variable.validation.type === "noValidation" || variable.secret)
    return {
      success: true,
      message: t("fields.value") + ": " + variable.value,
    };
  if (variable.validation.type === "int") {
    const parsedVariable = parseInt(variable.value);
    if (
      !Number.isNaN(parsedVariable) &&
      parsedVariable >= variable.validation.min &&
      parsedVariable <= variable.validation.max
    )
      return {
        success: true,
        message: t("fields.value") + ": " + variable.value,
      };
    else
      return {
        success: false,
        message:
          t("scratch.validationErrors.validationError") +
          " " +
          variable.value +
          " " +
          t("scratch.validationErrors.int") +
          " (" +
          variable.validation.min +
          ", " +
          variable.validation.max +
          ")",
      };
  } else if (variable.validation.type === "bool") {
    if (boolValues.includes(variable.value))
      return {
        success: true,
        message: t("fields.value") + ": " + variable.value,
      };
    else
      return {
        success: false,
        message:
          t("scratch.validationErrors.validationError") +
          ' "' +
          variable.value +
          '" ' +
          t("scratch.validationErrors.boolAndOneOf"),
      };
  } else if (variable.validation.type === "text") {
    const format = /[!~@#$%^&*()_+\-=\[\]{}<>;:.,'"\\|\/?]/g.test(
      variable.value
    );
    if (
      (variable.validation.allowSpecial ||
        (!variable.validation.allowSpecial && !format)) &&
      (variable.validation.allowSpaces ||
        (!variable.validation.allowSpaces && !variable.value.includes(" ")))
    )
      return {
        success: true,
        message: t("fields.value") + ": " + variable.value,
      };
    else
      return {
        success: false,
        message:
          t("scratch.validationErrors.validationError") +
          ' "' +
          variable.value +
          '" ' +
          t("scratch.validationErrors.textAndPasword"),
      };
  } else if (variable.validation.type === "oneof") {
    if (
      variable.validation.oneOfValues.length > 0 &&
      variable.validation.oneOfValues.includes(variable.value)
    )
      return {
        success: true,
        message: t("fields.value") + ": " + variable.value,
      };
    else
      return {
        success: false,
        message:
          t("scratch.validationErrors.validationError") +
          ' "' +
          variable.value +
          '" ' +
          t("scratch.validationErrors.boolAndOneOf"),
      };
  } else if (variable.validation.type === "regex") {
    const re = new RegExp(variable.validation.regexString);
    if (re.test(variable.value))
      return {
        success: true,
        message: t("fields.value") + ": " + variable.value,
      };
    else
      return {
        success: false,
        message:
          t("scratch.validationErrors.validationError") +
          ' "' +
          variable.value +
          '" ' +
          t("scratch.validationErrors.regex") +
          ' "' +
          variable.validation.regexString +
          ' ".',
      };
  } else if (variable.validation.type === "password") {
    if (
      variable.value.length >= variable.validation.minLen &&
      (variable.value.match(/[a-zA-Z]/g)?.length || 0) >=
        variable.validation.minLetters &&
      (!variable.validation.upAndLow ||
        variable.value !== variable.value.toLowerCase()) &&
      (variable.value.match(/[0-9]/g)?.length || 0) >=
        variable.validation.minDigits &&
      (variable.value.match(/[!~@#$%^&*()_+\-=\[\]{}<>;:.,'"\\|\/?]/g)
        ?.length || 0) >= variable.validation.minSpecial
    )
      return {
        success: true,
        message: t("fields.value") + ": " + variable.value,
      };
    else
      return {
        success: false,
        message:
          t("scratch.validationErrors.validationError") +
          ' "' +
          variable.value +
          '" ' +
          t("scratch.validationErrors.textAndPasword"),
      };
  } else
    return {
      success: false,
      message: "no idea why",
    };
};

watch(
  $$(variables),
  () => {
    dynamicInputsHandle<Variable>(
      variables,
      ["name", "value"],
      emptyVariables,
      "variables-dynamic",
      onVariableCreate
    );
  },
  { deep: true }
);

onBeforeMount(() => {
  let variablesP: NonNullable<z.infer<typeof PodType>["Variables"]> =
    (props.element?.Info.Variables as z.infer<typeof PodType>["Variables"]) ||
    {};
  Object.keys(variablesP || {}).forEach((variable) => {
    if (!variablesP[variable].System)
      variables.push({
        name: variable,
        value: variablesP[variable].CurrentValue,
        description: variablesP[variable].Description,
        secret: variablesP[variable].Secret,
        system: variablesP[variable].System,
        validation: validationDecoder(variablesP[variable].Validation),
      } as Variable);
  });
  variables.push({
    name: "",
    value: "",
    secret: false,
    system: false,
    validation: {
      type: "noValidation",
      min: 0,
      max: 0,
      allowSpaces: true,
      allowSpecial: true,
      oneOfValues: [],
      regexString: "",
      minLen: 0,
      minLetters: 0,
      upAndLow: false,
      minDigits: 0,
      minSpecial: 0,
    },
    description: "",
    errors: {},
  });
});

onMounted(() => {
  setTimeout(() => {
    if (inputRef && Array.isArray(inputRef) && inputRef.length)
      // @ts-ignore
      inputRef[0]?.focus();
  }, 200);
});
</script>

<template>
  <n-tabs
    v-if="selectedType"
    justify-content="space-evenly"
    type="line"
    animated
    v-model:value="selectedTab"
    style="margin-bottom: 0.5em"
  >
    <n-tab
      v-for="tab in tabsFactory(selectedType)"
      :key="tab.name"
      :name="tab.name"
    >
      {{ tab.value }}
    </n-tab>
  </n-tabs>
  <div style="display: flex; min-height: 30em">
    <div style="width: 100%">
      <div v-if="!selectedType">
        <div
          v-for="scratchMode in scratchModes"
          :key="scratchMode.value"
          class="grid"
        >
          <div class="grid-left">
            <n-button
              class="mode-button"
              secondary
              type="primary"
              :disabled="['elasticSearch', 'mongo'].includes(scratchMode.value)"
              @click="
                () => {
                  selectedType = scratchMode.value;
                  if (selectedType === 'config') selectedTab = 'variables';
                  reset();
                  emit('isTouched');
                }
              "
            >
              {{ scratchMode.label }}
            </n-button>
            <div class="grid-right">{{ scratchMode.description }}</div>
          </div>
        </div>
      </div>
      <div
        style="
          display: flex;
          flex-direction: column;
          width: 100%;
          height: 100%;
          justify-content: space-between;
        "
      >
        <div style="display: flex; flex-direction: column; height: 100%">
          <!-- <div
            v-if="selectedType && selectedTab === 'general'"
            class="element-from-scratch-row"
          >
            <ElementAddFromScratchRowHeader
              :name="t('fields.description')"
              :tip="t('scratch.docs.description')"
            />

            <div style="width: 80%; height: 100%; padding-right: 5em">
              <n-input
                v-model:value="description"
                type="textarea"
                :placeholder="t('fields.description')"
                style="height: 100%"
              />
            </div>
          </div> -->
          <div
            v-if="selectedType && selectedTab === 'description'"
            style="height: 100%"
          >
            <div style="width: 100%; height: 100%; padding: 0 5em 1em 5em">
              <n-input
                v-model:value="description"
                type="textarea"
                :placeholder="t('fields.description')"
                style="height: 100%"
              />
            </div>
          </div>
          <div
            v-if="selectedType && selectedTab === 'variables'"
            id="variables-dynamic"
            style="
              width: 100%;
              padding: 0 5em 0 5em;
              overflow-y: scroll;
              max-height: 27em;
            "
          >
            <n-table
              :bordered="false"
              :single-line="true"
              class="dynamic-table"
            >
              <thead>
                <tr>
                  <th style="width: 25%">
                    <ElementAddFromScratchTableHeader
                      :name="t('fields.name')"
                      :tip="t('scratch.docs.variables.name')"
                    />
                  </th>
                  <th style="width: 25%">
                    <ElementAddFromScratchTableHeader
                      :name="t('fields.value')"
                      :tip="t('scratch.docs.variables.value')"
                    />
                  </th>
                  <th style="width: 34%">
                    <ElementAddFromScratchTableHeader
                      :name="t('scratch.advanced')"
                      :tip="t('scratch.docs.variables.advanced')"
                    />
                  </th>
                  <th style="width: 10%">
                    <ElementAddFromScratchTableHeader
                      :name="'Status'"
                      :tip="t('scratch.docs.variables.status')"
                    />
                  </th>
                  <th style="width: 6%"></th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(vari, index) in variables" :key="index">
                  <td>
                    <n-input
                      v-model:value="vari.name"
                      type="text"
                      :placeholder="t('fields.name')"
                      :status="
                        vari.name !== undefined &&
                        (/^[a-zA-Z_]+\w*$/.test(vari.name) || vari.name === '')
                          ? undefined
                          : 'error'
                      "
                    />
                  </td>
                  <td>
                    <n-button
                      v-if="vari.secret"
                      class="secret-button"
                      :type="vari.secret && !vari.value ? 'warning' : 'primary'"
                      style="width: 100%"
                      @click="
                        () => {
                          secretOverwriteIndex = index;
                          secretOverwriteValue = '';
                          secretOverwriteModal = true;
                        }
                      "
                    >
                      {{
                        vari.secret && !vari.value
                          ? t("others.autogeneratedSecret")
                          : t("fields.secret")
                      }}
                    </n-button>
                    <n-auto-complete
                      v-else-if="
                        !vari.secret && vari.validation.type === 'bool'
                      "
                      :ref="
                        vari.name === props.selectedVariable
                          ? 'inputRef'
                          : undefined
                      "
                      v-model:value="vari.value"
                      :get-show="() => true"
                      :options="
                        ['true', 'false'].filter((el) =>
                          el.includes(vari.value)
                        )
                      "
                      :placeholder="t('fields.value')"
                    />
                    <n-auto-complete
                      v-else-if="
                        !vari.secret && vari.validation.type === 'oneof'
                      "
                      :ref="
                        vari.name === props.selectedVariable
                          ? 'inputRef'
                          : undefined
                      "
                      v-model:value="vari.value"
                      :get-show="() => true"
                      :options="
                        vari.validation.oneOfValues.filter((el) =>
                          el.includes(vari.value)
                        )
                      "
                      :placeholder="t('fields.value')"
                    />
                    <n-input
                      v-else
                      :ref="
                        vari.name === props.selectedVariable
                          ? 'inputRef'
                          : undefined
                      "
                      v-model:value="vari.value"
                      type="text"
                      :placeholder="t('fields.value')"
                    />
                  </td>
                  <td>
                    <n-button
                      dashed
                      style="width: 100%"
                      @click="
                        () => {
                          advancedVariable = JSON.parse(JSON.stringify(vari));
                          advancedVariableIndex = index;
                          advancedVariableShow = true;
                        }
                      "
                    >
                      {{ t("scratch.advanced") }}
                    </n-button>
                  </td>
                  <td style="text-align: center">
                    <div
                      style="
                        display: flex;
                        justify-content: center;
                        padding-right: 2em;
                      "
                    >
                      <n-tooltip
                        v-if="validationFunctionFactory(vari).success"
                        placement="bottom"
                        trigger="hover"
                      >
                        <template #trigger>
                          <n-icon :size="25" color="var(--successColor)">
                            <Mdi :path="mdiCheckCircle" />
                          </n-icon>
                        </template>
                        <span>
                          {{
                            vari.secret
                              ? "\{\{" + t("fields.secret") + "\}\}"
                              : validationFunctionFactory(vari).message
                          }}
                        </span>
                      </n-tooltip>
                      <n-tooltip v-else placement="bottom" trigger="hover">
                        <template #trigger>
                          <n-icon :size="25" color="var(--errorColor)">
                            <Mdi :path="mdiCloseCircle" />
                          </n-icon>
                        </template>
                        <span>
                          {{
                            vari.secret
                              ? "\{\{ secret \}\}"
                              : validationFunctionFactory(vari).message
                          }}
                        </span>
                      </n-tooltip>
                    </div>
                  </td>
                  <td>
                    <n-button
                      v-if="
                        vari.name + vari.value !== '' ||
                        index < variables.length - 1
                      "
                      type="primary"
                      quaternary
                      circle
                      style="margin-left: 5px"
                      @click="
                        () => {
                          variables.splice(index, 1);
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
          <div
            v-if="selectedType === 'domain'"
            style="
              display: flex;
              flex-direction: column;
              justify-content: space-between;
              flex-grow: 1;
            "
          >
            <ElementAddFromScratchDomain
              :element="props.element"
              :testElementName="testElementName"
              :selectedTab="selectedTab"
              :allowApplyResp="applyTooltipMessage"
              @confirm="(data: DomainManifestProps) => {
                domainName = data.domain;
                annotations = data.annotations;
                paths = data.paths;
                httpOnly = data.httpOnly;
                wwwRedirect = data.wwwRedirect;
                createElementFromScratch();
              }"
              @delete="
                () => {
                  deleteShow = true;
                }
              "
            />
          </div>
          <div
            v-if="selectedType === 'pod'"
            style="
              display: flex;
              flex-direction: column;
              justify-content: space-between;
              flex-grow: 1;
            "
          >
            <ElementAddFromScratchPod
              :element="props.element as Pod"
              :testElementName="testElementName"
              :selectedTab="selectedTab"
              :allowApplyResp="applyTooltipMessage"
              @confirm="(data: PodManifestProps) => {
                runCmd = data.runCmd;
                runAsUser = data.runAsUser;
                build = data.script;
                writableFileSystem = data.writableFileSystem;
                actions = data.actions;
                expositions = data.expositions;
                storages = data.storages;
                image = data.image;
                dockerfilePath = data.dockerfilePath;
                buildType = data.buildType;
                // scale = data.scale;
                createElementFromScratch();
              }"
              @delete="
                () => {
                  deleteShow = true;
                }
              "
            />
          </div>
        </div>
        <div
          v-if="selectedType === 'config'"
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
            @click="
              () => {
                deleteShow = true;
              }
            "
          >
            {{ t("actions.remove") }}
          </n-button>
          <n-button
            v-if="allowApplyBase"
            type="primary"
            secondary
            @click="createElementFromScratch"
          >
            {{ t("actions.apply") }}
          </n-button>
          <n-tooltip v-else>
            <template #trigger>
              <n-button type="primary" secondary disabled tag="div">
                {{ t("actions.apply") }}
              </n-button>
            </template>
            {{ applyTooltipMessage }}
          </n-tooltip>
        </div>
      </div>
    </div>
    <!-- <div v-if="selectedType" style="width: 20%; margin-left: 1em">
      <div style="display: flex; height: 100%">
        <n-divider
          vertical
          style="height: auto; margin-right: 1em"
          width="2px"
        />
        <div style="height: 100%">
          <h3>{{ tips[tipMode].title }}</h3>
          {{ tips[tipMode].tip }}
        </div>
      </div>
    </div> -->
  </div>
  <Modal
    v-model:show="openErrorModal"
    :title="t('objects.error')"
    style="width: 50em; height: 40em"
  >
    <Monaco :value="errorMessage" lang="toml" read-only lineNumbers="off" />
  </Modal>
  <Modal
    v-model:show="nameModal"
    :title="t('scratch.enterName')"
    style="width: 30rem"
    :showFooter="true"
    @positive-click="createClick"
  >
    <div class="element-adding-modal-row">
      <p style="width: 40%">{{ t("fields.elementName") }}</p>
      <n-input
        v-model:value="elementName"
        type="text"
        :placeholder="t('fields.elementName')"
        style="width: 60%"
        :status="testElementName ? undefined : 'error'"
        @keyup.enter="createClick"
      />
    </div>
    <div v-if="selectedType === 'pod'" class="element-adding-modal-row">
      <p style="width: 70%">{{ t("elements.options.addStopped") }}</p>
      <n-checkbox v-model:checked="addStopped" />
    </div>
  </Modal>
  <Modal
    v-model:show="deleteShow"
    :title="t('scratch.deleteElement')"
    style="width: 30rem"
    :showFooter="true"
    @positive-click="
      () => {
        deleteElement();
      }
    "
  >
    <p>{{ t("questions.sure") }}</p>
  </Modal>

  <Modal
    v-model:show="advancedVariableShow"
    :title="t('scratch.advanced')"
    style="width: 50rem"
    :showFooter="true"
    :touched="advancedTouched"
    @positive-click="
      () => {
        variables[advancedVariableIndex] = advancedVariable;
        advancedVariableShow = false;
      }
    "
  >
    <div class="element-from-scratch-row">
      <ElementAddFromScratchRowHeader
        :name="t('fields.description')"
        :tip="t('scratch.docs.variables.description')"
      />

      <div style="width: 80%; height: 100%; padding-right: 5em">
        <n-input
          v-model:value="advancedVariable.description"
          @input="advancedTouched = true"
          :placeholder="t('fields.description')"
        />
      </div>
    </div>
    <div class="element-from-scratch-row">
      <ElementAddFromScratchRowHeader
        :name="t('fields.secret')"
        :tip="t('scratch.docs.variables.secret')"
      />

      <div
        style="width: 80%; height: 100%; padding-right: 5em; text-align: center"
      >
        <n-button
          :dashed="advancedVariable.secret ? undefined : true"
          :type="advancedVariable.secret ? 'primary' : undefined"
          style="width: 100%"
          @click="
            () => {
              secretModal = true;
            }
          "
        >
          <template #icon>
            <n-icon>
              <mdi
                :path="
                  advancedVariable.secret
                    ? mdiLockOutline
                    : mdiLockOpenVariantOutline
                "
              />
            </n-icon>
          </template>
          {{
            advancedVariable.secret ? t("fields.secret") : t("fields.visible")
          }}
        </n-button>
      </div>
    </div>
    <div v-if="!advancedVariable.secret">
      <h3 style="text-align: center; margin: 1em 0">
        {{ t("fields.validation") }}
      </h3>
      <div class="element-from-scratch-row">
        <ElementAddFromScratchRowHeader
          :name="t('fields.type')"
          :tip="t('scratch.docs.variables.validation.type')"
        />

        <div style="width: 80%; height: 100%; padding-right: 5em">
          <n-select
            v-model:value="advancedVariable.validation.type"
            :options="validationFunctions"
          />
        </div>
      </div>
      <div
        v-if="advancedVariable.validation.type === 'int'"
        class="element-from-scratch-row"
      >
        <ElementAddFromScratchRowHeader
          name="Min"
          :tip="t('scratch.docs.variables.validation.min')"
        />

        <div style="width: 80%; height: 100%; padding-right: 5em">
          <n-input-number v-model:value="advancedVariable.validation.min" />
        </div>
      </div>
      <div
        v-if="advancedVariable.validation.type === 'int'"
        class="element-from-scratch-row"
      >
        <ElementAddFromScratchRowHeader
          name="Max"
          :tip="t('scratch.docs.variables.validation.max')"
        />

        <div style="width: 80%; height: 100%; padding-right: 5em">
          <n-input-number
            v-model:value="advancedVariable.validation.max"
            :min="advancedVariable.validation.min"
          />
        </div>
      </div>
      <div
        v-if="advancedVariable.validation.type === 'text'"
        class="element-from-scratch-row"
      >
        <ElementAddFromScratchRowHeader
          :name="t('scratch.allowSpaces')"
          :tip="t('scratch.docs.variables.validation.allowSpaces')"
        />

        <div
          style="
            width: 80%;
            height: 100%;
            padding-right: 5em;
            text-align: center;
          "
        >
          <n-checkbox
            v-model:checked="advancedVariable.validation.allowSpaces"
          />
        </div>
      </div>
      <div
        v-if="advancedVariable.validation.type === 'text'"
        class="element-from-scratch-row"
      >
        <ElementAddFromScratchRowHeader
          :name="t('scratch.allowSpecial')"
          :tip="t('scratch.docs.variables.validation.allowSpecial')"
        />

        <div
          style="
            width: 80%;
            height: 100%;
            padding-right: 5em;
            text-align: center;
          "
        >
          <n-checkbox
            v-model:checked="advancedVariable.validation.allowSpecial"
          />
        </div>
      </div>
      <div
        v-if="advancedVariable.validation.type === 'oneof'"
        class="element-from-scratch-row"
      >
        <ElementAddFromScratchRowHeader
          :name="t('fields.value', 2)"
          :tip="t('scratch.docs.variables.validation.oneOfValues')"
        />

        <div style="width: 80%; height: 100%; padding-right: 5em">
          <n-dynamic-input
            v-model:value="advancedVariable.validation.oneOfValues"
            :placeholder="t('fields.value', 2)"
          >
            <template #create-button-default>
              {{ t("actions.add") }}
            </template>
          </n-dynamic-input>
        </div>
      </div>
      <div
        v-if="advancedVariable.validation.type === 'regex'"
        class="element-from-scratch-row"
      >
        <ElementAddFromScratchRowHeader
          name="Regex"
          :tip="t('scratch.docs.variables.validation.allowSpecial')"
        />

        <div style="width: 80%; height: 100%; padding-right: 5em">
          <n-input
            v-model:value="advancedVariable.validation.regexString"
            placeholder="Regex"
          />
        </div>
      </div>
      <div
        v-if="advancedVariable.validation.type === 'password'"
        class="element-from-scratch-row"
      >
        <ElementAddFromScratchRowHeader
          :name="t('scratch.minLen')"
          :tip="t('scratch.docs.variables.validation.minLen')"
        />

        <div style="width: 80%; height: 100%; padding-right: 5em">
          <n-input-number
            v-model:value="advancedVariable.validation.minLen"
            :min="0"
          />
        </div>
      </div>
      <div
        v-if="advancedVariable.validation.type === 'password'"
        class="element-from-scratch-row"
      >
        <ElementAddFromScratchRowHeader
          :name="t('scratch.minLetters')"
          :tip="t('scratch.docs.variables.validation.minLetters')"
        />

        <div style="width: 80%; height: 100%; padding-right: 5em">
          <n-input-number
            v-model:value="advancedVariable.validation.minLetters"
            :min="0"
          />
        </div>
      </div>
      <div
        v-if="advancedVariable.validation.type === 'password'"
        class="element-from-scratch-row"
      >
        <ElementAddFromScratchRowHeader
          :name="t('scratch.upAndLow')"
          :tip="t('scratch.docs.variables.validation.upAndLow')"
        />

        <div
          style="
            width: 80%;
            height: 100%;
            padding-right: 5em;
            text-align: center;
          "
        >
          <n-switch v-model:value="advancedVariable.validation.upAndLow">
            <template #checked> {{ t("others.required") }} </template>
            <template #unchecked> {{ t("others.notRequired") }} </template>
          </n-switch>
        </div>
      </div>
      <div
        v-if="advancedVariable.validation.type === 'password'"
        class="element-from-scratch-row"
      >
        <ElementAddFromScratchRowHeader
          :name="t('scratch.minDigits')"
          :tip="t('scratch.docs.variables.validation.minDigits')"
        />

        <div style="width: 80%; height: 100%; padding-right: 5em">
          <n-input-number
            v-model:value="advancedVariable.validation.minDigits"
            :min="0"
          />
        </div>
      </div>
      <div
        v-if="advancedVariable.validation.type === 'password'"
        class="element-from-scratch-row"
      >
        <ElementAddFromScratchRowHeader
          :name="t('scratch.minSpecial')"
          :tip="t('scratch.docs.variables.validation.minSpecial')"
        />

        <div style="width: 80%; height: 100%; padding-right: 5em">
          <n-input-number
            v-model:value="advancedVariable.validation.minSpecial"
            :min="0"
          />
        </div>
      </div>
    </div>
  </Modal>

  <Modal
    v-model:show="secretModal"
    :title="t('fields.secret')"
    style="width: 30rem"
    :showFooter="true"
    @positive-click="
      () => {
        advancedVariable.secret = !advancedVariable.secret;
        if (!advancedVariable.secret) advancedVariable.value = '';
        secretModal = false;
      }
    "
  >
    <p>
      {{
        advancedVariable.secret
          ? t("questions.setVisible")
          : t("questions.setSecret")
      }}
    </p>
  </Modal>

  <Modal
    v-model:show="secretOverwriteModal"
    :title="t('fields.secret')"
    style="width: 30rem"
    :showFooter="true"
    @positive-click="
      () => {
        if (secretOverwriteValue.trim())
          variables[secretOverwriteIndex].value = secretOverwriteValue.trim();
        secretOverwriteModal = false;
      }
    "
  >
    <p style="margin-bottom: 0.5em">
      {{ t("questions.overwriteSecret") }}
    </p>
    <n-input
      v-model:value="secretOverwriteValue"
      :placeholder="t('fields.secret')"
    />
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

.element-from-scratch-row-header > p {
  margin-right: 0.5em;
}

.grid {
  display: flex;
  &:hover {
    background-color: rgba(gray, 0.05);
  }
}

.grid-left {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 8em;
  margin-left: 1em;
}

.grid-right {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 8em;
  margin-left: 2em;
}

.mode-button {
  width: 9em;
  height: 6em;
}

.icon {
  cursor: pointer;
  transition: all 0.2s ease-in-out;
  color: var(--n-color-target);
}

.expose-option {
  display: flex;
  align-items: center;
  width: 100%;
  margin-bottom: 0.5em;
}

.expose-option-header {
  width: 15%;
  display: flex;
  justify-content: flex-end;
  margin-right: 0.5em;
}

.element-adding-modal-row {
  display: flex;
  align-items: center;
  margin-bottom: 0.5em;
}

.secret-button:hover {
  background-color: var(--n-color) !important;
}

.secret-button:focus {
  background-color: var(--n-color) !important;
}
</style>
