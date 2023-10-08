<script setup lang="ts">
import { computed, ComputedRef } from "vue";
import { Column } from "@/components/utils/DataTable.vue";
import { z } from "zod";
import { useI18n } from "vue-i18n";
import { useMessage } from "naive-ui";
import { teamListObject, User } from "@/zodios/schemas/user";
import { useUserStore } from "@/store/userStore";

type User = z.infer<typeof User>;
type Team = z.infer<typeof teamListObject>;

const userStore = useUserStore();
const { t } = useI18n();
const message = useMessage();
let deleteConfirmationModal = $ref(false);
let editMembersModal = $ref(false);
let permissionModal = $ref(false);
let teamList = $ref<Team[]>();
let users = $ref<Record<string, string>>({});
let envs = $ref<Record<string, string>>({});
let selectedTeam = $ref<Team>();
let selectedTeamMembers = $ref<Record<string, boolean>>({});
let newTeamName = $ref<string>("");

const columns: ComputedRef<Column[]> = computed(() => [
  {
    title: t("fields.name"),
    key: "Name",
    template: "name",
    width: "20%",
  },
  {
    title: t("objects.member", 2),
    key: "Members",
    template: "members",
    width: "15%",
  },
  {
    title: t("objects.environment", 2),
    key: "Environment",
    template: "environment",
    width: "15%",
  },
  {
    title: t("objects.gitRepo", 2),
    key: "Repos",
    template: "repos",
    width: "15%",
  },
  {
    title: "",
    template: "actions",
    width: "35%",
  },
]);

const getUserList = () => {
  api.get("/user-list").then((res) => {
    for(const user of Object.values(res)) {
      users[user.ID] = user.Name;
    }
  });
};

const getTeamList = () => {
  api.get("/team-list").then((res) => {
    teamList = res.sort((a, b) => a.Name.localeCompare(b.Name));
  });
};

const getEnvMap = () => {
  api.get("/env-map").then((res) => {
    for(const env of Object.values(res)) {
      envs[env.ID] = env.Name;
    }
  });
};

onMounted(() => {
  getUserList();
  getTeamList();
  getEnvMap();
});

let newTeamModal = $ref(false);

const createNewTeam = () => {
  if (newTeamName)
    api
      .post("/team-create", {
        Name: newTeamName,
      })
      .then((res) => {
        if (res === "ok") {
          getTeamList();
          message.success(t("messages.teamCreated"));
          newTeamModal = !newTeamModal;
        }
      });
};

const deleteTeam = () => {
  api
    .post("/team-delete", {
      ID: selectedTeam?.ID as string,
    })
    .then((res) => {
      if (res === "ok") {
        getTeamList();
        message.success(t("messages.teamRemoved"));
        deleteConfirmationModal = false;
      }
    });
};

let renameTeamModal = $ref(false);
let renameInput = $ref<HTMLInputElement>();
watch($$(renameInput), () => renameInput?.focus());

const renameTeam = () => {
  api
    .post("/team-update", {
      ID: selectedTeam?.ID as string,
      Name: newTeamName,
    })
    .then((res) => {
      if (res === "ok") {
        message.success(t("messages.teamNameRenamed"));
        renameTeamModal = !renameTeamModal;
        getTeamList();
        return;
      }
      message.error(res);
    });
};

const handleEditMembers = (row: Team) => {
  
  api
    .get('/team-info', {
      queries: {
        teamID: row.ID as string,
      },
    }).then((res) => {
      selectedTeam = row;
      selectedTeamMembers = {};

      for(const member of Object.values(res.Members) as User[]) {
        selectedTeamMembers[member.ID] = true;
      }

      editMembersModal = true;
    })
}

const handleCheckedChange = (checked: boolean, key: string) => {
  if(checked) {
    api
      .post('/team-user-add', {
        TeamID: selectedTeam?.ID as string,
        UserID: key,
      })
      .then((res) => {
        if(res === 'ok') {
          message.success(t('messages.teamUserAdded'));
          getTeamList();
        } else {
          message.error(res);
        }
      })
  } else {
    api
      .post('/team-user-remove', {
        TeamID: selectedTeam?.ID as string,
        UserID: key,
      })
      .then((res) => {
        if(res === 'ok') {
          message.success(t('messages.teamUserRemoved'));
          getTeamList();
        } else {
          message.error(res);
        }
      })
  }
}

const handleClose = () => {
  permissionModal = false;
  selectedTeam = undefined;
}

</script>

