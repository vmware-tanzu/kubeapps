// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import React from "react";

// Code from https://github.com/rexxars/react-markdown/issues/69
function flatten(text: string, child: any): any {
  return typeof child === "string"
    ? text + child
    : React.Children.toArray(child.props.children).reduce(flatten, text);
}

const HeadingRenderer: React.FunctionComponent<{}> = (props: any) => {
  const children = React.Children.toArray(props.children);
  const text = children.reduce(flatten, "");
  const slug = text
    .toLowerCase()
    .replace(/[^a-zA-Z0-9_\s]/g, "") // remove punctuation
    .replace(/\s/g, "-"); // replace spaces with dash
  return React.createElement("h" + props.level, { id: slug }, props.children);
};

export default HeadingRenderer;
