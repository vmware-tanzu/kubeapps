import { shallow } from "enzyme";
import * as React from "react";
import * as ReactMarkdown from "react-markdown";
import OperatorDescription from "./OperatorDescription";

it("renders a description", () => {
  const wrapper = shallow(<OperatorDescription description="# Title!\n" />);
  const markdown = wrapper.find(ReactMarkdown);
  expect(markdown).toExist();
  expect(markdown.html()).toContain("Title!");
});
