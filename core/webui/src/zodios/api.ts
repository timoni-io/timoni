import { Zodios, ZodiosEndpointDefinition } from "@zodios/core";
// import { FakerPlugin } from "./plugins/faker";
import { EnvMap, RepoMap } from "./schemas/dashboard";
import {
  ElementResp,
  ElementRespExtended,
  ElementMapRespExtended,
  GitEnvS,
} from "./schemas/elements";
import { EnvInfo, EnvPod } from "./schemas/env";
import { ElementCommitList, TagList } from "./schemas/changeVersion";
import { EnvElementsInputs, Input } from "./schemas/inputs";
import { CommitList } from "./schemas/commitList";
import { EnvModRequest, EnvModResponse } from "./schemas/envMod";
import { Logs } from "./schemas/logs";
import { ImageList } from "./schemas/imageList";
import { UsersObject, userInfoSchema } from "./schemas/user";
import { defineGet, definePost } from "./utils";
import { infer as Infer } from "zod";
// import { ZodErrorPlugin } from "./plugins/zodError";
import useErrorMsg from "@/utils/errorMsg";
import useErrorMsgEnv from "@/utils/errorMsgEnv";
import { useToken } from "@/store/token";
import { ElementVersionsResponse } from "./schemas/env";
import { containerResponse } from "./schemas/containers";
import { RepoInfo, FileList } from "./schemas/repos";
import { teamListObject } from "./schemas/user";
import {
  SystemInfo,
  systemCert,
  systemResourcesObject,
} from "./schemas/system";
import { GitOpsMap } from "./schemas/repos";
// import _ from "lodash";
let { setErrorMsg, setErrorComunicate } = useErrorMsg();
let { setErrorMsgEnv, hideErrorMsgEnv } = useErrorMsgEnv();

const elementContainers = definePost("/env-element-containers", {
  queries: {
    env: z.string(),
  },
  response: containerResponse,
  request: z.object({
    App: z.string(),
    Elements: z.array(z.string()),
  }),
});

const envActiveStatus = definePost("/env-schedule-set", {
  response: z.string(),
  request: z.object({
    // Mode: z.number(),
    Location: z.string(),
    OnCrons: z.array(z.string()),
    OffCrons: z.array(z.string()),
    EnvID: z.string(),
    Active: z.boolean(),
  }),
  queries: {
    EnvID: z.string(),
  },
});

const getEnvInfo = defineGet("/env-info", {
  response: z.union([EnvInfo, z.string()]),
  queries: {
    env: z.string(),
  },
});
const changeEnvName = defineGet("/env-rename", {
  queries: {
    env: z.string(),
    newName: z.string(),
  },
  response: z.any(),
});
const rebuildElement = defineGet("/image-rebuild", {
  queries: {
    imageID: z.string(),
    envID: z.string(),
  },
  response: z.any(),
});
const restartElement = defineGet("/env-element-restart-pods", {
  queries: {
    env: z.string(),
    element: z.string(),
  },
  response: z.any(),
});
const restartPod = defineGet("/env-pod-restart", {
  queries: {
    env: z.string(),
    element: z.string(),
    pod: z.string(),
  },
  response: z.string(),
});
const rerenderEnv = defineGet("/env-rerender", {
  queries: {
    env: z.string(),
  },
  response: z.string(),
});
const cloneEnv = defineGet("/env-clone", {
  queries: {
    env: z.string(),
    targetName: z.string(),
  },
  response: z.any(),
});
const deleteEnv = defineGet("/env-delete", {
  queries: {
    id: z.string(),
  },
  response: z.any(),
});
// const getDashboardInfo = defineGet("/dashboard", { response: Dashboard });
const getEnvMap = defineGet("/env-map", {
  response: EnvMap,
  // queries: {
  //   env: z.string(),
  // },
});
const getRepoMap = defineGet("/git-repo-map", {
  response: RepoMap,
});

