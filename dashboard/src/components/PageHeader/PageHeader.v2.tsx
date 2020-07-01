import * as React from "react";

import "./PageHeader.v2.css";

interface IPageHeaderProps {
  children: Array<boolean | JSX.Element> | JSX.Element;
}

function PageHeader(props: IPageHeaderProps) {
  return <header className="kubeapps-header clr-row">{props.children}</header>;
}

export default PageHeader;
