import * as React from "react";
import { HashLink as Link } from "react-router-hash-link";

const LinkRenderer: React.SFC<{}> = (props: any) => {
  if (props.href.startsWith("#")) {
    return <Link to={props.href}>{props.children}</Link>;
  }
  return <a href={props.href}>{props.children}</a>;
};

export default LinkRenderer;
