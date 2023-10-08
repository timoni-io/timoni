export const RepoInfo = z.object({
  Name: z.string(),
  Description: z.string(),
  CloneURL: z.string(),
  RemoteURL: z.string(),
  DefaultBranch: z.string(),
  AppCreateLimitExceeded: z.boolean(),
  Owners: z.array(z.object({ Name: z.string(), Email: z.string() })),
  Error: z.string(),
  AccessToCode: z.string(),
});

export const FileList = z.array(
  z.object({ name: z.string(), isFile: z.enum(["file", "dir"]) })
);

export const GitOpsMap = z.record(z.record(z.array(z.string())));
