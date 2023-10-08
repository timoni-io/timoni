<script setup lang="ts">
import { useI18n } from "vue-i18n";
import { ElementMapStatus, PodType } from "@/zodios/schemas/elements";
import { dynamicInputsHandle, pick } from "@/utils/dynamicInputsHandle";
import { findDuplicates } from "@/utils/findDuplicates";
import { z } from "zod";
import {
  Action,
  Exposition,
  Probe,
  // Scale,
  Storage,
} from "../ElementAddFromScratch.vue";

interface XD {
  command: string;
}

const { t } = useI18n();

const emit = defineEmits(["tipMode", "confirm", "delete"]);
const props = defineProps<{
  element?: {
    Info: z.infer<typeof PodType>;
    Status: z.infer<typeof ElementMapStatus>;
  };
  testElementName: boolean;
  allowApplyResp: string;
  selectedTab: string;
}>();

const onExpose = (): Exposition => {
  return {
    port: 0,
    type: "http",
    metricsPath: "",
    name: "",
    probe: {
      disable: false,
      path: "",
      "initial-delay-seconds": 0,
      "period-seconds": 0,
      "timeout-seconds": 0,
      "success-threshold": 0,
      "failure-threshold": 0,
      "restart-on-fail": false,
    },
  };
};

let advancedShow = $ref<boolean>(false);
let advancedMode = $ref<string>("");
let advancedIndex = $ref<number>(-1);
let advancedCommand = $ref<Array<XD>>([]);
let advancedExposition = $ref<Exposition>(onExpose());
let advancedTouched = $ref<boolean>(false);

let dockerfileExplorer = $ref<boolean>(false);

let build = $ref(
  props.element &&
    props.element.Info &&
    props.element.Info.Build &&
    "Script" in props.element.Info.Build
    ? props.element.Info.Build.Script
    : ""
);
let dockerfilePath = $ref(
  props.element &&
    props.element.Info &&
    props.element.Info.Build &&
    "DockerfilePath" in props.element.Info.Build
    ? props.element.Info.Build.DockerfilePath
    : ""
);
let dockerfilePathTemp = $ref("");
let image = $ref(
  props.element &&
    props.element.Info &&
    props.element.Info.Build &&
    "Image" in props.element.Info.Build
    ? props.element.Info.Build.Image
    : ""
);

let runCmd = $ref<Array<string>>(
  props.element && props.element.Info && "RunCommand" in props.element.Info
    ? (props.element.Info.RunCommand as string[])
    : []
);
let runAsUser = $ref(
  props.element &&
    props.element.Info &&
    "RunAsUser" in props.element.Info &&
    props.element.Info.RunAsUser &&
    props.element.Info.RunAsUser !== null
    ? props.element.Info.RunAsUser[0]
    : null
);
let writableFileSystem = $ref(
  props.element && props.element.Info && "RunWritableFS" in props.element.Info
    ? props.element.Info.RunWritableFS
    : false
);
let actions = $ref<Action[]>([]);
let buildType = $ref(
  props.element &&
    props.element.Info &&
    props.element.Info.Build &&
    "Type" in props.element.Info.Build
    ? props.element.Info.Build.Type
    : "script"
);

const onStorageCreate = () => {
  return {
    path: "",
    type: "",
    class: "",
    maxSizeMB: 0,
    readOnly: true,
    name: "",
    login: "",
    password: "",
    remoteHost: "",
    remotePath: "",
    options: "",
  };
};

let nfsStorages = $ref<Array<Storage>>([]);
let blockStorages = $ref<Array<Storage>>([]);
let sharedStorages = $ref<Array<Storage>>([]);
let tempStorages = $ref<Array<Storage>>([]);
let cifsStorages = $ref<Array<Storage>>([]);

const onActionCreate = (): Action => {
  return {
    key: "",
    command: [],
  };
};

const onXDCreate = (): XD => {
  return {
    command: "",
  };
};

let expositions = $ref<Array<Exposition>>([]);

const protocolTypes = $ref([
  {
    label: "http",
    value: "http",
  },
  {
    label: "https",
    value: "https",
  },
  {
    label: "tcp",
    value: "tcp",
  },
  {
    label: "udp",
    value: "udp",
  },
]);

const buildTypes = $computed(() => {
  return [
    ...[
      {
        label: t("scratch.buildTypes.script"),
        value: "script",
      },
      {
        label: t("scratch.buildTypes.image"),
        value: "image",
      },
    ],
    ...(props.element?.Info?.SourceGit.RepoName
      ? [
          {
            label: t("scratch.buildTypes.dockerfile"),
            value: "dockerfile",
          },
        ]
      : []),
  ];
});

