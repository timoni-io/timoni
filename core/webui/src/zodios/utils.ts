import { z } from "zod";

type UnionToIntersectionFn<TUnion> = (
  TUnion extends TUnion ? (union: () => TUnion) => void : never
) extends (intersection: infer Intersection) => void
  ? Intersection
  : never;

type LastUnion<TUnion> = UnionToIntersectionFn<TUnion> extends () => infer Last
  ? Last
  : never;

type UnionToTuple<
  TUnion,
  TResult extends Array<unknown> = []
> = TUnion[] extends never[]
  ? TResult
  : UnionToTuple<
      Exclude<TUnion, LastUnion<TUnion>>,
      [...TResult, LastUnion<TUnion>]
    >;

type SchemaRecord = Record<string, z.ZodTypeAny>;

type SchemaRecordToParams<
  Type extends "Header" | "Query",
  Q extends SchemaRecord
> = UnionToTuple<
  {
    [Key in keyof Q]: { name: Key; type: Type; schema: Q[Key] };
  }[keyof Q]
>;

type Concat<T> = T extends [infer A, ...infer Rest]
  ? A extends any[]
    ? [...A, ...Concat<Rest>]
    : A
  : T;

const schemaRecordToParams = (
  type: "Header" | "Query",
  schemaRecord: SchemaRecord
) =>
  Object.entries(schemaRecord).map(([name, schema]) => ({
    name,
    type,
    schema,
  }));

type EndpointOptions = {
  response: z.ZodTypeAny;
  queries?: SchemaRecord;
  headers?: SchemaRecord;
};

type PostEndpointOptions = EndpointOptions & {
  request: z.ZodTypeAny;
  requestFormat?: string;
};

type EndpointDescription<
  M extends "get" | "post",
  P extends string,
  Res extends z.ZodTypeAny,
  ReqBody extends z.ZodTypeAny | unknown,
  Queries extends SchemaRecord,
  Headers extends SchemaRecord
> = Readonly<{
  method: M;
  path: P;
  response: Res;
  parameters: Concat<
    [
      ReqBody extends z.ZodTypeAny
        ? [{ name: "body"; type: "Body"; schema: ReqBody }]
        : [],
      Queries extends SchemaRecord
        ? SchemaRecordToParams<"Query", Queries>
        : [],
      Headers extends SchemaRecord
        ? SchemaRecordToParams<"Header", Headers>
        : []
    ]
  >;
}>;

export const defineGet = <P extends string, O extends EndpointOptions>(
  path: P,
  options: O
): EndpointDescription<
  "get",
  P,
  O["response"],
  unknown,
  O["queries"] extends SchemaRecord ? O["queries"] : {},
  O["headers"] extends SchemaRecord ? O["headers"] : {}
> => {
  return {
    method: "get",
    path,
    response: options.response,
    // @ts-ignore
    parameters: [
      ...schemaRecordToParams("Query", options.queries || {}),
      ...schemaRecordToParams("Header", options.headers || {}),
    ],
  };
};

export const definePost = <P extends string, O extends PostEndpointOptions>(
  path: P,
  options: O
): EndpointDescription<
  "post",
  P,
  O["response"],
  O["request"],
  O["queries"] extends SchemaRecord ? O["queries"] : {},
  O["headers"] extends SchemaRecord ? O["headers"] : {}
> => {
  return {
    method: "post",
    path,
    response: options.response,
    requestFormat: options.requestFormat,
    // @ts-ignore
    parameters: [
      { name: "body", type: "Body", schema: options.request },
      ...schemaRecordToParams("Query", options.queries || {}),
      ...schemaRecordToParams("Header", options.headers || {}),
    ],
  };
};
