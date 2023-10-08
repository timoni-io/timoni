export const Dashboard = z.object({
  ProjectList: z.array(
    z.object({
      Name: z.string(),
      Description: z.string(),
      CreatedTime: z.number().int(),
      ApplicationsStatus: z.record(z.number().int()),
      Favourite: z.boolean(),
      Public: z.boolean(),
      HasAccess: z.boolean(),
      Group: z.string(),
    })
  ),
  LastVisited: z.array(
    z.object({
      ID: z.string(),
      Name: z.string(),
      Status: z.number().int(),
      Variant: z.string(),
      UpdateMode: z.string(),
      Active: z.string(),
      Project: z.string(),
      LastVisitedTime: z.number().int(),
      Favourite: z.boolean(),
      // Group: z.string(),
    })
  ),
  Attention: z.array(z.any()),
  ClusterInfo: z.object({
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
    Nodes: z.object({ Total: z.number().int(), Ready: z.number().int() }),
    Resources: z.object({
      CPUGuaranteed: z.number().int(),
      RAMGuaranteed: z.number().int(),
    }),
  }),
});

export const Env = z.object({
  ID: z.string(),
  Name: z.string(),
  Status: z.record(z.number().int()),
  ElementsCount: z.number().int(),
  ToDelete: z.boolean(),
});

export const EnvMap = z.record(Env);

export const Repo = z.object({
  Name: z.string(),
  Local: z.boolean(),
  DefaultBranch: z.string(),
  Status: z.number().int(),
  StatusMessage: z.string(),
});

export const RepoMap = z.record(Repo);
