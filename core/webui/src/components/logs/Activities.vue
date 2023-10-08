<script setup lang="ts">
import { useApi } from "@/next-api";
// import { useLogsStore } from "@/store/logsStore";
// import moment from "moment";
import { useRoute } from "vue-router";
import { useLogs } from "@/store/logsStore";
import { useTimeFormatter } from "@/utils/formatTime";

// import { useEnvStore } from "@/store/envStore";

// const envStore = useEnvStore();
const route = useRoute();

interface Log {
  time: string;
  id: string;
  level: string;
  message: string;
  element: string;
  pod: string;
  env: string;
  version: string;
  user: string;
  project: string;
  details_string: string;
  details_number: number;
  env_id: string;
  tags_string?: any;
}

useApi();

// const logs = useLogsStore();

// onMounted(() => {
//   logs.liveArgs = {
//     EnvID: route.params.id as string,
//     Limit: 50,
//   };
// });

// let localLogList = $ref<typeof logs.LogsList>([]);

let showDetails = $ref(false);
let currentLog = $ref<Log>();

const detailsShow = (log: Log) => {
  currentLog = log;
  showDetails = true;
};

// const updateLogList = () => {
//   localLogList = JSON.parse(JSON.stringify(logs.LogsList));
//   // triggerRef($$(localLogList));
// };

// watchEffect(() => {
//   logs.LogsList = [];
//   logs.liveArgs = {
//     Subs: [
//       {
//         EnvID: route.params.id as string,
//         Limit: 10,
//       },
//     ],
//   };
// });

// logs.$subscribe((_) => {
//   updateLogList();
// });
const { dateFranekFormat } = useTimeFormatter();

// const relativeOrDistanceToNowLukaszFormatted = (time: any) => {
//   const timeFormatted = relativeOrDistanceToNow(
//     new Date(Math.floor(time / 1000000))
//   );
//   if (
//     ["mniej niż minuta temu", "less than a minute ago"].includes(timeFormatted)
//   )
//     return t("time.now").toLowerCase();
//   return timeFormatted;
// };

onMounted(() => {
  startLive();
});
let envId = $ref([
  {
    id: route.params.id as string,
    limit: 10,
    events: true as true,
  },
]);
let { logs, startLive } = useLogs($$(envId));
</script>
<template>
  <div v-if="logs" style="position: relative; height: 100%">
    <n-scrollbar
      style="position: absolute; inset: 0; height: 100%; overflow: hidden"
    >
      <div style="display: flex; flex-flow: column nowrap; gap: 0.5rem">
        <template v-for="(log, idx) in logs" :key="log.id">
          <div
            v-if="logs.length - idx < 10"
            :class="{
              error: log.level === 'ERROR',
              warn: log.level === 'WARN',
              fatal: log.level === 'FATAL',
            }"
            style="
              background-color: var(--cardColor);
              padding: 0 0.5rem;
              border-radius: 5px;
              padding-bottom: 0.5rem;
            "
          >
            <div>
              <div style="display: flex; gap: 0.5rem; align-items: center">
                <span style="margin: 0 -0.5rem">
                  <LogIcon :level="log.level" @click="detailsShow(log)" />
                </span>
                <span>•</span>
                <span>
                  <time>
                    {{ dateFranekFormat(log.time) }}
                    <!-- {{
                    log.time
                      ? moment(log.time / 1000000).fromNow()
                      : moment(Date.now()).fromNow()
                  }} -->
                  </time>
                </span>
                <span>•</span>
                <span>
                  <span
                    style="
                      color: #6ee7b7;
                      white-space: nowrap;
                      word-break: keep-all;
                    "
                    >{{ log.element }}</span
                  >
                </span>
              </div>

              {{ log.message.substring(0, 100) }}...
            </div>
          </div>
        </template>
      </div>
    </n-scrollbar>
  </div>

  <div
    v-else
    style="
      display: flex;
      justify-content: center;
      align-items: center;
      height: inherit;
      height: 100%;
    "
  >
    <Spinner :data="''" />
  </div>
  <!-- </n-card> -->
  <Modal
    v-model:show="showDetails"
    title="Log details"
    style="width: 1200px; max-height: 900px"
  >
    <LogDetails :log="currentLog" />
  </Modal>
</template>

<style scoped>
.n-card {
  /* background: var(--cardColor); */
  padding: 15px;
  border-radius: 5px;
}
.logs {
  /* display: grid; */
  grid-template-rows: 2rem 1fr;
  width: 100%;
  max-width: 100%;
  height: 100%;
}

table {
  text-align: left;
  width: 100%;
  border-spacing: 0 10px;
  height: 100%;
  max-height: 100%;
  overflow-y: scroll;
}

tbody {
  width: 100%;
}

tr {
  font-family: Menlo, Monaco, Consolas, "Liberation Mono", "Courier New",
    monospace;
  color: white;
  font-size: 0.8rem;
}

tr:hover {
  background-color: #2c2c2cff;
}

tr.spaceUnder > td {
  padding-bottom: 10em;
}

th {
  padding-left: 0.5rem;
  padding-right: 0.5rem;
}

.error {
  color: #f87272;
  /* background-color: #470000af; */
}

.fatal {
  color: rgb(235, 6, 235);
  /*background-color: rgba(78, 2, 78, 0.603); */
}

.warn {
  color: #fbbd23;
  /*  background-color: #382800b0; */
}

time {
  color: #fde68a;
  white-space: nowrap;
}

.incoming {
  display: flex;
  position: absolute;
  right: 0;
  bottom: 0px;
  left: 0;
  padding-bottom: 0.5rem;
  transition: transform 0.2s;
  justify-content: center;
}
.incoming.hide {
  transform: translateY(3rem);
}

/* :deep(.n-tabs-pane-wrapper > * > *) {
  max-height: calc(var(--h) - 2rem);
}

:deep(.n-tabs-nav) {
  padding-left: 0.5rem;
} */
</style>
<style>
.scrollbar.n-scrollbar {
  max-height: calc(var(--h) - 2.5rem);
}
</style>
