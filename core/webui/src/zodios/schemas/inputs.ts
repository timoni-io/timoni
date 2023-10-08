export const Input = z.object({
  Type: z.string(),
  Message: z.string(),
  Value: z.string(),
  Default: z.string(),
  MatchRegEx: z.string(),
  Min: z.number().int(),
  Max: z.number().int(),
  Options: z.union([z.array(z.string()), z.null()]),
});

export const EnvElementsInputs = z.union([z.record(Input), z.string()]);