const deleteRepo = defineGet("/git-repo-delete", {
  response: z.unknown(),
  queries: {
    name: z.string(),
  },
});
const getCommitList = defineGet("/git-repo-commit-list", {
  response: CommitList,
  queries: {
    name: z.string(),
    branch: z.string(),
  },
});
const createEnv = definePost("/env-create", {
  response: z.string(),
  request: z.object({
    name: z.string(),
  }),
});
const getCreateInputsCheck = defineGet("/env-create-inputs-check", {
  response: EnvElementsInputs,
  queries: {
    name: z.string(),
    branch: z.string(),
    commit: z.string(),
    env: z.string(),
    element: z.string(),
    project: z.string(),
  },
});
const getInputs = definePost("/env-element-inputs", {
  request: z.object({ Env: z.string(), Elements: z.array(z.string()) }),
  response: z.record(z.record(Input)),
});
const envElementsInputs = definePost("/env", {
  response: EnvElementsInputs,
  request: z.object({ Env: z.string(), Elements: z.array(z.string()) }),
});
const elementCommitList = defineGet("/env-element-commit-list", {
  response: ElementCommitList,
  queries: {
    env: z.string(),
    branch: z.string(),
    element: z.string(),
  },
});

const envMod = definePost("/env-mod", {
  request: EnvModRequest,
  response: EnvModResponse,
  queries: {
    env: z.string(),
  },
});

const elementScale = definePost("/env-element-scale", {
  request: z.object({
    CPUTargetProc: z.number(),
    Element: z.string(),
    EnvID: z.string(),
    NrOfPodsMax: z.number(),
    NrOfPodsMin: z.number(),
  }),
  response: z.any(), // to do
});
const elementManifest = definePost("/env-element-manifest", {
  request: z.object({ Env: z.string(), Elements: z.array(z.string()) }),
  response: z.record(z.string()),
});

const elementDockerFile = definePost("/env-element-docker-file", {
  request: z.object({ Env: z.string(), Elements: z.array(z.string()) }),
  response: z.record(z.string()),
});

const elementStaticScale = definePost("/env-element-static-scale", {
  request: z.record(z.string(), z.number()),
  response: z.any(), // to do
  queries: {
    env: z.string(),
  },
});
const elementVersions = defineGet("/env-element-versions", {
  response: ElementVersionsResponse,
  queries: {
    env: z.string(),
    element: z.string(),
  },
});
const logs = definePost("/logs", {
  response: Logs,
  request: z.object({
    Is: z.object({ env: z.array(z.string()), element: z.array(z.string()) }),
    TableSuffix: z.array(z.string()),
    Limit: z.number().int(),
    ID: z.union([z.null(), z.string()]),
    Time: z.union([z.null(), z.number().int()]),
    Type: z.string(),
  }),
});

// const getProjectMap = defineGet("/project/project-map", {
//   response: z.record(z.string()),
// });

const updateElementDebug = defineGet("/env-element-debug-update", {
  response: z.string(),
  queries: {
    env: z.string(),
    element: z.string(),
    state: z.string(),
  },
});

const createLocal = definePost("/git-repo-create", {
  response: z.string(),
  request: z.object({
    Name: z.string(),
  }),
});

const createRemote = definePost("/git-repo-create-remote", {
  response: z.union([
    z.string(),
    z.object({
      IsTaken: z.boolean(),
      GitRepoName: z.string(),
      Owners: z.array(
        z.object({
          Name: z.string(),
          Email: z.string(),
        })
      ),
    }),
  ]),
  request: z.object({
    URL: z.string(),
    // Branch: z.string().optional(),
    Name: z.string().optional(),
    // Public: z.boolean().optional(),
    Login: z.string().optional().optional(),
    Password: z.string().optional().optional(),
  }),
});

// images
const imageList = defineGet("/image-list", {
  response: ImageList,
});

const imageDeleteUnused = defineGet("/image-delete-unused", {
  response: z.string(),
});

// users
const userList = defineGet("/user-list", {
  response: UsersObject,
});

const updateUser = definePost("/user-update", {
  request: z.object({
    Email: z.string(),
		Name: z.string(),
		Teams: z.record(z.boolean())
  }),
  response: z.string(),
});

