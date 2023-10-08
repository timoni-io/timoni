import {
  formatDistanceToNow,
  formatRelative,
  differenceInDays,
  format,
} from "date-fns";
import pl from "date-fns/locale/pl";
import { useI18n } from "vue-i18n";

export const useTimeFormatter = () => {
  const { locale } = useI18n();

  const distanceToNow = (date: Date) => {
    if (date.getTime() - new Date().getTime() >= 0) {
      return `${locale.value === "pl" ? "za" : "in"} ${formatDistanceToNow(
        date,
        {
          locale: locale.value === "pl" ? pl : undefined,
        }
      )}`;
    } else {
      return `${formatDistanceToNow(date, {
        locale: locale.value === "pl" ? pl : undefined,
      })} ${locale.value === "pl" ? "temu" : "ago"}`;
    }
  };

  const relativeOrDistanceToNow = (date: Date) => {
    if (differenceInDays(date, new Date()) > 2) {
      return formatRelative(date, new Date(), {
        locale: locale.value === "pl" ? pl : undefined,
      });
    } else {
      return distanceToNow(date).replace("około", "").replace("about", "");
    }
  };

  const formatTime = (date: Date) => {
    return format(date, "HH:mm • d MMMM yy", {
      locale: locale.value === "pl" ? pl : undefined,
    });
  };

  const formatTimeWithZero = (date: Date) => {
    return format(date, "dd.MM.yyyy");
  };

  const dateFranekFormat = (time: any) => {
    const date = new Date(Math.floor(time / 1000000));
    return date.toLocaleString(locale.value === "pl" ? "pl-PL" : "en-US", {
      month: "short",
      day: "numeric",
      hour: "numeric",
      minute: "numeric",
      second: "numeric",
      fractionalSecondDigits: 3,
    });
  };

  return {
    distanceToNow,
    relativeOrDistanceToNow,
    formatTime,
    formatTimeWithZero,
    dateFranekFormat,
  };
};
