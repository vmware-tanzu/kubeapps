import { CdsButton } from "components/Clarity/clarity";
import Modal from "components/js/Modal/Modal";
import React from "react";
import LoadingWrapper from "../LoadingWrapper/LoadingWrapper.v2";
import "./ConfirmDialog.v2.css";

interface IConfirmDialogProps {
  modalIsOpen: boolean;
  loading: boolean;
  extraElem?: JSX.Element;
  confirmationText?: string;
  confirmationButtonText?: string;
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
}: IConfirmDialogProps) {
  return (
    <Modal showModal={modalIsOpen} onModalClose={closeModal}>
      {loading === true ? (
        <div className="confirmation-modal">
          <div>Loading, please wait</div>
          <LoadingWrapper loaded={false} />
        </div>
      ) : (
        <div className="confirmation-modal">
          <div>{confirmationText || "Are you sure you want to delete this?"}</div>
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