const aliasAdd = defineGet("/user/alias-add", {
  queries: {
    alias: z.string(),
    email: z.string(),
  },
  response: z.string(),
});

const aliasDelete = defineGet("/user/alias-remove", {
  queries: {
    alias: z.string(),
    email: z.string(),
  },
  response: z.string(),
});

// tags
const envCreateTag = defineGet("/env-tag-create", {
  queries: {
    env: z.string(),
    name: z.string(),
  },
  response: z.string(),
});

const envDeleteTag = defineGet("/env-tag-delete", {
  queries: {
    env: z.string(),
    name: z.string(),
  },
  response: z.string(),
});

const repoInfo = defineGet("/git-repo-info", {
  response: z.union([RepoInfo, z.string()]),
  queries: {
    name: z.string(),
  },
});
const commitInfoFileList = defineGet("/git-repo-commit-info-file-list", {
  response: z.array(z.string()),
  queries: {
    name: z.string(),
    branch: z.string(),
    commit: z.string(),
  },
});
const commitInfoFileDiff = defineGet("/git-repo-commit-info-file-diff", {
  response: z.string(),
  queries: {
    name: z.string(),
    branch: z.string(),
    commit: z.string(),
    path: z.string(),
  },
});
const fileList = defineGet("/git-repo-file-list", {
  response: FileList,
  queries: {
    name: z.string(),
    branch: z.string(),
    directory: z.string().optional(),
  },
});

const fileOpen = defineGet("/git-repo-file-open", {
  response: z.string(),
  queries: {
    name: z.string(),
    branch: z.string(),
    path: z.string(),
  },
});
const runAction = defineGet("/env-element-actions-run", {
  queries: {
    env: z.string(),
    element: z.string(),
    action: z.string(),
  },
  response: z.null(),
});

const branchList = defineGet("/git-repo-branch-list", {
  response: z.array(z.string()),
  queries: {
    name: z.string(),
    level: z
      .number()
      .optional()
      .transform((l) => l ?? 1),
  },
});

const tagList = defineGet("/git-repo-tag-list", {
  response: TagList,
  queries: {
    name: z.string(),
    level: z
      .number()
      .optional()
      .transform((l) => l ?? 1),
  },
});

// system
const systemVersion = defineGet("/system-version", {
  response: z.object({
    clusterDomain: z.string(),
    colorBar: z.string(),
    documentationShow: z.string(),
    k3os: z.string(),
    k3s: z.string(),
    kernel: z.string(),
    licenseServer: z.string(),
    name: z.string(),
    release: z.string(),
    systemCommit: z.string(),
    systemTime: z.string(),
  }),
});

const systemInfo = defineGet("/system-info", {
  response: SystemInfo,
});

const systemResources = defineGet("/system-resources", {
  response: systemResourcesObject,
});

const notificationUpdate = defineGet("/system/notification-update", {
  queries: {
    value: z.boolean(),
  },
  response: z.string(),
});

const systemCerts = defineGet("/system-certs", {
  response: z.array(systemCert),
});

const elementListManager = defineGet("/git-elements-list", {
  queries: {
    filter: z.string().optional(),
    type: z.string().optional(),
  },
  response: z.array(ElementResp),
});
const getElementListRepo = defineGet("/git-repo-element-list", {
  queries: {
    "git-repo-name": z.string(),
    branch: z.string(),
  },
  response: z.array(ElementRespExtended),
});
const getEnvListRepo = defineGet("/git-repo-env-map", {
  queries: {
    "git-repo-name": z.string(),
    branch: z.string(),
  },
  response: z.record(GitEnvS).nullish(),
});

const elementCreate = definePost("/env-element-create-from-git", {
  request: z.object({
    GitID: z.string(),
    EnvID: z.string(),
    Name: z.string(),
    DontStart: z.boolean(),
  }),
  response: z.string(),
});
const elementMap = defineGet("/env-element-map", {
  queries: {
    env: z.string(),
  },
  response: z.union([z.record(ElementMapRespExtended), z.string()]),
});
const userInfo = defineGet("/user-info", {
  response: userInfoSchema,
});
const variableList = defineGet("/env-variables", {
  response: z.any(),
  queries: {
    env: z.string(),
  },
});

