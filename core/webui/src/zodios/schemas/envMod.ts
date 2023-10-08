export const EnvModRequest = z.object({
  DeleteElements: z.array(z.string()),
  Elements: z.record(
    z.object({
      Branch: z.string(),
      Commit: z.string(),
      // Tags: z.any(),
      Inputs: z.object({}).catchall(z.any()).optional(),
    })
  ),
  Apply: z.boolean(),
});

export const EnvModResponse = z.record(
  z.object({
    Status: z.number().int(),
    Message: z.string(),
    Inputs: z.null(), // poprawić
    ElementType: z.string(),
  })
);
