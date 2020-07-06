import * as React from "react";

import "./PageHeader.v2.css";

function PageHeader(props: any) {
  return <header className="kubeapps-header">{props.children}</header>;
}

export default PageHeader;
