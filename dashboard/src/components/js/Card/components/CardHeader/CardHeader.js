// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import cs from "classnames";
import PropTypes from "prop-types";
import React from "react";
import "./CardHeader.scss";

const CardHeader = ({ children, noBorder }) => {
  const cssClass = cs("card-header", {
    "no-border": noBorder,
  });
  return <header className={cssClass}>{children}</header>;
};

CardHeader.propTypes = {
  children: PropTypes.oneOfType([
    PropTypes.arrayOf(PropTypes.element),
    PropTypes.element,
    PropTypes.string,
  ]).isRequired,
  noBorder: PropTypes.bool,
};

CardHeader.defaultProps = {
  noBorder: false,
};

export default CardHeader;
