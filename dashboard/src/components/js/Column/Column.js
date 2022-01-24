// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import PropTypes from "prop-types";
import React from "react";

// Constant with the different breakpoints for collapsing rows
const breakpoints = ["sm", "md", "lg", "xl"];

// Generate the CSS classes for responsive design based on the given property
const responsiveCss = (prefix, values) =>
  values
    .reduce((css, value, i) => {
      if (breakpoints[i] == null) return css;
      css = `${css} ${prefix}${breakpoints[i]}-${value}`;
      return css;
    }, "")
    .trim();

// Generate CSS clarity classes
const generateClarityCss = (prefix, property) => {
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

/**
 * https://v2.clarity.design/grid
 */
const Column = ({ children, listItem, offset, span }) => {
  const innerProps = {
    className: "",
  };

  if (span == null && offset == null) {
    // Ignore any other config
    innerProps.className = "clr-col";
  } else {
    // Prepare responsive columns
    innerProps.className = `${innerProps.className} ${generateClarityCss("col", span)}`;

    // Prepare offset
    innerProps.className = `${innerProps.className} ${generateClarityCss("offset", offset)}`;
  }

  // Add the role if the column is a list item
  if (listItem === true) {
    innerProps.role = "listitem";
  }

  // Trim the className before passing it
  innerProps.className = innerProps.className.trim();

  return <div {...innerProps}>{children}</div>;
};

Column.propTypes = {
  children: PropTypes.oneOfType([
    PropTypes.arrayOf(PropTypes.element),
    PropTypes.element,
    PropTypes.string,
  ]).isRequired,
  offset: PropTypes.oneOfType([PropTypes.arrayOf(PropTypes.number), PropTypes.number]),
  span: PropTypes.oneOfType([PropTypes.arrayOf(PropTypes.number), PropTypes.number]),
  // For accessibility. A lot of times, column is a list item inside a row
  listItem: PropTypes.bool,
};

Column.defaultProps = {
  listItem: false,
};

export default Column;
