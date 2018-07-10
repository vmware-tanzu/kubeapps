import * as React from "react";

import { IServiceBindingWithSecret } from "../../shared/ServiceBinding";
import BindingDetails from "../BindingList/BindingDetails";

interface IDeploymentBindingProps {
  bindingsWithSecrets: IServiceBindingWithSecret[];
}

interface IDeploymentBindingState {
  selectedBinding: IServiceBindingWithSecret | undefined;
}

class DeploymentBinding extends React.Component<IDeploymentBindingProps, IDeploymentBindingState> {
  public state: IDeploymentBindingState = {
    selectedBinding: undefined,
  };
  public render() {
    const { selectedBinding } = this.state;
    const bindingDetail = selectedBinding ? <BindingDetails {...selectedBinding} /> : <div />;
    return (
      <div>
        <p>[Optional] Select a service binding for your new app</p>
        <label htmlFor="bindings">Bindings</label>
        <select onChange={this.onBindingChange}>
          <option key="none" value="none">
            {" "}
            -- Select one --
          </option>
          {this.props.bindingsWithSecrets.map(b => (
            <option
              key={b.binding.metadata.name}
              selected={
                b.binding.metadata.name ===
                (selectedBinding && selectedBinding.binding.metadata.name)
              }
              value={b.binding.metadata.name}
            >
              {b.binding.metadata.name}
            </option>
          ))}
        </select>
        {bindingDetail}
      </div>
    );
  }

  public onBindingChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    this.setState({
      selectedBinding:
        this.props.bindingsWithSecrets.find(b => b.binding.metadata.name === e.target.value) ||
        undefined,
    });
  };
}

export default DeploymentBinding;
