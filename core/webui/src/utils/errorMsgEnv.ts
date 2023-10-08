const errorMsgEnv = ref<string | null>(null);

export default function useErrorMsgEnv() {
  const errorTime = ref(0);

  function setErrorMsgEnv(val: string | null) {
    errorMsgEnv.value = val;
    errorTime.value = Date.now();
  }

  function hideErrorMsgEnv() {
    setTimeout(() => {
      if (errorTime.value < Date.now() - 2000) errorMsgEnv.value = null;
    }, 2000);
  }

  return {
    errorMsgEnv,
    setErrorMsgEnv,
    hideErrorMsgEnv,
  };
}
