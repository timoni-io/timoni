let errorMsg = ref<string | null>(null);
let envError = ref<string | null>(null);
let errorTranslation = ref(false);
let errorComunicate = ref("");
export default function useErrorMsg() {
  function setErrorMsg(val: string | null, translation = false) {
    errorMsg.value = val;
    if (translation) {
      errorTranslation.value = true;
    } else {
      errorTranslation.value = false;
    }
  }
  function setEnvError(val: string | null) {
    envError.value = val;
  }
  function setErrorComunicate(val: string) {
    errorComunicate.value = val;
  }

  return {
    errorMsg,
    setErrorMsg,
    envError,
    setEnvError,
    errorTranslation,
    setErrorComunicate,
    errorComunicate,
  };
}
