import * as React from "react";

import { IOperatorNewProps } from "./OperatorNew";

const OperatorNew = React.lazy(() => import("./OperatorNew"));
const OperatorNewV2 = React.lazy(() => import("./OperatorNew.v2"));

interface IOperatorNewSelectorProps extends IOperatorNewProps {
  UI: string;
}

const OperatorNewSelector: React.FC<IOperatorNewSelectorProps> = props => (
  <React.Suspense fallback={null}>
    {props.UI === "clarity" ? <OperatorNewV2 {...props} /> : <OperatorNew {...props} />}
  </React.Suspense>
);

export default OperatorNewSelector;
