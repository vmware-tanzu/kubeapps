// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { shallow } from "enzyme";
import Input from "../js/Input";
import SearchFilter, { ISearchFilterProps } from "./SearchFilter";

const defaultProps: ISearchFilterProps = {
  value: "",
  placeholder: "search!",
  onChange: jest.fn(),
  submitFilters: jest.fn(),
};

jest.useFakeTimers();

it("should render a PageHeader", () => {
  const wrapper = shallow(<SearchFilter {...defaultProps} value="test" />);
  expect(wrapper).toMatchSnapshot();
  expect(wrapper.find(Input).prop("value")).toBe("test");
});

it("changes the filter", () => {
  const onChange = jest.fn();
  const wrapper = shallow(<SearchFilter {...defaultProps} value="test" onChange={onChange} />);
  wrapper.find(Input).simulate("change", { currentTarget: { value: "foo" } });
  jest.runAllTimers();
  expect(onChange).toHaveBeenCalledWith("foo");
});

it("should render a PageHeader (onSubmit)", () => {
  const onSubmit = jest.fn();
  const wrapper = shallow(<SearchFilter {...defaultProps} value="test" submitFilters={onSubmit} />);
  wrapper.find("form").simulate("submit", { preventDefault: jest.fn() });
  expect(onSubmit).toHaveBeenCalled();
});
