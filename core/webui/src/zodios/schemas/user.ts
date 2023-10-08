export const permissions = z.object({
  Glob_AccessToAdminZone: z.object({ Index: z.number(), IsSet: z.boolean() }),
  Glob_AccessToKube: z.object({ Index: z.number(), IsSet: z.boolean() }),
  Glob_AccessToWebUI: z.object({ Index: z.number(), IsSet: z.boolean() }),
  Glob_CreateAndDeleteEnvs: z.object({ Index: z.number(), IsSet: z.boolean() }),
  Glob_CreateAndDeleteGitRepos: z.object({ Index: z.number(), IsSet: z.boolean() }),
  Glob_ManageGlobalMemebers: z.object({ Index: z.number(), IsSet: z.boolean() }),
});
export const User = z.object({
  ID: z.string(),
  Email: z.string(),
  Name: z.string(),
  Theme: z.string(),
  NotificationsSend: z.boolean(),
  CreatedTimeStamp: z.number(),
  Logout: z.number(),
  LastActionTimeStamp: z.number(),
  CanCreate: z.boolean(),
  HideGitRepoLocal: z.boolean(),
  PermissionsGlobal: z.record(z.any()).nullable(),
  Teams: z.array(z.string())
});
export const teamListObject = z.object({
    ID: z.string(),
    Name: z.string(),
    NrOfMembers: z.number(),
    NrOfEnvironments: z.number(),
    NrOfGitRepos: z.number(),
    Perms: z.any(),
    Members: z.any().nullable(),
  }
);

export const userInfoSchema = z.object({
  ID: z.string(),
  Email: z.string(),
  Name: z.string(),
  Theme: z.string(),
  NotificationsSend: z.boolean(),
  AutoLogout: z.number(),
  CreatedTimeStamp: z.number(),
  Logout: z.number(),
  LastActionTimeStamp: z.number(),
  CanCreate: z.boolean(),
  HideGitRepoLocal: z.boolean(),
  Teams: z.array(z.string()).nullish(),
  PermissionsGlobal: permissions,
});

export const UsersObject = z.record(z.string(), User);
