import { useMessage } from "naive-ui";
import { useI18n } from "vue-i18n";
import { createModule } from ".";

export const ErrorMessagesModule = createModule({
  setup() {
    const message = useMessage();
    const { t } = useI18n();
    return { message, t };
  },
  onMessage(message) {
    if (message.Code === 1) {
      return;
    }
    if (message.Code !== 0) {
      this.message.error(this.t(`errors.${message.Code}`));
    }
  },
});
