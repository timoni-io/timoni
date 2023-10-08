import { createModule } from ".";
import { useLogsStore } from "@/store/logsStore";
// import { RequestMessage } from "@/next-api/types";
type GetLogDetails = ({
  time,
  envId,
}: {
  time: string;
  envId: string;
}) => Promise<any>;

type GetLogsVector = ({
  time,
  envId,
  count,
  Where,
}: {
  time: string;
  count: number;
  envId: string;
  Where?: any;
}) => Promise<any>;

export type GetLogs = {
  getLogDetails: GetLogDetails;
  getLogsVector: GetLogsVector;
};

export const GetLogsModule = createModule({
  setup(core, api) {
    const logsStore = useLogsStore();
    const getLogDetails: GetLogDetails = ({ time, envId }) =>
      core.sendWithResponse((id) => {
        logsStore.id.push({
          id,
          live: false,
          fullLog: true,
        });
        return {
          RequestID: id,
          Action: "Get",
          Args: {
            Querys: [
              {
                Query: {
                  Type: "ONE",
                  Time: time,
                  EnvID: envId,
                  FullLog: true,
                },
              },
            ],
          },
        };
      });

    const getLogsVector = (Queries: any) =>
      core.sendWithResponse((id) => {
        logsStore.id.push({
          id,
          live: false,
          fullLog: false,
        });
        return {
          RequestID: id,
          Action: "Get",
          Args: {
            Querys: Queries,
          },
        };
      });

    // @ts-ignore
    api.getLogDetails = getLogDetails;
    // @ts-ignore
    api.getLogsVector = getLogsVector;
  },
});

//   core
//   .sendWithResponse(
//     (id) => ({
//       RequestID: id,
//       Action: "Get",
//       Args: {
//         //@ts-ignore
//         Query: {
//           Type: "ONE",
//           Time: 1666177132777753065n,
//           EnvID: "image_builder",
//         },
//       },
//     }),
//     true
//   )
