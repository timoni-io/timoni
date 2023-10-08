import { ZodiosPlugin } from "@zodios/core";

export const ZodErrorPlugin: ZodiosPlugin = {
  name: "zod-error",
  async error(_api, _config, error) {
    throw new Error(error.cause as string);
  },
};
