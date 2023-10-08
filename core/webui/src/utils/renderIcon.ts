import { NIcon } from "naive-ui";
import Mdi from "@/components/Mdi.vue";

export function renderIcon(icon: string, color?: string) {
  return () =>
    h(NIcon, null, { default: () => h(Mdi, { path: icon, color: color }) });
}
