<script setup lang="ts">
// import { useI18n } from "vue-i18n";
// import { ComputedRef } from "vue";
// import { Column } from "@/components/utils/DataTable.vue";
// import { useTimeFormatter } from "@/utils/formatTime";
// import { Check } from "../../zodios/schemas/system";
// import { z } from "zod";

// const { t } = useI18n();
// const { relativeOrDistanceToNow, formatTime } = useTimeFormatter();

let columns = $ref([]);
let data = $ref([]);

api.get("/system-element-versions").then((res) => {
    columns = res.Columns;
    data = res.Data;

    // console.log(columns);
    // console.log(data);

});


</script>

<template>
    <div>
        <AdminTabs />
        <PageLayout>
            <div class="admin-cards">
                <n-card title="Element versions" size="small">
                    <Spinner :data="data">

                        <n-data-table :bordered="false" :columns="columns" :data="data" class="data-table" />

                    </Spinner>
                </n-card>
            </div>
        </PageLayout>
    </div>
</template>

<style lang="scss">
.data-table a {
    width: 100%;
    text-decoration: none;
    color: inherit;
    display: block;
}
</style>