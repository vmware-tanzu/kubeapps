// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { shallow } from "enzyme";
import Differential from "./Differential";
import DifferentialSelector from "./DifferentialSelector";

it("should use default values when first deploying", () => {
  const wrapper = shallow(
    <DifferentialSelector
      deploymentEvent="install"
      deployedValues=""
      defaultValues="foo"
      appValues="bar"
    />,
  );
  expect(wrapper.find(Differential).props()).toMatchObject({
    oldValues: "foo",
    newValues: "bar",
  });
});

it("should use deployed values when upgrading", () => {
  const wrapper = shallow(
    <DifferentialSelector
      deploymentEvent="upgrade"
      deployedValues="foobar"
      defaultValues="foo"
      appValues="bar"
    />,
  );
  expect(wrapper.find(Differential).props()).toMatchObject({
    oldValues: "foobar",
    newValues: "bar",
  });
});
