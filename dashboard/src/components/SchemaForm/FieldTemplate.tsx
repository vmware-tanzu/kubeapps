import * as React from "react";
import { FieldTemplateProps } from "react-jsonschema-form";

// adapted from https://jsfiddle.net/hdp1kgn6/1/
const FieldTemplate: React.SFC<FieldTemplateProps> = props => {
  const { id, classNames, label, children, displayLabel, required } = props;
  return (
    <div className={classNames}>
      {displayLabel && (
        <label htmlFor={id}>
          {label}
          {required && " (required)"}
        </label>
      )}
      {children}
    </div>
  );
};

export default FieldTemplate;