const exportToTOML = defineGet("/env-export-toml", {
  response: z.any(),
  queries: {
    env: z.string(),
  },
});

const podsList = defineGet("/env-pods", {
  response: z.record(EnvPod),

  queries: {
    env: z.string(),
  },
});
const getSecret = defineGet("/env-variable-get-secret", {
  response: z.string(),
  queries: {
    env: z.string(),
    element: z.string(),
    variable: z.string(),
  },
});
const changeVersion = defineGet("/env-element-version-change", {
  response: z.string().nullish(),
  queries: {
    env: z.string(),
    element: z.string(),
    branch: z.string(),
    commit: z.string(),
  },
});
const changeUpdateMode = defineGet("/env-element-update-mode-set", {
  response: z.string(),
  queries: {
    env: z.string(),
    element: z.string(),
    mode: z.string(),
  },
});
const createElementFromToml = definePost("/env-element-create-from-toml", {
  response: z.string().nullish(),
  request: z.string(),
  queries: {
    element: z.string(),
    env: z.string(),
    "dont-start": z.boolean(),
  },
  requestFormat: "text",
});

const updateElementFromToml = definePost("/env-element-update-from-toml", {
  response: z.string().nullish(),
  request: z.string(),
  queries: {
    element: z.string(),
    env: z.string(),
  },
  requestFormat: "text",
});

const deleteElement = defineGet("/env-element-delete", {
  response: z.any(),
  queries: {
    env: z.string(),
    element: z.string(),
  },
});
const managementSet = definePost("/env-gitops-set", {
  response: z.any(),
  request: z.object({
    Enabled: z.boolean(),
    GitRepoName: z.string(),
    BranchName: z.string(),
    FilePath: z.string(),
  }),
  queries: {
    env: z.string(),
  },
});
const envDomainTargets = defineGet("/env-domain-targets", {
  response: z.array(z.string()),
  queries: {
    env: z.string(),
  },
});
const updateDefaultBranch = defineGet("/git-repo-update", {
  response: z.literal("ok"),
  queries: {
    "git-repo": z.string(),
    default: z.string(),
  },
});
const repoUpdateAccess = defineGet("/git-repo-remote-access-update", {
  response: z.string(),
  queries: {
    "git-repo": z.string(),
    url: z.string(),
    login: z.string(),
    password: z.string(),
  },
});
const userLogin = defineGet("/user-login", {
  response: z.any(),
  queries: {
    email: z.string(),
    token: z.string().optional(),
  },
});
const permissionList = defineGet("/perms-list", {
  response: z.any(),
});
const userPermission = defineGet("/user/permissions", {
  response: z.any(),
  queries: {
    email: z.string(),
  },
});
// const userPermissionUpdate = definePost("/user/permissions-update", {
//   request: z.object({
//     ID: z.string(),
//     Permissions: z.any(),
//   }),
//   response: z.any(),
// });
const getInvitationQr = defineGet("/user-invite-qr", {
  queries: {
    Email: z.string(),
  },
  response: z.any(),
});
const createTeam = definePost("/team-create", {
  request: z.object({
    Name: z.string(),
  }),
  response: z.any(),
});
const getTeamList = defineGet("/team-list", {
  response: z.array(teamListObject),
});
const getTeamInfo = defineGet("/team-info", {
  queries: {
    teamID: z.string(),
  },
  response: z.any(),
});
const updateTeam = definePost("/team-update", {
  request: z.object({
    ID: z.string(),
    Name: z.string(),
  }),
  response: z.string(),
});
const teamUserAdd = definePost("/team-user-add", {
  request: z.object({
    TeamID: z.string(),
    UserID: z.string(),
  }),
  response: z.string(),
});
const teamUserRemove = definePost("/team-user-remove", {
  request: z.object({
    TeamID: z.string(),
    UserID: z.string(),
  }),
  response: z.string(),
});
const userInvite = defineGet("/user-invite", {
  queries: {
    Email: z.string(),
  },
  response: z.any(),
});
const teamDelete = definePost("/team-delete", {
  request: z.object({
    ID: z.string(),
  }),
  response: z.string(),
});
const teamPermissionSet = definePost("/team-perms-set", {
  request: z.object({
    TeamID: z.string(),
		Global: z.record(z.boolean()),
    Env: z.record(z.record(z.boolean())),
    GitRepo: z.record(z.record(z.boolean()))
  }),
  response: z.string(),
});
const envTeamAdd = defineGet("/env-team-add", {
  queries: {
    team: z.string(),
    env: z.string(),
  },
  response: z.string(),
});
const envTeamRemove = defineGet("/env-team-remove", {
  queries: {
    team: z.string(),
    env: z.string(),
  },
  response: z.string(),
});
const manageElement = definePost("/env-element-run-control", {
  request: z.object({
    EnvID: z.string(),
    Element: z.string().optional(),
    Control: z.number(),
  }),
  response: z.string(),
});
const dynamicSourceAdd = definePost("/env-dynamic-sources-add", {
  request: z.object({
    RepoName: z.string(),
    BranchName: z.string(),
    DirPath: z.string(),
  }),
  response: z.string(),
});
const dynamimcSourceMap = defineGet("/env-dynamic-sources-map", {
  response: z.any(),
});
const gitOpsRepoMap = defineGet("/gitops-repo-map", {
  response: GitOpsMap,
});
const envDynamicSourceDelete = defineGet("/env-dynamic-sources-delete", {
  queries: {
    envSourceID: z.string(),
  },
  response: z.string(),
});


