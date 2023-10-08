export const Logs = z.array(
  z.object({
    time: z.number().int(),
    id: z.string(),
    level: z.string(),
    message: z.string(),
    element: z.string(),
    pod: z.string(),
    env: z.string(),
    version: z.string(),
    user: z.string(),
    project: z.string(),
    details_string: z.string(),
    details_number: z.number(),
  })
);
