<script lang="ts" setup>
// imports
import { onMounted } from "vue";
import { ElementCommit, BranchList, TagList } from "@/zodios/schemas/changeVersion";
import { ElementMapRespExtended } from "@/zodios/schemas/elements";
import { z } from "zod";
import { useI18n } from "vue-i18n";
import { useRoute } from "vue-router";
import { useMessage } from "naive-ui";
import { ElementVersionsResponse } from "@/zodios/schemas/env";
import { useTimeFormatter } from "@/utils/formatTime";

const message = useMessage();
const { t } = useI18n();
const route = useRoute();
const { relativeOrDistanceToNow, formatTime } = useTimeFormatter();

// props
const props = defineProps<{
  elementName: string;
  element: z.infer<typeof ElementMapRespExtended> | null;
}>();

const emit = defineEmits<{
  (e: "closeDialog"): void;
}>();
// commits and branch list
type ElementCommits = z.infer<typeof ElementCommit>;
type BranchList2 = z.infer<typeof BranchList>;
type TagList2 = z.infer<typeof TagList>;

let elementCommits = $ref<ElementCommits[] | []>([]);
let branchList = $ref<BranchList2 | []>([]);
let tagList = $ref<TagList2 | []>([]);
let currentBranch = $ref<string>(props.element!.Info.SourceGit.BranchName);
let elementVersions = $ref<z.infer<typeof ElementVersionsResponse>>({});
let listFromBranchesAndCommits = $ref(true);
// get commits and branchs

const getElementCommits = () => {
  api
    .get("/env-element-commit-list", {
      queries: {
        env: route.params.id as string,
        branch: currentBranch as string,
        element: props.elementName as string,
      },
    })
    .then((res) => {
      elementCommits = res.Commits || [];
    });
};
api
  .get("/env-element-versions", {
    queries: {
      env: route.params.id as string,
      element: props.elementName,
    },
  })
  .then((res) => {
    elementVersions = res;
  });



const getBranchList = () => {
  api
    .get("/git-repo-branch-list", {
      queries: {
        name: props.element!.Info.SourceGit.RepoName,
        level: 2,
      },
    })
    .then((res) => {
      branchList = res;
      if(!branchList.includes(currentBranch)) {
        currentBranch = branchList[0];
      }
    });
};


const getTagList = () => {
  api
    .get("/git-repo-tag-list", {
      queries: {
        name: props.element!.Info.SourceGit.RepoName,
        level: 2,
      },
    })
    .then((res) => {
      tagList = res;
    });
};




const branchOptions = computed(() => {
  if (typeof branchList === "string") {
    return [{ label: branchList, value: branchList }];
  }
  return branchList.map((branch: string) => {
    return { label: branch, value: branch };
  });
});

// watch

watch(
  () => currentBranch,
  (newBranch, oldBranch) => {
    if (newBranch !== oldBranch) {
      getElementCommits();
    }
  }
);

// mounted
onMounted(() => {
  getElementCommits();
  getBranchList();
  getTagList();
});

// time transform
// const transformUnixTime = (time: number) => {
//   let date = new Date(time * 1000);
//   let months = [
//     t("time.monthsShort.Jan"),
//     t("time.monthsShort.Feb"),
//     t("time.monthsShort.Mar"),
//     t("time.monthsShort.Apr"),
//     t("time.monthsShort.May"),
//     t("time.monthsShort.Jun"),
//     t("time.monthsShort.Jul"),
//     t("time.monthsShort.Aug"),
//     t("time.monthsShort.Sep"),
//     t("time.monthsShort.Oct"),
//     t("time.monthsShort.Nov"),
//     t("time.monthsShort.Dec"),
//   ];
//   return (
//     date.getDate() +
//     " " +
//     months[date.getMonth()] +
//     " " +
//     date.getFullYear() +
//     " " +
//     date.getHours() +
//     ":" +
//     date.getMinutes() +
//     ":" +
//     date.getSeconds()
//   );
// };
let showConfirmModal = $ref(false);
let selectedCommit: string;
const selectCommit = (sha: string): void => {
  selectedCommit = sha;
  showConfirmModal = true;
};
const changeVersion = () => {
  api
    .get("/env-element-version-change", {
      queries: {
        env: route.params.id as string,
        element: props?.elementName as string,
        commit: selectedCommit as string,
        branch: currentBranch as string,
      },
    })
    .then((res) => {
      showConfirmModal = false;
      emit("closeDialog");
      if (res === "ok") {
        message.success(t("messages.elementVersionChange"));
      } else {
        message.error(res as string);
      }
    });
};

