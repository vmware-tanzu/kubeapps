import * as React from "react";
import { IServiceBinding, ServiceBinding } from "../../shared/ServiceBinding";

interface IRemoveBindingButtonProps {
  binding: IServiceBinding;
  onRemoveComplete?: () => Promise<any>;
}

export class RemoveBindingButton extends React.Component<IRemoveBindingButtonProps> {
  public render() {
    return (
      <div className="RemoveBindingButton">
        <button
          className="button button-small button-danger"
          onClick={this.handleRemoveBindingClick}
        >
          Remove
        </button>
      </div>
    );
  }

  private handleRemoveBindingClick = async () => {
    const { binding, onRemoveComplete } = this.props;
    const { name, namespace } = binding.metadata;
    await ServiceBinding.delete(name, namespace);
    if (onRemoveComplete) {
      await onRemoveComplete();
    }
  };
}
