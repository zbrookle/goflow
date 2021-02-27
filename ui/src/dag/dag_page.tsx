// import type { DAGProps } from "../typing/dag_types";
import { Container, Row, Col, Card, Nav } from "react-bootstrap";
import { OnOffButton } from "../buttons/on_off_button";
import { Switch, Route, useRouteMatch, useParams } from "react-router-dom";
import { RouterNavLink } from "../routing/router_nav";
import { DAG, fetchDAG } from "../backend/fetch_calls";
import { useState } from "react";
import { useComponentWillMount } from "../hooks/component_will_mount";

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
  let defaultPath = path.replace("/metrics", "");
  let defaultURL = url.replace("/metrics", "");
  let obj = { config: { Schedule: "", Name: "" } } as DAG;
  const [dag, setDAG] = useState<DAG>(obj);
  const [currentActiveRun, setCurrentActiveRun] = useState("N/A");
  useComponentWillMount(() => {
    fetchDAG(name).then((restDAG) => {
      console.log(restDAG);
      let dag = {
        config: restDAG.Config,
        isOn: restDAG.IsOn,
        DAGRuns: restDAG.DAGRuns,
      } as DAG;
      setDAG(dag);

      if (dag.DAGRuns.length !== 0) {
        setCurrentActiveRun(dag.DAGRuns[0].Name);
      }
    });
  });

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
            <Nav
              variant="tabs"
              defaultActiveKey={getPath(defaultURL, "metrics")}
            >
              <CardTab Path={defaultURL} Ref="metrics" Label="Metrics" />
              <CardTab Path={defaultURL} Ref="timeline" Label="Timeline" />
              <CardTab Path={defaultURL} Ref="runtimes" Label="Run Times" />
              <CardTab
                Path={defaultURL}
                Ref="resources"
                Label="Resource Usage"
              />
            </Nav>
          </Card.Header>
          <Switch>
            <Route path={getPath(defaultPath, "metrics")}>
              <Card.Body>
                <p>Current Job Name: {currentActiveRun}</p>
                <p>Schedule: {dag.config.Schedule}</p>
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
