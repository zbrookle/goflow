// import type { DAGProps } from "../typing/dag_types";
import { Container, Row, Col, Card, Nav } from "react-bootstrap";
import { OnOffButton } from "../buttons/on_off_button";
import { Switch, Route, useRouteMatch, useParams } from "react-router-dom";
import { RouterNavLink } from "../routing/router_nav";
import { fetchDAG, fetchDAGObject } from "../backend/fetch_calls";

// type DagInfoProps = DAGProps & {
//   MaxMemoryUsage: number;
//   Successes: number;
//   Failures: number;
// };

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
      <RouterNavLink
        link={link}
        text={props.Label}
        style={{ color: "black" }}
        hoverStyle={{ color: "black" }}
      />
    </Nav.Item>
  );
}

type DagPropName = {
  name: string;
};

function DagInfo() {
  let { path, url } = useRouteMatch();

  let { name } = useParams<DagPropName>();
  let defaultPath = path.replace("/metrics", "")
  let defaultURL = url.replace("/metrics", "")
  let dag = fetchDAGObject(name);

  // fetchDAG(name).then((data) => console.log(data));

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
          <OnOffButton Name={name} IsOn={dag.isOn} />
        </div>
        <Col>
          <h1>{name}</h1>
        </Col>
      </Row>
      <Row>
        <Card>
          <Card.Header>
            <Nav variant="tabs" defaultActiveKey={getPath(defaultURL, "metrics")}>
              <CardTab Path={defaultURL} Ref="metrics" Label="Metrics" />
              <CardTab Path={defaultURL} Ref="timeline" Label="Timeline" />
              <CardTab Path={defaultURL} Ref="runtimes" Label="Run Times" />
              <CardTab Path={defaultURL} Ref="resources" Label="Resource Usage" />
            </Nav>
          </Card.Header>
          <Switch>
            <Route path={getPath(defaultPath, "metrics")}>
              <Card.Body>
                <p>Current Job Name: {"test"}</p>
                <p>Schedule: {"dag.config.schedule"}</p>
                <p>Successes: {0}</p>
                <p>Failures: {0}</p>
                <p>Max Memory Usage: {0}</p>
                <p>Logs</p>
              </Card.Body>
            </Route>
            <Route path={getPath(defaultPath, "timeline")}>
              <Card.Body>Timeline!</Card.Body>
            </Route>
            <Route path={getPath(defaultPath, "runtimes")}>
              <Card.Body>Run run run!</Card.Body>
            </Route>
            <Route path={getPath(defaultPath, "resources")}>
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