const emptyNetworkValues = ["0"];

const exportObject = <T>(
  list: Array<T>,
  param: keyof T
): Record<keyof T, T> => {
  let resek: Record<keyof T, T> = {} as Record<keyof T, T>;
  list.forEach((el) => {
    resek[el[param] as keyof T] = el;
  });
  return resek;
};

const mergeStorages = (): Array<Storage> => {
  return [
    ...nfsStorages.map((el) => {
      el.type = "nfs";
      return el;
    }),
    ...blockStorages.map((el) => {
      el.type = "block";
      return el;
    }),
    ...sharedStorages.map((el) => {
      el.type = "shared";
      return el;
    }),
    ...tempStorages.map((el) => {
      el.type = "temp";
      return el;
    }),
    ...cifsStorages.map((el) => {
      el.type = "cifs";
      return el;
    }),
  ].filter((storage) => {
    return storage.path !== "";
  });
};

const actionsDuplicates = $computed<Array<string>>(() =>
  findDuplicates<string>(actions.map((action) => action.key))
);

const expositionsDuplicates = $computed<Array<number>>(() =>
  findDuplicates<number>(expositions.map((exposition) => exposition.port))
);

const storagesDuplicates = $computed<Array<string>>(() =>
  findDuplicates<string>(mergeStorages().map((storage) => storage.path))
);

const allowApply = $computed<boolean>(
  () =>
    actionsDuplicates.length +
      expositionsDuplicates.length +
      storagesDuplicates.length ===
      0 && !props.allowApplyResp
);

const applyTooltipMessage = (): string => {
  let sections = [];
  let resp = [];
  if (props.allowApplyResp.includes(t("scratch.removeDuplicates")))
    sections.push(t("scratch.tabs.variables"));
  if (actionsDuplicates.length) sections.push(t("scratch.tabs.actions"));
  if (expositionsDuplicates.length) sections.push(t("scratch.tabs.network"));
  if (storagesDuplicates.length) sections.push(t("scratch.tabs.storage"));
  if (sections.length)
    resp.push(
      t("scratch.removeDuplicates") +
        (t("scratch.removeDuplicates").includes("from") && sections.length > 1
          ? "s: "
          : ": ") +
        sections.join(", ")
    );
  // if (props.allowApplyResp.includes(t("scratch.emptySecret")))
  //   resp.push(t("scratch.emptySecret"));
  return resp.join(" ");
};

const storageFactory = (type: string): Array<Storage> => {
  switch (type) {
    case "temp":
      return tempStorages;
    case "block":
      return blockStorages;
    case "shared":
      return sharedStorages;
    case "nfs":
      return nfsStorages;
    case "cifs":
      return cifsStorages;
    default:
      return [];
  }
};

const commandPositionChange = <T>(
  index: number,
  arr: Array<T>,
  up: boolean
) => {
  const fromIndex = index;
  const toIndex = index + (up ? -1 : 1);
  const element = arr.splice(fromIndex, 1)[0];
  arr.splice(toIndex, 0, element);
  advancedTouched = true;
};

onBeforeMount(() => {
  const elInfo = props.element?.Info as z.infer<typeof PodType>;
  if (elInfo) {
    Object.keys(elInfo.Actions || {}).forEach((action) => {
      if (elInfo.Actions) {
        actions.push({
          key: action,
          command: elInfo.Actions[action],
        } as Action);
      }
    });

    Object.keys(elInfo.ExposePorts || {}).forEach((exposition) => {
      if (elInfo.ExposePorts) {
        expositions.push({
          port: parseInt(exposition),
          type: elInfo.ExposePorts[exposition].Type,
          metricsPath: elInfo.ExposePorts[exposition].MetricsPath,
          name: elInfo.ExposePorts[exposition].Name,
          probe: {
            disable: elInfo.ExposePorts[exposition].Probe.Disable,
          },
        } as Exposition);
      }
    });

    Object.keys(elInfo.Storage || {}).forEach((storage) => {
      if (elInfo.Storage) {
        storageFactory(elInfo.Storage[storage].Type).push({
          path: storage,
          type: "", // elInfo.Storage[storage].Type,
          maxSizeMB: elInfo.Storage[storage].MaxSizeMB,
          readOnly: elInfo.Storage[storage].ReadOnly,
          class: elInfo.Storage[storage].Class,
          name: elInfo.Storage[storage].Name,
          login: elInfo.Storage[storage].Login,
          password: elInfo.Storage[storage].Password,
          remoteHost: elInfo.Storage[storage].RemoteHost,
          remotePath: elInfo.Storage[storage].RemotePath,
          options: elInfo.Storage[storage].Options,
        } as Storage);
      }
    });

    // scale.disable = elInfo.Scale.Disable;
    // scale.min = elInfo.Scale.NrOfPodsMin;
    // scale.max = elInfo.Scale.NrOfPodsMax;
    // scale.targetCPU = elInfo.Scale.CPUTargetProc;
  }
  expositions.push(onExpose());
  actions.push(onActionCreate());
  ["nfs", "cifs", "temp", "block", "shared"].forEach((type) => {
    storageFactory(type).push(onStorageCreate());
  });
  advancedCommand.push(onXDCreate());
});

