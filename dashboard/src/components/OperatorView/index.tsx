import * as React from "react";

import { IOperatorViewProps } from "./OperatorView";

const OperatorView = React.lazy(() => import("./OperatorView"));
const OperatorViewV2 = React.lazy(() => import("./OperatorView.v2"));

interface IOperatorViewSelectorProps extends IOperatorViewProps {
  UI: string;
}

const OperatorViewSelector: React.FC<IOperatorViewSelectorProps> = props => (
  <React.Suspense fallback={null}>
    {props.UI === "clarity" ? <OperatorViewV2 {...props} /> : <OperatorView {...props} />}
  </React.Suspense>
);

export default OperatorViewSelector;
