export const Node = z.object({
  ID: z.string(),
  Name: z.string(),
  IP: z.string(),
  Ready: z.boolean(),
  KubeMaster: z.boolean(),
  Resources: z.object({
    CPU: z.object({
      Usage: z.number().int(),
      Guaranteed: z.number().int(),
      Max: z.number().int(),
      Capacity: z.number().int(),
    }),
    RAM: z.object({
      Usage: z.number().int(),
      Guaranteed: z.number().int(),
      Max: z.number().int(),
      Capacity: z.number().int(),
    }),
    Pods: z.object({
      Capacity: z.number().int(),
      Total: z.number().int(),
      Ready: z.number().int(),
    }),
    Environments: z.object({
      Total: z.number().int(),
      Ready: z.number().int(),
    }),
  }),
  ClickHouseIsReady: z.boolean().optional(),
});

export const Check = z.object({
  Name: z.string(),
  Status: z.string(),
  Message: z.string(),
  // Success: z.boolean(),
  // Traceback: z.array(
  //   z.object({
  //     FilePath: z.string(),
  //     LineNr: z.number().int(),
  //     FunctionName: z.string(),
  //   })
  // ),
  LastCheckTime: z.number().int(),
  LastUpdateTime: z.number().int(),
});

export const SystemInfo = z.object({
  Versions: z.object({
    clusterDomain: z.string(),
    colorBar: z.string(),
    name: z.string(),
    gitTag: z.string(),
  }),
  Nodes: z.record(z.string(), Node).nullable(),
  StatusByModules: z.array(Check),
  ImageBuilderQueue: z.array(z.any()).nullable(),
  ImageBuilderStatus: z
    .record(
      z.string(),
      z.object({
        PodName: z.string(),
        IP: z.string(),
        StatusUpdateTime: z.string(),
        StatusPodExist: z.boolean(),
        StatusHTTPAlive: z.boolean(),
        StatusBuilding: z.boolean(),
        Blueprint: z.any(),
      })
    )
    .nullable(),
  NotificationsSend: z.boolean(),
  DiskUsage: z.record(z.string(), z.number().int()).nullable(),
});

export const systemCert = z.object({
  Name: z.string(),
  IngressName: z.string(),
  IngressNamespace: z.string(),
  SecretName: z.string(),
  SecretExist: z.boolean(),
  SecretExpirationDaysLeft: z.number(),
  CertName: z.string(),
  CertNamespace: z.string(),
  CertExist: z.boolean(),
  CertReady: z.boolean(),
  Timoni: z.boolean(),
  UsedIn: z.any(),
});

export const systemResourcesObject = z.object({
  CPUCapacity: z.number(),
  RAMCapacity: z.number(),
  Resources: z.record(
    z.string(),
    z.object({
      Elements: z.record(
        z.string(),
        z.object({
          CPU: z.object({
            Guaranteed: z.number(),
            Usage: z.number(),
            Max: z.number(),
          }),
          RAM: z.object({
            Guaranteed: z.number(),
            Usage: z.number(),
            Max: z.number(),
          }),
        })
      ),
      Total: z.object({
        CPU: z.object({
          Guaranteed: z.number(),
          Usage: z.number(),
          Max: z.number(),
        }),
        RAM: z.object({
          Guaranteed: z.number(),
          Usage: z.number(),
          Max: z.number(),
        }),
      }),
    })
  ),
});
