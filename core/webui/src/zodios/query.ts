import { MaybeRef } from "@vueuse/core";
import { defineStore } from "pinia";
import { Ref } from "vue";

type QueryInfo = { data: unknown; age: number };

const useQueryStore = defineStore("query-store", {
  state: () => ({} as Record<string, QueryInfo>),
  actions: {
    put(key: string, props: unknown, data: unknown) {
      this.$state[JSON.stringify({ key, props })] = { data, age: Date.now() };
    },
    get(key: string, props: unknown) {
      return this.$state[JSON.stringify({ key, props })];
    },
  },
});

type Options = { props?: MaybeRef<object>; interval?: MaybeRef<number> };

declare global {
  type Unref<T> = T extends Ref<infer V> ? V : T;
}

export const useQuery = <Key extends string, O extends Options, R>(
  key: Key,
  fetchFn: (props: Unref<O["props"]>) => Promise<R>,
  options?: O
) => {
  const queryStore = useQueryStore();

  const fetchAndCacheData = () =>
    fetchFn(unref(options?.props as any)).then((data) => {
      queryStore.put(key, unref(options?.props), data);
    });

  watchEffect(() => {
    fetchAndCacheData();
  });

  useIntervalFn(() => {
    unref(options?.interval) &&
      queryStore.get(key, unref(options?.props))?.age >
        unref(options?.interval!) &&
      fetchAndCacheData();
  }, options?.interval);

  const data = computed(
    () => (queryStore.get(key, unref(options?.props)).data || null) as R | null
  );

  return { data, fetch: fetchAndCacheData };
};
