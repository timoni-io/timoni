<script setup lang="ts">
import { useUserStore } from "@/store/userStore";

const userStore = useUserStore();
// import { format } from "date-fns";

// useApi();
</script>

<template>
  <div>
    <EnvTab />
    <PageLayout style="height: calc(100vh - 5rem)">
      <LogsAdvanced v-if="userStore.havePermission('Env_View')" />
      <ErrorDialog v-else :msg="$t('messages.permissionDenied')" />
    </PageLayout>
  </div>
</template>

<style scoped>
.history-logs-table {
  font-family: Menlo, Monaco, Consolas, "Liberation Mono", "Courier New",
    monospace;
  font-size: 0.8rem;
  display: flex;
  gap: 0.5rem;
  align-items: center;
  flex-wrap: nowrap;
  white-space: nowrap;
}
.logs {
  /* display: grid; */
  grid-template-rows: 2rem 1fr;
  width: 100%;
  max-width: 100%;
  height: 100%;
}
.scrollbar {
  overflow-y: hidden;
  /* background-color: #000000; */
  background-color: #f87272;
}

table {
  text-align: left;
  width: 100%;
  border-spacing: 0;
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

th {
  padding-left: 0.5rem;
  padding-right: 0.5rem;
}

.error {
  color: #f87272;
  /*  background-color: #470000af; */
}

.fatal {
  color: rgb(235, 6, 235);
  /* background-color: rgba(78, 2, 78, 0.603); */
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
  bottom: 0;
  left: 0;
  padding-bottom: 0.5rem;
  transition: transform 0.2s;
  justify-content: center;
}
.incoming.hide {
  transform: translateY(3rem);
}
</style>
