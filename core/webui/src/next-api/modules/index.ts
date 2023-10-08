import { z } from "zod";
import { Core } from "../core";
import { Message } from "../types";

export const defineObject = <
  T extends z.AnyZodObject,
  Z extends z.AnyZodObject
>(shema: {
  output: T;
  input: Z;
}) => {
  return shema;
};

export type ObjectSchema = ReturnType<typeof defineObject>;
export type ApiShema = Record<string, ObjectSchema>;

type ApiModule<Context> = {
  setup?: (core: Core, api: object) => Context;
  addProperty?: (
    obj: string,
    schema: ObjectSchema,
    core: Core
  ) => [string, any];

  onMessage?: (message: Message) => void;
  onConnected?: () => unknown;
  onDisconnected?: (event: CloseEvent) => unknown;
  onError?: () => unknown;
  onReconnectFailed?: () => unknown;
} & ThisType<Context>;

export const createModule = <Context>(hooks: ApiModule<Context>) => hooks;
