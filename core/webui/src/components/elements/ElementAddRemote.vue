<script setup lang="ts">
import { useRoute } from "vue-router";
import { ComputedRef } from "vue";
import { Column } from "@/components/utils/DataTable.vue";
import { useI18n } from "vue-i18n";
import { useMessage } from "naive-ui";
// import { useEnv } from "@/store/envStore";

export interface IInput {
  Type: string;
  Message: string;
  Value: string;
  Default: string;
  MatchRegEx: string;
  Min: number;
  Max: number;
  Options: string[] | null;
}

type ProjectMap = ResType<"/git-repo-map">;

type Commit = {
  SHA: string;
  Message: string;
  Date: string;
  TimeStamp: number;
  AuthorName: string;
  AuthorEmail: string;
  AuthorInitials: string;
  Files: string[] | null;
};

const { t } = useI18n();
const emit = defineEmits(["creatingDone", "back"]);

interface IInputs {
  [k: string]: IInput;
}

interface IInputsReq {
  [k: string]: string;
}

const route = useRoute();
const message = useMessage();

// dodawanie elementu
const currentRef = ref<number>(1);

// const env = useEnv(computed(() => route.params.id as string));

let projects = ref<string[]>([]);
let branches = ref<string[] | string>([]);
let commits = ref<Commit[]>([]);
let elements = ref<string[]>([]);
let inputs = ref<IInputs>({});
let inputsReq = ref<IInputsReq>({});

let project = ref("");
let projectURL = ref("");
let branch = ref("");
let commit = ref("");
let element = ref("");
let errorInputs = ref("");
let errorProjects = ref("");
let errorCreating = ref("");

let formProject = ref("");
let elementName = ref("");
let elementNameTry = ref("");

let nameBusy = $ref(false);
let noName = $ref(false);

let formName = ref("");
let formBranch = ref("");
let formLogin = ref("");
let formPassword = ref("");
let firstSection = $ref(true);

let projectMap = $ref<ProjectMap>({});
const lol = $ref(null as HTMLInputElement | null);

const columns: ComputedRef<Column[]> = computed(() => [
  {
    title: t("time.date"),
    key: "Date",
    width: "25%",
  },
  {
    title: t("fields.author"),
    key: "AuthorName",
    width: "20%",
  },
  {
    title: "Commit",
    template: "commit",
    width: "15%",
  },
  {
    title: t("fields.message"),
    key: "Message",
    width: "40%",
  },
]);

// wybierania
const chooseProject = (projectName: string) => {
  if (firstSection) {
    api
      .get("/git-repo-map", {})
      .then((res) => {
        projectMap = res;
      })
      .then(() => {
        if (projectMap[projectName] !== undefined) {
          api
            .get("/git-repo-branch-list", {
              queries: {
                name: projectName,
                level: 2,
              },
            })
            .then((d) => {
              if (typeof d === "string") {
                errorProjects.value = d;
              } else {
                project.value = projectName;
                errorProjects.value = "";
                currentRef.value += 1;
                branches.value = d;
              }
            });
        } else {
          firstSection = false;
          projectURL.value = projectName;
        }
      });
  } else {
    api
      .post("/git-repo-create-remote", {
        URL: projectURL.value,
        // Branch: formBranch.value,
        Name: formName.value,
        // Public: false,
        Login: formLogin.value,
        Password: formPassword.value,
      })
      .then((res) => {
        if (typeof res === "string") {
          message.error("nieznany błąd xD");
          errorProjects.value = res;
        } else {
          api
            .get("/git-repo-branch-list", {
              queries: {
                name: formName.value,
                level: 2,
              },
            })
            .then((d) => {
              if (typeof d === "string") {
                errorProjects.value = d;
              } else {
                project.value = formName.value;
                errorProjects.value = "";
                currentRef.value += 1;
                branches.value = d;
              }
            });
        }
      });
  }
};

const chooseBranch = (branchName: string) => {
  api
    .get("/git-repo-commit-list", {
      queries: {
        name: project.value,
        branch: branchName,
      },
    })
    .then((d) => {
      branch.value = branchName;
      commits.value = d;
      // chooseCommit(commits.value[0].SHA, true);
    });
};

// const chooseCommit = (commitHash: string, skip: boolean) => {
//   api
//     .get("/git-repo-element-list", {
//       queries: {
//         name: project.value!,
//         branch: branch.value,
//         commit: commitHash,
//       },
//     })
//     .then((d) => {
//       elements.value = d.map((el) => {
//         return el["Element"];
//       });
//       commit.value = commitHash;
//       currentRef.value += skip ? 2 : 1;
//     });
// };