const systemElementVersions = defineGet("/system-element-versions", {
  queries: {
  },
  response: z.any(),
});


const defineEndpoints = <T extends ZodiosEndpointDefinition[]>(
  ...endpoints: T
) => endpoints;

const endpoints = defineEndpoints(
  systemElementVersions,
  getInvitationQr,
  envDynamicSourceDelete,
  gitOpsRepoMap,
  dynamimcSourceMap,
  dynamicSourceAdd,
  manageElement,
  envTeamRemove,
  teamDelete,
  teamUserAdd,
  updateTeam,
  getTeamList,
  createTeam,
  envTeamAdd,
  // userPermissionUpdate,
  teamPermissionSet,
  userPermission,
  permissionList,
  userLogin,
  updateDefaultBranch,
  repoUpdateAccess,
  managementSet,
  deleteElement,
  createElementFromToml,
  updateElementFromToml,
  changeVersion,
  getSecret,
  getEnvInfo,
  envElementsInputs,
  // getDashboardInfo,
  getEnvMap,
  getRepoMap,
  elementCommitList,
  getCommitList,
  // getElementList,
  getCreateInputsCheck,
  logs,
  createEnv,
  envMod,
  // getProjectMap,
  changeEnvName,
  createLocal,
  createRemote,
  cloneEnv,
  deleteEnv,
  updateElementDebug,
  rebuildElement,
  restartElement,
  restartPod,
  elementScale,
  elementStaticScale,
  elementVersions,
  elementManifest,
  elementDockerFile,
  imageList,
  imageDeleteUnused,
  userList,
  updateUser,
  aliasAdd,
  aliasDelete,
  getInputs,
  envActiveStatus,
  envCreateTag,
  envDeleteTag,
  elementContainers,
  rerenderEnv,
  repoInfo,
  fileList,
  fileOpen,
  runAction,
  commitInfoFileList,
  commitInfoFileDiff,
  branchList,
  tagList,
  systemVersion,
  systemInfo,
  notificationUpdate,
  systemResources,
  systemCerts,
  elementListManager,
  elementCreate,
  getElementListRepo,
  userInfo,
  elementMap,
  variableList,
  exportToTOML,
  podsList,
  deleteRepo,
  envDomainTargets,
  changeUpdateMode,
  getEnvListRepo,
  getTeamInfo,
  userInvite,
  teamUserRemove
);

