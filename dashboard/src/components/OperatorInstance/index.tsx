import * as React from "react";

import { IOperatorInstanceProps } from "./OperatorInstance";

const OperatorInstance = React.lazy(() => import("./OperatorInstance"));
const OperatorInstanceV2 = React.lazy(() => import("./OperatorInstance.v2"));

interface IOperatorInstanceSelectorProps extends IOperatorInstanceProps {
  UI: string;
}

const OperatorInstanceSelector: React.FC<IOperatorInstanceSelectorProps> = props => (
  <React.Suspense fallback={null}>
    {props.UI === "clarity" ? <OperatorInstanceV2 {...props} /> : <OperatorInstance {...props} />}
  </React.Suspense>
);

export default OperatorInstanceSelector;
