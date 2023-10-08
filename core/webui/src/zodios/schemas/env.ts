import { teamListObject } from "@/zodios/schemas/user";
export const Application = z.object({
  Name: z.string(),
  ProjectName: z.string(),
  Status: z.number().int(),
  StatusMessage: z.string(),
  Variant: z.string(),
  UpdateMode: z.string(),
  Active: z.string(),
  ScheduleEnabled: z.boolean(),
  OnCrons: z.null(),
  OffCrons: z.null(),
});

export const Version = z.object({
  ID: z.string(),
  UserEmail: z.string(),
  UserInitials: z.string(),
  Time: z.number().int(),
  Status: z.number().int(),
  Current: z.boolean(),
  Working: z.boolean(),
});

export const Tag = z.object({
  Statuses: z.any(),
  Elements: z.array(z.string()),
  Actions: z.array(z.string()).nullish(),
});

export const Element = z.object({
  Source: z.object({
    Name: z.string(),
    FileName: z.string(),
    EnvironmentID: z.string(),
    Error: z.string(),
    Status: z.number().int(),
    Warnings: z
      .record(
        z.array(
          z.object({
            Message: z.string(),
            Reason: z.string(),
            RestartCount: z.number().int(),
            ExitCode: z.number().int(),
          })
        )
      )
      .nullish(),
    StatusMessage: z.string(),
    Description: z.string(),
    CurrentVersionTimeUnix: z.number().int(),
    Type: z.string(),
    GitRepoName: z.string(),
    Branch: z.string(),
    Commit: z.string(),
  }),
  Status: z.number().int(),
  Tags: z.array(z.string()).nullish().optional(),
  StatusMessage: z.string(),
  ImageID: z.string(),
  ContainerTotal: z.number().int(),
  ContainerReady: z.number().int(),
  Scale: z.object({
    NrOfPodsMin: z.number().int(),
    NrOfPodsMax: z.number().int(),
    NrOfPodsCurrent: z.number().int(),
    Disable: z.boolean(),
    CPUTargetProc: z.number().int(),
    CPUTargetMinCores: z.number().int(),
    CPUTargetMaxCores: z.number().int(),
    CPURequestedCores: z.number().int(),
  }),
  Resources: z.object({
    CPU: z.object({
      CurrentUsageProcentage: z.number().int(),
      CurrentUsageValue: z.number().int(),
      LastSevenDaysAverageUsage: z.number().int(),
      LastSevenDaysMaxUsage: z.number().int(),
      LastSevenDaysMaxHitLimitRatio: z.number().int(),
      GuaranteedValue: z.number().int(),
      LimitValue: z.number().int(),
    }),
    RAM: z.object({
      CurrentUsageProcentage: z.number().int(),
      CurrentUsageValue: z.number().int(),
      LastSevenDaysAverageUsage: z.number().int(),
      LastSevenDaysMaxUsage: z.number().int(),
      LastSevenDaysMaxHitLimitRatio: z.number().int(),
      GuaranteedValue: z.number().int(),
      LimitValue: z.number().int(),
    }),
  }),
  Inputs: z.boolean(),
  Actions: z.record(z.string()).nullish(),
});

export const Image = z.object({
  BuildStatus: z.string(),
  BuildID: z.string(),
  TimeBegin: z.number().int(),
  TimeEnd: z.number().int(),
});

export const ElementVersionsResponse = z.record(
  z.object(
    //   {
    //   Environment: z.string(),
    //   ElementName: z.string(),
    //   TimeUnix: z.number().int(),
    //   UserEmail: z.string(),
    //   UserInitials: z.string(),
    //   Branch: z.string(),
    //   Commit: z.string(),
    //   Status: z.number().int(),
    //   StatusMessage: z.string(),
    //   Current: z.boolean(),
    //   Previous: z.boolean(),
    // }
    {
      SaveTimestamp: z.number().int(),
      UserEmail: z.string(),
      UserInitials: z.string(),
      SourceGit: z.object({
        RepoName: z.string(),
        BranchName: z.string(),
        CommitHash: z.string(),
        FilePath: z.string(),
      }),
      Status: z.number(),
      Current: z.boolean(),
      Previous: z.boolean(),
    }
  )
);

