<script lang="ts" setup>
// import router from "@/router";
import { useRepo } from "@/store/repoStore";
import { useMessage } from "naive-ui";
import { useRoute, useRouter } from "vue-router";
import { useUserStore } from "@/store/userStore";

const userStore = useUserStore();
const { t } = useI18n();
const message = useMessage();

const route = useRoute();
const router = useRouter();
const repo = $(useRepo(computed(() => route.params.name as string)));

let accessToken = $ref("");
watch($$(repo), () => {
  accessToken = repo?.AccessToCode as string;
});
// "curl -sfL '" +
// document.location.origin +
// "/api/git-on-board?token=" +
// JSON.parse(localStorage.getItem("token")!).token +
// "' | sh";

const copy = () => {
  navigator.clipboard.writeText(
    repo?.RemoteURL ? (repo?.RemoteURL as string) : (repo?.CloneURL as string)
  );
};
const access = () => {
  navigator.clipboard.writeText(accessToken);
};
// let timerangeSelected = $ref<[number, number]>([
//   parseInt(localStorage.getItem("timerangeSelectedFrom") as string) ||
//     Date.now(),
//   parseInt(localStorage.getItem("timerangeSelectedTo") as string) || Date.now(),
// ]);
// let timeMode = $ref<string>(localStorage.getItem("timeMode") || "relative");
// let timeSelected = $ref<number>(
//   parseInt(localStorage.getItem("timeSelected") as string) || 15
// );
// let timeUnitSelected = $ref<string>(
//   localStorage.getItem("timeUnitSelected") || "m"
// );

let timeChangeTrigger = $ref(0);

let showDeleteModal = $ref(false);
const deleteRepo = () => {
  api
    .get("/git-repo-delete", {
      queries: { name: route.params.name as string },
    })
    .then((res) => {
      if (res === "ok") {
        router.push(`/code`);
        message.success(`Repo deleted`);
      } else {
        message.error(res as string);
      }
    });
};
</script>

<template>
  <div>
    <RepoTabs />
    <PageLayout>
      <div v-if="userStore.havePermission('Repo_View')">
        <n-card
          size="small"
          style="
            float: right;
            display: flex;
            gap: 1rem;
            margin-bottom: 1rem;
            height: 4rem;
          "
        >
          <div style="float: right; display: flex; gap: 1rem">
            <TimeSetBtn
              @timeSettingChanged="timeChangeTrigger += 1"
              :id="'-repo-charts'"
              :size="'medium'"
              :timeUnitTo="'h'"
            />
            <Dropdown v-if="userStore.havePermission('Repo_Pull')">
              <template #trigger>
                <n-button type="primary" secondary>
                  Clone
                  <n-icon>
                    <Mdi :path="mdiChevronDown" />
                  </n-icon>
                </n-button>
              </template>
              <template #content>
                <div
                  style="
                    display: flex;
                    align-items: flex-end;
                    flex-flow: column;
                  "
                >
                  <div
                    style="
                      padding: 1rem;
                      display: flex;
                      justify-content: center;
                      align-items: center;
                    "
                  >
                    <p style="margin-right: 1rem">
                      {{ t("actions.addAccess") }}
                    </p>

                    <n-input :value="repo?.AccessToCode" style="width: 250px" />
                    <n-button @click="access">
                      <n-icon> <Mdi :path="mdiContentCopy" /> </n-icon
                    ></n-button>
                  </div>
                  <div
                    style="
                      padding: 1rem;
                      display: flex;
                      justify-content: center;
                      align-items: center;
                    "
                  >
                    <p style="margin-right: 1rem">Clone</p>
                    <n-input
                      :value="
                        repo?.RemoteURL ? repo?.RemoteURL : repo?.CloneURL
                      "
                      style="width: 250px"
                    />
                    <n-button @click="copy">
                      <n-icon> <Mdi :path="mdiContentCopy" /> </n-icon>
                    </n-button>
                  </div>
                </div>
              </template>
            </Dropdown>
            <n-button
              secondary
              type="error"
              @click="showDeleteModal = true"
              :disabled="
                !userStore.havePermission('Glob_CreateAndDeleteGitRepos')
              "
              >{{ $t("actions.deleteRepo") }}
              <template #icon>
                <n-icon>
                  <Mdi :path="mdiTrashCan" />
                </n-icon>
              </template>
            </n-button>
          </div>
        </n-card>
        <div class="charts-container" style="height: calc(100vh - 10.1rem)">
          <!-- <n-card
          size="small"
          style="margin-bottom: 1rem"
          class="chart-card"
          :title="t('objects.commit', 2)"
        >
          <CommitsStackedBarChart />
        </n-card> -->
          <!-- <n-card
          size="small"
          class="chart-card"
          style="grid-area: chart-1; background: #181b1f"
        >
          <MetricsIFrame
            :src="
              'http://' +
              VITE_API_HOST +
              ':32002/d-solo/sAfDUSNVz/commits_summary?orgId=1&from=1666593532833&to=1666615132833&panelId=2&var-title=Aktywność'
            "
            :refreshRate="5"
            :mode="timeMode"
            :timeWindow="timeSelected"
            :timeUnit="timeUnitSelected"
            :timerange="timerangeSelected"
          />
        </n-card> -->
          <n-card
            size="small"
            class="chart-card"
            style="grid-area: chart-1"
            :title="t('repositoryCharts.activity')"
          >
            <RepoActivity />
          </n-card>
          <n-card
            size="small"
            class="chart-card"
            style="grid-area: chart-2"
            :title="t('repositoryCharts.mostActiveUsers')"
            id="most-active-users-card"
          >
            <MostActiveUsers
              :timeChangeTrigger="timeChangeTrigger"
              :id="'-repo-charts'"
            />
          </n-card>
          <n-card
            size="small"
            class="chart-card"
            style="grid-area: chart-3"
            :title="t('repositoryCharts.commitsByHour')"
          >
            <CommitsByHours />
          </n-card>
          <n-card
            size="small"
            class="chart-card"
            style="grid-area: chart-4"
            :title="t('repositoryCharts.commitsByDayOfWeek')"
          >
            <CommitsByDayOfWeek />
          </n-card>
          <n-card
            size="small"
            class="chart-card"
            style="grid-area: chart-5"
            :title="t('repositoryCharts.numLinesOfCode')"
          >
            <NumLinesOfCode />
          </n-card>
          <n-card
            size="small"
            class="chart-card"
            style="grid-area: chart-6"
            :title="`${t('repositoryCharts.repositorySize')} (MB)`"
          >
            <RepositorySize />
          </n-card>
          <n-card
            size="small"
            class="chart-card"
            style="grid-area: chart-7"
            :title="'Wykres w czasie z liczbą środowiska która używa elementów z tego repo'"
          ></n-card>
        </div>
        <Modal
          v-model:show="showDeleteModal"
          show-footer
          :title="$t('actions.deleteRepo')"
          style="max-width: 20rem"
          @positive-click="deleteRepo()"
        >
          {{ $t("questions.sure") }}
        </Modal>
      </div>
      <ErrorDialog v-else :msg="$t('messages.permissionDenied')" />
    </PageLayout>
  </div>
</template>

<style scoped>
.icon-color {
  color: var(--primaryColor);
}

.charts-container {
  width: 100%;
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  margin-bottom: 1rem;
  gap: 1rem;
  grid-template-areas: "chart-1 chart-1 chart-2 chart-2" "chart-3 chart-3 chart-4 chart-4" "chart-5 chart-5 chart-6 chart-7";
}

.chart-card {
  min-height: 10rem;
}
</style>
