import * as React from "react";

const HEx = React.lazy(() => import("./HEx"));
const Clarity = React.lazy(() => import("./Clarity"));

export interface IUISelectorProps {
  UI: string;
}

const UISelector: React.FC<IUISelectorProps> = props => (
  <React.Suspense fallback={null}>
    {props.UI === "hex" && <HEx />}
    {props.UI === "clarity" && <Clarity />}
  </React.Suspense>
);

export default UISelector;
