<script setup lang="ts">
// import { useI18n } from "vue-i18n";
import NavbarTabs from "@/components/layout/nav/NavbarTabs.vue";
import { useRepo } from "@/store/repoStore";
import { useRoute, useRouter } from "vue-router";
import { envIcon } from "@/utils/iconFactory";
import { useI18n } from "vue-i18n";
import { useSpinner } from "@/store/spinner";
import { useRepoBranch } from "@/store/repoStore";
import { useRouteParam } from "@/utils/router";

const repoBranch = useRepoBranch();
const router = useRouter();
const route = useRoute();
let selectedBranch = $(useRouteParam("branch"));
const spinner = useSpinner();
const { t } = useI18n();
const repo = $(useRepo(computed(() => route.params.name as string)));
const currentTab = $computed(() => {
  if (route.path.split("/").length < 4) {
    return "repo";
  }
  if (route.path.split("/").length < 5) {
    return route.path.split("/")[3];
  }
  return route.path.split("/")[4];
});
// let includesReadme = $ref<Record<string, boolean>>({});

let includesReadme = $(
  useLocalStorage<Record<string, boolean>>("includesReadme", {})
);

const repoTabs = computed(() => [
  {
    name: repo?.Name,
    route: `repo`,
    path: `/code/${repo?.Name}`,
    icon: envIcon(3),
  },
  {
    name: t("objects.commit", 2),
    route: `commits`,
    path: `/code/${repo?.Name}/${
      repoBranch.selectedBranch
        ? repoBranch.selectedBranch
        : repo?.DefaultBranch
    }/commits`,
    icon: {
      icon: mdiGit,
      color: "",
      iconSize: undefined,
    },
  },
  {
    name: t("objects.file", 2),
    route: "files",
    path: `/code/${repo?.Name}/${
      repoBranch.selectedBranch
        ? repoBranch.selectedBranch
        : repo?.DefaultBranch
    }/files`,
    icon: {
      icon: mdiFileTree,
      color: "",
      iconSize: undefined,
    },
  },
  {
    name: t("objects.element", 2),
    route: "elements",
    path: `/code/${repo?.Name}/${
      repoBranch.selectedBranch
        ? repoBranch.selectedBranch
        : repo?.DefaultBranch
    }/elements`,
    icon: {
      icon: mdiApps,
      color: "",
      iconSize: undefined,
    },
  },
  {
    name: t("objects.environment", 2),
    route: "environments",
    path: `/code/${repo?.Name}/${
      repoBranch.selectedBranch
        ? repoBranch.selectedBranch
        : repo?.DefaultBranch
    }/environments`,
    icon: {
      icon: mdiCloud,
      color: "",
      iconSize: undefined,
    },
  },
  {
    name: "Read me",
    route: `readme`,
    path: `/code/${repo?.Name}/readme`,
    icon: {
      icon: mdiLanguageMarkdown,
      color: "",
      iconSize: undefined,
    },
  },
  {
    name: "Settings",
    route: `settings`,
    path: `/code/${repo?.Name}/settings`,
    icon: {
      icon: mdiCog,
      color: "",
      iconSize: undefined,
    },
  },
]);
const push = (path: string) => {
  if (route.path !== path) spinner.spinner = true;
  router.push(path);
};
watch(
  () => repo,
  () => {
    if (repo && repo.Name && selectedBranch)
      api
        .get("/git-repo-file-list", {
          queries: {
            name: repo!.Name,
            branch: selectedBranch! as string,
          },
        })
        .then((res) => {
          includesReadme[repo!.Name] = !!res.filter((file) =>
            file.name.toLowerCase().includes("readme.md")
          ).length;
        });
  }
);
</script>
<template>
  <NavbarTabs :activeTab="currentTab">
    <template #tabs>
      <NavTab
        v-for="tab in repoTabs"
        :key="tab.name"
        :name="tab.route"
        @click="push(`${tab.path}`)"
      >
        <div
          v-if="tab.name === repo?.Name"
          style="display: flex; align-items: center"
        >
          <n-icon-wrapper
            :size="17"
            :border-radius="24"
            :color=" tab.icon!.color"
            style="margin-right: 4px"
          >
            <n-icon :size="tab.icon!.iconSize">
              <Mdi :path="tab.icon!.icon" />
            </n-icon>
          </n-icon-wrapper>
          <p>{{ tab.name }}</p>
        </div>
        <div v-else style="display: flex; align-items: center">
          <n-icon-wrapper
            :size="17"
            :border-radius="24"
            :color="'inherit'"
            style="margin-right: 4px"
          >
            <n-icon
              :size="tab.icon!.iconSize"
              :color="currentTab === tab.route ? 'white' : ''"
            >
              <Mdi :path="tab.icon!.icon" />
            </n-icon>
          </n-icon-wrapper>
          <p>{{ tab.name }}</p>
        </div>
      </NavTab>
    </template>
  </NavbarTabs>
</template>
