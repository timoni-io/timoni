import { useThemeVars } from "naive-ui";
import { watchEffect } from "vue";

export const useThemeCssVars = (overrides?: Record<string, string>) => {
  const vars = $(useThemeVars());

  watchEffect(() => {
    for (const [key, value] of Object.entries({ ...vars, ...overrides })) {
      document.documentElement.style.setProperty(`--${key}`, value);
    }
    document.documentElement.style.setProperty(
      `--backgroundSuccess`,
      "#36ad6a9c"
    );
    document.documentElement.style.setProperty(
      `--backgroundError`,
      "#de576d9c"
    );
  });
};
