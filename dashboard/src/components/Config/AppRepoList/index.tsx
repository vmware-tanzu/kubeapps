import * as React from "react";

import { IAppRepoListProps } from "./AppRepoList";

const AppRepoList = React.lazy(() => import("./AppRepoList"));
const AppRepoListV2 = React.lazy(() => import("./AppRepoList.v2"));

interface IAppRepoListSelectorProps extends IAppRepoListProps {
  UI: string;
}

const AppRepoListSelector: React.FC<IAppRepoListSelectorProps> = props => (
  <React.Suspense fallback={null}>
    {props.UI === "clarity" ? <AppRepoListV2 {...props} /> : <AppRepoList {...props} />}
  </React.Suspense>
);

export default AppRepoListSelector;
