import { shallow } from "enzyme";
import * as React from "react";
import { ArrayFieldTemplateProps } from "react-jsonschema-form";

import ArrayFieldTemplate from "./ArrayFieldTemplate";

it("renders a label for the title", () => {
  const wrapper = shallow(<ArrayFieldTemplate {...{} as ArrayFieldTemplateProps} title="Test" />);
  expect(wrapper.find("label").text()).toBe("Test");
});

it("renders each element in the array", () => {
  const items = [
    {
      children: <div className="test1" />,
      hasMoveDown: true,
      hasMoveUp: true,
      index: 0,
      onDropIndexClick: (i: number) => {
        jest.fn();
      },
      onReorderClick: (i: number, j: number) => {
        jest.fn();
      },
    } as ArrayFieldTemplateProps["items"][0],
    {
      children: <div className="test2" />,
      hasMoveDown: true,
      hasMoveUp: true,
      index: 1,
      onDropIndexClick: (i: number) => {
        jest.fn();
      },
      onReorderClick: (i: number, j: number) => {
        jest.fn();
      },
    },
  ] as ArrayFieldTemplateProps["items"];
  const wrapper = shallow(<ArrayFieldTemplate {...{} as ArrayFieldTemplateProps} items={items} />);
  expect(wrapper.find(".test1").exists()).toBe(true);
  expect(wrapper.find(".test2").exists()).toBe(true);
  expect(wrapper).toMatchSnapshot();
});

it("renders the add item button if enabled", () => {
  const onAddClick = jest.fn();
  const wrapper = shallow(
    <ArrayFieldTemplate {...{} as ArrayFieldTemplateProps} canAdd={true} onAddClick={onAddClick} />,
  );
  const button = wrapper.find("button.button-primary");
  expect(button.exists()).toBe(true);
  button.simulate("click");
  expect(onAddClick).toBeCalled();
});
