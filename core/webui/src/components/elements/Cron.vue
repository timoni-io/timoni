<script setup lang="ts">
// import { useEnv } from "@/store/envStore";
import { useRoute } from "vue-router";
import { useMessage } from "naive-ui";
import { useI18n } from "vue-i18n";
import { useEnv } from "@/store/envStore";

defineProps<{ custom: boolean; inactive: boolean; manage: boolean }>();
const { t } = useI18n();
const message = useMessage();
const route = useRoute();
const env = useEnv(route.params.id as string);

let onCrons = $ref<string[]>([]);
let offCrons = $ref<string[]>([]);

let fromTime = $ref<string | null>();
let selectedDuration = $ref<string | null>();
let selectedHourDuration = $ref<string | null>();
let selectedMinuteDuration = $ref<string | null>();
let scheduleList = $ref<
  { fromTime: string; duration: string; days: string[] }[]
>([]);
let lengthOffCrons = $ref<number>();

let value = $ref<number>();
let tmpValue = $ref<number>();
let openCronModal = $ref(true);
let openCronSchedule = $ref(false);
let selectedDays = $ref<string[]>([]);
let selectedId = $ref<number | null>(null);
// let confirmChangeStatus = $ref(false);
let shouldLock = $ref(false);
let touched = $ref(false);
let scheduleTouched = $ref(false);
let cronSwitcher = $ref(true);
let backendActiveState = $ref(false);
// const updateTouched = () => {
//   touched = true;
// };
const touchedScheduleCreator = (val: string | null) => {
  fromTime = val;
  scheduleTouched = true;
};
// const updateSelectedHourDuration = () => {
//   scheduleTouched = true;
// };
// const updateSelectedMinuteDuration = (val: string | null) => {
//   scheduleTouched = true;
//   selectedDuration = val;
// };
// watch($$(selectedDuration), () => {
//   selectedHourDuration = selectedDuration?.split(":")[0];
//   selectedMinuteDuration = selectedDuration?.split(":")[1];
// });
// const envStore = useEnv(route.params.id as string);
const interval = $computed(() => {
  let tmpInterval: any[] = [];
  let days = {
    Pon: 0,
    Wt: 1,
    Sro: 2,
    Czw: 3,
    Pia: 4,
    Sob: 5,
    Nie: 6,
  };
  const sorter = ["Pon", "Wt", "Sro", "Czw", "Pia", "Sob", "Nie"];
  let weekP = 24 * 7 * 60;
  for (const id in scheduleList) {
    let item = scheduleList[id];
    let daysArr = item.days.length ? item.days : sorter;
    let start = parseToMin(item.fromTime);
    let duration = parseToMin(item.duration);
    daysArr.forEach((day) => {
      const offSet = days[day as keyof typeof days] * 24 * 60;
      let nextWeekS;
      if (offSet + start + duration > weekP) {
        nextWeekS = offSet + start + duration - weekP;
      }
      if (nextWeekS) {
        tmpInterval.push([0, nextWeekS, id], [offSet + start, weekP, id]);
      } else {
        tmpInterval.push([offSet + start, offSet + start + duration, id]);
      }
    });
  }
  tmpInterval = tmpInterval.sort((a, b) => a[0] - b[0]);
  if ($$(selectedId)) {
    tmpInterval = tmpInterval.filter(
      (interval) => interval[2] !== selectedId?.toString()
    );
  }
  return tmpInterval;
});
const durationTime = $computed(() => {
  let hour =
    selectedHourDuration!.toString().length === 1
      ? `0${selectedHourDuration}`
      : selectedHourDuration;
  if (!hour) hour = "00";
  let min =
    selectedMinuteDuration!.toString().length === 1
      ? `0${selectedMinuteDuration}`
      : selectedMinuteDuration;
  return `${hour}:${min}`;
});
onMounted(() => {
  //   let tmpScheduleList = [];
  let days = {
    1: "Pon",
    2: "Wt",
    3: "Sro",
    4: "Czw",
    5: "Pia",
    6: "Sob",
    0: "Nie",
  };
  let tmpScheduleList: any = [];

  type daysKeys = keyof typeof days;
  api
    .get("/env-info", { queries: { env: route.params.id as string } })
    .then((res) => {
      if (!is(z.string())(res)) {
        offCrons = res.Env.Schedule.OffCrons || [];
        onCrons = res.Env.Schedule.OnCrons || [];
        cronSwitcher = res.Env.Schedule.Active;
        backendActiveState = res.Env.Schedule.Active;
        // value = res.Env.Mode;
        tmpValue = value;
        for (const id in onCrons) {
          let itemOn = onCrons[parseInt(id)].split(" ");
          let itemOff = offCrons[parseInt(id)].split(" ");

          let hourOn = itemOn[1].length === 1 ? `0${itemOn[1]}` : itemOn[1];
          let minuteOn = itemOn[0].length === 1 ? `0${itemOn[0]}` : itemOn[0];

          let hourOff = itemOff[1].length === 1 ? `0${itemOff[1]}` : itemOff[1];
          let minuteOff =
            itemOff[0].length === 1 ? `0${itemOff[0]}` : itemOff[0];
          let startInMin = Number(hourOn) * 60 + Number(minuteOn);
          let stopInMin = Number(hourOff) * 60 + Number(minuteOff);

          let duration;
          let offSet;
          if (itemOn[4] !== "*") {
            if (Number(itemOn[4]) > Number(itemOff[4])) {
              offSet = 7 - Number(itemOn[4]) + Number(itemOff[4]);
            } else {
              offSet = Number(itemOff[4]) - Number(itemOn[4]);
            }
          }
          if (offSet) {
            if (startInMin > stopInMin && offSet === 0) {
              duration = 6 * 24 * 60 + 24 * 60 - startInMin + stopInMin;
            } else {
              if (startInMin > stopInMin) {
                duration =
                  24 * 60 - startInMin + stopInMin + (offSet - 1) * 24 * 60;
              } else {
                duration = stopInMin - startInMin + offSet * 24 * 60;
              }
            }
          } else {
            if (startInMin > stopInMin) {
              duration = 24 * 60 - startInMin + stopInMin;
            } else {
              duration = stopInMin - startInMin;
            }
          }

          let durHour =
            parseInt((duration / 60).toString()).toString().length === 1
              ? `0${parseInt((duration / 60).toString())}`
              : parseInt((duration / 60).toString());
          let durMin =
            (duration % 60).toString().length === 1
              ? `0${duration % 60}`
              : duration % 60;
          if (!tmpScheduleList.length) {
            tmpScheduleList.push({
              fromTime: `${hourOn}:${minuteOn}`,
              duration: `${durHour}:${durMin}`,
              days:
                itemOn[4] !== "*"
                  ? [days[parseInt(itemOn[4]) as daysKeys]]
                  : [],
            });
          } else {
            //
            let lastEl = tmpScheduleList[tmpScheduleList.length - 1];
            if (lastEl.fromTime === `${hourOn}:${minuteOn}`) {
              tmpScheduleList[tmpScheduleList.length - 1].days.push(
                days[parseInt(itemOn[4]) as daysKeys]
              );
            } else {
              tmpScheduleList.push({
                fromTime: `${hourOn}:${minuteOn}`,
                duration: `${durHour}:${durMin}`,
                days:
                  itemOn[4] !== "*"
                    ? [days[parseInt(itemOn[4]) as daysKeys]]
                    : [],
              });
            }
          }
        }
        scheduleList = tmpScheduleList;
        lengthOffCrons = scheduleList.length;
      }
    });
});

