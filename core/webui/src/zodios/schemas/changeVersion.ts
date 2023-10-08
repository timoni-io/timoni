export const ElementCommit = z.object({
  SHA: z.string(),
  Message: z.string(),
  Date: z.string(),
  TimeStamp: z.number().int(),
  AuthorName: z.string(),
  AuthorEmail: z.string(),
  AuthorInitials: z.string(),
  Files: z.array(z.string()),
});

export const ElementCommitList = z.object({
  GitRepo: z.string(),
  Branch: z.string(),
  Commits: z.array(ElementCommit).nullish(),
});

export const BranchList = z.union([z.array(z.string()), z.string()]);

export const TagList = z.array(
  z.object({
    Name: z.string(),
    CommitSHA: z.string(),
    CommitAuthorEmail: z.string(),
    CommitAuthorInitials: z.string(),
    TimeStamp: z.number().int(),
  })
);
