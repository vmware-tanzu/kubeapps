import * as React from "react";

import { IAppUpgradeProps } from "./AppUpgrade";

const AppUpgrade = React.lazy(() => import("./AppUpgrade"));
const AppUpgradeV2 = React.lazy(() => import("./AppUpgrade.v2"));

interface IAppUpgradeSelectorProps extends IAppUpgradeProps {
  UI: string;
}

const AppUpgradeSelector: React.FC<IAppUpgradeSelectorProps> = props => (
  <React.Suspense fallback={null}>
    {props.UI === "clarity" ? <AppUpgradeV2 {...props} /> : <AppUpgrade {...props} />}
  </React.Suspense>
);

export default AppUpgradeSelector;