// const buttons =
const scheduleDays = [
  {
    name: t("time.days.Mon"),
    value: "Pon",
  },
  {
    name: t("time.days.Tue"),
    value: "Wt",
  },
  {
    name: t("time.days.Wed"),
    value: "Sro",
  },
  {
    name: t("time.days.Thu"),
    value: "Czw",
  },
  {
    name: t("time.days.Fri"),
    value: "Pia",
  },
  {
    name: t("time.days.Sat"),
    value: "Sob",
  },
  {
    name: t("time.days.Sun"),
    value: "Nie",
  },
];
const addDayFn = (day: string): void => {
  if (selectedDays.includes(day)) {
    selectedDays = selectedDays.filter((days) => days !== day);
  } else {
    selectedDays.push(day);
  }
};
const removeSchedule = (idx: number) => {
  scheduleList.splice(idx, 1);
};
const parseToMin = (input: string) => {
  let [hour, min] = input.split(":").map((i) => Number(i));
  return hour * 60 + min;
};

const openScheduleDialog = (id: number) => {
  selectedId = id;
  fromTime = scheduleList[id].fromTime;
  selectedHourDuration = scheduleList[id].duration.split(":")[0];
  selectedMinuteDuration = scheduleList[id].duration.split(":")[1];
  selectedDuration = scheduleList[id].duration;
  selectedDays = scheduleList[id].days;
  openCronSchedule = true;
};
const openCronScheduleFn = () => {
  fromTime = null;
  selectedHourDuration = null;
  selectedMinuteDuration = null;
  selectedDays = [];
  selectedId = null;
  selectedDuration = null;
  openCronSchedule = true;
};
watch(
  () => openCronSchedule,
  () => {
    shouldLock = openCronSchedule ? true : false;
  }
);
watch(
  () => openCronSchedule,
  () => {
    if (!openCronSchedule) {
      scheduleTouched = false;
    }
  }
);
const changeActive = () => {
  let days = {
    Pon: 1,
    Wt: 2,
    Sro: 3,
    Czw: 4,
    Pia: 5,
    Sob: 6,
    Nie: 0,
  };
  let onCrons: any[] = [];
  let offCrons: any[] = [];

  scheduleList.forEach((list) => {
    let [hour, min] = JSON.stringify(list.fromTime)
      .replaceAll('"', "")
      .split(":")
      .map((i) => Number(i));
    if (list.fromTime) {
      if (list.days.length) {
        list.days.forEach((day) => {
          onCrons.push(`${min} ${hour} * * ${days[day as keyof typeof days]}`);
        });
      } else {
        onCrons.push(`${min} ${hour} * * *`);
      }
    }
    if (list.duration) {
      let [dHour, dMin] = JSON.stringify(list.duration)
        .replaceAll('"', "")
        .split(":")
        .map((i) => Number(i));

      let offTime = hour * 60 + min + (dHour * 60 + dMin);

      if (!list.days.length) {
        if (offTime > 24 * 60) {
          offTime = offTime - 24 * 60;
        }
        let offTimeMin = offTime % 60;
        let offTimeHour = parseInt((offTime / 60).toString());
        offCrons.push(
          `${offTimeMin} ${offTimeHour == 24 ? 0 : offTimeHour} * * *`
        );
      } else {
        if (offTime > 24 * 60) {
          list.days.forEach((day) => {
            let offSet =
              days[day as keyof typeof days] +
              parseInt((offTime / (24 * 60)).toString());
            if (offSet > 6) {
              offSet = offSet - 7;
            }
            let offTimeDay = offTime % (24 * 60);
            let offTimeMin = offTimeDay % 60;
            let offTimeHour = parseInt((offTimeDay / 60).toString());
            offCrons.push(
              `${offTimeMin} ${
                offTimeHour == 24 ? 0 : offTimeHour
              } * * ${offSet}`
            );
          });
        } else {
          let offTimeMin = offTime % 60;
          let offTimeHour = parseInt((offTime / 60).toString());
          list.days.forEach((day) => {
            offCrons.push(
              `${offTimeMin} ${offTimeHour == 24 ? 0 : offTimeHour} * * ${
                days[day as keyof typeof days]
              }`
            );
          });
        }
      }
    }
  });
  api
    .post(
      "/env-schedule-set",
      {
        // Mode: tmpValue as number,
        Location: "Europe/Warsaw",
        OnCrons: onCrons,
        OffCrons: offCrons,
        EnvID: route.params.id as string,
        Active: cronSwitcher,
      },
      {
        queries: {
          EnvID: route.params.id as string,
        },
      }
    )
    .then((res) => {
      // if (res === "permission denied") {
      //   message.error(t("messages.permissionDenied"));
      //   return;
      // }
      if (res === "ok") {
        value = tmpValue;
        message.success("Schedule changed");

        openCronModal = !openCronModal;
        lengthOffCrons = scheduleList.length;

        api
          .get("/env-info", { queries: { env: route.params.id as string } })
          .then((res) => {
            if (!is(z.string())(res)) {
              if (env.value.EnvInfo) env.value.EnvInfo.Env = res.Env;
              backendActiveState = res.Env.Schedule.Active;
            }
          });
      } else {
        message.error(res);
      }
    });
};
const createCron = () => {
  let days = {
    Pon: 0,
    Wt: 1,
    Sro: 2,
    Czw: 3,
    Pia: 4,
    Sob: 5,
    Nie: 6,
  };
  let shouldAdd = true;
  const sorter = ["Pon", "Wt", "Sro", "Czw", "Pia", "Sob", "Nie"];
  if (
    fromTime &&
    selectedHourDuration !== null &&
    selectedMinuteDuration !== null
  ) {
    if (selectedDays.length > 1) {
      selectedDays = selectedDays.sort((a, b) => {
        return sorter.indexOf(a) - sorter.indexOf(b);
      });
    }
    const inStart = parseToMin(fromTime);
    const inDuration = parseToMin(durationTime);
    const weekP = 24 * 7 * 60;
    const inDaysLength = selectedDays.length;
    if (selectedDays.length > 1) {
      for (let day in selectedDays) {
        let tmpDay = Number(day);
        if (tmpDay + 1 < inDaysLength) {
          let offSet = days[selectedDays[tmpDay] as keyof typeof days];
          let offSetPlus = days[selectedDays[tmpDay + 1] as keyof typeof days];
          if (
            offSet * 24 * 60 + inStart + inDuration >=
            offSetPlus * 24 * 60 + inStart
          ) {
            shouldAdd = false;
          }
        } else {
          let offSet = days[selectedDays[tmpDay] as keyof typeof days];
          let offSetPlus = days[selectedDays[0] as keyof typeof days];
          let nextWeekS;
          if (offSet * 24 * 60 + inStart + inDuration > weekP) {
            nextWeekS = offSet * 24 * 60 + inStart + inDuration - weekP;
          }
          if (nextWeekS && nextWeekS >= offSetPlus * 24 * 60 + inStart) {
            shouldAdd = false;
          }
        }
      }
    } else {
      // if mniej niz 2 dni
      const inDuration = parseToMin(durationTime);
      if (selectedDays.length === 1) {
        if (inDuration > 7 * 24 * 60 - 1) {
          shouldAdd = false;
        }
      } else {
        if (inDuration >= 24 * 60) {
          shouldAdd = false;
        }
      }
    }
    if (!shouldAdd) {
      // this.minuteError = false;
      // this.hourError = false;
      // this.sendMessageError = "Nieprawidłowy harmonogram";
      // this.sendError = true;
      return;
    }
    let tmpInterval = [];
    let daysArr = selectedDays.length ? selectedDays : sorter;
    if (!interval.length) {
      daysArr.forEach((day) => {
        const offSet = days[day as keyof typeof days] * 24 * 60;
        let nextWeekS;
        if (offSet + inStart + inDuration > weekP) {
          nextWeekS = offSet + inStart + inDuration - weekP;
        }
        if (nextWeekS) {
          tmpInterval.push([0, nextWeekS], [offSet + inStart, weekP]);
        } else {
          tmpInterval.push([offSet + inStart, offSet + inStart + inDuration]);
        }
      });
    }
    let sendErrorId: number | null;
    for (const id in interval) {
      const localInterval = interval[id];
      daysArr.forEach((day) => {
        const offSet = days[day as keyof typeof days] * 24 * 60;
        if (offSet + inStart < localInterval[0]) {
          if (localInterval[0] > offSet + inStart + inDuration) {
            tmpInterval.push([offSet + inStart, offSet + inStart + inDuration]);
          } else {
            shouldAdd = false;
            sendErrorId = interval[id][2];
          }
        } else {
          if (localInterval[1] < offSet + inStart) {
            let nextWeekS;
            if (offSet + inStart + inDuration > weekP)
              nextWeekS = offSet + inStart + inDuration - weekP;
            if (nextWeekS) {
              if (nextWeekS < interval[0][0]) {
                tmpInterval.push([offSet + inStart, weekP], [0, nextWeekS]);
              } else {
                shouldAdd = false;
                sendErrorId = interval[id][2];
              }
            } else {
              tmpInterval.push([
                offSet + inStart,
                offSet + inStart + inDuration,
              ]);
            }
          } else {
            shouldAdd = false;
            sendErrorId = interval[id][2];
          }
        }
      });
    }

    if (shouldAdd) {
      let value = {
        fromTime: fromTime,
        days: selectedDays,
        duration: durationTime,
      };
      // this.$emit("input", value, this.currentScheduleIdx);
      if (selectedId !== null) {
        scheduleList[selectedId!] = value;
      } else {
        scheduleList.push(value);
      }
      openCronSchedule = false;
      // this.minuteError = false;
      // this.hourError = false;
      return;
    } else {
      let el = scheduleList[sendErrorId!];

      let text = `${el.days.join(", ")} ${el.days.length ? "," : ""} from: ${
        el.fromTime
      }, duration: ${el.duration}`;
      // this.sendMessageError = "Harmonogram koliduje z regułą: " + text;
      message.error("Schedule conflicts with rule: " + text);
      // this.sendError = true;
    }
  } else {
    message.error("invalid schedule");
  }
};

