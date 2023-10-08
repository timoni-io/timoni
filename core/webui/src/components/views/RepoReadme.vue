<script lang="ts" setup>
import { useUserStore } from "@/store/userStore";

const userStore = useUserStore();
</script>
<template>
  <RepoTabs />
  <PageLayout>
    <RepoMd v-if="userStore.havePermission('Repo_View')" />
    <ErrorDialog v-else :msg="$t('messages.permissionDenied')" />
  </PageLayout>
</template>
