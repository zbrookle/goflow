import type { DAGProps } from "../typing/dag_types";
import { Container, Row, Col, Card, Nav } from "react-bootstrap";
import { OnOffButton } from "../buttons/on_off_button";
import { BrowserRouter, Switch, Route } from "react-router-dom";
import { RouterNavLink } from "../routing/router_nav";

type DagInfoProps = DAGProps & {
  JobName: string;
  MaxMemoryUsage: number;
  Successes: number;
  Failures: number;
};

type CardTabProps = {
  Ref: string;
  Label: string;
};

function CardTab(props: CardTabProps) {
  return (
    <Nav.Item>
      <RouterNavLink link={props.Ref} text={props.Label} />
    </Nav.Item>
  );
}

function DagInfo(props: DagInfoProps) {
  return (
    <Container style={{ marginTop: "1%" }}>
      <Row noGutters={true}>
        <div
          style={{
            alignItems: "center",
            paddingRight: "7px",
            justifyContent: "center",
            display: "flex",
          }}
        >
          <OnOffButton Name={props.Name} IsOn={props.IsOn} />
        </div>
        <Col>
          <h1>{props.Name}</h1>
        </Col>
      </Row>
      <Row>
        <Card>
          <Card.Header>
            <Nav variant="tabs" defaultActiveKey="#metrics">
              <CardTab Ref="#metrics" Label="Metrics" />
              <CardTab Ref="#timeline" Label="Timeline" />
              <CardTab Ref="#runtimes" Label="Run Times" />
              <CardTab Ref="#resources" Label="Resource Usage" />
            </Nav>
          </Card.Header>
          <Card.Body>Hello!</Card.Body>
          <Card.Footer className="text-muted">Last Updated:</Card.Footer>
        </Card>
      </Row>
    </Container>
  );
}

export { DagInfo };