// const changeStatus = () => {
//   api
//     .post(
//       "/env-active-status",
//       {
//         Status: tmpValue,
//         Location: "Europe/Warsaw",
//         OnCrons: onCrons,
//         OffCrons: offCrons,
//       },
//       {
//         queries: {
//           env: route.params.id as string,
//         },
//       }
//     )
//     .then((res) => {
//       if (res === "ok") {
//         value = tmpValue;
//         message.success("Schedule changed");
//         confirmChangeStatus = false;
//       }
//     });
// };
let hourDurationRef = $ref<any>(null);
let hourDurationListRef = $ref<any>(null);
let minuteDurationRef = $ref<any>(null);
let minuteDurationListRef = $ref<any>(null);
let openDurationModal = $ref(false);
const updateMinuteDuration = (i: string) => {
  selectedMinuteDuration = i;
};
const updateHourDuration = (i: string) => {
  selectedHourDuration = i;
};
watch([$$(selectedHourDuration), $$(selectedMinuteDuration)], () => {
  if (selectedHourDuration && selectedMinuteDuration) {
    selectedDuration = `${
      selectedHourDuration.length > 1
        ? selectedHourDuration
        : "0" + selectedHourDuration
    }:${
      selectedMinuteDuration.length > 1
        ? selectedMinuteDuration
        : "0" + selectedMinuteDuration
    }`;
    openDurationModal = false;
  }
});
let durationModal = $ref(null);
onClickOutside($$(durationModal), () => (openDurationModal = false));
const openDurationModalF = () => {
  openDurationModal = true;
  nextTick(() => {
    if (
      hourDurationListRef?.querySelector(".active") &&
      minuteDurationListRef?.querySelector(".active")
    ) {
      hourDurationRef?.scrollBy({
        top: hourDurationListRef?.querySelector(".active").offsetTop,
      });
      minuteDurationRef?.scrollBy({
        top: minuteDurationListRef?.querySelector(".active").offsetTop,
      });
    }
  });
};

