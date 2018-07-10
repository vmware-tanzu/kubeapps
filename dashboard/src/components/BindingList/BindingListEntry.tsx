import * as React from "react";
import { IK8sApiSecretResponse, IServiceBinding } from "../../shared/ServiceBinding";

import { RemoveBindingButton } from "../InstanceView/RemoveBindingButton";
import BindingDetails from "./BindingDetails";

interface IBindingEntryProps {
  binding: IServiceBinding;
  secret?: IK8sApiSecretResponse;
  removeBinding: (name: string, namespace: string) => Promise<boolean>;
}

interface IBindingEntryState {
  isExpanded: boolean;
}

export class BindingEntry extends React.Component<IBindingEntryProps, IBindingEntryState> {
  public state = {
    isExpanded: false,
  };

  public render() {
    const { binding, removeBinding, secret } = this.props;
    const { name, namespace } = binding.metadata;

    const condition = [...binding.status.conditions].shift();
    const currentStatus = condition ? (
      <div className="condition">
        <code>{condition.message}</code>
      </div>
    ) : (
      undefined
    );

    const rows = [
      <tr key={"row"}>
        <td>
          {namespace}/{name}
        </td>
        <td>{currentStatus}</td>
        <td style={{ display: "flex", justifyContent: "space-around" }}>
          <button className={"button button-primary button-small"} onClick={this.toggleExpand}>
            Expand/Collapse
          </button>
          <RemoveBindingButton binding={binding} removeBinding={removeBinding} />
        </td>
      </tr>,
    ];
    if (this.state.isExpanded) {
      rows.push(
        <tr key="info">
          <td colSpan={3}>
            <BindingDetails binding={binding} secret={secret} />
          </td>
        </tr>,
      );
    }
    return rows;
  }

  private toggleExpand = async () => this.setState({ isExpanded: !this.state.isExpanded });
}
