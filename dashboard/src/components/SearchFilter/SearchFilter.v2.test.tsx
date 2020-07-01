import { shallow } from "enzyme";
import * as React from "react";
import SearchFilter, { ISearchFilterProps } from "./SearchFilter.v2";

const defaultProps: ISearchFilterProps = {
  value: "",
  placeholder: "search!",
  onChange: jest.fn(),
  onSubmit: jest.fn(),
};

it("should render a PageHeader", () => {
  const wrapper = shallow(<SearchFilter {...defaultProps} value="test" />);
  expect(wrapper).toMatchSnapshot();
  expect(wrapper.find("input").prop("value")).toBe("test");
});

it("changes the filter", () => {
  const onChange = jest.fn();
  const wrapper = shallow(<SearchFilter {...defaultProps} value="test" onChange={onChange} />);
  wrapper.find("input").simulate("change", { currentTarget: { value: "foo" } });
  expect(onChange).toHaveBeenCalledWith("foo");
});

it("should render a PageHeader", () => {
  const onSubmit = jest.fn();
  const wrapper = shallow(<SearchFilter {...defaultProps} value="test" onSubmit={onSubmit} />);
  wrapper.find("form").simulate("submit", { preventDefault: jest.fn() });
  expect(onSubmit).toHaveBeenCalledWith("test");
});