const editDuration = (duration: string) => {
  let durationSplit = duration.split(":");
  return `${durationSplit[0]}h ${durationSplit[1]}min`;
};
</script>
<template>
  <div style="display: flex; gap: 5px; align-items: center">
    <PopModal
      :title="t('objects.schedule')"
      style="width: 40rem"
      :locked="shouldLock"
      :touched="touched"
      :show="openCronModal"
      v-if="lengthOffCrons !== undefined"
    >
      <template #trigger>
        <n-button
          size="tiny"
          strong
          secondary
          type="primary"
          :disabled="!manage"
        >
          {{
            !backendActiveState
              ? t("actions.noActive")
              : !lengthOffCrons
              ? t("actions.noActive")
              : inactive
              ? t("actions.noActive")
              : custom
              ? t("actions.custom")
              : t("actions.active")
          }}

          <template #icon>
            <n-icon class="icon">
              <mdi :path="mdiClockOutline" />
            </n-icon>
          </template>
        </n-button>
      </template>
      <template #content>
        <div>
          <!-- <n-radio-group
            v-model:value="tmpValue"
            style="width: 100%"
            @update:value="updateTouched"
          >
            <n-radio-button :value="0">{{
              t("actions.active")
            }}</n-radio-button>
            <n-radio-button :value="1">{{
              t("actions.noActive")
            }}</n-radio-button>
            <n-radio-button :value="2">{{
              t("actions.schedule")
            }}</n-radio-button>
          </n-radio-group> -->
          <div class="schedule-list">
            <div
              v-for="(schedule, idx) in scheduleList"
              :key="idx"
              class="schedule-card"
            >
              <p>
                Start:
                {{
                  schedule.days.length === 0 || schedule.days.length === 7
                    ? t("messages.everyDay")
                    : ""
                }}
                {{
                  schedule.days.length !== 7
                    ? schedule.days
                        .map(
                          (day) =>
                            scheduleDays.find(
                              (schedule) => schedule.value === day
                            )?.name
                        )
                        .join(", ")
                    : ""
                }}{{
                  schedule.days.length !== 0 && schedule.days.length !== 7
                    ? ","
                    : ""
                }}
                {{ t("others.from").toLowerCase() }}:
                {{ schedule.fromTime }}
                , {{ t("fields.duration") }}:
                {{ editDuration(schedule.duration) }}
              </p>
              <div>
                <Mdi
                  width="20"
                  :path="mdiClockOutline"
                  @click="openScheduleDialog(idx)"
                  style="cursor: pointer; margin-right: 5px"
                />
                <Mdi
                  style="cursor: pointer"
                  width="20"
                  :path="mdiTrashCan"
                  @click="removeSchedule(idx)"
                />
              </div>
            </div>
            <Btn style="width: 100%" @click="openCronScheduleFn"
              >{{ t("actions.add") }} <Mdi :path="mdiPlus" width="15" />
            </Btn>
          </div>
          <div
            style="
              display: flex;
              justify-content: space-between;
              align-items: flex-end;
            "
          >
            <div style="display: flex; gap: 0.5rem; align-items: center">
              <p>{{ t("actions.schedule") }}:</p>
              <n-switch v-model:value="cronSwitcher" :round="false">
                <template #checked> {{ t("actions.enabled") }} </template>
                <template #unchecked> {{ t("actions.disabled") }} </template>
              </n-switch>
            </div>

            <Btn @click="changeActive">
              {{ t("actions.apply") }}
            </Btn>
          </div>
        </div>
      </template>
    </PopModal>
    <Modal
      :title="t('objects.schedule')"
      v-model:show="openCronSchedule"
      style="width: 26rem"
      :touched="scheduleTouched"
    >
      <div style="display: flex; gap: 20px">
        <div style="">
          <p style="margin-bottom: 10px">{{ t("fields.startTime") }}:</p>

          <div
            style="
              display: flex;
              gap: 10px;
              align-items: center;
              position: relative;
            "
          >
            <Mdi :path="mdiClockTimeFiveOutline" width="20" />
            <n-time-picker
              :placeholder="
                t('actions.select') + ' ' + t('fields.time').toLowerCase()
              "
              format="HH:mm"
              v-model:formatted-value="fromTime"
              style="width: 150px"
              :actions="null"
              @update:formatted-value="touchedScheduleCreator"
            />
          </div>
        </div>
        <div style="">
          <p style="margin-bottom: 10px">{{ t("fields.duration") }}:</p>
          <div
            style="
              display: flex;
              gap: 10px;
              align-items: center;
              position: relative;
            "
          >
            <!-- <div
              style="
                position: relative;
                display: flex;
                align-items: center;
                gap: 10px;
              "
            >
              <label style="white-space: nowrap"
                >{{ t("time.hour", 2) }}:</label
              >
              <n-input
                :placeholder="
                  t('actions.select') + ' ' + t('fields.time').toLowerCase()
                "
                style="width: 130px"
                v-model:value="selectedHourDuration"
                @update:value="updateSelectedHourDuration"
              />
              <div style="position: absolute; right: 10px; top: 9px">
                <Mdi
                  :path="mdiClockTimeThreeOutline"
                  width="16"
                  style="color: rgba(255, 255, 255, 0.38)"
                />
              </div>
            </div> -->
            <Mdi :path="mdiTimerOutline" width="20" />

            <!-- <label style="white-space: nowrap">
              {{ t("time.minute", 2) }}:
            </label> -->
            <div style="position: relative">
              <n-input
                placeholder="Select time"
                style="width: 150px"
                :value="selectedDuration"
                @click="openDurationModalF"
              />
              <Mdi :path="mdiTimerOutline" width="20" class="input-clock" />
              <div
                v-if="openDurationModal"
                class="duration-modal"
                ref="durationModal"
              >
                <n-scrollbar style="height: 250px" ref="hourDurationRef">
                  <ul ref="hourDurationListRef">
                    <li
                      v-for="i in 168"
                      :key="i"
                      :class="
                        parseInt(selectedHourDuration as string) === i - 1 ? 'active' : ''
                      "
                      @click="updateHourDuration((i - 1).toString())"
                    >
                      {{ i - 1 > 9 ? i - 1 : `0${i - 1}` }}
                    </li>
                  </ul>
                </n-scrollbar>
                <n-scrollbar style="height: 250px" ref="minuteDurationRef">
                  <ul ref="minuteDurationListRef">
                    <li
                      v-for="i in 60"
                      :key="i"
                      :class="
                        parseInt(selectedMinuteDuration as string) === (i - 1 ) ? 'active' : ''
                      "
                      @click="updateMinuteDuration((i - 1).toString())"
                    >
                      {{ i - 1 > 9 ? i - 1 : `0${i - 1}` }}
                    </li>
                  </ul>
                </n-scrollbar>
              </div>
            </div>
            <!-- <n-time-picker
              :placeholder="
                t('actions.select') + ' ' + t('fields.time').toLowerCase()
              "
              format="HHH:mm"
              :hours="Array.from(Array(100).keys())"
              style="width: 130px"
              v-model:formatted-value="selectedDuration"
              @update:formatted-value="updateSelectedMinuteDuration"
            /> -->
          </div>
        </div>
      </div>
      <div>
        <p style="margin: 10px 0">{{ t("time.day", 2) }}:</p>
        <div>
          <n-button
            class="cron-button"
            v-for="day in scheduleDays"
            :key="day.name"
            :type="selectedDays.includes(day.value) ? 'primary' : undefined"
            @click="addDayFn(day.value)"
          >
            {{ day.name }}
          </n-button>
        </div>
      </div>
      <Btn style="width: 100%; margin-top: 15px" @click="createCron"
        >{{ t("actions.add") }} <Mdi :path="mdiPlus" width="15"
      /></Btn>
    </Modal>
    <!-- <Modal
      v-model:show="confirmChangeStatus"
      title="Change status"
      style="width: 40rem"
      :showFooter="true"
      @positive-click="changeStatus"
      @negative-click="confirmChangeStatus = false"
    >
      Are u sure?
    </Modal> -->
  </div>
</template>
<style>
.cron-button {
  border-radius: 0;
  margin-right: 1px;
}
</style>
<style scoped lang="scss">
.duration-modal {
  position: absolute;
  left: 0;
  top: 40px;
  display: flex;
  width: 100px;
  background: #48484e;
  border-radius: 5px;
  z-index: 2;
  & li {
    text-align: center;
    padding: 5px;
    margin: 5px;
    border-radius: 5px;
    font-size: 13px;
    &:hover {
      background: #59595e;
    }
    &.active {
      color: #3d9ddd;
      background: #59595e;
    }
  }
}
.input-clock {
  position: absolute;
  right: 12px;
  top: 10px;
  width: 15px;
  color: rgba(255, 255, 255, 0.38);
}
.schedule-list {
  display: grid;
  grid-template-columns: 1fr;
  gap: 0.5rem;
  padding: 1rem 0;
}
.schedule-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem 1rem;
}

.icon {
  cursor: pointer;
  transition: all 0.2s ease-in-out;
  color: var(--n-color-target);

  &:hover {
    filter: brightness(1.4);
  }
}
</style>
