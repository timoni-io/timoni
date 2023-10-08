import { defineStore } from "pinia";
import { z } from "zod";
import { userInfoSchema } from "@/zodios/schemas/user";

type Permissions = z.infer<typeof userInfoSchema>["PermissionsGlobal"];
type userTeams = z.infer<typeof userInfoSchema>["Teams"];
export const useUserStore = defineStore("userStore", {
  state: () => {
    return {
      email: "",
      userName: "",
      permissions: {} as Permissions,
      teams: [] as userTeams,
      HideGitRepoLocal: false,
      havePermission: function (permName: string) {
        return !!permName || true;
      },
    };
  },
});
