import { shallow } from "enzyme";
import * as React from "react";
import CapabiliyLevel, { AUTO_PILOT, BASIC_INSTALL } from "./OperatorCapabilityLevel";

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
    const satisfiedLevels = wrapper
      .find(".capabilityLevelIcon")
      .filterWhere(s => s.prop("stroke") === "#1598CB");
    expect(t.expectedLevels).toEqual(satisfiedLevels.length);
  });
});
