import { defineStore } from "pinia";

interface userSettings {
  lang: "pl" | "en";
  themeDark?: boolean;
  color?: { left: string; right: string };
  filtersReleasesProject?: string[];
  filtersReleasesStatus?: number[];
  opacity: number;
}

export const useUserSettings = defineStore("userSettings", {
  state: (): userSettings => {
    return {
      lang: "en",
      themeDark: true,
      color: {
        left: "#1C1C1C",
        right: "#121212",
      },
      opacity: 50,
    };
  },
  persist: true,
});
