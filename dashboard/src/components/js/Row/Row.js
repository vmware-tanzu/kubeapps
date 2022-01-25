// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import PropTypes from "prop-types";
import React from "react";

/**
 * https://v2.clarity.design/grid
 */
const Row = ({ list, children }) => {
  const innerProps = {
    className: "clr-row",
  };

  if (list === true) {
    innerProps["role"] = "list";
  }

  return <div {...innerProps}>{children}</div>;
};

Row.propTypes = {
  children: PropTypes.oneOfType([
    PropTypes.arrayOf(PropTypes.element),
    PropTypes.element,
    PropTypes.string,
  ]).isRequired,
  // For accessibility. A lot of times, rows include a list of elements
  list: PropTypes.bool,
};

Row.defaultProps = {
  list: false,
};

export default Row;
