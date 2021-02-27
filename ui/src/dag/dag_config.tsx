import { DAGConfigProps } from "../typing/dag_types";
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
  console.log(props.config);
  //   return <Table variant="dark", style >test!</Table>;
  return (
    <AppTable>
      <tbody>
        <th className="my-auto">Setting</th>
        <th className="my-auto">Value</th>
      </tbody>
    </AppTable>
  );
}
