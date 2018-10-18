import * as React from "react";
import { IServiceBindingWithSecret } from "../../shared/ServiceBinding";

import MessageDetails from "../MessageDetails";
import BindingDetails from "./BindingDetails";
import RemoveBindingButton from "./RemoveBindingButton";

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
      <tr className="row padding-t-small">
        <td className="col-2">{name}</td>
        <td className="col-2">{reason}</td>
        <td className="col-2">
          <button className="button" onClick={this.openModal}>
            Show message
          </button>
          <MessageDetails
            modalIsOpen={this.state.modalIsOpen}
            closeModal={this.closeModal}
            message={message}
          />
        </td>
        <td className="col-4">
          <BindingDetails {...bindingWithSecret} />
        </td>
        <td className="col-2">
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
