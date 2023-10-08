<script setup lang="ts">
import { computed, ComputedRef } from "vue";
import { Column } from "@/components/utils/DataTable.vue";
import { User, teamListObject } from "@/zodios/schemas/user";
import { z } from "zod";
import { useI18n } from "vue-i18n";
import { useTimeFormatter } from "@/utils/formatTime";
import { useMessage } from "naive-ui";
import { useUserStore } from "@/store/userStore";


type UsersList = z.infer<typeof User>;
type Teams = Record<string, string>;
type Team = z.infer<typeof teamListObject>;

const userStore = useUserStore();
const { t } = useI18n();
const distanceToNow = useTimeFormatter().distanceToNow;
const message = useMessage();
let usersList = $ref<UsersList[]>();
let teams = $ref<Teams>({});
let blacklistedId = $ref<string>('');
let teamsExtended = $ref<Array<Team>>();
let newUserEmail = $ref<string>('');
let newUserName = $ref<string>('');
let newUserTeams = $ref<Array<string>>([]);
let inviteMemberModal = $ref<boolean>(false);
let updateMemberModal = $ref<boolean>(false);
let inviteUserOpen = $ref<boolean>(false);
let qrToken = $ref(null);
let selectedTeam = $ref<Team>();
let permissionModal = $ref<boolean>(false);

const columns: ComputedRef<Column[]> = computed(() => [
  {
    title: t("fields.name"),
    key: "Name",
    template: "name",
    width: "20%",
  },
  {
    title: t("fields.email"),
    key: "Email",
    template: "email",
    width: "20%",
  },
  {
    title: t("fields.createdTimeStamp"),
    key: "CreatedTimeStamp",
    template: "createdTime",
    width: "15%",
  },
  {
    title: t("fields.lastActionTimeStamp"),
    key: "LastActionTimeStamp",
    template: "lastAction",
    width: "15%",
  },
  {
    title: t("fields.team", 2),
    template: "teams",
    width: "25%",
  },
  {
    title: "",
    template: "edit",
    width: "5%",
  },
]);

const getUserList = () => {
  api.get("/user-list").then((res) => {
    usersList = Object.values(res).sort((a, b) => a.Name.localeCompare(b.Name));
  });
};

const getTeamList = () => {
  api.get("/team-list").then((res) => {
    for (const team in res) {
      teams[res[team].ID] = res[team].Name;

      if(res[team].Name === 'Blacklisted') {
        blacklistedId = res[team].ID;
      }
    }

    teamsExtended = res;
  });
};

const getTeamColor = (name: string) => {
  switch(name) {
    case 'Administrators':
      return 'success';
    case 'Blacklisted':
      return 'error';
    default:
      return undefined;
  }
}

const areListsEqual = (list1: Array<string>, list2: Array<string>): boolean => {
  if (list1.length !== list2.length) {
    return false;
  }

  const sortedList1 = list1.slice().sort();
  const sortedList2 = list2.slice().sort();

  for (let i = 0; i < sortedList1.length; i++) {
    if (sortedList1[i] !== sortedList2[i]) {
      return false;
    }
  }

  return true;
}

const createNewUser = () => {
  if (!newUserEmail) {
    message.error(t("messages.userEmailRequired"));
    return;
  }
  api
    .get("/user-invite", {
      queries: {
        Email: newUserEmail as string,
      },
    })
    .then((res) => {
      if (res.error && res.error !== "empty token") {
        message.error(res.error);
        return;
      }

      if (res === 'ok') {
        getUserList();
        message.success(t("messages.userAdded"));
        inviteMemberModal = !inviteMemberModal;
      } else {
        message.error(res);
      }
    });
};

const updateUser = (user: UsersList) => {
  if (!newUserName) {
    message.error(t("messages.nameRequired"));
    return;
  }

  if (newUserName !== user.Name || !areListsEqual(user.Teams, newUserTeams)) {
    let teams: Record<string, boolean> = {};

    for (const key of newUserTeams) {
      teams[key] = true;
    }

    api
      .post("/user-update", {
          Email: user.Email,
          Name: newUserName,
          Teams: teams
      })
      .then(() => {
        getUserList();
        message.success(t("messages.userAdded"));
        updateMemberModal = false;
      });
  }
};

const getInvitationQr = (email: string) => {
  qrToken = null;
  api.get("/user-invite-qr", {
    queries: {
        Email: email,
      }
    })
    .then((res) => {
      qrToken = res.QRcode;
    });
};

const handleClose = () => {
  permissionModal = false;
  selectedTeam = undefined;
}

onBeforeMount(() => {
  getUserList();
  getTeamList();
});

</script>

