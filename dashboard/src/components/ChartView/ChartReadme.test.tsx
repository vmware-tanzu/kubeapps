import { mount, shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";
import * as ReactMarkdown from "react-markdown";
import { BrowserRouter } from "react-router-dom";
import { HashLink as Link } from "react-router-hash-link";

import itBehavesLike from "../../shared/specs";

import ChartReadme from "./ChartReadme";

const chartNamespace = "chart-namespace";
const version = "1.2.3";

context("when readme is not present", () => {
  itBehavesLike("aLoadingComponent", {
    component: ChartReadme,
    props: {
      getChartReadme: jest.fn(),
      hasError: false,
      version,
      chartNamespace,
    },
  });
});

describe("getChartReadme", () => {
  const spy = jest.fn();
  const props = {
    chartNamespace,
    hasError: false,
    version,
  };
  const wrapper = shallow(<ChartReadme getChartReadme={spy} {...props} />);

  it("gets triggered when mounting", () => {
    expect(spy).toHaveBeenCalledWith(chartNamespace, version);
  });

  it("gets triggered after changing version", () => {
    wrapper.setProps({ version: "1.2.4" });
    expect(spy).toHaveBeenCalledWith(chartNamespace, "1.2.4");
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

it("renders the ReactMarkdown ignoring comments", () => {
  const wrapper = mount(
    <ChartReadme
      getChartReadme={jest.fn()}
      hasError={false}
      version="1.2.3"
      readme={`<!-- This is a comment -->
      This is text`}
    />,
  );
  const html = wrapper.html();
  expect(html).toContain("This is text");
  expect(html).not.toContain("This is a comment");
});

it("renders the ReactMarkdown content with hash links", () => {
  const wrapper = mount(
    <BrowserRouter>
      <ChartReadme
        getChartReadme={jest.fn()}
        hasError={false}
        version="1.2.3"
        readme={`[section 1](#section-1)
      # Section 1`}
      />
    </BrowserRouter>,
  );
  expect(wrapper.find(Link)).toExist();
});
