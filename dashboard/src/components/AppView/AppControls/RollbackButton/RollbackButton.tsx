import * as React from "react";
import Modal from "react-modal";
import RollbackDialog from "./RollbackDialog";

export interface IRollbackButtonProps {
  cluster: string;
  namespace: string;
  releaseName: string;
  revision: number;
  rollbackApp: (
    cluster: string,
    namespace: string,
    releaseName: string,
    revision: number,
  ) => Promise<boolean>;
  loading: boolean;
  error?: Error;
}

interface IRollbackButtonState {
  modalIsOpen: boolean;
  loading: boolean;
}

class RollbackButton extends React.Component<IRollbackButtonProps> {
  public state: IRollbackButtonState = {
    modalIsOpen: false,
    loading: false,
  };

  public render() {
    return (
      <React.Fragment>
        <Modal
          className="centered-modal"
          isOpen={this.state.modalIsOpen}
          onRequestClose={this.closeModal}
          contentLabel="Modal"
        >
          <RollbackDialog
            onConfirm={this.handleRollback}
            loading={this.state.loading}
            closeModal={this.closeModal}
            currentRevision={this.props.revision}
          />
        </Modal>
        <button className="button" onClick={this.openModal}>
          Rollback
        </button>
      </React.Fragment>
    );
  }

  public openModal = () => {
    this.setState({
      modalIsOpen: true,
    });
  };

  public closeModal = async () => {
    this.setState({
      modalIsOpen: false,
    });
  };

  private handleRollback = async (revision: number) => {
    this.setState({ loading: true });
    const success = await this.props.rollbackApp(
      this.props.cluster,
      this.props.namespace,
      this.props.releaseName,
      revision,
    );
    // If there is an error it's catched at AppView level
    if (success) {
      this.setState({ loading: false });
      this.closeModal();
    }
  };
}

export default RollbackButton;
