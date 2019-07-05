import * as React from "react";
import { IServiceInstance } from "../../shared/ServiceInstance";
import ConfirmDialog from "../ConfirmDialog";

interface IDeprovisionButtonProps {
  disabled: boolean;
  instance: IServiceInstance;
  deprovision: (instance: IServiceInstance) => Promise<boolean>;
}

interface IDeprovisionButtonState {
  instance: IServiceInstance | undefined;
  isDeprovisioning: boolean;
  modalIsOpen: boolean;
}

class DeprovisionButton extends React.Component<IDeprovisionButtonProps, IDeprovisionButtonState> {
  public state: IDeprovisionButtonState = {
    instance: this.props.instance,
    isDeprovisioning: false,
    modalIsOpen: false,
  };

  public handleDeprovision = async () => {
    const { deprovision, instance } = this.props;
    this.setState({ isDeprovisioning: true, modalIsOpen: false });

    await deprovision(instance);
    this.setState({ isDeprovisioning: false });
  };

  public render() {
    return (
      <div className="DeprovisionButton">
        {this.state.isDeprovisioning && <div>Deprovisioning...</div>}
        <ConfirmDialog
          onConfirm={this.handleDeprovision}
          modalIsOpen={this.state.modalIsOpen}
          loading={false}
          closeModal={this.closeModal}
        />

        <button
          className="button button-primary button-danger"
          disabled={this.state.isDeprovisioning || this.props.disabled}
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
