import { createSharedComposable, WebSocketStatus } from "@vueuse/core";
import { ComputedRef } from "vue";
import { useCore } from "./core";
import { ExposeSchemas } from "./modules/exposeSchemas";
import {
  LiveLogs,
  SubscribeMessagesModule,
} from "./modules/subscribe/liveLogs";
import { GetLogsModule, GetLogs } from "./modules/getLogs";
import { ObjectSchema } from "./modules";
import { ErrorMessagesModule } from "./modules/errorMessages";

const modules = [SubscribeMessagesModule, GetLogsModule, ErrorMessagesModule];

type Api<T extends Record<string, ObjectSchema>> = ExposeSchemas<T> & {
  status: ComputedRef<WebSocketStatus>;
} & GetLogs &
  LiveLogs;

export const defineApi = <T extends Record<string, ObjectSchema>>(
  apiShema: T
): (() => Api<T>) =>
  // @ts-ignore
  createSharedComposable(() => {

    let url = `wss://${location.hostname}:${location.port}/j1/`;
    let devURL = import.meta.env.VITE_API_HOST
    if (typeof devURL !== 'undefined' && devURL !== null && devURL !== "") {
      url = `wss://${import.meta.env.VITE_API_HOST}/j1/`;
    }

    const core = useCore({
      url,
      isResponse(id, message) {
        return id === message.RequestID;
      },
      responseTimeout: 1000,
      reconnect: {
        delay: 1000,
        retries: 10,
      },
      onMessage(message) {
        // Message hook
        modules.forEach((module) => module.onMessage?.(message));
      },
      onConnected() {
        modules.forEach((module) => module.onConnected?.());
      },
      onReconnectFailed() {
        modules.forEach((module) => module.onReconnectFailed?.());
      },
      onDisconnected(_ws, event) {
        modules.forEach((module) => module.onDisconnected?.(event));
      },
    });

    const total: Partial<Api<any>> = { status: core.status };

    // Setup hook
    modules.forEach((module) => {
      // @ts-ignore
      Reflect.setPrototypeOf(module, module.setup?.(core, total) || module);
    });

    // Props hook
    Object.entries(apiShema).forEach(([name, objectSchema]) => {
      modules.forEach((module) => {
        if (module.addProperty) {
          const [key, value] = module.addProperty(name, objectSchema, core);
          total[key] = value;
        }
      });
    });

    return total;
  });
