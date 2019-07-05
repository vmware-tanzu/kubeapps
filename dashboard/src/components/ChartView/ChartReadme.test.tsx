import { mount, shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";
import * as ReactMarkdown from "react-markdown";

import itBehavesLike from "../../shared/specs";

import ChartReadme from "./ChartReadme";

context("when readme is not present", () => {
  itBehavesLike("aLoadingComponent", {
    component: ChartReadme,
    props: {
      getChartReadme: jest.fn(),
      hasError: false,
      version: "1.2.3",
    },
  });
});

describe("getChartReadme", () => {
  const spy = jest.fn();
  const wrapper = shallow(<ChartReadme getChartReadme={spy} hasError={false} version="1.2.3" />);

  it("gets triggered when mounting", () => {
    expect(spy).toHaveBeenCalledWith("1.2.3");
  });

  it("gets triggered after changing version", () => {
    wrapper.setProps({ version: "1.2.4" });
    expect(spy).toHaveBeenCalledWith("1.2.4");
  });

  it("does not get triggered when version doesn't change", () => {
    wrapper.setProps({ version: "1.2.4" });
    wrapper.setProps({ hasError: true });
    expect(spy).toHaveBeenCalledTimes(2);
  });
});

it("renders the ReactMarkdown content is readme is present", () => {
  const wrapper = shallow(
    <ChartReadme
      getChartReadme={jest.fn()}
      hasError={false}
      version="1.2.3"
      readme="# Markdown Readme"
    />,
  );
  const component = wrapper.find(ReactMarkdown);
  expect(component.props()).toMatchObject({ source: "# Markdown Readme" });
});

it("renders an error when hasError is set", () => {
  const wrapper = shallow(
    <ChartReadme getChartReadme={jest.fn()} hasError={true} version="1.2.3" />,
  );
  expect(wrapper.text()).toContain("No README found");
});

it("renders the ReactMarkdown content adding IDs for the titles", () => {
  const wrapper = mount(
    <ChartReadme
      getChartReadme={jest.fn()}
      hasError={false}
      version="1.2.3"
      readme="# _Markdown_ 'Readme_or_not'!"
    />,
  );
  const component = wrapper.find("#markdown-readme_or_not");
  expect(component).toExist();
});
