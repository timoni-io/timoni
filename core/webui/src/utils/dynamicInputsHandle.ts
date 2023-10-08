export function dynamicInputsHandle<T extends object>(
  dynamicInputs: Array<T>,
  keys: Array<keyof T>,
  emptyValue: Array<String>,
  id: string,
  onCreate: () => T
) {
  let empties: number = dynamicInputs
    .map((el) => {
      return Object.values(pick(el, keys)).join("");
    })
    .filter((el) => {
      return emptyValue.includes(el);
    }).length;
  if (empties === 0) {
    dynamicInputs.push(onCreate());
    let objDiv = document.getElementById(id);
    if (objDiv) {
      setTimeout(() => {
        if (objDiv) objDiv.scrollTop = objDiv.scrollHeight;
      }, 100);
    }
  } else if (empties === 2) {
    const removeInputs = dynamicInputs.filter((el) => {
      return emptyValue.includes(Object.values(pick(el, keys)).join(""));
    });
    for (const remInp of removeInputs) {
      const index = dynamicInputs.indexOf(remInp);
      dynamicInputs.splice(index, 1);
    }
  }
}

export const pick = <T extends object, K extends keyof T>(
  o: T,
  keys: K[]
): Pick<T, K> =>
  Object.fromEntries(
    Object.entries(o).filter(([key]) => keys.includes(key as K))
  ) as Pick<T, K>;
