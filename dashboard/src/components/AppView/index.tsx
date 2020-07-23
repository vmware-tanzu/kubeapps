import * as React from "react";

import { IAppViewProps } from "./AppView";

const AppView = React.lazy(() => import("./AppView"));
const AppViewV2 = React.lazy(() => import("./AppView.v2"));

interface IAppViewSelectorProps extends IAppViewProps {
  UI: string;
}

const AppViewSelector: React.FC<IAppViewSelectorProps> = props => (
  <React.Suspense fallback={null}>
    {props.UI === "clarity" ? <AppViewV2 {...props} /> : <AppView {...props} />}
  </React.Suspense>
);

export default AppViewSelector;
