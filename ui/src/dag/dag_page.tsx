import type { DAGProps } from "../typing/dag_types";
import { Container, Row, Col, Card, Nav } from "react-bootstrap";
import { OnOffButton } from "../buttons/on_off_button";
import { Switch, Route, useRouteMatch } from "react-router-dom";
import { RouterNavLink } from "../routing/router_nav";

type DagInfoProps = DAGProps & {
  JobName: string;
  MaxMemoryUsage: number;
  Successes: number;
  Failures: number;
};

function getPath(path: string, name: string) {
  return `${path}/${name}`;
}

type CardTabProps = {
  Ref: string;
  Label: string;
  Path: string;
};

function CardTab(props: CardTabProps) {
  let link = getPath(props.Path, props.Ref);

  return (
    <Nav.Item>
      <RouterNavLink link={link} text={props.Label} />
    </Nav.Item>
  );
}

function DagInfo(props: DagInfoProps) {
  let { path, url } = useRouteMatch();

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
            <Nav variant="tabs" defaultActiveKey={getPath(url, "metrics")}>
              <CardTab Path={url} Ref="metrics" Label="Metrics" />
              <CardTab Path={url} Ref="timeline" Label="Timeline" />
              <CardTab Path={url} Ref="runtimes" Label="Run Times" />
              <CardTab Path={url} Ref="resources" Label="Resource Usage" />
            </Nav>
          </Card.Header>
          <Switch>
            <Route exact path={getPath(path, "metrics")}>
              <Card.Body>Metrics!</Card.Body>
            </Route>
            <Route path={getPath(path, "timeline")}>
              <Card.Body>Timeline!</Card.Body>
            </Route>
            <Route path={getPath(path, "runtimes")}>
              <Card.Body>Run run run!</Card.Body>
            </Route>
            <Route path={getPath(path, "resources")}>
              <Card.Body>mems and cps!</Card.Body>
            </Route>
          </Switch>
          <Card.Footer className="text-muted">Last Updated:</Card.Footer>
        </Card>
      </Row>
    </Container>
  );
}

export { DagInfo };
