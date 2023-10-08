import { z } from "zod";

export interface Message {
  RequestID: string;
  Code: number;
  Data: unknown;
}

export type Action = "Live" | "Get";

export type Args =
  | {
      Subs: Array<{
        EnvID: string;
        Limit: number;
        Where?: any;
        Time?: string;
        Events?: boolean;
        FullLog?: boolean;
      }>;
    }
  | {
      Querys: Array<{
        Query: {
          Type: "ONE";
          Time: string;
          EnvID: string;
          FullLog?: boolean;
        };
      }>;
    };
//  {
//     EnvID: string;
//     Limit: number;
//     Where?: Where;
//     Time?: number;
//     Events?: boolean;
//   }[];
// | {
//     Query: {
//       Type: "ONE";
//       Time: BigInt;
//       EnvID: string;
//     };
//   }[]
// | {
//     Query: {
//       Type: "VECTOR";
//       Time: BigInt;
//       EnvID: string;
//       Direction: "AFTER" | "BEFORE";
//     };
//   }[];

export type Where =
  | {
      Type: "IS";
      Value: { Field: string; Value: string | number };
    }
  | {
      Type: "NOT";
      Value: Where;
    }
  | {
      Type: "OR";
      Value: Where[];
    };

export const WhereSchema: z.ZodType<Where> = z.lazy(() =>
  z.union([
    z.object({
      Type: z.literal("IS"),
      Value: z.object({ Field: z.string(), Value: z.string().or(z.number()) }),
    }),
    z.object({
      Type: z.literal("NOT"),
      Value: WhereSchema,
    }),
    z.object({
      Type: z.literal("OR"),
      Value: z.array(WhereSchema),
    }),
  ] as const)
);

export interface RequestMessage {
  RequestID: string;
  Action: Action;
  Args: Args;
}
