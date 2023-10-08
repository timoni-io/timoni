<script setup lang="ts">
import { useSlots, computed, h } from "vue";
import { DataTableColumns } from "naive-ui";
import { RouterLink } from "vue-router";
import pluck from "@/utils/pluck";

export type Column = {
  title: string;
  key?: string;
  template?: string;
  linkTo?: (o: any) => string;
  render?: Function;
} & Partial<DataTableColumns[number]>;

const props = defineProps<{
  columns: Column[];
  data?: object[];
  rowProps?: (row: { [key: string]: string }) => {
    style: string;
    onClick: () => void;
  };
}>();
const slots = useSlots();

const columnsWithTemplates = computed(() =>
  props.columns.map(({ key, template, linkTo, render, ...col }, idx) => {
    const slot = (template && slots[template]) || undefined;
    return {
      ...col,
      ...(slot
        ? {
            render: (row: any): any => slot?.(row),
          }
        : { key }),
      ...(linkTo && !slot
        ? {
            render: (row: any): any => {
              return h(
                RouterLink,
                { to: linkTo(row) },
                {
                  default: () => {
                    return (
                      row && key && (render?.(row, idx) || pluck(row, key))
                    );
                  },
                }
              );
            },
          }
        : { key }),
    };
  })
);
</script>

<template>
  <n-data-table
    :columns="(columnsWithTemplates as any)"
    :data="data"
    :row-props="rowProps"
    :bordered="false"
    class="data-table"
  />
</template>
<style lang="scss">
.data-table a {
  width: 100%;
  text-decoration: none;
  color: inherit;
  display: block;
}
</style>