export const ContainersInfo = z
  .array(
    z.object({
      PodName: z.string(),
      ElementName: z.string(),
      NodeName: z.string(),
      Status: z.number(),
      StatusMessage: z.string(),
      CheckResult: z.object({
        Status: z.number(),
        Msg: z.string(),
      }),
      CreationTime: z.number(),
      StartTime: z.number(),
      RestartCount: z.number(),
      Debug: z.boolean(),
      Warnings: z
        .union([
          z.record(
            z.array(
              z.object({
                Message: z.string(),
                Reason: z.string(),
                RestartCount: z.number().int(),
                ExitCode: z.number().int(),
              })
            )
          ),
          z.array(z.string()),
        ])
        .nullish(),
      CPUUsagePC: z.number(),
      CPUUsageProc: z.number(),
      RAMUsedMB: z.number(),
      RAMUsedProc: z.number(),
    })
  )
  .nullish();

// export const EnvInfo = z.union([
//   z.object({
//     Env: z.object({
//       Name: z.string(),
//       Status: z.number().int(),
//       StatusMessage: z.string(),
//       Active: z.string(),
//       ScheduleEnabled: z.boolean(),
//       ScheduleTimezone: z.string(),
//       OnCrons: z.union([z.array(z.string()), z.null()]),
//       OffCrons: z.union([z.array(z.string()), z.null()]),
//     }),
//     CurrentVersion: z.object({
//       ID: z.string(),
//       UserEmail: z.string(),
//       UserInitials: z.string(),
//       Time: z.number().int(),
//       Status: z.number().int(),
//       Current: z.boolean(),
//       Working: z.boolean(),
//     }),
//     Elements: z
//       .array(
//         z.object({
//           Source: z.object({
//             Name: z.string(),
//             FileName: z.string(),
//             EnvironmentID: z.string(),
//             Error: z.string(),
//             Status: z.number().int(),
//             Warnings: z
//               .record(
//                 z.array(
//                   z.object({
//                     Message: z.string(),
//                     Reason: z.string(),
//                     RestartCount: z.number().int(),
//                     ExitCode: z.number().int(),
//                   })
//                 )
//               )
//               .nullish(),
//             StatusMessage: z.string(),
//             Description: z.string(),
//             CurrentVersionTimeUnix: z.number().int(),
//             Type: z.string(),
//             GitRepoName: z.string(),
//             Branch: z.string(),
//             Commit: z.string(),
//           }),
//           Status: z.number().int(),
//           Tags: z.array(z.string()).nullish(),
//           StatusMessage: z.string(),
//           ImageID: z.string(),
//           Containers: ContainersInfo,
//           ContainerStatus: z.object({ 7: z.number().int().optional() }),
//           ContainerTotal: z.number().int(),
//           ContainerReady: z.number().int(),
//           Scale: z.object({
//             NrOfPodsMin: z.number().int(),
//             NrOfPodsMax: z.number().int(),
//             NrOfPodsCurrent: z.number().int(),
//             Disable: z.boolean(),
//             CPUTargetProc: z.number().int(),
//             CPUTargetMinCores: z.number().int(),
//             CPUTargetMaxCores: z.number().int(),
//             CPURequestedCores: z.number().int(),
//           }),
//           Resources: z.object({
//             CPU: z.object({
//               CurrentUsageProcentage: z.number().int(),
//               CurrentUsageValue: z.number().int(),
//               LastSevenDaysAverageUsage: z.number().int(),
//               LastSevenDaysMaxUsage: z.number().int(),
//               LastSevenDaysMaxHitLimitRatio: z.number().int(),
//               GuaranteedValue: z.number().int(),
//               LimitValue: z.number().int(),
//             }),
//             RAM: z.object({
//               CurrentUsageProcentage: z.number().int(),
//               CurrentUsageValue: z.number().int(),
//               LastSevenDaysAverageUsage: z.number().int(),
//               LastSevenDaysMaxUsage: z.number().int(),
//               LastSevenDaysMaxHitLimitRatio: z.number().int(),
//               GuaranteedValue: z.number().int(),
//               LimitValue: z.number().int(),
//             }),
//           }),
//           Inputs: z.boolean(),
//           Actions: z.record(z.string()).nullish(),
//         })
//       )
//       .nullish(),
//     // Actions: z.record(z.string()).nullish(),
//     Actions: z
//       .array(
//         z.object({
//           Name: z.string(),
//           Status: z.string(),
//           Error: z.string(),
//           Warnings: z.record(
//             z.array(
//               z.union([
//                 z.string(),
//                 z.object({
//                   Message: z.string(),
//                   Reason: z.string(),
//                   RestartCount: z.number().int(),
//                   ExitCode: z.number().int(),
//                 }),
//               ])
//             )
//           ),
//           GitRepoName: z.string(),
//           Branch: z.string(),
//           Commit: z.string(),
//           FileName: z.string(),
//           Tags: z.array(z.string()).nullish(),
//           ParentName: z.string(),
//           ActionName: z.string(),
//           TimeBegin: z.number(),
//           TimeEnd: z.number(),
//         })
//       )
//       .nullish(),
//     Images: z.record(
//       z.object({
//         BuildStatus: z.string(),
//         BuildID: z.string(),
//         TimeBegin: z.number().int(),
//         TimeEnd: z.number().int(),
//       })
//     ),
//     URLs: z.array(z.any()),
//   }),
//   z.string(),
// ]);

