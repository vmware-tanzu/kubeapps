import { shallow } from "enzyme";
import * as React from "react";

import ChartInfo from "./ChartInfo";
import { hapi } from "shared/hapi/release";

it("renders a app item", () => {
  const wrapper = shallow(
    <ChartInfo
      app={
        {
          name: "foo",
          chart: {
            metadata: {
              appVersion: "0.0.1",
              version: "1.0.0",
              icon: "icon.png",
              description: "test chart",
            },
          },
        } as hapi.release.Release
      }
    />,
  );
  expect(wrapper.find(".ChartInfo").exists()).toBe(true);
  expect(wrapper).toMatchSnapshot();
});
