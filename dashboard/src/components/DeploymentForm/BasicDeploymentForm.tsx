import * as React from "react";
import { IBasicFormParam } from "shared/types";

export interface IBasicDeploymentFormProps {
  params: IBasicFormParam[];
  handleBasicFormParamChange: (
    p: IBasicFormParam,
  ) => (e: React.FormEvent<HTMLInputElement>) => void;
}

class BasicDeploymentForm extends React.Component<IBasicDeploymentFormProps> {
  public render() {
    return this.props.params.map(param => {
      return this.renderParam(param);
    });
  }

  private renderParam(param: IBasicFormParam) {
    switch (param.name) {
      case "username":
        return (
          <div key={param.name}>
            <label htmlFor="username">Username</label>
            <input
              id="username"
              onChange={this.props.handleBasicFormParamChange(param)}
              value={param.value}
            />
          </div>
        );
    }
    return null;
  }
}

export default BasicDeploymentForm;
