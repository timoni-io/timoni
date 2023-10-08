import { useRoute } from "vue-router";

export const useRouteParam = (param: string) => {
  const route = useRoute();
  let paramVal = $shallowRef(route.params[param] as string);

  return computed({
    get() {
      return paramVal;
    },
    set(p) {
      history.pushState(
        null,
        "",
        Object.entries(route.params)
          .reduce(
            (path, [pName, pVal]) =>
              Array.isArray(pVal)
                ? path.replace(`:${pName}*`, pVal.join("/"))
                : path.replace(`:${pName}`, pVal),
            route.matched[0].path.replace(`:${param}`, p)
          )
          .replaceAll("*", "")
      );
      paramVal = p;
      triggerRef($$(paramVal));
    },
  });
};

export const useRouteParamArray = (param: string) => {
  const route = useRoute();
  let paramVal = $shallowRef(
    (Array.isArray(route.params[param])
      ? route.params[param]
      : ([route.params[param]] as string[])) as string[]
  );
  ``;

  return computed({
    get() {
      return paramVal;
    },
    set(p) {
      history.pushState(
        null,
        "",
        Object.entries(route.params)
          .reduce(
            (path, [pName, pVal]) =>
              Array.isArray(pVal)
                ? path.replace(`:${pName}*`, pVal.join("/"))
                : path.replace(`:${pName}`, pVal),

            route.matched[0].path.replace(`:${param}`, p.join("/"))
          )
          .replaceAll("*", "")
      );
      paramVal = p;
      triggerRef($$(paramVal));
    },
  });
};
