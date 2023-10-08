let fontSize = ref<number>(14);

// get default font size from local storage
let font = localStorage.getItem("defaultFontSize");
if (font) fontSize.value = parseInt(font);

export default function useDefaultFontSize() {
  function setFontSize(val: number) {
    localStorage.setItem("defaultFontSize", val + "px");
    fontSize.value = val;
  }

  return { fontSize, setFontSize };
}
