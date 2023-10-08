// import { computed } from "vue";
import { z } from "zod";
import { defineApi } from "@/next-api/api";
import { defineObject } from "@/next-api/modules";
import { WhereSchema } from "./types";

// const TimestampDate = z.number().transform((n) => new Date(n * 1000));

const LogBase = z.object({
  RequestID: z.string(),
});

const LogInput = LogBase.extend({
  Where: WhereSchema,
  Limit: z.number(),
});

const LogOutput = LogBase.extend({
  Code: z.number(),
  Data: z.array(
    z.object({
      time: z.number(),
      level: z.string(),
      message: z.string(),
      env_id: z.string(),
      element: z.string(),
      pod: z.string(),
      version: z.string(),
      project: z.string(),
      user_email: z.string(),
      tags_string: z.record(z.string()),
      tags_number: z.record(z.number()),
    })
  ),
});

const Log = defineObject({
  input: LogInput,
  output: LogOutput,
});

export const useApi = defineApi({
  Log,
});
