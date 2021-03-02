import { CdsButton } from "@cds/react/button";
import { CdsModal, CdsModalActions, CdsModalContent, CdsModalHeader } from "@cds/react/modal";
import React from "react";

interface IModalProps {
  children: JSX.Element;
  footer?: JSX.Element;
  hideCloseButton?: boolean;
  modalSize?: "default" | "sm" | "lg" | "xl" | undefined;
  showModal: boolean;
  staticBackdrop?: boolean;
  title?: string;
  onModalClose: () => any;
}

function Modal({
  children,
  footer,
  hideCloseButton,
  modalSize,
  showModal,
  title,
  onModalClose,
}: IModalProps) {
  modalSize = modalSize || "default";
  return (
    <>
      {showModal && (
        <CdsModal size={modalSize} closable={!hideCloseButton} onCloseChange={onModalClose}>
          <CdsModalHeader>{title}</CdsModalHeader>
          <CdsModalContent>{children}</CdsModalContent>
          <CdsModalActions>
            {footer}
            {!hideCloseButton && (
              <CdsButton action="outline" type="button" onClick={onModalClose}>
                Cancel
              </CdsButton>
            )}
          </CdsModalActions>
        </CdsModal>
      )}
    </>
  );
}

export default Modal;