<template>
  <PageLayout>
    <div
      v-if="userStore.havePermission('Glob_ManageGlobalMemebers')"
    >
      <n-card
        size="small"
        :title="t('fields.team', 2)"
        style="min-height: 90vh; position: relative"
      >
        <template #header-extra>
          <PopModal
            :title="t('fields.addTeam')"
            @positive-click="createNewTeam"
            @negative-click="() => {}"
            :show-footer="{
              positiveText: t('actions.confirm'),
              negativeText: t('actions.cancel'),
            }"
            :show="newTeamModal"
            :touched="newTeamName.length > 0"
            :width="'20rem'"
          >
            <template #trigger>
              <n-button
                size="tiny"
                secondary
                type="primary"
                @click="newTeamName = ''"
              >
                <div style="display: flex; align-items: center; gap: 2px">
                  <span style="margin-top: 1px">
                    {{ t("fields.addTeam") }}
                  </span>
                </div>
              </n-button>
            </template>
            <template #content>
              <n-input
                :placeholder="t('fields.teamName')"
                v-model:value="newTeamName"
                @keydown.enter="createNewTeam"
              />
            </template>
          </PopModal>
        </template>
        <data-table
          :columns="columns"
          :data="teamList"
          v-if="teamList"
        >
          <template #name="row">
            {{ row.Name }}
          </template>
          <template #members="row">
            {{ row.NrOfMembers }}
          </template>
          <template #environment="row">
            {{ row.NrOfEnvironments }}
          </template>
          <template #repos="row">
            {{ row.NrOfGitRepos }}
          </template>
          <template #actions="row">
            <div style="display: flex; gap: 5px; justify-content: flex-end;">
              <PopModal
                v-if="!['Administrators', 'Blacklisted'].includes(row.Name)"
                :title="t('actions.edit') + ' ' + t('fields.team').toLowerCase()"
                @positive-click="renameTeam"
                @negative-click="() => {}"
                :show-footer="{
                  positiveText: t('actions.confirm'),
                  negativeText: t('actions.cancel'),
                }"
                :show="renameTeamModal"
                :touched="false"
                :width="'20rem'"
              >
                <template #trigger>
                  <n-button
                    size="tiny"
                    secondary
                    type="primary"
                    @click="() => {
                      newTeamName = row.Name;
                      selectedTeam = row;
                    }"
                  >
                    <div style="display: flex; align-items: center; gap: 2px">
                      <span style="margin-top: 1px">
                        {{ t('fields.rename') + ' ' + t('fields.team').toLowerCase() }}
                      </span>
                    </div>
                  </n-button>
                </template>
                <template #content>
                  <div style="display: flex; flex-direction: column; gap: 5px">
                    <label>{{ t('fields.name') }}</label>
                    <n-input
                      :placeholder="t('fields.name')"
                      v-model:value="newTeamName"
                      @keydown.enter="renameTeam"
                    />
                  </div>
                </template>
              </PopModal>
              <n-button
                v-if="!['Administrators', 'Blacklisted'].includes(row.Name)"
                size="tiny"
                secondary
                type="error"
                @click="() => {
                  if(row.NrOfMembers) {
                    message.error(t('messages.membersInTeam'));
                    return;
                  }

                  selectedTeam = row;
                  deleteConfirmationModal = true;
                }"
              >
                <div style="display: flex; align-items: center; gap: 2px">
                  <span style="margin-top: 1px">
                    {{ t('actions.delete') + ' ' + t('fields.team').toLowerCase() }}
                  </span>
                </div>
              </n-button>
              <n-button
                v-if="!['Administrators', 'Blacklisted'].includes(row.Name)"
                size="tiny"
                secondary
                type="primary"
                @click="() => {
                  selectedTeam = row;
                  permissionModal = true;
                }"
              >
                <div style="display: flex; align-items: center; gap: 2px">
                  <span style="margin-top: 1px">
                    {{ t('actions.edit') + ' ' + t('fields.permission', 2).toLowerCase() }}
                  </span>
                </div>
              </n-button>
              <n-button
                size="tiny"
                secondary
                type="primary"
                @click="handleEditMembers(row)"
              >
                <div style="display: flex; align-items: center; gap: 2px">
                  <span style="margin-top: 1px">
                    {{ t('actions.edit') + ' ' + t('objects.member', 2).toLowerCase() }}
                  </span>
                </div>
              </n-button>
            </div>
          </template>
        </data-table>
      </n-card>
      <Modal
        v-model:show="deleteConfirmationModal"
        :title="t('messages.removeTeam')"
        style="width: 30rem"
        :showFooter="true"
        @positive-click="deleteTeam"
        @negative-click="deleteConfirmationModal = false"
      >
        {{ $t("questions.sure") }}
      </Modal>
      <Modal
        v-model:show="editMembersModal"
        :title="t('actions.edit') + ' ' + t('objects.member', 2)"
        style="width: 30rem"
      >
        <div style="display: flex; justify-content: center;">
          <div style="display: flex; flex-direction: column;">
            <n-checkbox
              v-for="key in Object.keys(users)"
              :key="key"
              v-model:checked="selectedTeamMembers[key]"
              @update:checked="(checked) => handleCheckedChange(checked, key)"
            >
              {{ users[key] }}
            </n-checkbox>
          </div>
        </div>
      </Modal>
      <PermissionsModal
        v-if="selectedTeam && permissionModal"
        :selectedTeam="(selectedTeam as Team)"
        :permissionModal="permissionModal"
        @close-modal="handleClose"
      />
    </div>
    <ErrorDialog v-else :msg="$t('messages.permissionDenied')" />
  </PageLayout>
</template>

<style scoped lang="scss">
.icon {
  cursor: pointer;
  transition: all 0.2s ease-in-out;

  &:hover {
    filter: brightness(1.4);
  }
}
.admin-user-input {
  display: flex;
  margin-bottom: 5px;
  width: 100%;
  justify-content: space-between;
  align-items: center;
}

.permission-input {
  display: grid;
  grid-template-columns: 1fr 6rem;
  margin-bottom: 6px;
  width: 100%;
  justify-content: space-between;
  align-items: center;
}
</style>
