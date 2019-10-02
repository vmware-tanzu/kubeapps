import * as React from "react";
import { IBasicFormParam } from "shared/types";

export interface IBasicDeploymentFormProps {
  params: { [name: string]: IBasicFormParam };
  handleBasicFormParamChange: (
    name: string,
    p: IBasicFormParam,
  ) => (e: React.FormEvent<HTMLInputElement>) => void;
}

class BasicDeploymentForm extends React.Component<IBasicDeploymentFormProps> {
  public render() {
    return Object.keys(this.props.params).map(paramName => {
      return this.renderParam(paramName, this.props.params[paramName]);
    });
  }

  private renderParam(name: string, param: IBasicFormParam) {
    switch (name) {
      case "username":
        return (
          <div key={name}>
            <label htmlFor="username">Username</label>
            <input
              id="username"
              onChange={this.props.handleBasicFormParamChange(name, param)}
              value={param.value}
            />
          </div>
        );
    }
    return null;
  }
}

export default BasicDeploymentForm;
