<script setup lang="ts">
import { teamListObject } from "@/zodios/schemas/user";
import { z } from "zod";
import { useI18n } from "vue-i18n";
import { useMessage } from "naive-ui";

type Team = z.infer<typeof teamListObject>;

const props = defineProps<{
    selectedTeam: Team;
    permissionModal: boolean;
}>();

const emit = defineEmits<{
  (e: "close-modal"): void;
}>();

interface Permission {
  [key: string]: {
    Index: number;
    IsSet: boolean;
  };
}

interface Permissions {
  Global: Permission;
  Envs: Permission;
  GitRepos: Permission;
}

let selectedPermissions = $ref<Permissions>({
  Global: {
    Glob_AccessToAdminZone: {
      Index: 1,
      IsSet: false
    },
    Glob_AccessToKube: {
      Index: 2,
      IsSet: false
    },
    Glob_AccessToWebUI: {
      Index: 0,
      IsSet: false
    },
    Glob_CreateAndDeleteEnvs: {
      Index: 4,
      IsSet: false
    },
    Glob_CreateAndDeleteGitRepos: {
      Index: 5,
      IsSet: false
    },
    Glob_ManageGlobalMemebers: {
      Index: 3,
      IsSet: false
    }
  },
  Envs: {
    Env_CopyAndViewSecrets: {
      Index: 17,
      IsSet: false
    },
    Env_ElementFullManage: {
      Index: 9,
      IsSet: false
    },
    Env_ElementStartStopRestart: {
      Index: 7,
      IsSet: false
    },
    Env_ElementTerminal: {
      Index: 8,
      IsSet: false
    },
    Env_ElementVersionChangeOnly: {
      Index: 6,
      IsSet: false
    },
    Env_ManageCluster: {
      Index: 3,
      IsSet: false
    },
    Env_ManageGitOPS: {
      Index: 4,
      IsSet: false
    },
    Env_ManageMembers: {
      Index: 14,
      IsSet: false
    },
    Env_ManageMetrics: {
      Index: 16,
      IsSet: false
    },
    Env_ManageSchedule: {
      Index: 2,
      IsSet: false
    },
    Env_ManageTags: {
      Index: 5,
      IsSet: false
    },
    Env_Rename: {
      Index: 1,
      IsSet: false
    },
    Env_View: {
      Index: 0,
      IsSet: false
    },
    Env_ViewLogsBuild: {
      Index: 11,
      IsSet: false
    },
    Env_ViewLogsEvents: {
      Index: 10,
      IsSet: false
    },
    Env_ViewLogsRuntime: {
      Index: 12,
      IsSet: false
    },
    Env_ViewMembers: {
      Index: 13,
      IsSet: false
    },
    Env_ViewMetrics: {
      Index: 15,
      IsSet: false
    },
  },
  GitRepos: {
    Repo_LocalManage: {
      Index: 4,
      IsSet: false
    },
    Repo_Pull: {
      Index: 1,
      IsSet: false
    },
    Repo_Push: {
      Index: 2,
      IsSet: false
    },
    Repo_RemoteManage: {
      Index: 3,
      IsSet: false
    },
    Repo_SettingsManage: {
      Index: 5,
      IsSet: false
    },
    Repo_View: {
      Index: 0,
      IsSet: false
    }
  }
})

const message = useMessage();
const { t } = useI18n();
let permissionModal = $ref(props.permissionModal);
let newEnvs = $ref<Array<string>>([]);
let newRepos = $ref<Array<string>>([]);
let envs = $ref<Record<string, string>>({});
let repos = $ref<Array<string>>([]);

const updatePermissions = () => {
  let globalPermissions: Record<number, boolean> = {};
  let envPermissions: Record<number, boolean> = {};
  let repoPermissions: Record<number, boolean> = {};

  for(const globalPermission of Object.values(selectedPermissions.Global)) {
    globalPermissions[globalPermission.Index] = globalPermission.IsSet;
  }

  for(const envPermission of Object.values(selectedPermissions.Envs)) {
    envPermissions[envPermission.Index] = envPermission.IsSet;
  }

  for(const repoPermission of Object.values(selectedPermissions.GitRepos)) {
    repoPermissions[repoPermission.Index] = repoPermission.IsSet;
  }

  api
    .post('/team-perms-set', {
      TeamID: props.selectedTeam?.ID as string,
      Global: globalPermissions,
      Env: {
        [newEnvs.map((env) => 'id:' + env).join(';')]: envPermissions
      },
      GitRepo: {
        [newRepos.map((repo) => 'id:' + repo).join(';')]: repoPermissions
      },
    })
    .then((res) => {
      if(res === 'ok') {
        message.success(t('messages.permissionUpdated'));
        emit('close-modal');
      } else {
        message.error(res);
      }
    })
}

const getEnvMap = () => {
  api.get("/env-map").then((res) => {
    for(const env of Object.values(res)) {
      envs[env.ID] = env.Name;
    }

    envs['*'] = '*';
  });
};

