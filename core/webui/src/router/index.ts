import { createRouter, createWebHistory } from "vue-router";

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: "/",
      redirect: () => {
        return { path: "/env" };
      },
    },
    {
      name: "Home",
      path: "/env",
      component: () => import("@/components/views/Home.vue"),
    },
    {
      name: "Environment",
      path: "/env/:id",
      component: () => import("@/components/views/EnvironmentPage.vue"),
    },
    {
      name: "Environment elements",
      path: "/env/elements/:id",
      component: () => import("@/components/views/EnvironmentElements.vue"),
    },
    {
      name: "Environment pods",
      path: "/env/pods/:id",
      component: () => import("@/components/views/EnvironmentPods.vue"),
    },
    {
      name: "Environment statistics",
      path: "/env/statistics/:id",
      component: () => import("@/components/views/EnvironmentStatistics.vue"),
    },
    {
      name: "Environment metrics",
      path: "/env/metrics/:id",
      component: () => import("@/components/views/EnvironmentMetrics.vue"),
    },
    {
      name: "Environment activities",
      path: "/env/history/:id",
      component: () => import("@/components/views/EnvironmentHistory.vue"),
    },
    {
      name: "Environment infra",
      path: "/env/infrastructure/:id",
      component: () => import("@/components/views/EnvironmentInfra.vue"),
    },
    {
      name: "Environment inputs",
      path: "/env/variables/:id",
      component: () => import("@/components/views/EnvironmentInputs.vue"),
    },
    {
      name: "Environment topology",
      path: "/env/topology/:id",
      component: () => import("@/components/views/EnvironmentTopology.vue"),
    },
    // {
    //   name: "Environment settings",
    //   path: "/env/settings/:id",
    //   component: () => import("@/components/views/EnvironmentSettings.vue"),
    // },
    {
      name: "Environment members",
      path: "/env/members/:id",
      component: () => import("@/components/views/EnvironmentOperators.vue"),
    },
    {
      name: "admin-platform",
      path: "/admin/platform",
      component: () => import("@/components/views/AdminPlatform.vue"),
    },
    {
      name: "admin-users-list",
      path: "/users-list",
      component: () => import("@/components/views/AdminUsersList.vue"),
    },
    {
      name: "admin-teams-list",
      path: "/teams-list",
      component: () => import("@/components/views/AdminTeamsList.vue"),
    },
    {
      name: "admin-image-list",
      path: "/admin/image-list",
      component: () => import("@/components/views/AdminImageList.vue"),
    },
    {
      name: "admin-logs",
      path: "/admin/logs",
      component: () => import("@/components/views/AdminLogs.vue"),
    },
    {
      name: "element-versions",
      path: "/admin/element-versions",
      component: () => import("@/components/views/AdminElementVersions.vue"),
    },
    {
      name: "admin-histogram",
      path: "/admin/histogram",
      component: () => import("@/components/views/AdminHistogram.vue"),
    },
    //
    {
      name: "admin-status",
      path: "/admin/status",
      component: () => import("@/components/views/AdminStatus.vue"),
    },
    // {
    //   name: "admin-certificates",
    //   path: "/admin/certificates",
    //   component: () => import("@/components/views/AdminCertificates.vue"),
    // }, commented after removal of '/api/system-certs'
    {
      name: "admin-building-images",
      path: "/admin/building-images",
      component: () => import("@/components/views/AdminBuildingImages.vue"),
    },
    {
      name: "admin-resources",
      path: "/admin/resources",
      component: () => import("@/components/views/AdminResources.vue"),
    },
    {
      name: "admin-login",
      path: "/login",
      component: () => import("@/components/views/Login.vue"),
    },
    {
      name: "admin-user",
      path: "/user",
      component: () => import("@/components/views/User.vue"),
    },
    {
      name: "repositories",
      path: "/code",
      component: () => import("@/components/views/Repositories.vue"),
    },
    {
      name: "repo",
      path: "/code/:name",
      component: () => import("@/components/views/RepoPage.vue"),
    },
    {
      name: "commits",
      path: "/code/:name/:branch/commits",
      component: () => import("@/components/views/RepoCommits.vue"),
    },
    {
      name: "elements",
      path: "/code/:name/:branch/elements",
      component: () => import("@/components/views/RepoElements.vue"),
    },
    {
      name: "repoEnvironments",
      path: "/code/:name/:branch/environments",
      component: () => import("@/components/views/RepoEnvironments.vue"),
    },
    {
      name: "repository-files",
      path: "/code/:name/:branch/files/:filepath*",
      component: () => import("@/components/views/RepoFiles.vue"),
    },
    {
      name: "repository-readme",
      path: "/code/:name/readme",
      component: () => import("@/components/views/RepoReadme.vue"),
    },
    {
      name: "repository-settings",
      path: "/code/:name/settings",
      component: () => import("@/components/views/RepoSettings.vue"),
    },
    {
      name: "coming-soon",
      path: "/coming-soon",
      component: () => import("@/components/views/ComingSoon.vue"),
    },
  ],
});

export default router;
