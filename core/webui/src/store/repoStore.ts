import { MaybeRef } from "@vueuse/core";
import { defineStore } from "pinia";
import { RepoInfo } from "@/zodios/schemas/repos";
// export type RepoInfo = ResType<"/git-repo-info">;
type RepoStore = { RepoInfos: Record<string, z.infer<typeof RepoInfo>> };
import { z } from "zod";

export const useRepoStore = defineStore("repoStore", {
  state: () => ({ RepoInfos: {} } as RepoStore),
  actions: {
    async fetchRepo(repoName: string) {
      await api
        .get("/git-repo-info", { queries: { name: repoName } })
        .then((res) => {
          if (typeof res === "string") {
            window.location.href = "/code";
          }
          if (typeof res === "object")
            this.RepoInfos = {
              ...this.RepoInfos,
              [repoName]: res,
            };
        })
        .catch((e) => {
          return Promise.reject(e);
        });
    },
  },
});

export const useRepo = (repoName: MaybeRef<string>, once = false) => {
  const repoStore = useRepoStore();

  watchEffect(() => {
    if (unref(repoName)) repoStore.fetchRepo(unref(repoName));
  });
  if (once)
    useIntervalFn(() => {
      if (unref(repoName)) repoStore.fetchRepo(unref(repoName));
    }, 5000);

  return computed(
    () =>
      (repoStore.RepoInfos[unref(repoName)] || null) as z.infer<
        typeof RepoInfo
      > | null
  );
};
export const useRepoBranch = defineStore("repoBranch", {
  state: () => {
    return {
      selectedBranch: "",
    };
  },
});