const api = new Zodios("/api", endpoints);

declare global {
  type ResType<Path extends string> = {
    // @ts-ignore
    [K in keyof typeof endpoints]: typeof endpoints[K]["path"] extends Path
    ? // @ts-ignore
    Infer<typeof endpoints[K]["response"]>
    : never;
  }[keyof typeof endpoints];
}

// api.use(FakerPlugin);
// api.use(ZodErrorPlugin);

const originalGet = api.get;
api.get = async function () {
  try {
    // @ts-ignore
    let res = await originalGet.call(this, ...arguments);
    if (res === "permission denied") {
      if (arguments[0] === "/env-element-map") {
        return;
      }
      setErrorComunicate("messages.permissionDenied");
    }
    return res;
  } catch (e: any) {

    if (
      e.message.split("received:").length &&
      (e.message.split("received:")[1] === '\n"`session` is invalid"' ||
        e.message.split("received:")[1] === '\n"`token` is invalid"' ||
        e.message.split("received:")[1] === '\n"`session` is expired"' ||
        e.message === "$c:31")
    ) {
      localStorage.removeItem('user-session');
      window.location.href = "/login";

    } else if (
      e.message.split("received:").length &&
      e.message.split("received:")[1] === '\n"permission denied"'
    ) {
      // setErrorMsg("messages.permissionDenied", true);
      return;

    } else {
      setErrorMsg("Received an unsupported response. Repeating query, please wait a moment.");
      throw e;
    }

  }
};

const originalPost = api.post;
api.post = async function () {
  try {
    // @ts-ignore
    let res = await originalPost.call(this, ...arguments);
    if (res === "permission denied") {
      setErrorComunicate("messages.permissionDenied");
    }
    return res;
  } catch (e: any) {
    if (
      e.message.split("received:").length &&
      (e.message.split("received:")[1] === '\n"`session` is invalid"' ||
        e.message.split("received:")[1] === '\n"`token` is invalid"' ||
        e.message.split("received:")[1] === '\n"`session` is expired"' ||
        e.message === "$c:31")
    ) {
      window.location.href = "/login";
    } else {
      setErrorMsg("Received an unsupported response. Repeating query, please wait a moment.");
      throw e;
    }
  }
};

// api.axios.defaults.timeout = 2000;
// let gerror = false;
api.axios.interceptors.request.use(
  function (config) {
    const token = useToken();

    const params = new URLSearchParams(window.location.search);
    if (params.get("token")) {
      token.token = params.get("token") as string;
    }
    if (!token.token && window.location.pathname !== "/login") {
      // window.location.href = "/login";
    }
    // config["url"] += `?token=${token.token}`; // please let me die
    if (
      localStorage.getItem("user-session")
    ) {
      config.headers!["Session"] = localStorage.getItem("user-session");
    } else if (window.location.pathname !== "/login")
      window.location.href = "/login";

    return config;
  },
  function (error) {
    // Do something with request error
    return Promise.reject(error);
  }
);

api.axios.interceptors.response.use(
  function (resp) {
    // gerror = false;
    if (JSON.stringify(resp.data).charAt(0) === "{")
      return Promise.resolve(resp); // w przypadku braku bita na początku
    if (resp.data.substring(1) === '"`token` is invalid"') {
      setErrorMsgEnv(resp.data.substring(1));
    }

    hideErrorMsgEnv();
    setErrorMsg(null);
    // setErrorMsgEnv(null);
    try {
      return Promise.resolve({
        ...resp,
        data: JSON.parse(resp.data.substring(1)),
      });
    } catch (err) {
      return Promise.resolve({
        ...resp,
        data: resp.data,
      });
    }
  },
  function (error) {
    // Do something with request error
    if (!error.config.url.includes("/env-info")) {
      setErrorMsg(error.message);
    } else setErrorMsgEnv(error.message);
    // }
    return Promise.reject(error);
  }
);

export { api };
