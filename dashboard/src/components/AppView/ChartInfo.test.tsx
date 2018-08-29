import { shallow } from "enzyme";
import * as React from "react";

import { hapi } from "shared/hapi/release";
import ChartInfo from "./ChartInfo";

it("renders a app item", () => {
  const wrapper = shallow(
    <ChartInfo
      app={
        {
          chart: {
            metadata: {
              appVersion: "0.0.1",
              description: "test chart",
              icon: "icon.png",
              version: "1.0.0",
            },
          },
          name: "foo",
        } as hapi.release.Release
      }
    />,
  );
  expect(wrapper.find(".ChartInfo").exists()).toBe(true);
  expect(wrapper).toMatchSnapshot();
});
