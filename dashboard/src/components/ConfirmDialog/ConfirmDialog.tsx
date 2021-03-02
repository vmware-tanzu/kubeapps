import { CdsButton } from "@cds/react/button";
import { CdsModal, CdsModalActions, CdsModalContent, CdsModalHeader } from "@cds/react/modal";
import Alert from "components/js/Alert";
import React from "react";
import LoadingWrapper from "../LoadingWrapper/LoadingWrapper";
import "./ConfirmDialog.css";

interface IConfirmDialogProps {
  modalIsOpen: boolean;
  loading: boolean;
  extraElem?: JSX.Element;
  headerText?: string;
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
  headerText,
  confirmationButtonText,
  confirmationText,
  onConfirm,
  closeModal,
  error,
}: IConfirmDialogProps) {
  return (
    <>
      {modalIsOpen && (
        <CdsModal closable={true} onCloseChange={closeModal}>
          <CdsModalHeader>{headerText}</CdsModalHeader>
          {error && <Alert theme="danger">An error ocurred: {error.message}</Alert>}
          {loading === true ? (
            <>
              <CdsModalContent>
                <span>Loading, please wait</span>
                <LoadingWrapper loaded={false} />
              </CdsModalContent>
            </>
          ) : (
            <>
              <CdsModalContent>
                <p>{confirmationText}</p>
                <p>{extraElem}</p>
              </CdsModalContent>
              <CdsModalActions>
                <CdsButton action="outline" type="button" onClick={closeModal}>
                  Cancel
                </CdsButton>
                <CdsButton status="danger" type="submit" onClick={onConfirm}>
                  {confirmationButtonText || "Delete"}
                </CdsButton>
              </CdsModalActions>
            </>
          )}
        </CdsModal>
      )}
    </>
  );
}

export default ConfirmDialog;
