import * as React from "react";

import { IAccessURLTableProps } from "./AccessURLTable";

const AccessURLTable = React.lazy(() => import("./AccessURLTable"));
const AccessURLTableV2 = React.lazy(() => import("./AccessURLTable.v2"));

interface IAccessURLTableSelectorProps extends IAccessURLTableProps {
  UI: string;
}

const AccessURLTableSelector: React.FC<IAccessURLTableSelectorProps> = props => (
  <React.Suspense fallback={null}>
    {props.UI === "clarity" ? <AccessURLTableV2 {...props} /> : <AccessURLTable {...props} />}
  </React.Suspense>
);

export default AccessURLTableSelector;
