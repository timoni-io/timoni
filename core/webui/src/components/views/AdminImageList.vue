<script lang="ts" setup>
import { useI18n } from "vue-i18n";
import { z } from "zod";
import { ImageList } from "@/zodios/schemas/imageList";
import type { CollapseProps } from "naive-ui";

const { t } = useI18n();

type ImageObject = z.infer<typeof ImageList>;

let imagesList = ref<ImageObject>({});
let expandedNames = $ref<string[]>([]);
let imagesKeys = $ref<string[] | undefined>(undefined);
let clickedRow = $ref<string | null>(null);
let deleteModal = $ref<boolean>(false);
let deleteModalLoading = $ref<boolean>(false);
let deleteResponse = $ref<string>("");
let deleteResponseModal = $ref<boolean>(false);

const handleItemHeaderClick: CollapseProps["onItemHeaderClick"] = ({
  name,
  expanded,
}) => {
  if (!expanded) {
    expandedNames = expandedNames.filter((n) => {
      return n !== name;
    });
  } else {
    expandedNames.push(name);
  }
};

const handleClickRow = (id: string) => {
  id === clickedRow ? (clickedRow = null) : (clickedRow = id);
};

const handleDeleteUnusedImages = () => {
  deleteModalLoading = true;
  api.get("/image-delete-unused").then((res) => {
    deleteResponse = res;
  }).catch((err) => {
    deleteResponse = err;
  }).finally(() => {
    deleteModalLoading = false;
    deleteModal = false;
    deleteResponseModal = true;
  });
}

// window width
let screenWidth = $ref<number>(0);
onBeforeMount(() => {
  screenWidth = window.innerWidth;
});

onMounted(() => {
  api.get("/image-list").then((res) => {
    imagesList.value = res;
    imagesKeys = Object.keys(imagesList.value);
  });
  window.addEventListener("resize", () => {
    screenWidth = window.innerWidth;
  });
});
</script>

<template>
  <div>
    <AdminTabs />
    <PageLayout>
      <n-card class="admin-cards" size="small">
        <Spinner :data="imagesKeys">
          <n-space
            class="card-header"
            style="
              gap: none;
              justify-content: space-between;
              margin-bottom: 1rem;
            "
            v-if="imagesKeys?.length"
          >
            <div style="display: flex; gap: 0.5rem">
              <n-button
                secondary
                type="primary"
                size="tiny"
                @click="
                  () => {
                    if (imagesKeys) expandedNames = imagesKeys;
                  }
                "
                >{{ t("actions.expandAll") }}</n-button
              ><n-button
                secondary
                type="primary"
                size="tiny"
                @click="
                  () => {
                    expandedNames = [];
                  }
                "
                >{{ t("actions.collapseAll") }}</n-button
              >
              <n-button
                secondary
                type="error"
                size="tiny"
                @click="
                  () => {
                    deleteModal = true;
                  }
                "
                >{{ t("actions.deleteUnusedImages") }}
              </n-button>
            </div>
          </n-space>
          <n-collapse
            :expanded-names="expandedNames"
            @item-header-click="handleItemHeaderClick"
            v-if="imagesKeys?.length"
          >
            <n-collapse-item
              v-for="(val, key) in imagesList"
              :key="key"
              :title="(key as string)"
              :name="(key as string)"
            >
              <div
                v-for="image in val?.Images"
                :key="image.ID"
                class="interactive-row"
                :class="{ 'clicked-row': clickedRow === image.ID }"
                @click="handleClickRow(image.ID)"
              >
                <div>{{ image.ID }}</div>
                <div class="tags-container">
                  <n-tag
                    v-if="image?.BuildStatus && screenWidth > 1100"
                    size="tiny"
                    :bordered="false"
                  >
                    {{ t("fields.buildStatus") }}:
                    {{ image?.BuildStatus }}</n-tag
                  >
                  <n-tooltip
                    v-if="image?.BuildStatus && screenWidth <= 1100"
                    trigger="hover"
                  >
                    <template #trigger>
                      <n-tag size="tiny" :bordered="false">
                        {{
                          image?.BuildStatus.slice(0, 1).toUpperCase() +
                          image?.BuildStatus.slice(1)
                        }}
                      </n-tag>
                    </template>
                    {{ t("fields.buildStatus") }}: {{ image?.BuildStatus }}
                  </n-tooltip>
                  <n-tag
                    v-if="image?.SizeWithParents"
                    size="tiny"
                    :bordered="false"
                    >{{ t("fields.sizeWithParents") }}:
                    {{ image?.SizeWithParents }}</n-tag
                  >
                  <n-tag
                    v-if="image?.SizeWithoutParent"
                    size="tiny"
                    :bordered="false"
                    >{{ t("fields.sizeWithoutParent") }}:
                    {{ image?.SizeWithoutParent }}</n-tag
                  >
                </div>
              </div>
              <template #header-extra>
                <div class="tags-container" style="padding-right: 0.5rem">
                  <n-tag v-if="val?.ImagesCount" size="tiny" :bordered="false"
                    >{{ t("fields.numberOfImages") }}:
                    {{ val?.ImagesCount }}</n-tag
                  >
                  <n-tag
                    v-if="val?.ImagesSizeMBWithoutParent"
                    size="tiny"
                    :bordered="false"
                    >{{ t("fields.imagesSizeMBWithoutParent") }}:
                    {{ val?.ImagesSizeMBWithoutParent }}</n-tag
                  >
                </div>
              </template>
            </n-collapse-item>
          </n-collapse>
          <div v-else class="base-alert">
            {{ t("messages.imageListEmpty") }}
          </div>
        </Spinner>
      </n-card>
    </PageLayout>
  </div>
  <Modal
    v-model:show="deleteModal"
    :title="t('actions.deleteUnusedImages')"
    :touched="false"
    :showFooter="true"
    :loading="deleteModalLoading"
    style="width: auto; min-width: 30em;"
    @positive-click="handleDeleteUnusedImages"
  >
    {{ t("questions.deleteUnusedImages") }}
  </Modal>
  <Modal
    v-model:show="deleteResponseModal"
    :title="t('actions.deleteUnusedImages')"
    :touched="false"
    style="width: auto; min-width: 30em;"
  >
    <pre class="response-modal">
      {{ deleteResponse }}
    </pre>
  </Modal>
</template>

<style scoped lang="scss">
.tags-container {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
  justify-content: end;
}

.interactive-row {
  display: flex;
  align-items: center;
  gap: 1rem;
  justify-content: space-between;
  padding: 0.2rem 0.5rem;
  padding-left: 1.3rem;
  transition: all 0.2s ease-in-out;
  border-radius: 0.15rem;
  &:hover {
    color: var(--primaryColorHover);
    background-color: rgba(gray, 0.05);
  }
}
.clicked-row {
  background-color: rgba(#18a058, 0.15);
  color: var(--primaryColorHover);
  &:hover {
    background-color: rgba(#18a058, 0.15);
    color: var(--primaryColorHover);
  }
}

.response-modal {
  word-break: break-all;
  overflow-wrap: break-word;
  white-space: pre-line;
  max-height: 70vh;
  overflow-y: auto;
}
</style>
