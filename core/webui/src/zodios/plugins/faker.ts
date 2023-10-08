import { ZodiosPlugin } from "@zodios/core";
import { AxiosResponse, AxiosResponseHeaders, InternalAxiosRequestConfig } from "axios";
import { generateMock } from "@anatine/zod-mock";

const axiosResData = (data: any): AxiosResponse => ({
  data,
  status: 200,
  statusText: "ok",
  headers: {} as AxiosResponseHeaders,
  config: {} as InternalAxiosRequestConfig,
});

const isObject = (value: any) => typeof value === "object" && value !== null;

const undefToNull = (value: any): any => {
  if (value === undefined) {
    // Replace undef with null
    return null;
  }
  if (Array.isArray(value)) {
    return value.map(undefToNull);
  }
  if (isObject(value)) {
    return Object.fromEntries(
      Object.entries(value).map(([key, value]) => [key, undefToNull(value)])
    );
  }
  return value;
};
export const FakerPlugin: ZodiosPlugin = {
  name: "faker",
  async response(api, config) {
    const endpoint = api.find((e) => e.path === config.url)!;
    const mock = generateMock(endpoint.response, {});
    return axiosResData(undefToNull(mock));
  },
  async error(api, config) {
    const endpoint = api.find((e) => e.path === config.url)!;
    const mock = generateMock(endpoint.response);
    return axiosResData(undefToNull(mock));
  },
};
