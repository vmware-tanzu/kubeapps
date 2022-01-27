// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import PropTypes from "prop-types";
import React from "react";

const Checkbox = ({ value, label, id, ...otherProps }) => (
  <div className="clr-checkbox-wrapper">
    <input type="checkbox" value={value} checked={value} id={id} {...otherProps} />
    <label className="clr-control-label" htmlFor={id}>
      {label}
    </label>
  </div>
);

Checkbox.propTypes = {
  name: PropTypes.string.isRequired,
  id: PropTypes.string.isRequired,
  label: PropTypes.string.isRequired,
  value: PropTypes.bool,
  otherProps: PropTypes.object,
};

export default Checkbox;
