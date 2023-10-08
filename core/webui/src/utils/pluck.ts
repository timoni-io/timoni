const pluck = (obj: any, path: string): string => {
  const keys = path.split(".");
  const res = keys.reduce((oobj, key) => {
    return oobj[key] ? oobj[key] : {};
  }, obj);

  return typeof res === "string" ? res : "";
};

export default pluck;
