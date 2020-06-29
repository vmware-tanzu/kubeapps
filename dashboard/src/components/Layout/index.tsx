import * as React from "react";
import UISelector from "./UISelector";

const Layout = React.lazy(() => import("./Layout"));
const LayoutV2 = React.lazy(() => import("./Layout.v2"));

interface ILayoutSelectorProps {
  children: JSX.Element;
  UI: string;
  headerComponent: React.ComponentClass<any> | React.StatelessComponent<any>;
}

const LayoutSelector: React.FC<ILayoutSelectorProps> = props => (
  <React.Suspense fallback={null}>
    <UISelector UI={props.UI} />
    {props.UI === "clarity" ? <LayoutV2 {...props} /> : <Layout {...props} />}
  </React.Suspense>
);

export default LayoutSelector;
