// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { mount } from "enzyme";
import { act } from "react-dom/test-utils";
import BooleanParam, { IBooleanParamProps } from "./BooleanParam";

const defaultProps = {
  handleBasicFormParamChange: jest.fn(),
  id: "foo",
  label: "Enable Metrics",
  param: {
    title: "Enable Metrics",
    type: "boolean",
    currentValue: false,
    defaultValue: false,
    deployedValue: false,
    hasProperties: false,
    key: "enableMetrics",
    schema: {
      type: "boolean",
    },
  },
} as IBooleanParamProps;

it("should render a boolean param with title and description", () => {
  const wrapper = mount(<BooleanParam {...defaultProps} />);
  const s = wrapper.find("input").findWhere(i => i.prop("type") === "checkbox");
  expect(s.prop("checked")).toBe(defaultProps.param.currentValue);
});

it("should send a checkbox event to handleBasicFormParamChange", () => {
  const handler = jest.fn();
  const handleBasicFormParamChange = jest.fn().mockReturnValue(handler);
  const wrapper = mount(
    <BooleanParam {...defaultProps} handleBasicFormParamChange={handleBasicFormParamChange} />,
  );
  const s = wrapper.find("input").findWhere(i => i.prop("type") === "checkbox");
  const event = { currentTarget: { checked: true } } as React.FormEvent<HTMLInputElement>;
  act(() => {
    (s.prop("onChange") as any)(event);
  });
  s.update();
  expect(handleBasicFormParamChange).toHaveBeenCalledWith(defaultProps.param);
  expect(handler).toHaveBeenCalledWith({ currentTarget: { type: "checkbox", value: "true" } });
});
