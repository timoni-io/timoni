<script setup lang="ts">

const props = defineProps<{
  activeTab?: string;
}>();

let logo = `https://${location.hostname}:${location.port}/logo.svg`

provide("tab", props.activeTab);
// const openSpinner = () => {
//   console.log("spinner");
// };
</script>

<template>
  <div class="tabs" style="z-index: 10000">
    <div style="flex: 1; background: red; display: block; width: fit-content">
      <div>
        <div style="background: #fab900; min-width: 100%; height: 40px">
          <img
            :src="logo"
            alt=""
            width="70"
            style="margin-left: 15px; margin-top: 5px; opacity: 0.5"
          />
        </div>
      </div>
    </div>
    <div id="nav-tabs" style="display: flex">
      <!-- <slot name="tabs"></slot> -->
    </div>
    <div class="navbar-right"></div>
    <!-- <div style="flex: 0 1; background: red">
    </div>

    <div id="nav-tabs" style="display: flex">
      <slot name="tabs"></slot>
    </div>

    <div id="nav-suffix" style="background: red">
      <slot name="suffix"></slot>

    </div> -->
  </div>
</template>

<style lang="scss">
.navbar-right {
  flex: 1;
  display: flex;
  justify-content: flex-end;
  height: 40px;
  background-color: var(--accentColor);
}
.tabs {
  padding: 0;
  position: absolute;
  display: flex;
  align-items: center;
  justify-content: space-between;
  z-index: 10;
  top: 0;
  left: 0;
  width: 100%;
  margin-bottom: 1rem;
  backdrop-filter: blur(100px);
  // background: red;
  // background-color: var(--accentColor);

  & .n-tabs-tab__label {
    padding-left: 1rem;
    padding-right: 1rem;
    height: 100%;
  }

  & .n-tabs-tab-wrapper:has(.n-tabs-tab--active) {
    background-color: #191919;
    --n-tab-text-color-active: white;
  }

  .n-tabs-tab-wrapper {
    border-left: 1px solid #ebebeb1a;
    --n-tab-text-color: black;
    --n-tab-text-color-hover: black;
    border-left: 1px solid #31313152;
  }
  .n-tabs-tab-wrapper:hover {
    background-color: var(--accentColorDarken);
  }

  .n-tabs-tab-wrapper:last-child {
    border-right: 1px solid #31313152;
  }

  &__dark {
    background-color: #191919;
  }

  &__light {
    background-color: rgba(245, 245, 245, 0.9);
  }
}
</style>
