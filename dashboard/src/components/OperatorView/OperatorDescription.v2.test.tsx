import { shallow } from "enzyme";
import * as React from "react";
import ReactMarkdown from "react-markdown";
import OperatorDescription from "./OperatorDescription.v2";

it("renders a description", () => {
  const wrapper = shallow(<OperatorDescription description="# Title!" />);
  const markdown = wrapper.find(ReactMarkdown);
  expect(markdown).toExist();
  expect(markdown.html()).toContain('<h1 id="title">Title!</h1>');
});
