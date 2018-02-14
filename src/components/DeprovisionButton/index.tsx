import * as React from "react";
import { IServiceInstance } from "../../shared/ServiceInstance";
import ConfirmDialog from "../ConfirmDialog";

interface IDeprovisionButtonProps {
  instance: IServiceInstance;
  deprovision: (instance: IServiceInstance) => Promise<{}>;
}

interface IDeprovisionButtonState {
  error: string | undefined;
  instance: IServiceInstance | undefined;
  isDeprovisioning: boolean;
  modalIsOpen: boolean;
}

class DeprovisionButton extends React.Component<IDeprovisionButtonProps, IDeprovisionButtonState> {
  public state: IDeprovisionButtonState = {
    error: undefined,
    instance: this.props.instance,
    isDeprovisioning: false,
    modalIsOpen: false,
  };

  public handleDeprovision = async () => {
    const { deprovision, instance } = this.props;
    this.setState({ isDeprovisioning: true, modalIsOpen: false });

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
        <ConfirmDialog
          onConfirm={this.handleDeprovision}
          modalIsOpen={this.state.modalIsOpen}
          closeModal={this.closeModal}
        />

        <button
          className="button button-primary button-small button-danger"
          disabled={this.state.isDeprovisioning}
          onClick={this.openModel}
        >
          Deprovision
        </button>
      </div>
    );
  }

  public openModel = () => {
    this.setState({
      modalIsOpen: true,
    });
  };

  public closeModal = async () => {
    this.setState({
      modalIsOpen: false,
    });
  };
}

export default DeprovisionButton;
