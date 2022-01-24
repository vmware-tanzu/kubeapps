// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import PropTypes from "prop-types";
import React from "react";

const Input = ({ children, ...fieldProps }) => (
  <div className="clr-input-wrapper">
    <input className="clr-input" {...fieldProps} />
    {children}
  </div>
);

Input.propTypes = {
  // Optional item to add. It could be a help section or an error message
  children: PropTypes.node,
  name: PropTypes.string.isRequired,
  placeholder: PropTypes.string,
  type: PropTypes.string,
};

Input.defaultProps = {
  children: null,
};

export default Input;
