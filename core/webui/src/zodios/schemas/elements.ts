export const ElementRespExtended = z.object({
  Element: z.object({
    ID: z.string(),
    Name: z.string(),
    Type: z.string(),
    Source: z.object({
      RepoName: z.string(),
      BranchName: z.string(),
      CommitHash: z.string(),
      FilePath: z.string(),
    }),
    Description: z.string(),
    Favorite: z.boolean(),
    UsageCount: z.number().int(),
    UsageTime: z.number().int(),
    Error: z.string(),
    FileContent: z.string(),
  }),
  Usage: z
    .array(
      z.object({
        EnvID: z.string(),
        EnvName: z.string(),
        ElementName: z.string(),
      })
    )
    .nullish(),
});

export const ElementResp = z.object({
  ID: z.string(),
  Name: z.string(),
  Type: z.string(),
  Source: z.object({
    RepoName: z.string(),
    BranchName: z.string(),
    CommitHash: z.string(),
    FilePath: z.string(),
  }),
  Description: z.string(),
  Favorite: z.boolean(),
  UsageCount: z.number().int(),
  UsageTime: z.number().int(),
  Error: z.string(),
  FileContent: z.string(),
});

const ElementBase = z.object({
  Name: z.string(),
  EnvironmentID: z.string(),
  AutoUpdate: z.boolean().nullish(),
  Description: z.string(),
  SourceGit: z.object({
    RepoName: z.string(),
    BranchName: z.string(),
    CommitHash: z.string(),
    FilePath: z.string(),
  }),
  Errors: z.array(z.string()).nullish(),
  ToDelete: z.boolean(),
  Variables: z
    .record(
      z.object({
        Description: z.string(),
        Validation: z.string(),
        Secret: z.boolean(),
        System: z.boolean(),
        CurrentValue: z.string(),
        Errors: z.record(z.number()).nullish(),
        ResolvedValue: z.string(),
      })
    )
    .nullish(),
  UserEmail: z.string(),
  UserInitials: z.string(),
  Unschedulable: z.boolean(),
  Stopped: z.boolean(),
  SaveTimestamp: z.number().int(),
});

export const DomainType = ElementBase.extend({
  Type: z.literal("domain"),
  Domain: z.string(),
  // ExternalProtocol: z.string(),
  Auth: z.string(),
  UploadLimit: z.number().int(),
  Timeout: z.number().int(),
  // BuffersNumber: z.number().int(),
  // BufferSize: z.number().int(),
  // HeaderBuffersNumber: z.number().int(),
  // HeaderBufferSize: z.number().int(),
  // StartPath: z.string(),
  Paths: z
    .record(
      z.object({
        ElementName: z.string(),
        Port: z.number().int(),
        Prefix: z.string(),
        Label: z.string(),
      })
    )
    .nullish(),
  Annotations: z.record(z.string()).nullish(),
  HttpOnly: z.boolean(),
  WWWredirect: z.boolean(),
  // URL: z.string(),
});

export const ConfigType = ElementBase.extend({
  Type: z.literal("config"),
});

// export const MongoType = ElementBase.extend({
//   Type: z.literal("mongodb"),
//   Version: z.string(),
//   BackupAzureStorageAccountName: z.string(),
//   BackupAzureStorageAccountKey: z.string(),
//   BackupAzureStorageAccountContainer: z.string(),
//   MembersCount: z.number().int(),
//   StorageSize: z.number().int(),
//   CPUlimit: z.number().int(),
//   RAMlimit: z.number().int(),
//   Users: z.record(
//     z.object({
//       Password: z.string(),
//       Database: z.string(),
//       Roles: z.array(z.string()),
//     })
//   ),
// });