const chooseElement = (elementSelected: string) => {
  api
    .get("/env-create-inputs-check", {
      queries: {
        name: project.value,
        branch: branch.value,
        commit: commit.value,
        env: route.params.id as string,
        element: elementSelected,
        project: project.value,
      },
    })
    .then((res) => {
      if (typeof res === "string") {
        errorInputs.value = res;
      } else {
        inputs.value = res;
        element.value = elementSelected;
        elementName.value = elementSelected;
        errorInputs.value = "";
        currentRef.value += 1;
      }
    });
};

interface IEnvModReq {
  DeleteElements: string[];
  Elements: {
    [k: string]: {
      FileName: string;
      Project: string;
      Branch: string;
      Commit: string;
      Tags: {
        [k: string]: unknown;
      };
      Inputs: {
        [k: string]: unknown;
      };
      [k: string]: unknown;
    };
  };
  Apply: boolean;
}

const createElement = () => {
  let postReq: IEnvModReq = {
    DeleteElements: [],
    Elements: {},
    Apply: true,
  };
  postReq.Elements[elementName.value] = {
    FileName: element.value,
    Project: project.value,
    Branch: branch.value,
    Commit: commit.value,
    Tags: {},
    Inputs: inputsReq.value,
  };

  if (elementName.value === "") {
    noName = true;
    return;
  }

  let difference = Object.keys(inputs.value).filter(
    (x) => !Object.keys(inputsReq.value).includes(x)
  );
  if (difference.length)
    errorCreating.value = "Wprowadź poprawne wartości w polach: " + difference;
  else if (
    // env?.Elements.map((el) => {
    //   return el.Source.Name;
    // }).includes(elementName.value)
    false
  ) {
    elementNameTry.value = elementName.value;
    nameBusy = true;
  } else {
    api
      .post("/env-mod", postReq, {
        queries: {
          env: route.params.id as string,
        },
      })
      .then((res: any) => {
        if (res[elementName.value].Status === 0) emit("creatingDone");
        else errorCreating.value = res[elementName.value].Message;
      });
  }
};

// resety
const resetToProjects = () => {
  currentRef.value = 1;
  project.value = "";
  branch.value = "";
  commit.value = "";
  element.value = "";
  errorInputs.value = "";
  projectURL.value = "";

  projects.value = [];
  branches.value = [];
  commits.value = [];
  elements.value = [];

  // api.get("/git-repo-map").then((d) => {
  //   projects.value = Object.values(d);
  // });
};

const resetToBranches = () => {
  currentRef.value = 2;
  branch.value = "";
  commit.value = "";
  element.value = "";
  errorInputs.value = "";

  commits.value = [];
  elements.value = [];
};

const resetToCommits = () => {
  currentRef.value = 3;
  commit.value = "";
  element.value = "";
  errorInputs.value = "";

  elements.value = [];
};

const resetToElements = () => {
  currentRef.value = 4;
  element.value = "";
  errorInputs.value = "";
};

const updateInput = (res: string[]) => {
  inputsReq.value[res[0]] = "" + res[1];
};

onBeforeMount(() => {
  resetToProjects();
});

onMounted(() => {
  lol?.focus();
});
</script>

