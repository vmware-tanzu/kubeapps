import * as React from "react";

import ChartView from ".";

it("is a React Component", () => {
  expect(ChartView).toBeDefined();
  expect(ChartView.prototype instanceof React.Component).toBe(true);
});
