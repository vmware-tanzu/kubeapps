// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import actions from "actions";
import LoadingWrapper from "components/LoadingWrapper/LoadingWrapper";
import ReactMarkdown from "react-markdown";
import { HashLink as Link } from "react-router-hash-link";
import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import PackageReadme from "./PackageReadme";

const defaultProps = {
  error: undefined,
  readme: "",
};

const kubeaActions = { ...actions.kube };
beforeEach(() => {
  actions.availablepackages = {
    ...actions.availablepackages,
  };
});

afterEach(() => {
  actions.kube = { ...kubeaActions };
});

it("behaves as a loading component", () => {
  const wrapper = mountWrapper(defaultStore, <PackageReadme {...defaultProps} />);
  expect(wrapper.find(LoadingWrapper)).toExist();
});

it("renders the ReactMarkdown content is readme is present", () => {
  const props = {
    ...defaultProps,
    readme: "# Markdown Readme",
  };
  const wrapper = mountWrapper(defaultStore, <PackageReadme {...props} />);
  const component = wrapper.find(ReactMarkdown);
  expect(component.html()).toEqual('<h1 id="markdown-readme">Markdown Readme</h1>');
});

it("renders the ReactMarkdown content with github flavored markdown (table)", () => {
  const props = {
    ...defaultProps,
    readme: "|h1|h2|\n|-|-|\n|foo|bar|",
  };
  const wrapper = mountWrapper(defaultStore, <PackageReadme {...props} />);
  const component = wrapper.find(ReactMarkdown);
  expect(component.props()).toMatchObject({ children: props.readme });
  expect(component.find("table th").first().text()).toBe("h1");
  expect(component.find("table th").last().text()).toBe("h2");
  expect(component.find("table td").first().text()).toBe("foo");
  expect(component.find("table td").last().text()).toBe("bar");
});

it("renders a not found error when error is set", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <PackageReadme {...defaultProps} error={"not found"} />,
  );
  expect(wrapper.text()).toContain("No README found");
});

it("renders an alert when error is set", () => {
  const wrapper = mountWrapper(defaultStore, <PackageReadme {...defaultProps} error={"Boom!"} />);
  expect(wrapper.text()).toContain("Unable to fetch package README: Boom!");
});

it("renders the ReactMarkdown content adding IDs for the titles", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <PackageReadme {...defaultProps} readme="# _Markdown_ 'Readme_or_not'!" />,
  );
  const component = wrapper.find("#markdown-readme_or_not");
  expect(component).toExist();
});

it("renders the ReactMarkdown ignoring comments", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <PackageReadme
      {...defaultProps}
      readme={`<!-- This is a comment -->
      This is text`}
    />,
  );
  const html = wrapper.html();
  expect(html).toContain("This is text");
  expect(html).not.toContain("This is a comment");
});

it("renders the ReactMarkdown content with hash links", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <PackageReadme
      {...defaultProps}
      readme={`[section 1](#section-1)
      # Section 1`}
    />,
  );
  expect(wrapper.find(Link)).toExist();
});