<template>
  <n-space vertical>
    <Teleport to=".n-card-header__main">
      <n-button strong secondary type="tertiary" @click="emit('back')">
        <n-icon size="14px"><mdi :path="mdiArrowLeft" /></n-icon>
        Wróć
      </n-button>
    </Teleport>
    <n-steps vertical :current="(currentRef as number)" status="process">
      <n-step title="Wybierz projekt">
        <p style="color: red">{{ errorProjects }}</p>
        <div v-if="currentRef === 1 && firstSection">
          <div class="form">
            <p class="form-label">URL</p>
            <n-input
              v-model:value="formProject"
              autofocus
              ref="lol"
              placeholder="np. https://google.com"
              @keyup.enter="chooseProject(formProject)"
            />
          </div>
          <div style="display: flex; justify-content: end">
            <n-button
              secondary
              type="primary"
              class="add-project"
              style="margin-top: 5px"
              @click="chooseProject(formProject)"
              >Dodaj URL</n-button
            >
          </div>
        </div>
        <div v-if="currentRef === 1 && !firstSection">
          <div class="form">
            <p class="form-label">URL</p>
            <p class="form-input">{{ projectURL }}</p>
          </div>
          <div class="form">
            <p class="form-label">Nazwa</p>
            <n-input
              class="form-input"
              v-model:value="formName"
              type="text"
              placeholder="Nazwa"
            />
          </div>
          <div class="form">
            <p class="form-label">Branch</p>
            <n-input
              class="form-input"
              v-model:value="formBranch"
              type="text"
              placeholder="Branch"
            />
          </div>
          <div class="form">
            <p class="form-label">Login</p>
            <n-input
              class="form-input"
              v-model:value="formLogin"
              type="text"
              placeholder="Login"
            />
          </div>
          <div class="form">
            <p class="form-label">Hasło</p>
            <n-input
              class="form-input"
              v-model:value="formPassword"
              type="password"
              placeholder="Hasło"
            />
          </div>
          <div style="display: flex; justify-content: end">
            <n-button
              secondary
              type="primary"
              class="add-project"
              style="margin-top: 5px; width: 30%"
              @click="chooseProject(formProject)"
              >Dodaj projekt</n-button
            >
          </div>
        </div>
        <div v-if="project !== ''" class="add-summary">
          <span>{{ project }}</span>
          <n-button type="error" @click="resetToProjects" class="back-button">
            Wstecz
          </n-button>
        </div>
      </n-step>
      <n-step title="Wybierz gałąź">
        <template v-if="currentRef === 2">
          <n-button
            v-for="(branch, id) in branches"
            strong
            secondary
            type="info"
            :key="id"
            class="addBrick"
            @click="chooseBranch(branch)"
          >
            {{ branch }}
          </n-button>
        </template>
        <div v-if="branch !== ''" class="add-summary">
          <span>{{ branch }}</span>
          <n-button type="error" @click="resetToBranches" class="back-button">
            Wstecz
          </n-button>
        </div>
      </n-step>
      <n-step title="Wybierz wersję">
        <template v-if="currentRef === 3">
          <data-table
            :bordered="false"
            :single-line="false"
            :columns="columns"
            :data="commits"
            :pagination="{
              pageSize: 10,
            }"
          >
            <template #commit="commit">
              <div style="display: flex; justify-content: center">
                <n-button
                  strong
                  secondary
                  type="info"
                  size="medium"
                  style="width: 100px"
                  >{{ commit.SHA.slice(0, 8) }}</n-button
                >
              </div>
            </template>
          </data-table>
        </template>
        <div v-if="commit !== ''" class="add-summary">
          <span>{{ commit.slice(0, 8) }}</span>
          <n-button type="error" @click="resetToCommits" class="back-button">
            Wstecz
          </n-button>
        </div>
      </n-step>
      <n-step title="Wybierz element">
        <p style="color: red">{{ errorInputs }}</p>
        <template v-if="currentRef === 4">
          <n-button
            v-for="(element, id) in elements"
            strong
            secondary
            type="info"
            :key="id"
            class="addBrick"
            @click="chooseElement(element)"
          >
            {{ element }}
          </n-button>
        </template>
        <div v-if="element !== ''" class="add-summary">
          <span>{{ element }}</span>
          <n-button type="error" @click="resetToElements" class="back-button">
            Wstecz
          </n-button>
        </div>
      </n-step>
      <n-step title="Dane wejściowe">
        <template v-if="currentRef === 5">
          <div style="display: flex">
            <p style="width: 15%">{{ t("fields.elementName") }}</p>
            <div style="width: 85%">
              <n-input
                v-model:value="elementName"
                :placeholder="t('fields.elementName')"
                :status="
                  (noName && elementName.length === 0) ||
                  (nameBusy && elementNameTry === elementName)
                    ? 'error'
                    : undefined
                "
              />
              <p v-if="noName && elementName.length === 0" style="color: red">
                Name is required
              </p>
              <p
                v-if="nameBusy && elementNameTry === elementName"
                style="color: red"
              >
                Wybrana nazwa jest zajęta
              </p>
            </div>
          </div>
          <h3 v-if="Object.keys(inputs).length" style="margin-top: 1em">
            Inputs
          </h3>
          <p style="color: red">{{ errorCreating }}</p>
          <ElementInput
            v-for="(input, k) in inputs"
            @updateInput="updateInput"
            :key="k"
            :name="(k as string)"
            :input="input"
            :edit="false"
            style="margin-top: 1em"
          />
          <div
            style="display: flex; margin-top: 1em; justify-content: flex-end"
          >
            <n-button type="error" @click="createElement" class="apply-button">
              Dodaj
            </n-button>
          </div>
        </template>
      </n-step>
    </n-steps>
  </n-space>
</template>

<style scoped>
.addBrick {
  margin-right: 5px;
  margin-bottom: 4px;
}

td {
  text-align: center;
}

.add-summary {
  display: flex;
  justify-content: space-between;
}

.back-button {
  top: -2em;
}

.n-step-content {
  height: 10em !important;
}

.add-project {
  margin-left: 10px;
  align-self: end;
}

.form {
  display: flex;
  gap: 1rem;
  width: 100%;
}

.form-label {
  width: 15%;
  align-self: center;
}

.form-input {
  margin-top: 5px;
}
</style>