const when = (current: boolean, previous: boolean): string => {
  if (current) return t("commits.current");
  if (previous) return t("commits.previous");
  return t("commits.former");
};

const getStatus = (status: number) => {
  switch (true) {
    case status === 0:
      return "creating";
    case status === 1:
      return "deploying";
    case status === 2:
      return "failed";
    case status === 3:
      return "runningReady";
    case status === 4:
      return "terminating";
    case status === 5:
      return "building";
    case status === 6:
      return "disabled";
    case status === 7:
      return "createFailed";
    default:
      return "unknownStatus";
  }
};
</script>

<template>
  <div style="height: 78vh">

    <div v-if="tagList.length > 0" style="padding-bottom:30px">
      <n-space vertical>
        <n-radio-group v-model:value="listFromBranchesAndCommits" size="small">
          <n-radio-button :value="true">{{ t("fields.BranchesAndCommits") }}</n-radio-button>
          <n-radio-button :value="false">{{ t("fields.Tags") }}</n-radio-button>
        </n-radio-group>
      </n-space>
    </div>

    <n-select v-model:value="currentBranch" :options="branchOptions" style="margin-bottom: 10px"
      v-if="listFromBranchesAndCommits" />
    <n-scrollbar style="height: 66vh" v-if="listFromBranchesAndCommits">

      <n-table v-if="elementCommits.length" :bordered="false" :single-line="true">
        <thead>
          <tr>
            <th>{{ t("fields.author") }}</th>
            <th>{{ t("fields.time") }}</th>
            <th>Hash</th>
            <th>{{ t("fields.message") }}</th>
            <th style="min-width: 240px">
              {{ t("fields.version") }} / {{ t("fields.status") }}
            </th>
            <th>{{ t("fields.actions") }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="commit in elementCommits" :key="commit.SHA" :class="
            elementVersions[commit.SHA] && elementVersions[commit.SHA].Current
              ? 'current'
              : ''
          ">
            <td>
              <n-tooltip trigger="hover">
                <template #trigger>
                  <n-tag round type="info">
                    {{ commit.AuthorInitials }}
                  </n-tag>
                </template>
                {{ commit.AuthorEmail }}
              </n-tooltip>
            </td>
            <td>
              {{ relativeOrDistanceToNow(new Date(commit.TimeStamp * 1000)) }}
            </td>
            <td>{{ commit.SHA.slice(0, 8) }}</td>
            <td>{{ commit.Message }}</td>
            <td>
              <n-tooltip v-if="elementVersions[commit.SHA]" placement="top" trigger="hover">
                <template #trigger>
                  <div style="display: flex; flex-flow: nowrap">
                    <p v-if="elementVersions[commit.SHA].Current">
                      {{ t("commits.current") }}&nbsp;•&nbsp;
                    </p>
                    <p v-else-if="elementVersions[commit.SHA].Previous">
                      {{ t("commits.previous") }}&nbsp;•&nbsp;
                    </p>
                    <p v-else>{{ t("commits.former") }}&nbsp;•&nbsp;</p>
                    <!-- {{
                              moment(
                                elementVersions[commit.SHA].Version * 1000
                              ).fromNow()
                            }} -->
                    <p>
                      {{
                        relativeOrDistanceToNow(
                          new Date(
                            elementVersions[commit.SHA].SaveTimestamp * 1000
                          )
                        )
                      }}
                    </p>
                  </div>
                </template>
                <p>
                  <strong>{{
                    when(
                      elementVersions[commit.SHA].Current,
                      elementVersions[commit.SHA].Previous
                    )
                  }}
                    status:</strong>
                  {{
                    t(
                      `elements.status.${getStatus(
                        elementVersions[commit.SHA].Status
                      )}`
                    )
                  }}
                </p>
                <p>
                  <strong>{{ t("fields.releaseDate") }}:</strong>
                  {{
                    formatTime(
                      new Date(elementVersions[commit.SHA].SaveTimestamp * 1000)
                    )
                  }}
                </p>
                <p>
                  <strong>{{ t("fields.userEmail") }}:</strong>
                  {{ elementVersions[commit.SHA].UserEmail }}
                </p>
              </n-tooltip>
            </td>
            <td>
              <n-button strong secondary type="primary" v-if="element!.Info.SourceGit.CommitHash !== commit.SHA"
                @click="() => selectCommit(commit.SHA)">
                {{ t("actions.apply") }}
              </n-button>
            </td>
          </tr>
        </tbody>
      </n-table>
      <div v-else style="
                  display: flex;
                  justify-content: center;
                  align-items: center;
                  height: 50vh;
                ">
        <n-spin :size="80" />
      </div>

    </n-scrollbar>
    <n-scrollbar v-else style="height: 66vh">

      <!-- tags: -->

      <n-table :bordered="false" :single-line="true">
        <thead>
          <tr>
            <th>{{ t("fields.author") }}</th>
            <th>{{ t("fields.time") }}</th>
            <th>Tag</th>
            <th style="min-width: 240px">
              {{ t("fields.version") }} / {{ t("fields.status") }}
            </th>
            <th>{{ t("fields.actions") }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="tag in tagList" :key="tag.Name" :class="
            elementVersions[tag.CommitSHA] && elementVersions[tag.CommitSHA].Current
              ? 'current'
              : ''
          ">
            <td>
              <n-tooltip trigger="hover">
                <template #trigger>
                  <n-tag round type="info">
                    {{ tag.CommitAuthorInitials }}
                  </n-tag>
                </template>
                {{ tag.CommitAuthorEmail }}
              </n-tooltip>
            </td>
            <td>
              {{ relativeOrDistanceToNow(new Date(tag.TimeStamp * 1000)) }}
            </td>
            <td>{{ tag.Name }}</td>
            <td>


              <n-tooltip v-if="elementVersions[tag.CommitSHA]" placement="top" trigger="hover">
                <template #trigger>
                  <div style="display: flex; flex-flow: nowrap">
                    <p v-if="elementVersions[tag.CommitSHA].Current">
                      {{ t("commits.current") }}&nbsp;•&nbsp;
                    </p>
                    <p v-else-if="elementVersions[tag.CommitSHA].Previous">
                      {{ t("commits.previous") }}&nbsp;•&nbsp;
                    </p>
                    <p v-else>{{ t("commits.former") }}&nbsp;•&nbsp;</p>
                    <p>
                      {{
                        relativeOrDistanceToNow(
                          new Date(
                            elementVersions[tag.CommitSHA].SaveTimestamp * 1000
                          )
                        )
                      }}
                    </p>
                  </div>
                </template>
                <p>
                  <strong>{{
                    when(
                      elementVersions[tag.CommitSHA].Current,
                      elementVersions[tag.CommitSHA].Previous
                    )
                  }}
                    status:</strong>
                  {{
                    t(
                      `elements.status.${getStatus(
                        elementVersions[tag.CommitSHA].Status
                      )}`
                    )
                  }}
                </p>
                <p>
                  <strong>{{ t("fields.releaseDate") }}:</strong>
                  {{
                    formatTime(
                      new Date(elementVersions[tag.CommitSHA].SaveTimestamp * 1000)
                    )
                  }}
                </p>
                <p>
                  <strong>{{ t("fields.userEmail") }}:</strong>
                  {{ elementVersions[tag.CommitSHA].UserEmail }}
                </p>
              </n-tooltip>


            </td>
            <td>
              <n-button strong secondary type="primary" v-if="element!.Info.SourceGit.CommitHash !== tag.CommitSHA"
                @click="() => selectCommit(tag.CommitSHA)">
                {{ t("actions.apply") }}
              </n-button>
            </td>
          </tr>
        </tbody>
      </n-table>

    </n-scrollbar>
    <Modal v-model:show="showConfirmModal" style="width: 20rem" :title="t('elements.actions.changeVersion')" showFooter
      @positive-click="changeVersion" @negative-click="showConfirmModal = false">
      <div>
        {{ t("questions.sure") }}
      </div>
    </Modal>
  </div>
</template>
<style scoped>
.current td {
  background: #37504d;
}

.table-container {
  max-height: 75vh;
  overflow-x: hidden;
  overflow-y: scroll;
}
</style>