watch(
  $$(advancedCommand),
  () => {
    dynamicInputsHandle<XD>(
      advancedCommand,
      ["command"],
      [""],
      "advanced-command-dynamic",
      onXDCreate
    );
  },
  { deep: true }
);

watch(
  $$(actions),
  () => {
    dynamicInputsHandle<Action>(
      actions,
      ["key"],
      [""],
      "actions-dynamic",
      onActionCreate
    );
  },
  { deep: true }
);

watch(
  $$(expositions),
  () => {
    dynamicInputsHandle<Exposition>(
      expositions,
      ["port"],
      emptyNetworkValues,
      "network-dynamic",
      onExpose
    );
  },
  { deep: true }
);

watch(
  $$(blockStorages),
  () => {
    dynamicInputsHandle<Storage>(
      blockStorages,
      ["path", "maxSizeMB"],
      ["0"],
      "", // TODO
      onStorageCreate
    );
  },
  { deep: true }
);

watch(
  $$(tempStorages),
  () => {
    dynamicInputsHandle<Storage>(
      tempStorages,
      ["path", "maxSizeMB"],
      ["0"],
      "", // TODO
      onStorageCreate
    );
  },
  { deep: true }
);

watch(
  $$(nfsStorages),
  () => {
    dynamicInputsHandle<Storage>(
      nfsStorages,
      ["path", "readOnly", "remoteHost", "remotePath", "options"],
      ["true"],
      "", // TODO
      onStorageCreate
    );
  },
  { deep: true }
);

watch(
  $$(cifsStorages),
  () => {
    dynamicInputsHandle<Storage>(
      cifsStorages,
      ["path", "readOnly", "login", "password", "options"],
      ["true"],
      "", // TODO
      onStorageCreate
    );
  },
  { deep: true }
);

watch(
  $$(sharedStorages),
  () => {
    dynamicInputsHandle<Storage>(
      sharedStorages,
      ["path", "maxSizeMB", "name", "class"],
      ["0"],
      "", // TODO
      onStorageCreate
    );
  },
  { deep: true }
);
</script>

