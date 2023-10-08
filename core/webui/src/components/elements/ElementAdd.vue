<script setup lang="ts">
// import Configurator from "@/components/configurator/ElementAddRemote.vue";

enum ConfiguratorMode {
  REMOTE = "remote",
  LOCAL = "local",
  SCRATCH = "scratch",
  ELEMENTHUB = "elementhub",
  NONE = "",
}

const emit = defineEmits(["creatingDone", "isTouched"]);

let addingMode = $ref<ConfiguratorMode>(ConfiguratorMode.NONE);
// let title = $ref("");

const setCurrentMode = (mode: ConfiguratorMode): void => {
  if (
    (mode === ConfiguratorMode.REMOTE || mode === ConfiguratorMode.LOCAL) &&
    document.querySelector(".n-card-header__main")
  ) {
    // title = document.querySelector(".n-card-header__main")!.innerHTML;
    document.querySelector(".n-card-header__main")!.innerHTML = "";
  }
  addingMode = mode;
};

onBeforeMount(() => {
  setCurrentMode(ConfiguratorMode.ELEMENTHUB);
});

const done = (data: string) => {
  emit("creatingDone", data);
};

// const back = () => {
//   addingMode = ConfiguratorMode.NONE;
//   document.querySelector(".n-card-header__main")!.innerHTML = title;
// };

const isTouched = () => {
  emit("isTouched");
};
</script>

<template>
  <div v-if="addingMode === ConfiguratorMode.NONE">
    <div class="start-step">
      <div class="start-step__box">
        <n-button size="small" @click="setCurrentMode(ConfiguratorMode.REMOTE)">
          <template #icon>
            <n-icon> <mdi :path="mdiRemote" /> </n-icon>
          </template>

          Remote
        </n-button>

        Import the code of an existing project to the cloud, the code can be
        downloaded from remote Git repository.
      </div>

      <div class="start-step__box">
        <n-button size="small" @click="setCurrentMode(ConfiguratorMode.LOCAL)">
          <template #icon>
            <n-icon> <mdi :path="mdiCoffee" /> </n-icon>
          </template>

          Local
        </n-button>

        Use element from local project.
      </div>

      <div class="start-step__box">
        <n-button
          size="small"
          @click="setCurrentMode(ConfiguratorMode.SCRATCH)"
        >
          <template #icon>
            <n-icon> <mdi :path="mdiPlus" /> </n-icon>
          </template>

          From scratch
        </n-button>

        Create brand new element from scratch.
      </div>
      <div class="start-step__box">
        <n-button
          size="small"
          @click="setCurrentMode(ConfiguratorMode.ELEMENTHUB)"
        >
          <template #icon>
            <n-icon> <mdi :path="mdiHubspot" /> </n-icon>
          </template>

          ElementHub
        </n-button>

        To będzie ficzer.
      </div>
    </div>
  </div>
  <ElementAddElementHub
    v-else-if="addingMode === ConfiguratorMode.ELEMENTHUB"
    @advanced="addingMode = ConfiguratorMode.NONE"
    @creatingDone="done"
    @isTouched="isTouched"
  />
  <!-- <Configurator v-else-if="addingMode === ConfiguratorMode.REMOTE" /> -->
  <!-- <ElementAddRemote
    v-else-if="addingMode === ConfiguratorMode.REMOTE"
    @creatingDone="done"
    @back="back"
  />
  <ElementAddLocal
    v-else-if="addingMode === ConfiguratorMode.LOCAL"
    @creatingDone="done"
    @back="back"
  /> -->
</template>

<style lang="scss" scoped>
.start-step {
  margin-bottom: 24px;

  &__box {
    display: grid;
    grid-template-columns: 125px 1fr;
    gap: 26px;

    align-items: center;

    margin-bottom: 16px;

    &:last-child {
      margin-bottom: 0;
    }
  }
}
</style>
