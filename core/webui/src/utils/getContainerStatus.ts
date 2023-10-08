export function getContainerStatus(status: number) {
  switch (true) {
    case status == 0:
      return { color: "#afafaf", label: "New" };
    case status == 1:
      return {
        color: "#f0e320",
        label: "Pending",
      };
    case status == 2:
      return {
        color: "var(--warningColor)",
        label: "Creating",
      };
    case status == 3:
      return {
        color: "var(--warningColor)",
        label: "Running",
      };
    case status == 4:
      return {
        color: "var(--successColor)",
        label: "Succeeded",
      };
    case status == 5:
      return {
        color: "var(--errorColor)",
        label: "Failed",
      };
    case status == 6:
      return {
        color: "var(--infoColor)",
        label: "Terminating",
      };
    case status == 7:
      return {
        color: "var(--successColor)",
        label: "Ready",
      };
    default:
      return {
        color: "#afafaf",
        label: "Unknown status",
      };
  }
}

export function actionStatusToNumber(status: string): number {
  switch (status) {
    case "Pending":
      return 1;
    case "Running":
      return 3;
    case "Succeeded":
      return 4;
    case "Failed":
      return 5;
    case "Terminating":
      return 6;
    default:
      return 0;
  }
}
