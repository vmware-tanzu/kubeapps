// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import cs from "classnames";
import PropTypes from "prop-types";
import React from "react";
import "./Spinner.scss";

/**
 * https://v2.clarity.design/spinners
 */
const Spinner = ({ center, inline, inverse, medium, small, text }) => {
  const cssClass = cs("spinner", {
    "spinner-center": center && !inline,
    "spinner-inline": inline,
    "spinner-inverse": inverse,
    "spinner-md": medium,
    "spinner-sm": small,
  });

  const containerCssClass = cs("spinner-container", {
    "spinner-center": center && !inline,
  });

  return (
    <span className={containerCssClass}>
      <span className={cssClass}>{text}</span>
    </span>
  );
};

Spinner.propTypes = {
  center: PropTypes.bool,
  inline: PropTypes.bool,
  inverse: PropTypes.bool,
  medium: PropTypes.bool,
  small: PropTypes.bool,
  text: PropTypes.string,
};

Spinner.defaultProps = {
  center: true,
  inline: false,
  inverse: false,
  medium: false,
  small: false,
  text: "Loading...",
};

export default Spinner;
