export const CommitList = z.array(
  z.object({
    SHA: z.string(),
    Message: z.string(),
    Date: z.string(),
    TimeStamp: z.number().int(),
    AuthorName: z.string(),
    AuthorEmail: z.string(),
    AuthorInitials: z.string(),
    Files: z.array(z.string()).or(z.null()),
  })
);
