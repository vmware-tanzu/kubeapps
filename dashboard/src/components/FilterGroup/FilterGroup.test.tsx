import * as React from "react";

import MultiCheckbox from "components/js/MultiCheckbox";
import { act } from "react-dom/test-utils";
import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import FilterGroup from "./FilterGroup";

const defaultProps = {
  name: "test",
  options: ["foo", "bar"],
  onChange: jest.fn(),
};

it("renders a multicheckbox", () => {
  const wrapper = mountWrapper(defaultStore, <FilterGroup {...defaultProps} />);
  expect(wrapper).toMatchSnapshot();
});

it("calls onChange function", () => {
  const onChange = jest.fn();
  const wrapper = mountWrapper(defaultStore, <FilterGroup {...defaultProps} onChange={onChange} />);
  act(() => {
    wrapper.find(MultiCheckbox).prop("onChange")({ target: { value: "foo" } });
  });
  expect(onChange).toHaveBeenCalledWith(["foo"]);
  // Force re-render
  wrapper.setProps({ ...defaultProps, onChange });
  // Adds a new item to the filter
  act(() => {
    wrapper.find(MultiCheckbox).prop("onChange")({ target: { value: "bar" } });
  });
  expect(onChange).toHaveBeenCalledWith(["foo", "bar"]);
  // Force re-render
  wrapper.setProps({ ...defaultProps, onChange });
  // Removes an item
  act(() => {
    wrapper.find(MultiCheckbox).prop("onChange")({ target: { value: "foo" } });
  });
  expect(onChange).toHaveBeenCalledWith(["bar"]);
});
