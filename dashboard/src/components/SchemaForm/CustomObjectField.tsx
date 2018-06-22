import * as React from "react";
import AceEditor from "react-ace";
import { FieldProps } from "react-jsonschema-form";
import ObjectField from "react-jsonschema-form/lib/components/fields/ObjectField";

import "brace/mode/json";
import "brace/theme/xcode";

interface ICustomObjectFieldState {
  rawValue: string;
}

// CustomObjectField that handles the rendering of JSON Schema object
// definitions where properties are undefined. This can be, for example, to
// facilitate input of additionalProperties or patternProperties. This will
// render an AceEditor for raw JSON input.
//
// See also https://github.com/mozilla-services/react-jsonschema-form/issues/44.
class CustomObjectField extends React.Component<FieldProps, ICustomObjectFieldState> {
  public state: ICustomObjectFieldState = {
    rawValue:
      (this.props.schema.default as string | undefined) || JSON.stringify({ foo: "bar" }, null, 2),
  };

  public render() {
    const { name, schema } = this.props;
    const { rawValue } = this.state;
    // if the schema has properties, render the upstream ObjectField
    if (schema.properties) {
      return <ObjectField {...this.props} />;
    }
    const label = schema.title || name;
    return (
      <div>
        <label htmlFor={label}>{label} (JSON)</label>
        <AceEditor
          className="margin-b-big"
          mode="json"
          theme="xcode"
          name={name}
          width="100%"
          height="100px"
          onChange={this.handleValueChange}
          setOptions={{ showPrintMargin: false }}
          value={rawValue}
        />
      </div>
    );
  }

  public handleValueChange = (value: string) => {
    this.setState({ rawValue: value });
    try {
      const json = JSON.parse(value);
      this.props.onChange(json);
    } catch {
      // do nothing
    }
  };
}

export default CustomObjectField;
