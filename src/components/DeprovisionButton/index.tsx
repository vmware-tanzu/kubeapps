import * as React from "react";
import { IServiceInstance } from "../../shared/ServiceCatalog";

interface IDeprovisionButtonProps {
  instance: IServiceInstance;
  deprovision: (instance: IServiceInstance) => Promise<{}>;
}

interface IDeprovisionButtonState {
  error: string | undefined;
  instance: IServiceInstance | undefined;
  isDeprovisioning: boolean;
}

class DeprovisionButton extends React.Component<IDeprovisionButtonProps, IDeprovisionButtonState> {
  public state: IDeprovisionButtonState = {
    error: undefined,
    instance: this.props.instance,
    isDeprovisioning: false,
  };

  public handleDeprovision = async () => {
    const { deprovision, instance } = this.props;
    this.setState({ isDeprovisioning: true });

    try {
      await deprovision(instance);
      this.setState({ isDeprovisioning: false });
    } catch (err) {
      this.setState({ isDeprovisioning: false, error: err.toString() });
    }
  };

  public render() {
    return (
      <div className="DeprovisionButton">
        {this.state.isDeprovisioning && <div>Deprovisioning...</div>}
        <button
          className="button button-primary"
          disabled={this.state.isDeprovisioning}
          onClick={this.handleDeprovision}
        >
          Deprovision
        </button>
      </div>
    );
  }
}

export default DeprovisionButton;
