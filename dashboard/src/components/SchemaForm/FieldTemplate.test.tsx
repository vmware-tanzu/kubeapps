import { shallow } from "enzyme";
import * as React from "react";
import { FieldTemplateProps } from "react-jsonschema-form";

import FieldTemplate from "./FieldTemplate";

it("sets the classNames in the parent div", () => {
  const wrapper = shallow(<FieldTemplate {...{} as FieldTemplateProps} classNames="test-class" />);
  expect(wrapper.props().className).toBe("test-class");
});

it("renders the label element with the right id and label", () => {
  const wrapper = shallow(
    <FieldTemplate {...{} as FieldTemplateProps} id="test" label="Test" displayLabel={true} />,
  );
  const label = wrapper.find("label");
  expect(label.exists()).toBe(true);
  expect(label.props().htmlFor).toBe("test");
  expect(label.text()).toBe("Test");
});

it("omits the label if displayLabel false", () => {
  const wrapper = shallow(
    <FieldTemplate {...{} as FieldTemplateProps} id="test" label="Test" displayLabel={false} />,
  );
  const label = wrapper.find("label");
  expect(label.exists()).toBe(false);
});

it("renders the child element", () => {
  const wrapper = shallow(
    <FieldTemplate {...{} as FieldTemplateProps}>
      <input type="text" />
    </FieldTemplate>,
  );
  expect(wrapper.find("input[type='text']").exists()).toBe(true);
});
