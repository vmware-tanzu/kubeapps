import * as React from "react";

import { IErrorBoundaryProps } from "./ErrorBoundary";

const ErrorBoundary = React.lazy(() => import("./ErrorBoundary"));
const ErrorBoundaryV2 = React.lazy(() => import("./ErrorBoundary.v2"));

interface IErrorBoundarySelectorProps extends IErrorBoundaryProps {
  UI: string;
}

const ErrorBoundarySelector: React.FC<IErrorBoundarySelectorProps> = props => (
  <React.Suspense fallback={null}>
    {props.UI === "clarity" ? <ErrorBoundaryV2 {...props} /> : <ErrorBoundary {...props} />}
  </React.Suspense>
);

export default ErrorBoundarySelector;
