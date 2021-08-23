import PropTypes from "prop-types";
import React from "react";
import "./MultiCheckbox.scss";
import { assignCssVariables } from "./utils/cssVariables";

const MultiCheckbox = ({ options, span, value, ...otherProps }) => (
  <div className="multicheckbox-wrapper" style={assignCssVariables(span)}>
    {options.map((opt, i) => (
      <div key={i} className="clr-checkbox-wrapper">
        <input
          {...otherProps}
          type="checkbox"
          value={opt}
          id={`${otherProps.name}-${i}`}
          checked={value.includes(opt)}
        />
        <label className="clr-control-label" htmlFor={`${otherProps.name}-${i}`}>
          {opt}
        </label>
      </div>
    ))}
  </div>
);

MultiCheckbox.propTypes = {
  span: PropTypes.oneOfType([PropTypes.arrayOf(PropTypes.number), PropTypes.number]),
  value: PropTypes.array,
  options: PropTypes.arrayOf(PropTypes.string).isRequired,
};

export default MultiCheckbox;
