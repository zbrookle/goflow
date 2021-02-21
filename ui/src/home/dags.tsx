import { useEffect, useState } from "react";
import Table from "react-bootstrap/Table";
import Switch from "bootstrap-switch-button-react";
import CSS from "csstype";

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

type DAGProps = {
  Name: string;
  Schedule: string;
  LastRunTime: string;
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
        <CenterColHead>Schedule</CenterColHead>
        <CenterColHead>Name</CenterColHead>
        <CenterColHead>Last Run Time</CenterColHead>
        <CenterColHead>Success/Failures</CenterColHead>
        <CenterColHead>Actions</CenterColHead>
      </tr>
    </thead>
  );
}

function DAG(props: DAGProps) {
  const [dagActive, switchDag] = useState(false);
  var date = new Date();
  const [lastRunTime] = useState(date.toISOString());

  return (
    <tr>
      <CenteredCol>
        <Switch
          size="sm"
          checked={dagActive}
          onChange={() => switchDag(!dagActive)}
        />
      </CenteredCol>
      <CenteredCol>{props.Schedule}</CenteredCol>
      <CenteredCol>{props.Name}</CenteredCol>
      <CenteredCol>{lastRunTime}</CenteredCol>
      <CenteredCol>Success/failures</CenteredCol>
    </tr>
  );
}

function DAGContainer() {
  const [dags, setDAGs] = useState<Record<string, DAGProps>>({});
  useEffect(() => {
    const intervalId = setInterval(() => {
      fetch("http://localhost:8080/dags")
        .then((res) => res.json())
        .then((data) => {
          var record: Record<string, DAGProps> = {};
          data.forEach((dag: any) => {
            record[dag.Config.Name] = {
              Name: dag.Config.Name,
              Schedule: dag.Config.Schedule,
              LastRunTime: dag.MostRecentExecution,
            };
          });
          setDAGs(record);
        });
    }, 5000); // TODO Make this number changeable in the UI
    return () => clearInterval(intervalId);
  }, []);

  return (
    <div>
      <h1>My DAGs</h1>
      <h2>(Dag Search Bar Here)</h2>
      <Table responsive bordered variant="dark" style={styles.table} size="2">
        <DAGColumnHeaders />
        <tbody>
          {Object.entries(dags).map((t, k) => (
            <DAG
              key={t[1].Name}
              Name={t[1].Name}
              Schedule={t[1].Schedule}
              LastRunTime={t[1].LastRunTime}
            />
          ))}
        </tbody>
      </Table>
    </div>
  );
}

export { DAGContainer };
