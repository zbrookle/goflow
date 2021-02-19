import React from "react";
import Navbar from "react-bootstrap/Navbar";
import Nav from "react-bootstrap/Nav";
import {DAGContainer} from "./dags";

function Header() {
  return (
      <Navbar bg="dark" variant="dark" expand="lg">
        <Navbar.Toggle aria-controls="basic-navbar-nav"></Navbar.Toggle>
        <Navbar.Brand href="#home">GoFlow</Navbar.Brand>
        <Navbar.Collapse>
          <Nav>
            <Nav.Link href="#home">Home</Nav.Link>
            <Nav.Link href="#metrics">Metrics</Nav.Link>
            <Nav.Link href="">Settings</Nav.Link>
            <Nav.Link href="">Documentation</Nav.Link>
          </Nav>
        </Navbar.Collapse>
      </Navbar>
  );
}

function App() {
  return (
    <div>
      <Header />
      <DAGContainer />
    </div>
  );
}

export default App;
