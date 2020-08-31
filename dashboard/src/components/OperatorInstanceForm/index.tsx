import * as React from "react";

import { IOperatorInstanceFormProps } from "./OperatorInstanceForm";

const OperatorInstanceForm = React.lazy(() => import("./OperatorInstanceForm"));
const OperatorInstanceFormV2 = React.lazy(() => import("./OperatorInstanceForm.v2"));

interface IOperatorInstanceFormSelectorProps extends IOperatorInstanceFormProps {
  UI: string;
}

const OperatorInstanceFormSelector: React.FC<IOperatorInstanceFormSelectorProps> = props => (
  <React.Suspense fallback={null}>
    {props.UI === "clarity" ? (
      <OperatorInstanceFormV2 {...props} />
    ) : (
      <OperatorInstanceForm {...props} />
    )}
  </React.Suspense>
);

export default OperatorInstanceFormSelector;
