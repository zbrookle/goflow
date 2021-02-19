import React from 'react';
// import logo from './logo.svg';
// import './App.css';
import Navbar from 'react-bootstrap/Navbar'
import Nav from 'react-bootstrap/Nav'
import Container from 'react-bootstrap/Container'
import Row from 'react-bootstrap/Row'
import Col from 'react-bootstrap/Col'

function Header() {
  return <div>
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
    </div>
}

function DAGContainer() {
  return <Container>
    <Row>
      <Col>on/off</Col>
      <Col>name</Col>
      <Col>icon?</Col>
      <Col>last runtime</Col>
      <Col>Success/failures</Col>
    </Row>
  </Container>
}

function App() {
  return (
    <div><Header />
    <DAGContainer /></div>
  );
}

export default App;