<template>
  <div>
    <div v-if="selectedTab === 'build'" class="element-from-scratch-row">
      <ElementAddFromScratchRowHeader :name="t('fields.type')" :tip="''" />
      <div style="width: 80%; padding-right: 5em">
        <n-select v-model:value="buildType" :options="buildTypes" />
      </div>
    </div>
    <div
      v-if="selectedTab === 'build'"
      style="display: flex; align-items: center; margin-bottom: 5px"
    >
      <ElementAddFromScratchRowHeader
        :name="buildTypes.find((bt) => bt.value === buildType)?.label || ''"
        :tip="''"
      />
      <div style="width: 80%; padding-right: 5em">
        <Monaco
          v-if="buildType === 'script'"
          style="height: 1px; min-height: 24em"
          v-model:value="build"
          lang="dockerfile"
        />
        <n-button
          v-if="buildType === 'dockerfile'"
          :dashed="!dockerfilePath.length"
          :type="dockerfilePath.length ? 'primary' : undefined"
          :secondary="!!dockerfilePath.length"
          style="width: 100%"
          @click="
            () => {
              dockerfileExplorer = true;
            }
          "
        >
          <n-icon size="14px" style="margin-right: 5px"
            ><mdi :path="mdiPencil"
          /></n-icon>
          {{
            dockerfilePath.length
              ? dockerfilePath.slice(0, 60)
              : t("scratch.configure")
          }}
        </n-button>
        <n-input
          v-if="buildType === 'image'"
          v-model:value="image"
          :placeholder="t('scratch.buildTypes.image')"
        />
      </div>
    </div>
    <div
      v-if="
        selectedTab === 'build' &&
        build &&
        buildType === 'dockerfile' &&
        props.element &&
        props.element.Info.Build.DockerfilePath
      "
      style="display: flex; align-items: center; margin-bottom: 5px"
    >
      <ElementAddFromScratchRowHeader :name="t('fields.preview')" :tip="''" />
      <div style="width: 80%; padding-right: 5em">
        <Monaco
          style="height: 1px; min-height: 24em"
          v-model:value="build"
          read-only
          lang="dockerfile"
        />
      </div>
    </div>
    <div v-if="selectedTab === 'general'" class="element-from-scratch-row">
      <ElementAddFromScratchRowHeader
        :name="t('scratch.runCmd')"
        :tip="t('scratch.docs.runCmd')"
      />
      <!-- <div style="width: 80%; padding-right: 5em">
        <n-input
          v-model:value="runCmd"
          type="text"
          :placeholder="t('scratch.runCmd')"
        />
      </div> -->
      <div style="width: 80%; padding-right: 5em">
        <n-button
          :dashed="!runCmd.length"
          :type="runCmd.length ? 'primary' : undefined"
          :secondary="!!runCmd.length"
          style="width: 100%"
          @click="
            () => {
              advancedMode = 'runCommand';
              advancedCommand = [];
              runCmd.forEach((el) => {
                advancedCommand.push({
                  command: el,
                });
              });
              advancedTouched = false;
              advancedShow = true;
            }
          "
        >
          <n-icon size="14px" style="margin-right: 5px"
            ><mdi :path="mdiPencil"
          /></n-icon>
          {{
            runCmd.length
              ? runCmd.join(" ").slice(0, 60)
              : t("scratch.configure")
          }}
        </n-button>
      </div>
    </div>
    <div v-if="selectedTab === 'general'" class="element-from-scratch-row">
      <ElementAddFromScratchRowHeader
        :name="t('scratch.runAsUser')"
        :tip="t('scratch.docs.runAsUser')"
      />
      <div style="width: 80%; padding-right: 5em">
        <n-input-number v-model:value="runAsUser" :min="0" />
      </div>
    </div>
    <div v-if="selectedTab === 'general'" class="element-from-scratch-row">
      <ElementAddFromScratchRowHeader
        :name="t('scratch.writableFileSystem')"
        :tip="t('scratch.docs.writableFileSystem')"
      />
      <div style="width: 80%">
        <n-checkbox v-model:checked="writableFileSystem" />
      </div>
    </div>
    <div v-if="selectedTab === 'actions'">
      <div
        id="actions-dynamic"
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
              <th style="width: 47%">
                <ElementAddFromScratchTableHeader
                  :name="t('scratch.actionName')"
                  :tip="t('scratch.docs.actions.name')"
                />
              </th>
              <th style="width: 47%">
                <ElementAddFromScratchTableHeader
                  :name="t('scratch.actionCmd')"
                  :tip="t('scratch.docs.actions.command')"
                />
              </th>
              <th style="width: 6%"></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(action, index) in actions" :key="index">
              <td>
                <n-input
                  v-model:value="action.key"
                  style="margin-right: 12px"
                  :status="
                    actionsDuplicates.includes(action.key)
                      ? 'warning'
                      : undefined
                  "
                  :placeholder="t('scratch.actionName')"
                />
              </td>
              <td>
                <!-- <n-input
                  v-model:value="action.value"
                  :placeholder="t('scratch.actionCmd')"
                /> -->
                <n-button
                  :dashed="!action.command.length"
                  :type="action.command.length ? 'primary' : undefined"
                  :secondary="!!action.command.length"
                  style="width: 100%"
                  @click="
                    () => {
                      advancedMode = 'action';
                      advancedIndex = index;
                      advancedCommand = [];
                      action.command.forEach((el) => {
                        advancedCommand.push({
                          command: el,
                        });
                      });
                      advancedTouched = false;
                      advancedShow = true;
                    }
                  "
                >
                  <n-icon size="14px" style="margin-right: 5px"
                    ><mdi :path="mdiPencil"
                  /></n-icon>
                  {{
                    action.command.length
                      ? action.command.join(" ").slice(0, 60)
                      : t("scratch.configure")
                  }}
                </n-button>
              </td>
              <td>
                <n-button
                  v-if="action.key || index < actions.length - 1"
                  type="primary"
                  quaternary
                  circle
                  style="margin-left: 5px"
                  @click="
                    () => {
                      actions.splice(index, 1);
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
    <div
      id="network-dynamic"
      v-if="selectedTab === 'network'"
      style="padding: 0 5em 0 5em; overflow-y: scroll; max-height: 27em"
    >
      <n-table class="dynamic-table" :bordered="false" :single-line="true">
        <thead>
          <tr>
            <th style="width: 10%">
              <ElementAddFromScratchTableHeader
                name="Port"
                :tip="t('scratch.docs.network.port')"
              />
            </th>
            <th style="width: 10%">
              <ElementAddFromScratchTableHeader
                :name="t('fields.type')"
                :tip="t('scratch.docs.network.type')"
              />
            </th>
            <!-- <th style="width: 30%">{{ t("fields.name") }}</th>
            <th style="width: 29%">{{ t("scratch.metricsPath") }}</th>
            <th style="width: 5%">Probe</th> -->
            <th style="width: 64%">{{ t("scratch.advanced") }}</th>
            <th style="width: 6%"></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(exposition, index) in expositions" :key="index">
            <td>
              <n-input-number
                v-model:value="exposition.port"
                :min="0"
                :max="65535"
                :status="
                  expositionsDuplicates.includes(exposition.port)
                    ? 'warning'
                    : undefined
                "
              />
            </td>
            <td>
              <n-select
                v-model:value="exposition.type"
                :options="protocolTypes"
              />
            </td>
            <!-- <td><n-input v-model:value="exposition.name" /></td>
            <td><n-input v-model:value="exposition.metricsPath" /></td>
            <td><n-checkbox v-model:checked="exposition.advanced.disable" /></td> -->
            <td>
              <n-button
                dashed
                style="width: 100%"
                @click="
                  () => {
                    advancedMode = 'network';
                    advancedExposition = JSON.parse(JSON.stringify(exposition));
                    advancedIndex = index;
                    advancedTouched = false;
                    advancedShow = true;
                  }
                "
              >
                {{ t("scratch.advanced") }}
              </n-button>
            </td>
            <td>
              <n-button
                v-if="
                  !emptyNetworkValues.includes(
                    Object.values(
                      pick(exposition, ['port', 'metricsPath', 'name'])
                    ).join('')
                  ) || index < expositions.length - 1
                "
                type="primary"
                quaternary
                circle
                style="margin-left: 5px"
                @click="
                  () => {
                    expositions.splice(index, 1);
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
      v-if="selectedTab === 'storage'"
      style="padding: 0 5em 0 5em; overflow-y: scroll; max-height: 30em"
    >
      <n-table
        v-for="storage in [
          { name: 'temp', list: tempStorages },
          { name: 'block', list: blockStorages },
        ]"
        :key="storage.name"
        class="storage-table"
        :bordered="false"
        :single-line="true"
      >
        <thead>
          <tr>
            <th style="width: 47%">
              <ElementAddFromScratchTableHeader
                :name="t('scratch.path') + ': ' + storage.name"
                :tip="t('scratch.docs.storage.type.' + storage.name)"
              />
            </th>
            <th style="width: 47%">
              <ElementAddFromScratchTableHeader
                :name="t('scratch.maxSizeMB')"
                :tip="t('scratch.docs.storage.maxSizeMB')"
              />
            </th>
            <th style="width: 6%"></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(block, index) in storage.list" :key="index">
            <td>
              <n-input
                v-model:value="block.path"
                :status="
                  storagesDuplicates.includes(block.path)
                    ? 'warning'
                    : undefined
                "
              />
            </td>
            <td>
              <n-input-number v-model:value="block.maxSizeMB" :min="0" />
            </td>
            <td>
              <n-button
                v-if="
                  Object.values(pick(block, ['path', 'maxSizeMB'])).join('') !==
                    '0' || index < storage.list.length - 1
                "
                type="primary"
                quaternary
                circle
                style="margin-left: 5px"
                @click="
                  () => {
                    storage.list.splice(index, 1);
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

      <n-table class="storage-table" :bordered="false" :single-line="true">
        <thead>
          <tr>
            <th style="width: 20%">
              <ElementAddFromScratchTableHeader
                :name="t('scratch.path') + ': nfs'"
                :tip="t('scratch.docs.storage.type.nfs')"
              />
            </th>
            <th style="width: 6%">
              <ElementAddFromScratchTableHeader
                :name="t('scratch.readOnly')"
                :tip="t('scratch.docs.storage.readOnly')"
              />
            </th>
            <th style="width: 24%">
              <ElementAddFromScratchTableHeader
                :name="t('scratch.remoteHost')"
                :tip="t('scratch.docs.storage.remoteHost')"
              />
            </th>
            <th style="width: 24%">
              <ElementAddFromScratchTableHeader
                :name="t('scratch.remotePath')"
                :tip="t('scratch.docs.storage.remotePath')"
              />
            </th>
            <th style="width: 20%">
              <ElementAddFromScratchTableHeader
                :name="t('scratch.options')"
                :tip="t('scratch.docs.storage.nfsoptions')"
              />
            </th>
            <th style="width: 6%"></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(nfs, index) in nfsStorages" :key="index">
            <td>
              <n-input
                v-model:value="nfs.path"
                :status="
                  storagesDuplicates.includes(nfs.path) ? 'warning' : undefined
                "
              />
            </td>
            <td style="text-align: center">
              <n-checkbox v-model:checked="nfs.readOnly" />
            </td>
            <td><n-input v-model:value="nfs.remoteHost" /></td>
            <td><n-input v-model:value="nfs.remotePath" /></td>
            <td><n-input v-model:value="nfs.options" /></td>
            <td>
              <n-button
                v-if="
                  Object.values(
                    pick(nfs, [
                      'path',
                      'readOnly',
                      'remoteHost',
                      'remotePath',
                      'options',
                    ])
                  ).join('') !== 'true' || index < nfsStorages.length - 1
                "
                type="primary"
                quaternary
                circle
                style="margin-left: 5px"
                @click="
                  () => {
                    nfsStorages.splice(index, 1);
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

      <n-table class="storage-table" :bordered="false" :single-line="true">
        <thead>
          <tr>
            <th style="width: 20%">
              <ElementAddFromScratchTableHeader
                :name="t('scratch.path') + ': cifs'"
                :tip="t('scratch.docs.storage.type.cifs')"
              />
            </th>
            <th style="width: 6%">
              <ElementAddFromScratchTableHeader
                :name="t('scratch.readOnly')"
                :tip="t('scratch.docs.storage.readOnly')"
              />
            </th>
            <th style="width: 24%">
              <ElementAddFromScratchTableHeader
                name="Login"
                :tip="t('scratch.docs.storage.login')"
              />
            </th>
            <th style="width: 24%">
              <ElementAddFromScratchTableHeader
                :name="t('scratch.password')"
                :tip="t('scratch.docs.storage.password')"
              />
            </th>
            <th style="width: 20%">
              <ElementAddFromScratchTableHeader
                :name="t('scratch.options')"
                :tip="t('scratch.docs.storage.cifsoptions')"
              />
            </th>
            <th style="width: 6%"></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(cifs, index) in cifsStorages" :key="index">
            <td>
              <n-input
                v-model:value="cifs.path"
                :status="
                  storagesDuplicates.includes(cifs.path) ? 'warning' : undefined
                "
              />
            </td>
            <td style="text-align: center">
              <n-checkbox v-model:checked="cifs.readOnly" />
            </td>
            <td><n-input v-model:value="cifs.login" /></td>
            <td><n-input v-model:value="cifs.password" /></td>
            <td><n-input v-model:value="cifs.options" /></td>
            <td>
              <n-button
                v-if="
                  Object.values(
                    pick(cifs, [
                      'path',
                      'readOnly',
                      'login',
                      'password',
                      'options',
                    ])
                  ).join('') !== 'true' || index < cifsStorages.length - 1
                "
                type="primary"
                quaternary
                circle
                style="margin-left: 5px"
                @click="
                  () => {
                    cifsStorages.splice(index, 1);
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

      <n-table class="storage-table" :bordered="false" :single-line="true">
        <thead>
          <tr>
            <th style="width: 25%">
              <ElementAddFromScratchTableHeader
                :name="t('scratch.path') + ': shared'"
                :tip="t('scratch.docs.storage.type.shared')"
              />
            </th>
            <th style="width: 19%">
              <ElementAddFromScratchTableHeader
                :name="t('scratch.maxSizeMB')"
                :tip="t('scratch.docs.storage.maxSizeMB')"
              />
            </th>
            <th style="width: 25%">
              <ElementAddFromScratchTableHeader
                :name="t('fields.name')"
                :tip="t('scratch.docs.storage.name')"
              />
            </th>
            <th style="width: 25%">
              <ElementAddFromScratchTableHeader
                :name="t('scratch.class')"
                :tip="t('scratch.docs.storage.class')"
              />
            </th>
            <th style="width: 6%"></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(shared, index) in sharedStorages" :key="index">
            <td>
              <n-input
                v-model:value="shared.path"
                :status="
                  storagesDuplicates.includes(shared.path)
                    ? 'warning'
                    : undefined
                "
              />
            </td>
            <td>
              <n-input-number v-model:value="shared.maxSizeMB" :min="0" />
            </td>
            <td><n-input v-model:value="shared.name" /></td>
            <td><n-input v-model:value="shared.class" /></td>
            <td>
              <n-button
                v-if="
                  Object.values(
                    pick(shared, ['path', 'maxSizeMB', 'name', 'class'])
                  ).join('') !== '0' || index < sharedStorages.length - 1
                "
                type="primary"
                quaternary
                circle
                style="margin-left: 5px"
                @click="
                  () => {
                    sharedStorages.splice(index, 1);
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

    <!-- <div v-if="selectedTab === 'scaling'">
      <div
        style="
          display: flex;
          flex-direction: column;
          width: 100%;
          padding-right: 5em;
        "
      >
        <div class="element-from-scratch-row">
          <p class="element-from-scratch-row-header">
            {{ t("scratch.scale.min") }}
          </p>
          <n-input-number v-model:value="scale.min" style="width: 85%" />
        </div>
        <div class="element-from-scratch-row">
          <p class="element-from-scratch-row-header">
            {{ t("scratch.scale.max") }}
          </p>
          <n-input-number v-model:value="scale.max" style="width: 85%" />
        </div>
        <div class="element-from-scratch-row">
          <p class="element-from-scratch-row-header">
            {{ t("scratch.scale.targetCPU") }}
          </p>
          <n-input-number v-model:value="scale.targetCPU" style="width: 85%" />
        </div>
        <div class="element-from-scratch-row">
          <p class="element-from-scratch-row-header">
            {{ t("scratch.scale.disable") }}
          </p>
          <div style="width: 85%">
            <n-checkbox v-model:checked="scale.disable" />
          </div>
        </div>
      </div>
    </div> -->
  </div>
  <div
    style="display: flex; margin-top: 5px"
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
          runCmd: runCmd.filter((el) => el),
          runAsUser: runAsUser,
          script: build,
          writableFileSystem: writableFileSystem,
          actions: actions,
          expositions: exportObject<Exposition>(
            expositions.filter((exp) => exp.port !== 0),
            'port'
          ),
          storages: exportObject<Storage>(mergeStorages(), 'path'),
          image: image,
          buildType: buildType,
          dockerfilePath: dockerfilePath,
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
    v-model:show="advancedShow"
    :title="t('scratch.advanced')"
    :touched="advancedTouched"
    style="width: 80rem; height: 40em"
    :show-footer="true"
    @positive-click="
      () => {
        if (advancedMode === 'network')
          expositions[advancedIndex] = JSON.parse(
            JSON.stringify(advancedExposition)
          );
        if (advancedMode === 'runCommand')
          runCmd = advancedCommand.map((el) => el.command);
        if (advancedMode === 'action')
          actions[advancedIndex].command = advancedCommand.map(
            (el) => el.command
          );
        advancedShow = false;
      }
    "
  >
    <div v-if="advancedMode === 'runCommand' || advancedMode === 'action'">
      <div
        id="advanced-command-dynamic"
        style="
          width: 100%;
          padding: 0 5em 0 5em;
          overflow-y: scroll;
          max-height: 22em;
        "
      >
        <n-table class="dynamic-table" :bordered="false" :single-line="true">
          <thead>
            <tr>
              <th style="width: 3%"></th>
              <th style="width: 51%">{{ t("scratch.actionCmd") }}</th>
              <th style="width: 6%"></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(command, index) in advancedCommand" :key="index">
              <td style="text-align: center">
                {{ index }}
              </td>
              <td>
                <n-input
                  v-model:value="command.command"
                  @input="advancedTouched = true"
                  :placeholder="t('scratch.actionCmd')"
                />
              </td>
              <td>
                <div style="display: flex; width: 5em">
                  <n-button
                    v-if="advancedCommand.length > 1"
                    type="primary"
                    quaternary
                    circle
                    style="margin-left: 5px"
                    :disabled="index === 0"
                    @click="commandPositionChange(index, advancedCommand, true)"
                  >
                    <template #icon>
                      <n-icon>
                        <mdi :path="mdiArrowUpThin" />
                      </n-icon>
                    </template>
                  </n-button>
                  <n-button
                    v-if="advancedCommand.length > 1"
                    type="primary"
                    quaternary
                    circle
                    style="margin-left: 5px"
                    :disabled="index === advancedCommand.length - 1"
                    @click="
                      commandPositionChange(index, advancedCommand, false)
                    "
                  >
                    <template #icon>
                      <n-icon>
                        <mdi :path="mdiArrowDownThin" />
                      </n-icon>
                    </template>
                  </n-button>
                  <n-button
                    v-if="command.command || index < advancedCommand.length - 1"
                    type="primary"
                    quaternary
                    circle
                    style="margin-left: 5px"
                    @click="
                      () => {
                        advancedCommand.splice(index, 1);
                      }
                    "
                  >
                    <template #icon>
                      <n-icon>
                        <mdi :path="mdiWindowClose" />
                      </n-icon>
                    </template>
                  </n-button>
                </div>
              </td>
            </tr>
          </tbody>
        </n-table>
      </div>
    </div>
    <div v-if="advancedMode === 'network'">
      <div class="element-from-scratch-row">
        <ElementAddFromScratchRowHeader
          :name="t('fields.name')"
          :tip="t('scratch.docs.expositionName')"
        />

        <div style="width: 80%; height: 100%; padding-right: 5em">
          <n-input
            v-model:value="advancedExposition.name"
            @input="advancedTouched = true"
            :placeholder="t('fields.name')"
          />
        </div>
      </div>
      <div class="element-from-scratch-row">
        <ElementAddFromScratchRowHeader
          :name="t('scratch.metricsPath')"
          :tip="t('scratch.docs.metricsPath')"
        />

        <div style="width: 80%; height: 100%; padding-right: 5em">
          <n-input
            v-model:value="advancedExposition.metricsPath"
            @input="advancedTouched = true"
            :placeholder="t('scratch.metricsPath')"
          />
        </div>
      </div>
      <h3 style="display: flex; justify-content: center; margin: 1em 0">
        Probe
      </h3>
      <div class="element-from-scratch-row">
        <ElementAddFromScratchRowHeader
          :name="t('scratch.probe.disable')"
          :tip="t('scratch.docs.advanced.disable')"
        />

        <div style="width: 80%; height: 100%; padding-right: 5em">
          <n-checkbox
            v-model:checked="advancedExposition.probe.disable"
            @update:checked="advancedTouched = true"
          />
        </div>
      </div>
      <div class="element-from-scratch-row">
        <ElementAddFromScratchRowHeader
          :name="t('scratch.probe.path')"
          :tip="t('scratch.docs.advanced.path')"
        />

        <div style="width: 80%; height: 100%; padding-right: 5em">
          <n-input
            v-model:value="advancedExposition.probe.path"
            @input="advancedTouched = true"
            :placeholder="t('scratch.probe.path')"
          />
        </div>
      </div>
      <div
        v-for="field in [
          'initial-delay-seconds',
          'period-seconds',
          'timeout-seconds',
          'success-threshold',
          'failure-threshold',
        ]"
        :key="field"
        class="element-from-scratch-row"
      >
        <ElementAddFromScratchRowHeader
          :name="t('scratch.probe.' + field)"
          :tip="t('scratch.docs.advanced.' + field)"
        />

        <div style="width: 80%; height: 100%; padding-right: 5em">
          <n-input-number
            v-model:value="advancedExposition.probe[field as keyof Omit<Probe, 'disable' | 'path' | 'restart-on-fail'>]"
            @input="advancedTouched = true"
            :placeholder="t('scratch.probe.' + field)"
            :min="0"
          />
        </div>
      </div>
      <div class="element-from-scratch-row">
        <ElementAddFromScratchRowHeader
          :name="t('scratch.probe.restart-on-fail')"
          :tip="t('scratch.docs.advanced.restart-on-fail')"
        />

        <div style="width: 80%; height: 100%; padding-right: 5em">
          <n-checkbox
            v-model:checked="advancedExposition.probe['restart-on-fail']"
            @update:checked="advancedTouched = true"
            :placeholder="t('scratch.probe.restart-on-fail')"
          />
        </div>
      </div>
    </div>
  </Modal>

  <Modal
    v-model:show="dockerfileExplorer"
    :title="t('fields.findDockerfile')"
    :touched="false"
    style="width: 80rem; height: 60em"
    :show-footer="true"
    @positive-click="
      () => {
        dockerfilePath = dockerfilePathTemp;
        dockerfileExplorer = false;
      }
    "
  >
    <RepoFiles
      :repo="props.element?.Info.SourceGit.RepoName"
      :branch="props.element?.Info.SourceGit.BranchName"
      :dockerfile="[]"
      @exportDocker="(rfPath: string) => {
          dockerfilePathTemp = rfPath;
        }"
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

// .storage-table {
//   margin-bottom: 0.5em;
//   background-color: grey;
// }

// .storage-table tr {
//   background-color: grey;
// }

// .storage-table td {
//   background-color: grey;
// }
</style>
