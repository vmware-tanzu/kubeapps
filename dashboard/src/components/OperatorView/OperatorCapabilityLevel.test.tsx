// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { shallow } from "enzyme";
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
    expect(wrapper.find(".color-icon-info").length).toEqual(t.expectedLevels);
  });
});
