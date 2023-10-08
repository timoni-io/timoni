<template>
  <div style="text-align: center">
    <p>{{ $t("messages.viewportSizeNotSupported") }}</p>
    <p>{{ $t("messages.minimumScreenWidthIs") }}: 1200px</p>
  </div>
</template>

<style scoped>
div {
  position: absolute;
  inset: 0;
  background-color: black;
  z-index: 20000;
  display: grid;
  place-content: center;
  font-size: large;
}

@media screen and (min-width: 1200px) {
  div {
    display: none;
  }
}
</style>
