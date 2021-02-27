import { DAGConfigProps, getConfigKeys } from "../typing/dag_types";
import { AppTable } from "../tables/table";

//   <Table responsive bordered variant="dark" style={styles.table} size="2">
//     <DAGColumnHeaders />
//     <tbody>
//       {Object.entries(dags).map((t, k) => {
//         return <DAGRow key={t[1].config.Name} dag={t[1]} />;
//       })}
//     </tbody>
//   </Table>

export function DAGConfigBody(props: DAGConfigProps) {
  let settingNames = getConfigKeys(props.config);
  let config = props.config
  let configStringRecord: Record<string, string> = {}
  Object.entries(settingNames).forEach((t) => {
    let configKey = t[1]
    let settingString = configKey as string
    if (config[configKey] !== null) {
        if (configKey === "Command") {
          configStringRecord["Command"] = config.Command.join(" ")
        } else {
          configStringRecord[settingString] = config[configKey].toString()
        }
    }
  })

  return (
    <AppTable>
      <tbody>
        <th>Setting</th>
        <th>Value</th>
        {Object.entries(configStringRecord).map((t) => {
          return (
            <tr>
              <td>{t[0]}</td>
              <td>{t[1]}</td>
            </tr>
          );
        })}
      </tbody>
    </AppTable>
  );
}
