import { useState } from "react";
import Container from "react-bootstrap/Container";
import Row from "react-bootstrap/Row";
import Col from "react-bootstrap/Col";
import Table from "react-bootstrap/Table";
import Switch from "bootstrap-switch-button-react";

const styles = {
  row: {
    // padding: 8,
  },
  col: {
    padding: 3,
  },
};

const ColOffsets = {
  Status: 0,
  Name: 0,
  RunTime: 0,
  Successes: 0,
};

type DAGProps = {
  Name: string;
};

function CenterColHead(props: any) {
    return <th className="my-auto">{props.children}</th>
}

function CenteredCol(props: any) {
  return <td className="my-auto">{props.children}</td>;
}

function DAGColumnHeaders() {
  return (
    <thead>
      <tr style={styles.row}>
        <CenterColHead>Status</CenterColHead>
        <CenterColHead>Name</CenterColHead>
        <CenterColHead>Last Run Time</CenterColHead>
        <CenterColHead>Success/Failures</CenterColHead>
      </tr>
    </thead>
  );
}

function DAG(props: DAGProps) {
  const [dagActive, switchDag] = useState(false);
  var date = new Date();
  const [lastRunTime] = useState(date.toISOString());

  return (
    <tr style={styles.row}>
      <CenteredCol colOffset={ColOffsets.Status}>
        <Switch checked={dagActive} onChange={() => switchDag(!dagActive)} />
      </CenteredCol>
      <CenteredCol xs={ColOffsets.Name}>{props.Name}</CenteredCol>
      <CenteredCol xs={ColOffsets.RunTime}>{lastRunTime}</CenteredCol>
      <CenteredCol xs={ColOffsets.Successes}>Success/failures</CenteredCol>
    </tr>
  );
}

function DAGContainer() {
  let dags: Record<number, string> = { 1: "test" };
  for (var i = 0; i < 40; i++) {
    dags[i] = "test" + i.toString();
  }
  return (
    <div>
      My DAGs
      <p />
      (Dag Search Bar Here)
      <Table bordered variant="dark">
        <DAGColumnHeaders />
        <tbody>
          {Object.entries(dags).map((t, k) => (
            <DAG Name={t[1]} />
          ))}
        </tbody>
      </Table>
    </div>
  );
}

export { DAGContainer };
