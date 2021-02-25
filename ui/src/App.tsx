import React from "react";
import Navbar from "react-bootstrap/Navbar";
import Nav from "react-bootstrap/Nav";
import { DAGContainer } from "./home/dags";
import { BrowserRouter, Switch, Route, Link } from "react-router-dom";
function Header() {
  return (
    <Navbar bg="dark" variant="dark" expand="lg" sticky="top">
      <Navbar.Toggle aria-controls="basic-navbar-nav"></Navbar.Toggle>
      <Navbar.Brand href="#home">GoFlow</Navbar.Brand>
      <Navbar.Collapse>
        <Nav>
          <Link to="/">Home</Link>
          <Nav.Link href="#metrics">Metrics</Nav.Link>
          <Nav.Link href="">Settings</Nav.Link>
          <Link to="/dag">DAG</Link>
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
          <Route path="/">
            <DAGContainer />
          </Route>
        </Switch>
      </div>
    </BrowserRouter>
  );
}

export default App;
