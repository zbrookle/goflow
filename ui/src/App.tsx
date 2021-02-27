import React from "react";
import Navbar from "react-bootstrap/Navbar";
import Nav from "react-bootstrap/Nav";
import { DAGContainer } from "./home/dag_list";
import { BrowserRouter, Switch, Route } from "react-router-dom";
import { LinkContainer } from "react-router-bootstrap";
import { DagInfo } from "./dag/dag_page";
import { RouterNavLink } from "./routing/router_nav";

type HeaderNavLinkProps = {
  link: string;
  text: string;
}

function HeaderNavLink(props: HeaderNavLinkProps) {
  return <RouterNavLink link={props.link} text={props.text}/>
}

function Header() {
  return (
    <Navbar bg="dark" variant="dark" expand="lg" sticky="top">
      <Navbar.Toggle aria-controls="basic-navbar-nav"></Navbar.Toggle>
      <LinkContainer to={"/home"}>
        <Navbar.Brand>GoFlow</Navbar.Brand>
      </LinkContainer>
      <Navbar.Collapse>
        <Nav>
          <HeaderNavLink link={"/home"} text="Home" />
          <HeaderNavLink link={"/metrics"} text="Metrics" />
          <HeaderNavLink link={"/settings"} text="Settings" />
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
          <Route path="/dag/:name">
            <DagInfo />
          </Route>
        </Switch>
      </div>
    </BrowserRouter>
  );
}

export default App;
