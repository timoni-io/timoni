import { MaybeRef } from "@vueuse/core";
import { defineStore } from "pinia";
import { EnvMap, RepoMap } from "../zodios/schemas/dashboard";
import { useRoute } from "vue-router";
import { z } from "zod";
import useErrorMsg from "@/utils/errorMsg";
import { EnvInfo } from "@/zodios/schemas/env";
let { setEnvError, setErrorMsg } = useErrorMsg();

export type EnvInfo = Exclude<ResType<"/env-info">, string>;
export type EnvElement = Exclude<ResType<"/env-element-map">, string>;
type EnvStore = {
  EnvInfos: Record<string, EnvInfo>;
  EnvElements: Record<string, EnvElement>;
  Metrics: z.infer<typeof EnvInfo>["Metrics"];
  Alerts: z.infer<typeof EnvInfo>["Alerts"];
  URLs: z.infer<typeof EnvInfo>["URLs"];
};
export type Dashboard = ResType<"/dashboard">;

export const useEnvStore = defineStore("envStore", {
  state: () => ({
    EnvInfos: {},
    EnvElements: {},
    Alerts: {},
    Metrics: {},
    URLs: {},
  } as EnvStore),
  actions: {
    async fetchEnv(envId: string) {
      await api
        .get("/env-info", { queries: { env: envId } })
        .then((res) => {
          if (is(z.string())(res)) {
            // window.location.href = "/env";
            setEnvError(res);
            throw new Error("Received an unsupported response. Repeating query, please wait a moment.");
          }
          this.EnvInfos = {
            ...this.EnvInfos,
            [envId]: res,
          };
          this.Alerts = res.Alerts;
          this.Metrics = res.Metrics;
          this.URLs = res.URLs;
        })
        .catch((e) => {
          setErrorMsg(e);
          return Promise.reject(e);
        });
    },
    async fetchElements(envId: string) {
      await api
        .get("/env-element-map", { queries: { env: envId } })
        .then((res) => {
          if (is(z.string())(res)) {
            // window.location.href = "/env";
            throw new Error("env-element-map error");
          }
          this.EnvElements = { ...this.EnvElements, [envId]: res };
        })
        .catch((e) => {
          return Promise.reject(e);
        });
    },
  },
});

export const useDashboardStore = defineStore("dashboard", {
  state: () => ({
    envMap: undefined as z.infer<typeof EnvMap> | undefined,
    repoMap: undefined as z.infer<typeof RepoMap> | undefined,
  }),
  actions: {
    async fetchEnvMap() {
      await api.get("/env-map").then((res) => {
        this.envMap = res;
      });
    },
    async fetchRepoMap() {
      await api.get("/git-repo-map").then((res: z.infer<typeof RepoMap>) => {
        this.repoMap = res;
      });
    },
  },
});

export const useEnv = createSharedComposable((envId: MaybeRef<string>) => {
  const route = useRoute();

  const envStore = useEnvStore();

  watchEffect(() => {
    if (unref(envId)) envStore.fetchEnv(unref(envId));
    if (unref(envId) && route.path.includes("elements"))
      envStore.fetchElements(unref(envId));
  });

  useIntervalFn(() => {
    if (
      unref(envId) &&
      !route.path.includes("elements") &&
      !route.path.includes("variables")
    )
      envStore.fetchEnv(unref(envId));
    if (
      unref(envId) &&
      (route.path.includes("elements") || route.path.includes("variables"))
    )
      envStore.fetchElements(unref(envId));
  }, 1000);

  const refetch = () => {
    envStore.fetchEnv(unref(envId));
  };

  return computed(() => {
    return {
      EnvInfo: (envStore.EnvInfos[unref(envId)] || null) as EnvInfo | null,
      EnvElements: (envStore.EnvElements[unref(envId)] ||
        null) as EnvElement | null,
      Alerts: envStore.Alerts,
      Metrics: envStore.Metrics,
      URLs: envStore.URLs,
      refetch
    };
  });
});

export const useEnvElements = createSharedComposable(
  (envId: MaybeRef<string>) => {
    const envStore = useEnvStore();
    watchEffect(() => {
      envStore.fetchElements(unref(envId));
    });
    return computed(() => {
      return {
        EnvInfo: (envStore.EnvInfos[unref(envId)] || null) as EnvInfo | null,
        EnvElements: (envStore.EnvElements[unref(envId)] ||
          null) as EnvElement | null,
        Metrics: {},
      };
    });
  }
);

export const useDashboard = createSharedComposable(() => {
  const dashboardStore = useDashboardStore();
  const route = useRoute();

  if (route.path === "/env") dashboardStore.fetchEnvMap();
  if (route.path === "/code") dashboardStore.fetchRepoMap();

  useIntervalFn(() => {
    if (route.path === "/env") dashboardStore.fetchEnvMap();
    if (route.path === "/code") dashboardStore.fetchRepoMap();
  }, 1000);

  return dashboardStore;
});

// export const useEnvStore = defineStore("envStore", () => {
//   const state = ref<EnvStore>({ EnvList: [], EnvInfo: null, ProjectList: [] });
//   const route = useRoute();

//   setInterval(() => {
//     if (route.name?.toString() === "Home") {
//       api.get("/dashboard", {}).then((res) => {
//         this.EnvList = res.LastVisited || [];
//         state.value.ProjectList = res.ProjectList || [];
//       });
//     }
//     if (route.name?.toString() === "Environment") {
//       api
//         .get("/env-info", { queries: { env: route.params.id as string } })
//         .then((res) => {
//           state.value.EnvInfo = res;
//         });
//     }
//   }, 1000);

//   return { state };
// });