const getRepoMap = () => {
  api.get("/git-repo-map").then((res) => {
    repos = Object.keys(res);
    repos.push('*');
  });
};

const getTeamInfo = () => {
  api
    .get('/team-info', {
      queries: {
        teamID: props.selectedTeam.ID as string,
      },
    }).then((res) => {
      Object.keys(res.Perm.Global).forEach((key) => {
        selectedPermissions.Global[key] = res.Perm.Global[key];
      })

      const envPerm = Object.keys(res.Perm.Envs)[0];
      const repoPerm = Object.keys(res.Perm.GitRepos)[0];

      newEnvs = envPerm ? envPerm.replaceAll('id:', '').split(';') : [];
      newRepos = repoPerm ? repoPerm.replaceAll('id:', '').split(';') : [];

      if(envPerm) {
        Object.keys(res.Perm.Envs[envPerm]).forEach((key) => {
          selectedPermissions.Envs[key] = res.Perm.Envs[envPerm][key];
        })
      }

      if(repoPerm) {
        Object.keys(res.Perm.GitRepos[repoPerm]).forEach((key) => {
          selectedPermissions.GitRepos[key] = res.Perm.GitRepos[repoPerm][key];
        })
      }
    })
}

onMounted(() => {
  getEnvMap();
  getRepoMap();
  getTeamInfo();
});

watch(
  () => permissionModal,
  (newPermissionModal) => {
    if(!newPermissionModal) {
      emit("close-modal");
    }
  }
);

</script>

<template>
    <Modal
        v-model:show="permissionModal"
        :title="t('fields.permission', 2) + ' - ' + selectedTeam?.Name"
        style="max-width: 60rem; max-height: 85vh; overflow-y: auto;"
        :showFooter="true"
        @positive-click="updatePermissions"
        @negative-click="permissionModal = false"
    >
        <div
            style="display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 2rem"
        >
            <div>
                <h2 style="padding-bottom: 44px; font-size: 1.2rem">
                    Global
                </h2>
                <div v-for="key in Object.keys(selectedPermissions.Global)" :key="key">
                    <div class="permission-input" v-if="key !== 'Glob_AllGl'">
                      <div style="padding: 3px;">
                        <n-switch
                          v-model:value="selectedPermissions.Global[key].IsSet"
                          :round="false"
                          style="width: 70px !important;"
                          >
                          <template #checked> True </template>
                          <template #unchecked> False </template>
                        </n-switch>
                      </div>
                      <div>
                        <p style="margin: 5px 0;">
                            {{ t("permissions." + key) }}
                        </p>
                      </div>
                    </div>
                </div>
            </div>
            <div>
                <h2 style="padding-bottom: 0.4rem; font-size: 1.2rem">
                    Envs
                </h2>
                <n-select
                    v-model:value="newEnvs"
                    multiple
                    :options="Object.keys(envs).map((key) => ({
                    label: envs[key],
                    value: key,
                    }))"
                    style="margin-bottom: 10px;"
                    :placeholder="t('actions.select') + ' ' + t('objects.environment', 2).toLowerCase()"
                />
                <div>
                  <div v-for="key in Object.keys(selectedPermissions.Envs)" :key="key">
                      <div class="permission-input" v-if="key !== 'Glob_AllGl' && newEnvs.length > 0">
                        <div style="padding: 3px;">
                          <n-switch
                              v-model:value="selectedPermissions.Envs[key].IsSet"
                              :round="false"
                          >
                              <template #checked> True </template>
                              <template #unchecked> False </template>
                          </n-switch>
                        </div>
                        <p style="margin: 5px 0;">
                          {{ t("permissions." + key) }}
                        </p>
                      </div>
                  </div>
                </div>
            </div>
            <div>
                <h2 style="padding-bottom: 0.4rem; font-size: 1.2rem">
                    GitRepos
                </h2>
                <n-select
                    v-model:value="newRepos"
                    multiple
                    :options="repos.map((key) => ({
                    label: key,
                    value: key,
                    }))"
                    style="margin-bottom: 10px;"
                    :placeholder="t('actions.select') + ' ' + t('objects.repo', 2).toLowerCase()"
                />
                <div v-for="key in Object.keys(selectedPermissions.GitRepos)" :key="key">
                    <div class="permission-input" v-if="key !== 'Glob_AllGl' && newRepos.length > 0">
                      <div>
                        <div style="padding: 3px;">
                          <n-switch
                              v-model:value="selectedPermissions.GitRepos[key].IsSet"
                              :round="false"
                              style="width: 70px;"
                          >
                              <template #checked> True </template>
                              <template #unchecked> False </template>
                          </n-switch>
                        </div>
                      </div>
                      <p style="margin: 5px 0;">
                          {{ t("permissions." + key) }}
                      </p>
                    </div>
                </div>
            </div>
        </div>
    </Modal>
</template>

<style scoped lang="scss">
.permission-input {
  display: flex;
  gap: 5px;
  align-items: center;
}
</style>