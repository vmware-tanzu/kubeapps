// Copyright 2018-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import { CdsModal, CdsModalActions, CdsModalContent, CdsModalHeader } from "@cds/react/modal";
import AlertGroup from "components/AlertGroup";
import LoadingWrapper from "components/LoadingWrapper";
import { DeleteError, FetchWarning } from "shared/types";
import "./ConfirmDialog.css";

interface IConfirmDialogProps {
  modalIsOpen: boolean;
  loading: boolean;
  extraElem?: JSX.Element;
  headerText?: string;
  confirmationText: string;
  confirmationButtonText?: string;
  error?: Error;
  size?: "sm" | "default" | "lg" | "xl";
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
  size,
}: IConfirmDialogProps) {
  return (
    <>
      {modalIsOpen && (
        <CdsModal size={size || "default"} closable={true} onCloseChange={closeModal}>
          {headerText && <CdsModalHeader>{headerText}</CdsModalHeader>}
          {error &&
            (error.constructor === FetchWarning ? (
              <AlertGroup withMargin={false} status="warning">
                There is a problem with this package: {error["message"]}.
              </AlertGroup>
            ) : error.constructor === DeleteError ? (
              <AlertGroup withMargin={false} status="danger">
                Unable to delete the application. Received: {error["message"]}.
              </AlertGroup>
            ) : (
              <AlertGroup withMargin={false} status="danger">
                An error occurred: {error["message"]}.
              </AlertGroup>
            ))}
          {loading === true ? (
            <div className="center">
              <CdsModalContent>
                <LoadingWrapper loadingText="Loading, please wait" loaded={false} />
              </CdsModalContent>
            </div>
          ) : (
            <>
              <CdsModalContent>
                <p>{confirmationText}</p>
                {extraElem && <p>{extraElem}</p>}
              </CdsModalContent>
              <CdsModalActions>
                <CdsButton type="button" onClick={closeModal} action="outline">
                  Cancel
                </CdsButton>

                <CdsButton type="button" onClick={onConfirm} status="danger">
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
