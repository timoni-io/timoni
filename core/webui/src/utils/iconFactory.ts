import chroma from "chroma-js";

export default function iconFactory(level: string) {
  let style = getComputedStyle(document.body);
  switch (level) {
    case "WARN": {
      return {
        color: style.getPropertyValue("--warningColorHover"),
        icon: mdiAlertCircleOutline,
      };
    }
    case "INFO": {
      return {
        color: style.getPropertyValue("--infoColorHover"),
        icon: mdiInformation,
      };
    }
    case "ERROR": {
      return {
        color: style.getPropertyValue("--errorColorHover"),
        icon: mdiAlert,
      };
    }
    case "DEBUG": {
      return {
        color: style.getPropertyValue("--primaryColorHover"),
        icon: mdiDecagram,
      };
    }
    case "FATAL": {
      return {
        color: "#EB06EB",
        icon: mdiSkullCrossbonesOutline,
      };
    }
    default: {
      return {
        color: "currentColor",
        icon: mdiInformation,
      };
    }
  }
}

export interface Icon {
  icon: string;
  background: string;
  color: string;
  iconSize: number;
}

export function envIcon(status: number): Icon {
  switch (status) {
    case 0:
      return {
        icon: mdiAutorenew,
        background: "var(--warningColorSuppl)",
        color: "var(--warningColor)",
        iconSize: 15,
      };

    case 1:
      return {
        icon: mdiAutorenew,
        background: "var(--warningColorSuppl)",
        color: "var(--warningColor)",
        iconSize: 15,
      };
    case 2:
      return {
        icon: mdiExclamationThick,
        background: "var(--errorColorSuppl)",
        color: "var(--errorColor)",
        iconSize: 15,
      };
    case 3:
      return {
        icon: mdiMenuRight,
        background: "var(--backgroundSuccess)",
        color: "var(--successColor)",
        iconSize: 20,
      };
    case 4:
      return {
        icon: mdiDeleteClock,
        background: chroma("purple").darken().hex(),
        color: "purple",
        iconSize: 13,
      };
    case 5:
      return {
        icon: mdiAutorenew,
        background: "var(--warningColorSuppl)",
        color: "var(--warningColor)",
        iconSize: 15,
      };
    case 6:
      return {
        icon: mdiPause,
        background: "var(--iconColorHover)",
        color: "var(--iconColor)",
        iconSize: 13,
      };
    default:
      return {
        icon: mdiCrosshairsQuestion,
        background: "var(--infoColorSuppl)",
        color: "",
        iconSize: 10,
      };
  }
}

export function podIcon(status: number): Icon {
  switch (status) {
    case 0:
      return {
        icon: mdiAutorenew,
        background: "var(--warningColorSuppl)",
        color: "var(--warningColor)",
        iconSize: 15,
      };
    case 1:
      return {
        icon: mdiAutorenew,
        background: "var(--warningColorSuppl)",
        color: "var(--warningColor)",
        iconSize: 15,
      };
    case 2:
      return {
        icon: mdiAutorenew,
        background: "var(--warningColorSuppl)",
        color: "var(--warningColor)",
        iconSize: 15,
      };
    case 3:
      return {
        icon: mdiMenuRight,
        background: "var(--backgroundSuccess)",
        color: "var(--successColor)",
        iconSize: 20,
      };
    case 4:
      return {
        icon: mdiMenuRight,
        background: "var(--backgroundSuccess)",
        color: "var(--successColor)",
        iconSize: 20,
      };
    case 5:
      return {
        icon: mdiExclamationThick,
        background: "var(--errorColorSuppl)",
        color: "var(--errorColor)",
        iconSize: 15,
      };
    case 6:
      return {
        icon: mdiDeleteClock,
        background: chroma("purple").darken().hex(),
        color: "purple",
        iconSize: 13,
      };
    case 7:
      return {
        icon: mdiMenuRight,
        background: "var(--backgroundSuccess)",
        color: "var(--successColor)",
        iconSize: 20,
      };
    default:
      return {
        icon: mdiCrosshairsQuestion,
        background: "var(--infoColorSuppl)",
        color: "",
        iconSize: 10,
      };
  }
}

export function envContainers(status: number) {
  switch (status) {
    case 0:
      return {
        color: chroma("purple"),
      };

    case 1:
      return {
        color: "#3EB77F",
      };

    case 2:
      return {
        color: "#D26D09",
      };

    case 3:
      return {
        color: "#ff0000",
      };
  }
}
