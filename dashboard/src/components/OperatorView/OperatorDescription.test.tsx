// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { shallow } from "enzyme";
import ReactMarkdown from "react-markdown";
import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import OperatorDescription from "./OperatorDescription";

it("renders a description", () => {
  const wrapper = shallow(<OperatorDescription description="# Title!" />);
  const markdown = wrapper.find(ReactMarkdown);
  expect(markdown).toExist();
  expect(markdown.html()).toContain('<h1 id="title">Title!</h1>');
});

it("renders the ReactMarkdown content with github flavored markdown (table)", () => {
  const props = {
    description: "|h1|h2|\n|-|-|\n|foo|bar|",
  };
  const wrapper = mountWrapper(defaultStore, <OperatorDescription {...props} />);
  const component = wrapper.find(ReactMarkdown);
  expect(component.find("table th").first().text()).toBe("h1");
  expect(component.find("table th").last().text()).toBe("h2");
  expect(component.find("table td").first().text()).toBe("foo");
  expect(component.find("table td").last().text()).toBe("bar");
});
