import { useEffect, useState } from "react";
import Table from "react-bootstrap/Table";
import { OnOffButton } from "../buttons/on_off_button";
import CSS from "csstype";
import { DAGProps } from "../typing/dag_types";
import { fetchDAGs, DAG } from "../backend/fetch_calls";
import { RouterNavLink } from "../routing/router_nav";

const styles = {
  table: {
    marginLeft: "1%",
    width: "98%",
  },
};

const tdStyles: CSS.Properties = {
  wordWrap: "normal",
  margin: 0,
  whiteSpace: "pre-line",
};

function CenterColHead(props: any) {
  return (
    <th className="my-auto" style={tdStyles}>
      {props.children}
    </th>
  );
}

function CenteredCol(props: any) {
  return (
    <td className="my-auto" style={tdStyles}>
      {props.children}
    </td>
  );
}

function DAGColumnHeaders() {
  return (
    <thead>
      <tr>
        <CenterColHead>Status</CenterColHead>
        <CenterColHead>Namespace</CenterColHead>
        <CenterColHead>Schedule</CenterColHead>
        <CenterColHead>Name</CenterColHead>
        <CenterColHead>Last Run Time</CenterColHead>
        <CenterColHead>Success/Failures</CenterColHead>
        <CenterColHead>Actions</CenterColHead>
      </tr>
    </thead>
  );
}

function DAGRow(props: DAGProps) {
  var date = new Date();
  const [lastRunTime] = useState(date.toISOString());

  return (
    <tr>
      <CenteredCol>
        <OnOffButton Name={props.dag.config.Name} IsOn={props.dag.isOn} />
      </CenteredCol>
      <CenteredCol>{props.dag.config.Namespace}</CenteredCol>
      <CenteredCol>{props.dag.config.Schedule}</CenteredCol>
      <CenteredCol>
        <RouterNavLink
          link={`/dag/${props.dag.config.Name}/metrics`}
          text={props.dag.config.Name}
          style={{ color: "#999D9F" }}
          hoverStyle={{ color: "#cccecf" }}
        />
      </CenteredCol>
      <CenteredCol>{lastRunTime}</CenteredCol>
      <CenteredCol>Success/failures</CenteredCol>
    </tr>
  );
}

function DAGContainer() {
  const [dags, setDAGs] = useState<Record<string, DAG>>({});
  useEffect(() => {
    const intervalId = setInterval(() => {
      fetchDAGs().then((data) => {
        var record: Record<string, DAG> = {};
        data.forEach((restDAG: any) => {
          let dag = { config: restDAG.Config } as DAG;
          record[restDAG.Config.Name] = dag;
        });
        setDAGs(record);
      });
    }, 10); // TODO Make this number changeable in the UI
    return () => clearInterval(intervalId);
  }, []);
  return (
    <div>
      <h1>My DAGs</h1>
      <h2>(Dag Search Bar Here)</h2>
      <Table responsive bordered variant="dark" style={styles.table} size="2">
        <DAGColumnHeaders />
        <tbody>
          {Object.entries(dags).map((t, k) => {
            return <DAGRow key={t[1].config.Name} dag={t[1]} />;
          })}
        </tbody>
      </Table>
    </div>
  );
}

export { DAGContainer };