export const EnvInfo = z.object({
  Env: z.object({
    ID: z.string(),
    Name: z.string(),
    ClusterName: z.string(),
    Status: z.record(z.number().int()),
    Schedule: z.object({
      Active: z.boolean(),
      Timezone: z.object({}).nullish(),
      OnCrons: z.array(z.string()).nullish(),
      OffCrons: z.array(z.string()).nullish(),
    }),
    GitOps: z.object({
      Enabled: z.boolean(),
      GitRepoName: z.string(),
      BranchName: z.string(),
      FilePath: z.string(),
    }),
    ToDelete: z.boolean(),
    CreationTime: z.number().int(),
    LastChangeTime: z.number().int(),
    Elements: z.record(z.string()).nullish(),
    Tags: z.array(z.string()).nullish(),
    Teams: teamListObject.nullish(),
    GlobalVariableCache: z.null(),
    Resources: z.object({
      // CodeBranches: z.number().int(),
      // CodeStorage: z.number().int(),
      // EnvCount: z.number().int(),
      // PodsMin: z.number().int(),
      // PodsMax: z.number().int(),
      CPUUsageAvgCores: z.number(),
      CPUReservedCores: z.number(),
      CPUMaxCores: z.number(),
      RAMUsageAvgMB: z.number(),
      RAMReservedMB: z.number(),
      RAMMaxMB: z.number(),
      // StorageTemporary: z.number().int(),
      // StoragePersistent: z.number().int(),
    }),
    Readiness: z.object({
      ElementReady: z.number(),
      ElementNotReady: z.number(),
      PodReady: z.number(),
      PodNotReady: z.number(),
    }),
  }),
  Alerts: z.record(
    z.object({
      Messages: z.array(z.string()).nullish(),
      Variables: z.record(z.record(z.number())),
    })
  ),
  Metrics: z.object({
    Enabled: z.boolean(),
  }),
  MostChangedElements: z.array(
    z.object({
      Name: z.string(),
      TotalChanges: z.number().int(),
      Successes: z.number().int(),
      Failures: z.number().int(),
      InProgress: z.number().int(),
    })
  ),
  Teams: z.record(z.boolean()),
  Members: z.record(z.record(z.boolean())),
  URLs: z.record(z.string()),
});

export const EnvPod = z.object({
  PodName: z.string(),
  ElementName: z.string(),
  Type: z.string(),
  NodeName: z.string(),
  Status: z.number().int(),
  CreationTime: z.number().int(),
  ReadyTime: z.number().int(),
  RestartCount: z.number().int(),
  Debug: z.boolean(),
  Alerts: z.array(z.string()).nullish(),
  CPUUsagePC: z.number(),
  CPUUsageProc: z.number(),
  RAMUsedMB: z.number(),
  RAMUsedProc: z.number(),
});
