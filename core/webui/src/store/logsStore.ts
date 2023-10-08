import { useApi } from "@/next-api";
import { Args } from "@/next-api/types";
import { Ref } from "vue";
import { z } from "zod";

export const LogSchema = z.object({
  time: z.string(),
  id: z.string(),
  level: z.string(),
  message: z.string(),
  element: z.string(),
  pod: z.string(),
  env: z.string(),
  version: z.string(),
  user: z.string(),
  project: z.string(),
  details_string: z.string(),
  details_number: z.number(),
  env_id: z.string(),
  tags_string: z.any(),
});

export type ILog = z.infer<typeof LogSchema> & { Event: boolean };

interface LogsStore {
  LogsList: ILog[] | null;
  liveArgs: Args | null;
  LogsDescripton: any;
  liveId: string;
  incomingCount: number;
  isLive: boolean;
  id: { id: string; live: boolean; fullLog: boolean }[];
}

const betterFasterLogsStore = shallowReactive<LogsStore>({
  LogsList: null,
  LogsDescripton: "",
  liveArgs: null,
  liveId: "",
  isLive: false,
  incomingCount: 0,
  id: [],
});

export const useLogsStore = () => {
  return betterFasterLogsStore;
};

// export const useLogsStore = defineStore("logsStore", {
//   state: (): LogsStore => {
//     return {
//       LogsList: null,
//       LogsDescripton: "",
//       liveArgs: null,
//     };
//   },
// actions: {
//   fetchLogs() {
//     // let dataLogs = ref<ILog[]>([]);
//     api
//       .post("/logs", {
//         Is: {
//           env: ["env-002"],
//           element: ["aaa", "bbb"],
//         },
//         TableSuffix: ["container", "system"],
//         Limit: 60,
//         ID: null,
//         Time: null,
//         Type: "latest",
//       })
//       .then((data) => {
//         // console.log(data);
//         // dataLogs.value = data;
//         this.LogsList.push(...data);
//         if (this.LogsList.length > 20) {
//           this.LogsList.splice(0, this.LogsList.length - 20);
//         }
//       });
//   },
// },
// });

type UseLogsResult = {
  logs: Ref<ILog[]>;
  startLive: () => void;
  stopLive: () => void;
  loadMore: (count: number) => Promise<ILog[]>;
  isLive: Ref<boolean>;
  // incomingLogs: Ref<ILog[]>;
  isLoading: Ref<boolean>;
  incomingCount: Ref<number>;
  refresh: () => void;
  clear: () => void;
  noMoreLogs: Ref<boolean>;
};

const LOGS_LIMIT = 40;

export const useLogs = (
  envId: Ref<
    { id: string; imageIdList?: string[]; limit?: number; events?: true }[]
  >
): UseLogsResult => {
  const api = useApi();
  const logsStore = useLogsStore();

  let logs = $shallowRef<ILog[]>([]);
  let incomingLogs = <ILog[]>[];
  let incomingCount = $ref(0);
  let isLive = $ref(true);
  let isLoading = $ref(true);
  let clearLogs = $ref(false);
  let noMoreLogs = $ref(false);

  setTimeout(() => {
    isLoading = false;
  }, 2000);
  watch(
    () => logsStore.incomingCount,
    () => {
      incomingCount = logsStore.incomingCount;
    }
  );
  watch(
    () => logsStore.LogsList,
    () => {
      const data = { LogsList: logsStore.LogsList };
      isLoading = false;
      if (clearLogs) {
        logs = data.LogsList || [];
        clearLogs = false;
        triggerRef($$(logs));
        return;
      }

      if (isLive) {
        if (!logs) logs = [];
        if (data.LogsList) {
          logs = logs.concat(data.LogsList);
        }
        logs.splice(0, logs.length - LOGS_LIMIT);
      } else if (data.LogsList) {
        // console.log(logsStore);
        // incomingCount += data.LogsList.length;
      }
    }
  );

  const clear = () => {
    clearLogs = true;
  };
  // const updateLiveArgs = (newEnvId) => {

  // }
  watch(
    envId,
    (newEnvId, oldEnvId) => {
      // isLoading = true;
      noMoreLogs = false;
      if (
        (oldEnvId && oldEnvId[0].id !== newEnvId[0].id) ||
        newEnvId[0].events !== newEnvId[0].events
      ) {
        clearLogs = true;
        logs = [];
      }
      if (
        newEnvId.find((el) => (el.id || "").includes("image_builder"))
          ?.imageIdList?.length === 0
      ) {
        return;
      }

      logsStore.liveArgs = {
        // @ts-ignore
        Subs: [
          ...newEnvId.map((el) => {
            if ((el.id || "").includes("image_builder")) {
              return {
                EnvID: el.id,
                Limit: LOGS_LIMIT,
                FullLog: true,
                ...(el.imageIdList?.length && {
                  Where: {
                    Type: "OR",
                    Value: [
                      ...el.imageIdList.map((image: string) => {
                        return {
                          Type: "IS",
                          Value: {
                            Field: "imageid",
                            Value: image,
                          },
                        };
                      }),
                    ],
                  },
                }),
              };
            }
            if (el.events) {
              return {
                EnvID: el.id,
                Limit: LOGS_LIMIT,
                Events: true,
              };
            }
            return {
              EnvID: el.id,
              Limit: el.limit ? el.limit : LOGS_LIMIT,
            };
          }),
        ],
      };
    },
    { immediate: true, deep: true }
  );

  watch($$(isLive), (isLive) => {
    if (isLive) {
      logs = logs
        .concat(incomingLogs)
        .slice(logs.length - LOGS_LIMIT, logs.length);
      incomingLogs = [];
      incomingCount = 0;
      logsStore.isLive = true;
      logsStore.incomingCount = 0;
    } else {
      logsStore.isLive = false;
    }
  });

  const startLive = () => {
    isLive = true;
  };

  const stopLive = () => {
    isLive = false;
  };

  const loadMore = async (count: number) => {
    const fromTime = logs[0]?.time;
    const resList = await api.getLogsVector(
      envId.value.map((env) => {
        return {
          Query: {
            Type: "VECTOR",
            Time: fromTime,
            EnvID: env.id,
            Direction: "BEFORE",
            Limit: count,
            ...(env.events ? { Events: true } : {}),
            ...(env.id === "image_builder"
              ? {
                  Where: {
                    Type: "OR",
                    Value: env.imageIdList?.map((image) => {
                      return {
                        Type: "IS",
                        Value: {
                          Field: "imageid",
                          Value: image,
                        },
                      };
                    }),
                  },
                }
              : {}),
          },
        };
      }) as any
    );
    if (!resList.Data.length) {
      noMoreLogs = true;
    }

    if (Array.isArray(resList.Data)) {
      logs = resList.Data.concat(logs);
      logsStore.isLive = false;
    }

    return resList.Data || [];
  };

  const refresh = () => {
    api.refreshLogsSubscribtion?.();
  };

  return {
    logs: $$(logs),
    startLive,
    stopLive,
    loadMore,
    refresh,
    isLive: $$(isLive),
    // incomingLogs: $$(incomingLogs),
    incomingCount: $$(incomingCount),
    isLoading: $$(isLoading),
    noMoreLogs: $$(noMoreLogs),
    clear,
  };
};
