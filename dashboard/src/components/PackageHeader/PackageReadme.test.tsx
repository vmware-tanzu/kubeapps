// Copyright 2021-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import actions from "actions";
import LoadingWrapper from "components/LoadingWrapper/LoadingWrapper";
import ReactMarkdown from "react-markdown";
import { HashLink as Link } from "react-router-hash-link";
import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import PackageReadme, { IPackageReadmeProps } from "./PackageReadme";

const defaultProps: IPackageReadmeProps = {
  error: undefined,
  readme: "",
  isFetching: false,
};

const kubeActions = { ...actions.kube };
beforeEach(() => {
  actions.availablepackages = {
    ...actions.availablepackages,
  };
});

afterEach(() => {
  actions.kube = { ...kubeActions };
});

it("behaves as a loading component if it's fetching with readme", () => {
  const props: IPackageReadmeProps = {
    ...defaultProps,
    isFetching: true,
    readme: "foo",
  };
  const wrapper = mountWrapper(defaultStore, <PackageReadme {...props} />);

  expect(wrapper.find(LoadingWrapper)).toExist();
});

it("behaves as a loading component if it's fetching without readme", () => {
  const props: IPackageReadmeProps = {
    ...defaultProps,
    isFetching: true,
    readme: "",
  };
  const wrapper = mountWrapper(defaultStore, <PackageReadme {...props} />);

  expect(wrapper.find(LoadingWrapper)).toExist();
});

it("renders the ReactMarkdown content is readme is present", () => {
  const props: IPackageReadmeProps = {
    ...defaultProps,
    readme: "# Markdown Readme",
  };
  const wrapper = mountWrapper(defaultStore, <PackageReadme {...props} />);

  const component = wrapper.find(ReactMarkdown);
  expect(component.html()).toEqual('<h1 id="markdown-readme">Markdown Readme</h1>');
});

it("renders the ReactMarkdown content with github flavored markdown (table)", () => {
  const props: IPackageReadmeProps = {
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

it("renders a message if no readme is fetched", () => {
  const props: IPackageReadmeProps = {
    ...defaultProps,
    readme: "",
  };
  const wrapper = mountWrapper(defaultStore, <PackageReadme {...props} />);

  expect(wrapper.text()).toContain("This package does not contain a README file.");
});

it("renders an alert when error is set", () => {
  const props: IPackageReadmeProps = {
    ...defaultProps,
    error: "Boom!",
  };
  const wrapper = mountWrapper(defaultStore, <PackageReadme {...props} />);

  expect(wrapper.text()).toContain("Unable to fetch the package's README: Boom!");
});

it("renders the ReactMarkdown content adding IDs for the titles", () => {
  const props: IPackageReadmeProps = {
    ...defaultProps,
    readme: "# _Markdown_ 'Readme_or_not'!",
  };
  const wrapper = mountWrapper(defaultStore, <PackageReadme {...props} />);

  const component = wrapper.find("#markdown-readme_or_not");
  expect(component).toExist();
});

it("renders the ReactMarkdown ignoring comments", () => {
  const props: IPackageReadmeProps = {
    ...defaultProps,
    readme: `<!-- This is a comment -->
    This is text`,
  };
  const wrapper = mountWrapper(defaultStore, <PackageReadme {...props} />);

  const html = wrapper.html();
  expect(html).toContain("This is text");
  expect(html).not.toContain("This is a comment");
});

it("renders the ReactMarkdown content with hash links", () => {
  const props: IPackageReadmeProps = {
    ...defaultProps,
    readme: `[section 1](#section-1)
    # Section 1`,
  };
  const wrapper = mountWrapper(defaultStore, <PackageReadme {...props} />);

  expect(wrapper.find(Link)).toExist();
});
