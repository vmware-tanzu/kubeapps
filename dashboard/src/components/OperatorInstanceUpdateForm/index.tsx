import * as React from "react";

import { IOperatorInstanceUpgradeFormProps } from "./OperatorInstanceUpdateForm";

const OperatorInstanceUpdateForm = React.lazy(() => import("./OperatorInstanceUpdateForm"));
const OperatorInstanceUpdateFormV2 = React.lazy(() => import("./OperatorInstanceUpdateForm.v2"));

interface IOperatorInstanceUpdateFormSelectorProps extends IOperatorInstanceUpgradeFormProps {
  UI: string;
}

const OperatorInstanceUpdateFormSelector: React.FC<IOperatorInstanceUpdateFormSelectorProps> = props => (
  <React.Suspense fallback={null}>
    {props.UI === "clarity" ? (
      <OperatorInstanceUpdateFormV2 {...props} />
    ) : (
      <OperatorInstanceUpdateForm {...props} />
    )}
  </React.Suspense>
);

export default OperatorInstanceUpdateFormSelector;
