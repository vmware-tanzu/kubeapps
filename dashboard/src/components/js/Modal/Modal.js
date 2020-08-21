import React from "react";
import PropTypes from "prop-types";

import { get } from "lodash-es";

export const modalSizes = {
  sm: "modal-sm",
  md: "modal-md",
  lg: "modal-lg",
  xl: "modal-xl",
};

const Modal = ({
  title,
  children,
  footer,
  hideCloseButton,
  modalSize,
  onModalClose,
  showModal,
  staticBackdrop,
}) => {
  const onClose = () => onModalClose();

  if (!showModal) {
    return null;
  }

  const size = get(modalSizes, modalSize, modalSizes["md"]);
  const closeOnBackdrop = staticBackdrop ? () => {} : onModalClose;

  return (
    <div className="modal">
      <div
        className={`modal-dialog ${size}`}
        aria-modal={true}
        role="dialog"
        aria-hidden={showModal}
      >
        <div className="modal-content">
          <div className="modal-header">
            {!hideCloseButton && (
              <button
                aria-label="Close the current modal"
                className="close"
                type="button"
                onClick={onClose}
              >
                <clr-icon aria-hidden="true" shape="close" />
              </button>
            )}
            <h3 className="modal-title">{title}</h3>
          </div>
          <div className="modal-body">{children}</div>
          <div className="modal-footer">{footer}</div>
        </div>
      </div>
      <div onClick={closeOnBackdrop} className="modal-backdrop" aria-hidden="true" />
    </div>
  );
};

Modal.defaultProps = {
  modalSize: "md",
  hideCloseButton: false,
  showModal: false,
  staticBackdrop: true,
};

Modal.propTypes = {
  title: PropTypes.string,
  children: PropTypes.node.isRequired,
  footer: PropTypes.node,
  hideCloseButton: PropTypes.bool.isRequired,
  modalSize: PropTypes.string.isRequired,
  onModalClose: PropTypes.func.isRequired,
  showModal: PropTypes.bool.isRequired,
  staticBackdrop: PropTypes.bool,
};

export default Modal;
