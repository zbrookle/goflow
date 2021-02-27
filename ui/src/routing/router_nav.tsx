import { LinkContainer } from "react-router-bootstrap";
import { Nav } from "react-bootstrap";
import { CSSProperties, useState } from "react";

type RouterNavLinkProps = {
  link: string;
  text: string;
  style?: CSSProperties;
  hoverStyle?: CSSProperties;
};

function getStyle(style: CSSProperties | undefined) {
  if (style === undefined) {
    style = {};
  }
  return style;
}

export function RouterNavLink(props: RouterNavLinkProps) {
  let defaultStyle = getStyle(props.style);
  let hoverStyle = getStyle(props.hoverStyle);

  const [style, setStyle] = useState(defaultStyle);
  return (
    <LinkContainer
      onMouseEnter={() => setStyle(hoverStyle)}
      onMouseLeave={() => setStyle(defaultStyle)}
      style={style}
      to={props.link}
    >
      <Nav.Link>{props.text}</Nav.Link>
    </LinkContainer>
  );
}