<template>
  <PageLayout>
    <div
      v-if="userStore.havePermission('Glob_ManageGlobalMemebers')"
    >
      <n-card
        size="small"
        :title="t('objects.member', 2)"
        style="min-height: 90vh; position: relative"
      >
        <template #header-extra>
          <PopModal
            :title="t('fields.addMember')"
            @positive-click="createNewUser"
            @negative-click="() => {}"
            :show-footer="{
              positiveText: t('actions.confirm'),
              negativeText: t('actions.cancel'),
            }"
            :show="inviteMemberModal"
            :touched="newUserEmail.length > 0"
            :width="'20rem'"
          >
            <template #trigger>
              <n-button
                size="tiny"
                secondary
                type="primary"
                @click="newUserEmail = ''"
              >
                <div style="display: flex; align-items: center; gap: 2px">
                  <span style="margin-top: 1px">
                    {{ t("fields.addMember") }}
                  </span>
                </div>
              </n-button>
            </template>
            <template #content>
              <n-input
                :placeholder="t('fields.userEmail')"
                v-model:value="newUserEmail"
                @keydown.enter="createNewUser"
              />
            </template>
          </PopModal>
        </template>
        <data-table
          :columns="columns"
          :data="usersList"
          v-if="usersList"
        >
          <template #name="row">
              <span :style="row.Teams.includes(blacklistedId) ? 'text-decoration: line-through;' : ''">{{ row.Name }}</span>
          </template
          >
          <template #email="row">
            <span :style="row.Teams.includes(blacklistedId) ? 'text-decoration: line-through;' : ''">{{ row.Email }}</span>
          </template>
          <template #createdTime="row">
            {{ distanceToNow(new Date(parseInt(row.CreatedTimeStamp) * 1000)) }}
          </template>
          <template #lastAction="row">
            <span v-if="!row.LastActionTimeStamp">
              {{ t("messages.never") }}
            </span>
            <span v-else>
              {{
                distanceToNow(
                  new Date(parseInt(row.LastActionTimeStamp) * 1000)
                )
              }}
            </span>
          </template>
          <template #teams="row">
            <div>
              <div
                v-if="!row.Teams.includes(blacklistedId)"
                style="display: flex; gap: 5px"
              >
                <div
                  v-for="team in row.Teams.slice().sort((a: string, b: string) => teams[a].localeCompare(teams[b]))"
                  :key="team"
                >
                  <n-tag
                    v-if="!['Administrators', 'Blacklisted'].includes(teams[team])"
                    :type="getTeamColor(teams[team])"
                    @click="() => {
                      selectedTeam = teamsExtended?.find((el) => el.ID === team);
                      permissionModal = true;
                    }"
                    style="cursor: pointer;"
                  >
                    {{ teams[team] }}
                  </n-tag>
                  <n-tooltip
                    v-else  
                    trigger="hover"
                  >
                    <template #trigger>
                      <n-tag
                        :type="getTeamColor(teams[team])"
                      >
                        {{ teams[team] }}
                      </n-tag>
                    </template>
                    {{ t('messages.teamNotEditable') }}
                  </n-tooltip>
                </div>
              </div>
              <n-tooltip
                v-else  
                trigger="hover"
              >
                <template #trigger>
                  <n-tag
                    :type="getTeamColor('Blacklisted')"
                  >
                    Blacklisted
                  </n-tag>
                </template>
                {{ t('messages.teamNotEditable') }}
              </n-tooltip>
            </div>
          </template>
          <template #edit="row">
            <div style="display: flex; gap: 5px; justify-content: flex-end;">
              <PopModal
                :title="t('actions.showInvitationQr')"
                @positive-click="() => {}"
                @negative-click="() => {}"
                :width="'20rem'"
                :show="inviteUserOpen"
                @close="inviteUserOpen = false"
              >
                <template #trigger>
                  <div>
                    <n-button
                      v-if="!row.LastActionTimeStamp && !row.Teams.includes(blacklistedId)"
                      style="float: left;"
                      @click="getInvitationQr(row.Email)"
                      secondary
                      type="primary"
                      size="tiny"
                      
                    >
                      <n-icon class="icon"> <mdi :path="mdiShieldLinkVariant" /> </n-icon>&nbsp; {{ t("fields.inviteMember") }}
                    </n-button>
                  </div>
                </template>
                <template #content>
                  <div style="display: flex; flex-direction: column; gap: 0.5rem">
                    <img :src="`data:image/png;base64,${qrToken}`" />
                  </div>
                </template>                      
              </PopModal>
              <PopModal
                :title="t('actions.editUser')"
                @positive-click="updateUser(row)"
                @negative-click="() => {}"
                :show-footer="{
                  positiveText: t('actions.confirm'),
                  negativeText: t('actions.cancel'),
                }"
                :show="updateMemberModal"
                :touched="row.Name !== newUserName || !areListsEqual(row.Teams, newUserTeams)"
                :width="'20rem'"
              >
                <template #trigger>
                  <n-button
                    size="tiny"
                    secondary
                    type="primary"
                    @click="() => {
                      newUserName = row.Name;
                      newUserTeams = row.Teams;
                    }"
                  >
                    <div style="display: flex; align-items: center; gap: 2px">
                      <span style="margin-top: 1px">
                        {{ t('actions.editUser') }}
                      </span>
                    </div>
                  </n-button>
                </template>
                <template #content>
                  <div style="display: flex; flex-direction: column; gap: 5px">
                    <label>{{ t('fields.name') }}</label>
                    <n-input
                      :placeholder="t('fields.name')"
                      v-model:value="newUserName"
                      @keydown.enter="updateUser(row)"
                    />
                    <label>{{ t('fields.selectTeam') }}</label>
                    <n-select
                      v-model:value="newUserTeams"
                      multiple
                      :options="Object.keys(teams).map((key) => ({
                        label: teams[key],
                        value: key,
                      }))"
                    >
                    </n-select>
                  </div>
                </template>
              </PopModal>
            </div>
          </template>
        </data-table>
      </n-card>
    </div>
    <ErrorDialog v-else :msg="$t('messages.permissionDenied')" />
    <PermissionsModal
      v-if="selectedTeam"
      :selectedTeam="(selectedTeam as Team)"
      :permissionModal="permissionModal"
      @close-modal="handleClose"
    />
  </PageLayout>
</template>

<style scoped lang="scss">
</style>
