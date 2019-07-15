import { JSONSchema6 } from "json-schema";
import * as React from "react";
import Form, { ISubmitEvent } from "react-jsonschema-form";

import ArrayFieldTemplate from "./ArrayFieldTemplate";
import CustomObjectField from "./CustomObjectField";
import FieldTemplate from "./FieldTemplate";
import "./SchemaForm.css";

interface ISchemaFormProps {
  schema: JSONSchema6;
  onSubmit?: (result: ISubmitEvent<any>) => void;
}

class SchemaForm extends React.Component<ISchemaFormProps> {
  public render() {
    return (
      <Form
        className="SchemaForm"
        schema={this.props.schema}
        onSubmit={this.props.onSubmit}
        autocomplete="off"
        FieldTemplate={FieldTemplate}
        ArrayFieldTemplate={ArrayFieldTemplate}
        fields={{ ObjectField: CustomObjectField }}
        noValidate={true}
      >
        {this.props.children}
      </Form>
    );
  }
}

export default SchemaForm;
