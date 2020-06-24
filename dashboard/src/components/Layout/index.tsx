import * as React from "react";

const Layout = React.lazy(() => import("./Layout"));
const LayoutV2 = React.lazy(() => import("./Layout.v2"));

interface ILayoutSelectorProps {
  UI: string;
  headerComponent: React.ComponentClass<any> | React.StatelessComponent<any>;
}

const LayoutSelector: React.FC<ILayoutSelectorProps> = props => (
  <React.Suspense fallback={null}>
    {props.UI === "hex" && <Layout {...props} />}
    {props.UI === "clarity" && <LayoutV2 {...props} />}
  </React.Suspense>
);

export default LayoutSelector;