export const MongoType = ElementBase.extend({
  Type: z.literal("mongodb"),
  Version: z.any(),
  BackupAzureStorageAccountName: z.string(),
  BackupAzureStorageAccountKey: z.string(),
  BackupAzureStorageAccountContainer: z.string(),
  MembersCount: z.number().int(),
  StorageSize: z.number().int(),
  CPUlimit: z.number().int(),
  RAMlimit: z.number().int(),
  Users: z.record(
    z.object({
      Password: z.string(),
      Database: z.string(),
      Roles: z.array(z.string()),
    })
  ),
  // Paths: z.record(z.object({
  //   ElementName: z.string(),
  //   Port: z.number().int(),
  //   Prefix: z.string(),
  // })).nullish(),
  // Annotations: z.null(),
});

export const ElasticsearchType = ElementBase.extend({
  Type: z.literal("elasticsearch"),
  Version: z.string(),
  ExternalIP: z.boolean(),
  XpackSecurity: z.boolean(),
  Backup: z
    .record(
      z.object({
        Type: z.string(),
        ReadOnly: z.boolean(),
        RemoteServer: z.string(),
        RemotePath: z.string(),
        AzureStorageAccountName: z.string(),
        AzureStorageAccountKey: z.string(),
        AzureStorageAccountContainer: z.string(),
      })
    )
    .nullish(),
  NodeSets: z.array(
    z.object({
      Name: z.string(),
      Count: z.number().int(),
      PodTemplate: z.object({
        Spec: z.record(z.any()),
      }),
    })
  ),
  CPUGuaranteed: z.number().int(),
  CPUMax: z.number().int(),
  RAMGuaranteed: z.number().int(),
  RAMMax: z.number().int(),
  Storage: z.number().int(),
});

export const PodType = ElementBase.extend({
  Type: z.literal("pod"),
  ParentSource: z
    .object({
      RepoName: z.string(),
      BranchName: z.string(),
      CommitHash: z.string(),
      FilePath: z.string(),
    })
    .nullish(),
  Build: z.object({
    Type: z.string(),
    ImageID: z.string(),
    Script: z.string(),
    Image: z.string(),
    RootPath: z.string(),
    Variables: z.record(z.string()).nullish(),
    DockerfilePath: z.string(),
  }),
  RunCommand: z.array(z.string()).nullish(),
  RunAsUser: z.array(z.number().int()).nullish(),
  RunWritableFS: z.boolean(),
  StaticFilesPath: z.string(),
  CapabilityAddBindService: z.boolean(),
  CPUUsageAvgCores: z.number(),
  RAMUsageAvgMB: z.number(),
  RunProbe: z
    .object({
      Exec: z.array(z.string()),
      InitialDelaySeconds: z.number().int(),
      PeriodSeconds: z.number().int(),
      TimeoutSeconds: z.number().int(),
      SuccessThreshold: z.number().int(),
      FailureThreshold: z.number().int(),
      RestartOnFail: z.boolean(),
    })
    .nullish(),
  Storage: z
    .record(
      z.object({
        Type: z.string(),
        Class: z.string(),
        MaxSizeMB: z.number().int(),
        Name: z.string(),
        Login: z.string(),
        Password: z.string(),
        RemoteHost: z.string(),
        RemotePath: z.string(),
        Options: z.string(),
        ReadOnly: z.boolean(),
      })
    )
    .nullish(),
  Stateful: z.boolean(),
  ExposePorts: z
    .record(
      z.object({
        MetricsPath: z.string(),
        Name: z.string(),
        PortAliases: z.array(z.number().int()).nullish(),
        Type: z.string(),
        Probe: z.object({
          Disable: z.boolean(),
          ErrorMsg: z.string(),
          FailureThreshold: z.number().int(),
          InitialDelaySeconds: z.number().int(),
          Path: z.string(),
          PeriodSeconds: z.number().int(),
          RestartOnFail: z.boolean(),
          SuccessThreshold: z.number().int(),
          TimeoutSeconds: z.number().int(),
        }),
      })
    )
    .nullish(),
  ExposePortsHeadless: z.boolean(),
  StickyCookie: z.string(),
  ServiceAccount: z.object({
    Name: z.string(),
    Secret: z.string(),
  }),
  ApplyVariablesOnFiles: z.array(z.string()).nullish(),
  Schedule: z.string(),
  Actions: z.record(z.array(z.string()).nullish()).nullish(),
  CPUReservedPC: z.number().int(),
  CPULimitPC: z.number().int(),
  RAMReservedMB: z.number().int(),
  RAMLimitMB: z.number().int(),
  Scale: z.object({
    MaxOnePod: z.boolean(),
    NrOfPodsMin: z.number().int(),
    NrOfPodsMax: z.number().int(),
    NrOfPodsCurrent: z.number().int(),
    CPUTargetProc: z.number().int(),
  }),
  HostAliases: z.record(z.array(z.string())).nullish(),
});

