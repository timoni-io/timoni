<script setup lang="ts">
import type { MenuOption } from "naive-ui";
import { renderIcon } from "@/utils/renderIcon";

const selectedItems = ref<string[]>([]);

const toggleItem = (item: string) => {
  selectedItems.value.includes(item)
    ? selectedItems.value.splice(selectedItems.value.indexOf(item), 1)
    : (selectedItems.value = [item]);
};

// @ts-ignore
const menuOptions = ref<MenuOption[]>([
  {
    label: "Zmiana wersji",
    key: "version",
    icon: renderIcon(mdiSwapVerticalBold),
  },
  {
    type: "divider",
    key: "divider-0",
  },
  {
    label: "Logi",
    key: "logs",
    icon: renderIcon(mdiCardText),
  },
  {
    label: "Git",
    key: "git",
    icon: renderIcon(mdiGit),
  },
  {
    label: "Historia",
    key: "history",
    icon: renderIcon(mdiHistory),
  },
  {
    label: "Manifest",
    key: "manifest",
    icon: renderIcon(mdiFileCog),
  },
  {
    label: "Kontenery",
    key: "containers",
    icon: renderIcon(mdiAppsBox),
  },
  {
    label: "Dockerfile",
    key: "logs",
    icon: renderIcon(mdiDocker),
  },
]);
</script>

<template>
  <navbar-start>
    <n-space align="center">
      <n-button tertiary @click="$router.push('/')">
        <template #icon>
          <n-icon>
            <mdi :path="mdiArrowLeft" />
          </n-icon>
        </template>
        Home
      </n-button>
      <h1>Środowisko: X-Wiki</h1>
      <n-button quaternary strong secondary type="primary" size="tiny">
        <template #icon>
          <n-icon><mdi :path="mdiDotsHorizontal" /></n-icon>
        </template>
      </n-button>
    </n-space>
  </navbar-start>
  <PageLayout>
    <div class="home-page">
      <!-- Elementy -->
      <div class="items-section">
        <n-card title="Elementy" size="small">
          <template #header-extra>
            <div class="header__extra">
              <n-input placeholder="Filtruj elementy" />
              <n-button strong secondary type="primary" size="tiny">
                <template #icon>
                  <n-icon><mdi :path="mdiPlus" /></n-icon>
                </template>
              </n-button></div
          ></template>
        </n-card>
        <n-layout has-sider sider-placement="right">
          <n-layout-content>
            <div class="items-grid">
              <ElementCard
                element-id="0"
                status="ok"
                :selected="selectedItems.includes('0')"
                @click="toggleItem('0')"
              />
              <ElementCard
                element-id="1"
                status="ok"
                :selected="selectedItems.includes('1')"
                @click="toggleItem('1')"
              />
              <ElementCard
                element-id="2"
                status="error"
                :selected="selectedItems.includes('2')"
                @click="toggleItem('2')"
              />
              <ElementCard
                element-id="3"
                status="ok"
                :selected="selectedItems.includes('3')"
                @click="toggleItem('3')"
              />
            </div>
          </n-layout-content>
          <n-layout-sider
            collapse-mode="width"
            :collapsed-width="0"
            :width="240"
            :collapsed="!selectedItems.length"
          >
            <div class="element-menu">
              <n-tag style="margin: 1rem; margin-bottom: 0px">postgresql</n-tag>
              <n-menu :value="''" :options="(menuOptions as MenuOption[])" />
            </div>
          </n-layout-sider>
        </n-layout>
      </div>
    </div>
  </PageLayout>
</template>

<style scoped>
h1 {
  padding: 0;
  margin: 0;
}
.home-page {
  display: grid;
  grid-template-columns: 1fr;
  gap: 1rem;
}
.items-section {
  width: 100%;
  min-width: 0;
}

.n-layout {
  padding: 0.5rem;
}
.items-grid {
  --colums: 4;
  display: grid;
  grid-template-columns: repeat(var(--colums), 1fr);
  gap: 1rem;
}

@media screen and (max-width: 1000px) {
  .items-grid {
    --colums: 2;
  }
}
@media screen and (max-width: 600px) {
  .items-grid {
    --colums: 1;
  }
}

.header__extra {
  display: flex;
  gap: 1rem;
}

.element-menu {
  overflow: hidden;
}
</style>
