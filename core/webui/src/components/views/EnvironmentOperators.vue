<script setup lang="ts">
import { computed, ComputedRef } from "vue";
import { Column } from "@/components/utils/DataTable.vue";
import { useI18n } from "vue-i18n";
import { useEnv } from "@/store/envStore";
import { useRoute } from "vue-router";
import { useUserStore } from "@/store/userStore";
const userStore = useUserStore();
const route = useRoute();
const env = useEnv(route.params.id as string);

const { t } = useI18n();

const columns: ComputedRef<Column[]> = computed(() => [
  {
    title: t("fields.name"),
    key: "Name",
    width: "30%",
  },
  {
    title: t("fields.team", 2),
    template: "teams",
    width: "30%",
  },
]);

interface UserT {
  Name: string;
  Teams: Array<string>
}

let teamList = $ref<Array<string>>();
let userList = $ref<Array<UserT>>([]);
watch(
  () => env.value,
  () => {
    teamList = Object.keys(env.value.EnvInfo?.Teams || {}).sort((a, b) => a.localeCompare(b));
    userList =  Object.entries(env.value.EnvInfo?.Members || {}).map(([name, teams]) => ({
      Name: name,
      Teams: Object.entries(teams).map(([team]) => team)
    }));
  }
);

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
</script>

<template>
  <div>
    <EnvTab />
    <PageLayout>
      <div v-if="userStore.havePermission('Env_ViewMembers')">
        <n-card
          :title="t('fields.team', 2)"
          size="small"
          style="height: calc(12vh); margin-bottom: 1rem"
        >
          <div style="display: flex; gap: 1rem">
            <n-tag
              v-for="team in teamList"
              :key="team"
              :type="getTeamColor(team)"
            >
              {{ team }}
            </n-tag>
          </div>
        </n-card>
        <n-card
          :title="t('objects.member', 2)"
          size="small"
          style="height: calc(78vh)"
        >
          <data-table :columns="columns" :data="userList">
            <template #teams="user">
              <div>
                <n-tag
                  v-for="team in user.Teams"
                  :type="getTeamColor(team)"
                  :key="team"
                >
                  {{ team }}
                </n-tag>
              </div>
            </template>
          </data-table>
        </n-card>
      </div>

      <ErrorDialog v-else :msg="$t('messages.permissionDenied')" />
    </PageLayout>
  </div>
</template>

<style scoped lang="scss">
.icon {
  cursor: pointer;
  transition: all 0.2s ease-in-out;

  &:hover {
    filter: brightness(1.4);
  }
}
</style>
