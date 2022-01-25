// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import ReactMarkdown from "react-markdown";
import { Link } from "react-router-dom";
import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import AppNotes, { IAppNotesProps } from "./AppNotes";

const defaultProps = {
  notes: "",
  title: "",
} as IAppNotesProps;

it("renders the ReactMarkdown content in readme if present", () => {
  const props = {
    ...defaultProps,
    notes: "# Markdown Readme",
  };
  const wrapper = mountWrapper(defaultStore, <AppNotes {...props} />);
  const component = wrapper.find(ReactMarkdown);
  expect(component.html()).toEqual('<h1 id="markdown-readme">Markdown Readme</h1>');
});

it("renders the ReactMarkdown content with github flavored markdown (table)", () => {
  const props = {
    notes: "|h1|h2|\n|-|-|\n|foo|bar|",
  };
  const wrapper = mountWrapper(defaultStore, <AppNotes {...props} />);
  const component = wrapper.find(ReactMarkdown);
  expect(component.props()).toMatchObject({ children: props.notes });
  expect(component.find("table th").first().text()).toBe("h1");
  expect(component.find("table th").last().text()).toBe("h2");
  expect(component.find("table td").first().text()).toBe("foo");
  expect(component.find("table td").last().text()).toBe("bar");
});

it("renders the ReactMarkdown content adding IDs for the titles", () => {
  const props = {
    ...defaultProps,
    notes: "# _Markdown_ 'Readme_or_not'!",
  };
  const wrapper = mountWrapper(defaultStore, <AppNotes {...props} />);
  const component = wrapper.find("#markdown-readme_or_not");
  expect(component).toExist();
});

it("renders the ReactMarkdown ignoring comments", () => {
  const props = {
    ...defaultProps,
    notes: `<!-- This is a comment -->
    This is text`,
  };
  const wrapper = mountWrapper(defaultStore, <AppNotes {...props} />);
  const html = wrapper.html();
  expect(html).toContain("This is text");
  expect(html).not.toContain("This is a comment");
});

it("renders the ReactMarkdown content with hash links", () => {
  const props = {
    ...defaultProps,
    notes: `[section 1](#section-1)
    # Section 1`,
  };
  const wrapper = mountWrapper(defaultStore, <AppNotes {...props} />);
  expect(wrapper.find(Link)).toExist();
});
