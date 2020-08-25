import * as React from "react";

import { IOperatorListProps } from "./OperatorList";

const OperatorList = React.lazy(() => import("./OperatorList"));
const OperatorListV2 = React.lazy(() => import("./OperatorList.v2"));

interface IOperatorListSelectorProps extends IOperatorListProps {
  UI: string;
}

const OperatorListSelector: React.FC<IOperatorListSelectorProps> = props => (
  <React.Suspense fallback={null}>
    {props.UI === "clarity" ? <OperatorListV2 {...props} /> : <OperatorList {...props} />}
  </React.Suspense>
);

export default OperatorListSelector;
