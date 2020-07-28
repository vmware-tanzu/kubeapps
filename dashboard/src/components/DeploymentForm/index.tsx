import * as React from "react";

import { IDeploymentFormProps } from "./DeploymentForm";

const DeploymentForm = React.lazy(() => import("./DeploymentForm"));
const DeploymentFormV2 = React.lazy(() => import("./DeploymentForm.v2"));

interface IDeploymentFormSelectorProps extends IDeploymentFormProps {
  UI: string;
}

const DeploymentFormSelector: React.FC<IDeploymentFormSelectorProps> = props => (
  <React.Suspense fallback={null}>
    {props.UI === "clarity" ? <DeploymentFormV2 {...props} /> : <DeploymentForm {...props} />}
  </React.Suspense>
);

export default DeploymentFormSelector;