export const ActionType = ElementBase.extend({
  Type: z.literal("action"),
  ActionName: z.string(),
  Status: z.string(),
  TimeBegin: z.number().int(),
  TimeEnd: z.number().int(),
  ActionToken: z.string(),
  ParentName: z.string(),
  ImageID: z.string(),
  RunCommand: z.array(z.string()),
  RunAsUser: z.number().int().nullish(),
  RunWritableFS: z.boolean(),
  Storage: z
    .record(
      z.object({
        Type: z.string(),
        Class: z.string(),
        MaxSizeMB: z.number().int(),
        Name: z.string(),
        Login: z.string(),
        Password: z.string(),
        RemoteHost: z.string(),
        RemotePath: z.string(),
        Options: z.string(),
        ReadOnly: z.boolean(),
      })
    )
    .nullish(),
  HostAliases: z.record(z.array(z.string())).nullish(),
  ApplyVariablesOnFiles: z.array(z.string()).nullish(),
  Scale: z.object({
    MaxOnePod: z.boolean(),
    NrOfPodsMin: z.number().int(),
    NrOfPodsMax: z.number().int(),
    NrOfPodsCurrent: z.number().int(),
    CPUTargetProc: z.number().int(),
  }),
});

export const ElementMapResp = z.union([
  DomainType,
  ConfigType,
  PodType,
  ElasticsearchType,
  MongoType,
  ActionType,
]);

export const ElementMapStatus = z.object({
  State: z.number().int(),
  Alerts: z.array(z.string()).nullish(),
  NewerVersion: z.boolean(),
  Next: z.object({
    SourceGit: z.object({
      RepoName: z.string(),
      BranchName: z.string(),
      CommitHash: z.string(),
      FilePath: z.string(),
    }),
    StepCurrent: z.number().int(),
    StepCount: z.number().int(),
    State: z.number().int(),
    Message: z.string(),
  }).nullish(),
  // CPUUsedProc: z.number().int(),
  // CPUUsedPC: z.number().int(),
  // CPUUsedLastSevenDaysAvgPC: z.number().int(),
  // CPUUsedLastSevenDaysMaxPC: z.number().int(),
  // CPULastSevenDaysMaxHitLimitRatio: z.number().int(),
  // RAMUsedProc: z.number().int(),
  // RAMUsedMB: z.number().int(),
  // RAMUsedLastSevenDaysAvgMB: z.number().int(),
  // RAMUsedLastSevenDaysMaxMB: z.number().int(),
  // RAMLastSevenDaysMaxHitLimitRatio: z.number().int(),
  PodCount: z.number().int(),
});

export const ElementMapRespExtended = z.object({
  Info: ElementMapResp,
  Status: ElementMapStatus,
});

export const GitEnvS = z.object({
  Name: z.string(),
  FileContent: z.string(),
  Description: z.string(),
  Tags: z.array(z.string()).nullish(),
  Teams: z.array(z.string()).nullish(),
  Source: z.object({
    RepoName: z.string(),
    BranchName: z.string(),
    CommitHash: z.string(),
    FilePath: z.string(),
  }),
  Element: z
    .record(
      z.object({
        RepoName: z.string(),
        BranchName: z.string(),
        CommitHash: z.string(),
        FilePath: z.string(),
      })
    )
    .nullish(),
  Error: z.string(),
});
