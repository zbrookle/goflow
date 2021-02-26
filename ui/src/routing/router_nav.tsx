import { LinkContainer } from "react-router-bootstrap";
import { Nav } from "react-bootstrap";

type RouterNavLinkProps = {
  link: string;
  text: string;
};

export function RouterNavLink(props: RouterNavLinkProps) {
  return (
    <LinkContainer to={props.link}>
      <Nav.Link>{props.text}</Nav.Link>
    </LinkContainer>
  );
}
