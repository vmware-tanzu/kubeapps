import * as React from "react";

import { IChartViewProps } from "./ChartView";

const ChartView = React.lazy(() => import("./ChartView"));
const ChartViewV2 = React.lazy(() => import("./ChartView.v2"));

interface IChartViewSelectorProps extends IChartViewProps {
  UI: string;
}

const ChartViewSelector: React.FC<IChartViewSelectorProps> = props => (
  <React.Suspense fallback={null}>
    {props.UI === "clarity" ? <ChartViewV2 {...props} /> : <ChartView {...props} />}
  </React.Suspense>
);

export default ChartViewSelector;
