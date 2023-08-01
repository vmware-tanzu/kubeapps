// Copyright 2020-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { HTMLAttributes } from "react";

export interface IColumnProps {
  children: React.ReactNode;
  isListItem?: boolean;
  offset?: number | number[];
  span?: number | number[];
}

// Constant with the different breakpoints for collapsing rows
const breakpoints = ["sm", "md", "lg", "xl"];

// Generate the CSS classes for responsive design based on the given property
const responsiveCss = (prefix: string, values: number[]) =>
  values
    .reduce((css, value, i) => {
      if (breakpoints[i] == null) return css;
      css = `${css} ${prefix}${breakpoints[i]}-${value}`;
      return css;
    }, "")
    .trim();

// Generate CSS clarity classes
const generateClarityCss = (prefix: string, property: number | number[]): string => {
  let cssClass = "";

  if (property == null) return cssClass;

  if (typeof property === "number") {
    cssClass = `clr-${prefix}-${property}`;
  } else {
    // Responsive array
    cssClass = responsiveCss(`clr-${prefix}-`, property);
  }

  return cssClass;
};

export function Column({ children, isListItem, offset, span }: IColumnProps) {
  const innerProps: HTMLAttributes<HTMLDivElement> = {
    className: "clr-col",
    role: isListItem ? "listitem" : "",
  };

  // If span is defined, add the class
  if (span) {
    // Prepare responsive columns
    innerProps.className = `${innerProps.className} ${generateClarityCss("col", span)}`;
  }

  // If offset is defined, add the class
  if (offset) {
    // Prepare offset
    innerProps.className = `${innerProps.className} ${generateClarityCss("offset", offset)}`;
  }

  // Trim the className before passing it
  innerProps.className = innerProps.className?.trim();

  return <div {...innerProps}>{children}</div>;
}
