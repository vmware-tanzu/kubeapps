import * as React from "react";
import { X } from "react-feather";
import * as Modal from "react-modal";

// TODO(andresmgot) This is a temporary component
// A form should be rendered instead asking for the operator
// details and creating a Subscription

interface IOperatorInstallationProps {
  name: string;
  namespaced: boolean;
  closeModal: () => void;
  modalIsOpen: boolean;
}

class OperatorInstallation extends React.Component<IOperatorInstallationProps> {
  public render() {
    return (
      <Modal
        className="centered-modal"
        isOpen={this.props.modalIsOpen}
        onRequestClose={this.props.closeModal}
        contentLabel="Modal"
      >
        <div className="container">
          <div className="row">
            <div className="col-10">
              <h5>Install {this.props.name}</h5>
            </div>
            <div className="col-2 text-r">
              <a onClick={this.props.closeModal} style={{ color: "inherit" }}>
                <X />
              </a>
            </div>
          </div>
          Install the operator by running the following command:
          <section className="AppNotes Terminal elevation-1 margin-v-big">
            <div className="Terminal__Top type-small">
              <div className="Terminal__Top__Buttons">
                <span className="Terminal__Top__Button Terminal__Top__Button--red" />
                <span className="Terminal__Top__Button Terminal__Top__Button--yellow" />
                <span className="Terminal__Top__Button Terminal__Top__Button--green" />
              </div>
            </div>
            <div className="Terminal__Tab">
              <pre className="Terminal__Code">
                <code>kubectl create -f https://operatorhub.io/install/{this.props.name}.yaml</code>
              </pre>
            </div>
          </section>
          {this.props.namespaced ? (
            <span>
              This Operator will be installed in the <code>my-{this.props.name}</code> namespace and
              will be usable from this namespace only.
            </span>
          ) : (
            <span>
              This Operator will be installed in the <code>operators</code> namespace and will be
              usable from all namespaces in the cluster.
            </span>
          )}
        </div>
      </Modal>
    );
  }
}

export default OperatorInstallation;
