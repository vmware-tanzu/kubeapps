import * as React from "react";

import { IHeaderProps } from "./Header";

const Header = React.lazy(() => import("./Header"));
const HeaderV2 = React.lazy(() => import("./Header.v2"));

interface IHeaderSelectorProps extends IHeaderProps {
  UI: string;
}

const HeaderSelector: React.FC<IHeaderSelectorProps> = props => (
  <React.Suspense fallback={null}>
    {props.UI === "clarity" ? <HeaderV2 {...props} /> : <Header {...props} />}
  </React.Suspense>
);

export default HeaderSelector;
