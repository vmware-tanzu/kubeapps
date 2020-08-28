import { shallow } from "enzyme";
import * as React from "react";
import CapabiliyLevel, { AUTO_PILOT, BASIC_INSTALL } from "./OperatorCapabilityLevel.v2";

[
  {
    name: "basic install level",
    expectedLevels: 1,
    level: BASIC_INSTALL,
  },
  {
    name: "auto pilot level",
    expectedLevels: 5,
    level: AUTO_PILOT,
  },
  {
    name: "unknown level",
    expectedLevels: 0,
    level: "Foo",
  },
].forEach(t => {
  it(t.name, () => {
    const wrapper = shallow(<CapabiliyLevel level={t.level} />);
    expect(wrapper.find(".color-icon-info").length).toEqual(t.expectedLevels);
  });
});
