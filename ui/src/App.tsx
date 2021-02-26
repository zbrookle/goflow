import React from "react";
import Navbar from "react-bootstrap/Navbar";
import Nav from "react-bootstrap/Nav";
import { DAGContainer } from "./home/dag_list";
import { BrowserRouter, Switch, Route } from "react-router-dom";
import { LinkContainer } from "react-router-bootstrap";
import { DagInfo } from "./dag/dag_page";
import { RouterNavLink } from "./routing/router_nav";

function Header() {
  return (
    <Navbar bg="dark" variant="dark" expand="lg" sticky="top">
      <Navbar.Toggle aria-controls="basic-navbar-nav"></Navbar.Toggle>
      <LinkContainer to={"/home"}>
        <Navbar.Brand>GoFlow</Navbar.Brand>
      </LinkContainer>
      <Navbar.Collapse>
        <Nav>
          <RouterNavLink link={"/home"} text="Home" />
          <RouterNavLink link={"/metrics"} text="Metrics" />
          <RouterNavLink link={"/settings"} text="Settings" />
          <RouterNavLink link={"/dag/metrics"} text="DAG" />
          <Nav.Link href="https://github.com/zbrookle/goflow">
            Documentation
          </Nav.Link>
        </Nav>
      </Navbar.Collapse>
    </Navbar>
  );
}

function App() {
  return (
    <BrowserRouter>
      <div>
        <Header />
        <Switch>
          <Route path="/home">
            <DAGContainer />
          </Route>
          <Route path="/dag">
            <DagInfo
              Namespace="default"
              Name="test-dag"
              Schedule="* * * *"
              LastRunTime="Never"
              IsOn={false}
              JobName={"test-run"}
              MaxMemoryUsage={20}
              Successes={10}
              Failures={5}
            />
          </Route>
        </Switch>
      </div>
    </BrowserRouter>
  );
}

export default App;
