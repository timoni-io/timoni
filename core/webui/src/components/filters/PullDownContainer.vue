<script setup lang="ts">
import { useI18n } from "vue-i18n";

const defaultTagName = "Admin-1";
let openCreateFilterModal = $ref(false);
let openCreateCondition = $ref(false);
let selectName = $ref(null);
let selectCondition = $ref(null);
let selectValue = $ref(null);
let tagName = $ref(defaultTagName);
let searchTags = $ref("");
const { t } = useI18n();

const globalTags = $ref<
  Array<{
    name: string;
    favorite: boolean;
    conditions: Array<{
      name: string;
      condition: string;
      value: string | null;
    }>;
  }>
>([
  {
    name: "siema",
    favorite: false,
    conditions: [
      {
        name: "asd",
        condition: "dsa",
        value: "",
      },
    ],
  },
]);
const filteredGlobalTags = $computed(() => {
  return globalTags.filter((el) =>
    searchTags
      .trim()
      .split(" ")
      .some((tag) => el.name.includes(tag))
  );
});
let conditions = $ref<
  Array<{ name: string; condition: string; value: string | null }>
>([]);

const addCondition = () => {
  if (selectName && selectCondition) {
    conditions.push({
      name: selectName,
      condition: selectCondition,
      value: selectValue,
    });
    openCreateCondition = false;
  }
};

const openCreateConditionDialog = () => {
  selectName = null;
  selectCondition = null;
  selectValue = null;
  openCreateCondition = true;
};

const openCreateFilterDialog = () => {
  openCreateFilterModal = true;
  conditions = [];
  tagName = defaultTagName;
};

const optionsName = [
  {
    label: "Everybody's Got Something to Hide Except Me and My Monkey",
    value: "song0",
  },
  {
    label: "Drive My Car",
    value: "song1",
  },
  {
    label: "Norwegian Wood",
    value: "song2",
  },
  {
    label: "You Won't See",
    value: "song3",
  },
  {
    label: "Nowhere Man",
    value: "song4",
  },
];
const conditionName = [
  {
    label: "is",
    value: "is",
  },
  {
    label: "not is",
    value: "not is",
  },
  {
    label: "exist",
    value: "exist",
  },
  {
    label: "not exist",
    value: "not exist",
  },
];

const createTag = () => {
  if (conditions.length) {
    globalTags.push({
      favorite: false,
      name: tagName,
      conditions,
    });
    conditions = [];

    tagName = defaultTagName;
    openCreateFilterModal = false;
  }
};
type iDontKnow = unknown;
const changeFav = (name: iDontKnow) => {
  return name;
};
</script>
<template>
  <Dropdown>
    <template #trigger>
      <n-button type="primary" strong secondary size="tiny">
        <n-icon size="14">
          <mdi :path="mdiPail" />
        </n-icon>
        <span style="padding-left: 0.2rem">{{ t("navbar.filters") }}</span>
      </n-button>
    </template>
    <template #content>
      <div class="container">
        <n-input
          style="width: 100%; margin-bottom: 10px"
          placeholder="Search"
          v-model:value="searchTags"
        />
        <n-tag
          type="primary"
          v-for="condition in filteredGlobalTags"
          :key="condition.name"
          style="
            width: 100%;
            margin-bottom: 10px;
            display: flex;
            justify-content: center;
            position: relative;
            cursor: pointer;
          "
          class="tag-element"
        >
          <div style="display: flex; align-items: center; width: 100%">
            {{ condition.name }}
          </div>
          <p
            @click.prevent="changeFav(condition.name)"
            class="star-icon"
            style="position: absolute; right: 5px; top: calc(50% - 10px)"
          >
            <n-icon class="star" size="18"><mdi :path="mdiStar" /></n-icon>
            <n-icon class="star-outline" size="18"
              ><mdi :path="mdiStarOutline"
            /></n-icon>
          </p>
        </n-tag>
        <n-button
          type="tertiary"
          class="add-button"
          @click="openCreateFilterDialog"
        >
          <div class="button-content">
            <n-icon>
              <mdi :path="mdiPlusThick" />
            </n-icon>
            Add filter
          </div>
        </n-button>
      </div>
    </template>
  </Dropdown>
  <Modal
    class="modal-container"
    v-model:show="openCreateFilterModal"
    title="Create filter"
    style="width: 30rem"
    :touched="tagName !== defaultTagName || conditions.length > 0"
  >
    <div class="input-container">
      <p style="margin-bottom: 10px">Filter name</p>
      <n-input
        v-model:value="tagName"
        style="margin-bottom: 10px"
        placeholder=""
      />
    </div>
    <n-tag
      type="primary"
      v-for="condition in conditions"
      :key="condition.name"
      style="
        width: 100%;
        text-align: center;
        margin-bottom: 10px;
        display: flex;
        justify-content: center;
      "
    >
      {{ condition.name }}
    </n-tag>
    <!-- <div v-if="!openCreateCondition"> -->
    <div class="modal-container">
      <n-button
        type="tertiary"
        class="add-button"
        @click="openCreateConditionDialog"
      >
        <div class="button-content">
          <n-icon>
            <mdi :path="mdiPlusThick" />
          </n-icon>
          Create condition
        </div>
      </n-button>
    </div>

    <n-button
      type="tertiary"
      @click="createTag"
      style="
        display: block;
        margin-right: 0;
        margin-left: auto;
        margin-top: 20px;
      "
    >
      <div class="button-content">
        <n-icon>
          <mdi :path="mdiContentSave" />
        </n-icon>
        <p>Save</p>
      </div>
    </n-button>
  </Modal>

  <Modal
    class="modal-container"
    v-model:show="openCreateCondition"
    title="Create condition"
    style="width: 23rem"
    :touched="selectName !== null"
  >
    <div class="input-container">
      <p>Field</p>
      <n-select
        v-model:value="selectName"
        filterable
        :options="optionsName"
        :clearable="false"
      />
    </div>

    <div v-if="selectName" class="input-container">
      <p>Operator</p>
      <n-select
        v-model:value="selectCondition"
        filterable
        :options="conditionName"
        :clearable="false"
      />
    </div>

    <div
      v-if="
        selectName && (selectCondition === 'is' || selectCondition === 'not is')
      "
      class="input-container"
    >
      <p>Value</p>
      <n-select
        v-model:value="selectValue"
        filterable
        :options="optionsName"
        :clearable="false"
      />
    </div>

    <n-button
      type="tertiary"
      class="add-button"
      @click="addCondition"
      style="margin-top: 20px"
      v-if="
        selectName &&
        selectCondition &&
        (selectValue ||
          selectCondition === 'exist' ||
          selectCondition === 'not exist')
      "
    >
      <div class="button-content">
        <n-icon>
          <mdi :path="mdiPlusThick" />
        </n-icon>
        Add Condition
      </div>
    </n-button>
  </Modal>
</template>
<style scoped lang="scss">
.container {
  width: 200px;
  padding: 4px 8px;
}
.add-button {
  width: 100%;
}
.button-content {
  display: flex;
  gap: 5px;
  align-items: center;
  justify-content: center;
}
.modal-container {
  width: 100%;
}
.tag-element:hover .star {
  display: block;
}
.star {
  display: none;
}
.tag-element:hover .star-outline {
  display: none;
}
.input-container {
  margin-top: 10px;
  display: flex;
  gap: 10px;
  align-items: center;
  justify-content: center;
  & p {
    white-space: nowrap;
    flex: 0 22%;
  }
}
</style>
