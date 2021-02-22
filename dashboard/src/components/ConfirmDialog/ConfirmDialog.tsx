import { CdsButton } from "@cds/react/button";
import Alert from "components/js/Alert";
import Modal from "components/js/Modal/Modal";
import React from "react";
import LoadingWrapper from "../LoadingWrapper/LoadingWrapper";
import "./ConfirmDialog.css";

interface IConfirmDialogProps {
  modalIsOpen: boolean;
  loading: boolean;
  extraElem?: JSX.Element;
  confirmationText: string;
  confirmationButtonText?: string;
  error?: Error;
  onConfirm: () => any;
  closeModal: () => any;
}

function ConfirmDialog({
  modalIsOpen,
  loading,
  extraElem,
  confirmationButtonText,
  confirmationText,
  onConfirm,
  closeModal,
  error,
}: IConfirmDialogProps) {
  return (
    <Modal showModal={modalIsOpen} onModalClose={closeModal}>
      {error && <Alert theme="danger">An error ocurred: {error.message}</Alert>}
      {loading === true ? (
        <div className="confirmation-modal">
          <span>Loading, please wait</span>
          <LoadingWrapper loaded={false} />
        </div>
      ) : (
        <div className="confirmation-modal">
          <span>{confirmationText}</span>
          {extraElem}
          <div className="confirmation-modal-buttons">
            <CdsButton action="outline" type="button" onClick={closeModal}>
              Cancel
            </CdsButton>
            <CdsButton status="danger" type="submit" onClick={onConfirm}>
              {confirmationButtonText || "Delete"}
            </CdsButton>
          </div>
        </div>
      )}
    </Modal>
  );
}

export default ConfirmDialog;
