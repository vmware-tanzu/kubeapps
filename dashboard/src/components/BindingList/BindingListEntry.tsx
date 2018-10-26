import * as React from "react";
import { IServiceBindingWithSecret } from "../../shared/ServiceBinding";

import TerminalModal from "../TerminalModal";
import BindingDetails from "./BindingDetails";
import RemoveBindingButton from "./RemoveBindingButton";

import "./BindingList.css";

interface IBindingListEntryProps {
  bindingWithSecret: IServiceBindingWithSecret;
  removeBinding: (name: string, namespace: string) => Promise<boolean>;
}

interface IBindingListEntryState {
  modalIsOpen: boolean;
}

class BindingListEntry extends React.Component<IBindingListEntryProps, IBindingListEntryState> {
  public state = {
    modalIsOpen: false,
  };

  public render() {
    const {
      bindingWithSecret,
      bindingWithSecret: { binding },
    } = this.props;
    const { name } = binding.metadata;

    let reason = <span />;
    let message = "";
    const condition = [...binding.status.conditions]
      .sort((a, b) => a.lastTransitionTime.localeCompare(b.lastTransitionTime))
      .pop();
    if (condition) {
      reason = <code>{condition.reason}</code>;
      message = condition.message;
    }

    return (
      <tr>
        <td>{name}</td>
        <td>{reason}</td>
        <td>
          <button className="button button-small" onClick={this.openModal}>
            Show message
          </button>
          <TerminalModal
            modalIsOpen={this.state.modalIsOpen}
            closeModal={this.closeModal}
            title="Status Message"
            message={message}
          />
        </td>
        <td>
          <BindingDetails {...bindingWithSecret} />
        </td>
        <td>
          <RemoveBindingButton {...this.props} />
        </td>
      </tr>
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

export default BindingListEntry;
