import * as React from "react";

import { IServiceBindingWithSecret } from "../../shared/ServiceBinding";

interface IRemoveBindingButtonProps {
  bindingWithSecret: IServiceBindingWithSecret;
  removeBinding: (name: string, ns: string) => Promise<boolean>;
}

class RemoveBindingButton extends React.Component<IRemoveBindingButtonProps> {
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
    const {
      removeBinding,
      bindingWithSecret: { binding },
    } = this.props;
    const { name, namespace } = binding.metadata;
    removeBinding(name, namespace);
  };
}

export default RemoveBindingButton;
