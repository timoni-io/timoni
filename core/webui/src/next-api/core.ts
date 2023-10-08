// import { useWatchScoped } from "@/utils/scope";
import {
  createSharedComposable,
  useWebSocket,
  UseWebSocketOptions,
  UseWebSocketReturn,
} from "@vueuse/core";
import { nanoid } from "nanoid";
import { computed, reactive, ref } from "vue";
import { Message, RequestMessage } from "./types";
// import JSONbig from "json-bigint";

const useReactiveUrlWebsocket = (
  url: string,
  options?: UseWebSocketOptions
) => {
  const websocketResult = reactive<UseWebSocketReturn<any>>(
    {} as unknown as UseWebSocketReturn<any>
  );

  Object.assign(websocketResult, useWebSocket(url, options));

  websocketResult.open();

  return websocketResult;
};

export type Core = ReturnType<typeof useCore>;

export const useCore = createSharedComposable(
  ({
    url,
    isResponse,
    responseTimeout,
    onMessage,
    reconnect,
    ...options
  }: {
    url: string;
    isResponse: (id: string, message: Message) => boolean;
    responseTimeout: number;
    onMessage: (message: Message, ws: WebSocket, event: MessageEvent) => void;
    reconnect: { retries: number; delay: number };
    onConnected: () => unknown;
    onDisconnected: (ws: WebSocket, event: CloseEvent) => unknown;
    onReconnectFailed: () => unknown;
  }) => {
    const reconnectCount = ref(reconnect.retries);

    const wsResult = useReactiveUrlWebsocket(url, {
      onMessage(ws: WebSocket, event: MessageEvent) {
        onMessage(JSON.parse(event.data), ws, event);
      },
      autoReconnect: {
        retries() {
          return --reconnectCount.value > 0;
        },
        delay: reconnect.delay,
        onFailed() {
          options.onReconnectFailed();
        },
      },
      onConnected() {
        reconnectCount.value = reconnect.retries;
        options.onConnected();
      },
      onDisconnected(ws: WebSocket, e: CloseEvent) {
        options.onDisconnected(ws, e);
      },
      immediate: false,
    });

    const sendWithResponse = (
      calcMessage: (id: string) => RequestMessage
    ): Promise<Message> => {
      const id = nanoid();
      const message = calcMessage(id);
      wsResult.send(JSON.stringify(message));

      return new Promise((resolve, reject) => {
        const onCancel = () =>
          reject(new Error("Timeout, no response from backend 🫶"));
        const timeout = setTimeout(onCancel, responseTimeout);

        const onMessage = (message: MessageEvent<string>) => {
          if (isResponse(id, JSON.parse(message.data))) {
            // clear side effects
            wsResult.ws?.removeEventListener("message", onMessage);
            clearTimeout(timeout);

            const res: Message = JSON.parse(message.data);

            if (res.Code !== 0) {
              reject(
                new Error(
                  `Backend Error Code: ${res.Code}, Message: ${res.Data}`
                )
              );
              return;
            }

            // reslove promise
            resolve(JSON.parse(message.data));
          }
        };

        wsResult.ws?.addEventListener("message", onMessage);
      });
    };

    return { sendWithResponse, status: computed(() => wsResult.status) };
  }
);
