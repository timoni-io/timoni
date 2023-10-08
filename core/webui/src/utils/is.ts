import { z } from "zod";

export const is =
  <T extends z.ZodTypeAny>(schema: T) =>
  (value: any): value is z.infer<T> =>
    schema.safeParse(value).success;
