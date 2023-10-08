export const containerResponse = z.array(
  z.object({
    PodName: z.string(),
    ElementName: z.string(),
    NodeName: z.string(),
    Status: z.number().int(),
    CheckResult: z.object({ Status: z.number().int(), Msg: z.string() }),
    CreationTime: z.number().int(),
    StartTime: z.number().int(),
    RestartCount: z.number().int(),
    Debug: z.boolean(),
    Alerts: z.union([z.null(), z.array(z.string())]),
    CPUUsagePC: z.number().int(),
    CPUUsageProc: z.number().int(),
    RAMUsedMB: z.number().int(),
    RAMUsedProc: z.number().int(),
  })
);
