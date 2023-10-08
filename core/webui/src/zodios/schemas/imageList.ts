export const ImageList = z.record(
  z.string(),
  z.object({
    Name: z.string(),
    ImagesCount: z.number().int(),
    ImagesSizeMBWithoutParent: z.number().int(),
    Images: z.array(
      z.object({
        ID: z.string(),
        ParentImageID: z.string(),
        BuildStatus: z.string(),
        SizeWithParents: z.number().int(),
        SizeWithoutParent: z.number().int(),
        Registred: z.boolean(),
        BuildSHA: z.string(),
      })
    ),
  })
);
