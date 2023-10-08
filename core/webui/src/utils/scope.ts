import { effectScope, watch, WatchOptions } from "vue";

export const useWatchScoped = <T, S>(
  watchFn: () => S,
  scopedFn: (v: S) => T,
  options?: WatchOptions
) => {
  watch(
    watchFn,
    (v, __, onCleanup) => {
      const scope = effectScope();
      scope.run(() => {
        scopedFn(v);
      });
      onCleanup(() => scope.stop());
    },
    options
  );
};
