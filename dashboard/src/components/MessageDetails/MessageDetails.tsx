import * as React from "react";
import * as Modal from "react-modal";

import "./MessageDetails.css";

interface IMessageDetailsProps {
  modalIsOpen: boolean;
  closeModal: () => Promise<any>;
  message: string;
}

class MessageDetails extends React.Component<IMessageDetailsProps> {
  public render() {
    return (
      <div className="MessageDetails">
        <Modal
          isOpen={this.props.modalIsOpen}
          onRequestClose={this.props.closeModal}
          contentLabel="Modal"
          className="Terminal"
          style={{
            overlay: {
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
            },
            content: {
              maxWidth: "80%",
            },
          }}
        >
          <div className="Terminal__Top type-small">
            <div className="Terminal__Top__Buttons">
              <a>
                <span
                  className="Terminal__Top__Button Terminal__Top__Button--red"
                  onClick={this.props.closeModal}
                />
              </a>
            </div>
            <div className="Terminal__Top__Title">Message</div>
          </div>
          <div className="Terminal__Tab">
            <pre className="Terminal__Code">
              <code>{this.props.message}</code>
            </pre>
          </div>
        </Modal>
      </div>
    );
  }
}

export default MessageDetails;
