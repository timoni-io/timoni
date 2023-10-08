import { createModule } from "..";
import { ref } from "vue";
import { debounce } from "radash";
import { useLogsStore } from "@/store/logsStore";
import { Args } from "@/next-api/types";

export type LiveLogs = {
  refreshLogsSubscribtion: () => void;
};

export const SubscribeMessagesModule = createModule({
  setup(core, api) {
    const lastArgsHash = ref("");

    const logs = useLogsStore();

    const updateSubscribtion = debounce({ delay: 2 }, (args: Args) => {
      if (lastArgsHash.value !== JSON.stringify(args)) {
        core.sendWithResponse((id) => {
          logs.id.push({
            id,
            live: true,
            fullLog: false,
          });
          logs.liveId = id;
          logs.isLive = true;
          logs.incomingCount = 0;
          return {
            RequestID: id,
            Action: "Live",
            Args: args,
          };
        });
        lastArgsHash.value = JSON.stringify(args);
      }
    });

    watch(
      () => logs.liveArgs,
      (args) => {
        args && updateSubscribtion(args);
      }
    );

    const refreshSubscribtion = () => {
      logs.liveArgs &&
        core.sendWithResponse((id) => {
          logs.id.push({
            id,
            live: true,
            fullLog: false,
          });
          logs.liveId = id;
          logs.isLive = true;
          logs.incomingCount = 0;
          return {
            RequestID: id,
            Action: "Live",
            Args: logs.liveArgs!,
          };
        });
    };

    // @ts-ignore
    api.refreshLogsSubscribtion = refreshSubscribtion;

    return { updateSubscribtion, logs };
  },
  onMessage(message) {
    // Subscribtion response
    if (message.Code === 1) {
      this.logs.LogsList = [];
      return;
    }
    if (message.Code === 0 && Array.isArray(message.Data)) {
      let logId = this.logs.id.find((el) => el.id === message.RequestID);
      if (!logId?.fullLog) this.logs.LogsList = message.Data;
      if (logId?.live && !this.logs.isLive) {
        this.logs.incomingCount = this.logs.incomingCount + message.Data.length;
      }
    }
  },
});
