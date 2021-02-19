import { useState } from "react";
import Container from "react-bootstrap/Container";
import Row from "react-bootstrap/Row";
import Col from "react-bootstrap/Col";
import Switch from "bootstrap-switch-button-react";

const styles = {
    row: {
        padding: 8
    },
    col: {
        padding: 3
    }
}

const ColOffsets = {
    Status: 2,
    Name: 1,
    RunTime: 0,
    Successes: 4
}

type DAGProps = {
  Name: string;
};

function DAG(props: DAGProps) {
  const [dagActive, switchDag] = useState(false);
  var date = new Date()
  const [lastRunTime] = useState(date.toISOString())

  return (
    <Row style={styles.row}>
      <Col xs={ColOffsets.Status}>
        <Switch checked={dagActive} onChange={() => switchDag(!dagActive)} />
      </Col>
      <Col xs={ColOffsets.Name}>{props.Name}</Col>
      <Col xs={ColOffsets.RunTime}>{lastRunTime}</Col>
      <Col xs={ColOffsets.Successes}>Success/failures</Col>
    </Row>
  );
}

function DAGContainer() {
  let dags: Record<number, string> = {1: "test"}
  for (var i = 0; i < 13; i ++) {
     dags[i] = "test" + i.toString();
  }
  return (
    <Container>
        { Object.entries(dags).map((t,k) => <DAG Name={t[1]} />) }          
    </Container>
  );
}

export { DAGContainer }