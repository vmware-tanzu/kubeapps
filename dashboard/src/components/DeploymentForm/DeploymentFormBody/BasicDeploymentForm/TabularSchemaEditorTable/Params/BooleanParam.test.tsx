// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { mount } from "enzyme";
import { IBasicFormParam } from "shared/types";
import BooleanParam from "./BooleanParam";

//TODO(agamez): enable back this test suite ASAP
/* eslint-disable jest/no-disabled-tests */

const param = { path: "enableMetrics", value: true, type: "boolean" } as unknown as IBasicFormParam;
const defaultProps = {
  id: "foo",
  label: "Enable Metrics",
  param,
  handleBasicFormParamChange: jest.fn(),
};

it.skip("should render a boolean param with title and description", () => {
  const wrapper = mount(<BooleanParam {...defaultProps} />);
  const s = wrapper.find(".react-switch").first();
  expect(s.prop("checked")).toBe(defaultProps.param.value);
});

it.skip("should send a checkbox event to handleBasicFormParamChange", () => {
  const handler = jest.fn();
  const handleBasicFormParamChange = jest.fn().mockReturnValue(handler);
  const wrapper = mount(
    <BooleanParam {...defaultProps} handleBasicFormParamChange={handleBasicFormParamChange} />,
  );
  const s = wrapper.find(".react-switch").first();

  (s.prop("onChange") as any)(false);

  expect(handleBasicFormParamChange.mock.calls[0][0]).toEqual({
    path: "enableMetrics",
    type: "boolean",
    value: true,
  });
  expect(handler.mock.calls[0][0]).toMatchObject({
    currentTarget: { value: "false", type: "checkbox", checked: false },
  });
});
