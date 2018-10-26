import * as React from "react";
import * as Modal from "react-modal";

import "./TerminalModal.css";

interface ITerminalModalProps {
  modalIsOpen: boolean;
  closeModal: () => Promise<any>;
  title: string;
  message: string;
}

const TerminalModal: React.SFC<ITerminalModalProps> = props => {
  return (
    <div className="MessageDetails">
      <Modal
        isOpen={props.modalIsOpen}
        onRequestClose={props.closeModal}
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
                onClick={props.closeModal}
              />
            </a>
          </div>
          <div className="Terminal__Top__Title">{props.title}</div>
        </div>
        <div className="Terminal__Tab">
          <pre className="Terminal__Code">
            <code>{props.message}</code>
          </pre>
        </div>
      </Modal>
    </div>
  );
};

export default TerminalModal;
