import * as React from "react";

import TerminalModal from "../../components/TerminalModal";
import { IServiceBindingWithSecret } from "../../shared/ServiceBinding";
import "./BindingDetails.css";

interface IBindingDetailsState {
  modalIsOpen: boolean;
}

class BindingDetails extends React.Component<IServiceBindingWithSecret, IBindingDetailsState> {
  public state = {
    modalIsOpen: false,
  };

  public render() {
    const { binding, secret } = this.props;
    const { secretName } = binding.spec;

    let secretDataArray: string[][] = [];
    if (secret) {
      const secretData = Object.keys(secret.data).map(k => [k, atob(secret.data[k])]);
      secretDataArray = [...secretData];
    }
    let message = "";
    if (secretDataArray.length > 0) {
      message = "";
      secretDataArray.forEach(statusPair => {
        const [key, value] = statusPair;
        message += `${key}: ${value}\n`;
      });
    } else {
      message = "The secret is empty";
    }
    return (
      <dl className="BindingDetails container margin-normal">
        <dt key={secretName}>
          {secretName} <a onClick={this.openModal}>(show)</a>
        </dt>
        <TerminalModal
          modalIsOpen={this.state.modalIsOpen}
          closeModal={this.closeModal}
          title={`Secret: ${secretName}`}
          message={message}
        />
      </dl>
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
}

export default BindingDetails;
