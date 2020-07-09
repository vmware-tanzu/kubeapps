import * as React from "react";

import { IAppListProps } from "./AppList.v2";

const AppList = React.lazy(() => import("./AppList"));
const AppListV2 = React.lazy(() => import("./AppList.v2"));

interface IAppListSelectorProps extends IAppListProps {
  UI: string;
}

const AppListSelector: React.FC<IAppListSelectorProps> = props => (
  <React.Suspense fallback={null}>
    {props.UI === "clarity" ? <AppListV2 {...props} /> : <AppList {...props} />}
  </React.Suspense>
);

export default AppListSelector;
