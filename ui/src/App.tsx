import React from 'react';
// import logo from './logo.svg';
// import './App.css';
import Navbar from 'react-bootstrap/Navbar'
import Nav from 'react-bootstrap/Nav'

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

function App() {
  return (
    <Header />
  );
}

export default App;
