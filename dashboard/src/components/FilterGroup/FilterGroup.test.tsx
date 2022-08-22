// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import MultiCheckbox from "components/MultiCheckbox";
import { act } from "react-dom/test-utils";
import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import FilterGroup from "./FilterGroup";

const defaultProps = {
  name: "test",
  options: ["foo", "bar"],
  currentFilters: [],
  disabled: false,
  onAddFilter: jest.fn(),
  onRemoveFilter: jest.fn(),
};

it("renders a multicheckbox", () => {
  const wrapper = mountWrapper(defaultStore, <FilterGroup {...defaultProps} />);
  expect(wrapper).toMatchSnapshot();
});

it("calls onChange function", () => {
  const currentFilters: string[] = [];
  const onAddFilter = jest.fn((_n, f) => currentFilters.push(f));
  const onRemoveFilter = jest.fn();
  const wrapper = mountWrapper(
    defaultStore,
    <FilterGroup
      {...defaultProps}
      currentFilters={currentFilters}
      onAddFilter={onAddFilter}
      onRemoveFilter={onRemoveFilter}
    />,
  );
  act(() => {
    (wrapper.find(MultiCheckbox).prop("onChange") as any)({ target: { value: "foo" } });
  });
  expect(onAddFilter).toHaveBeenCalledWith("test", "foo");
  // Force re-render
  wrapper.setProps({ ...defaultProps, onAddFilter, onRemoveFilter });
  // Adds a new item to the filter
  act(() => {
    (wrapper.find(MultiCheckbox).prop("onChange") as any)({ target: { value: "bar" } });
  });
  expect(onAddFilter).toHaveBeenCalledWith("test", "bar");
  // Force re-render
  wrapper.setProps({ ...defaultProps, onAddFilter, onRemoveFilter });
  // Removes an item
  act(() => {
    (wrapper.find(MultiCheckbox).prop("onChange") as any)({ target: { value: "foo" } });
  });
  expect(onRemoveFilter).toHaveBeenCalledWith("test", "foo");
});
