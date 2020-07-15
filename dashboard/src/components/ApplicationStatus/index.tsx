import * as React from "react";

import { IApplicationStatusProps } from "./ApplicationStatus";

const ApplicationStatus = React.lazy(() => import("./ApplicationStatus"));
const ApplicationStatusV2 = React.lazy(() => import("./ApplicationStatus.v2"));

interface IApplicationStatusSelectorProps extends IApplicationStatusProps {
  UI: string;
}

const ApplicationStatusSelector: React.FC<IApplicationStatusSelectorProps> = props => (
  <React.Suspense fallback={null}>
    {props.UI === "clarity" ? <ApplicationStatusV2 {...props} /> : <ApplicationStatus {...props} />}
  </React.Suspense>
);

export default ApplicationStatusSelector;
