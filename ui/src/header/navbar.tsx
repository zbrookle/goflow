// import Navbar from "react-bootstrap/Navbar";

import { MutableRefObject, useRef } from "react";

type LinkProps = {
  link: string
  display: string
}

function HeaderLink(props: LinkProps) {
  return (
    <li>
      <a href={props.link}>{props.display}</a>
    </li>
  );
}

export function GoflowNavbar() {
  return (
    <ul>
      <HeaderLink link="#home" display="Home" />
      <li>
        <a href="news.asp">Metrics</a>
      </li>
      <li>
        <a href="contact.asp">Documentation</a>
      </li>
      <li>
        <a href="about.asp">Settings</a>
      </li>
      <li>
        <a href="about.asp">Add New Dags</a>
      </li>
    </ul>
  );
}

export {}