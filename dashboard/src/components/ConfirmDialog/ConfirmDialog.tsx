// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsModal, CdsModalActions, CdsModalContent, CdsModalHeader } from "@cds/react/modal";
import Alert from "components/js/Alert";
import { DeleteError, FetchWarning } from "shared/types";
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
              <Alert theme="warning">
                There is a problem with this package: {error["message"]}
              </Alert>
            ) : error.constructor === DeleteError ? (
              <Alert theme="danger">
                Unable to delete the application. Received: {error["message"]}
              </Alert>
            ) : (
              <Alert theme="danger">An error occurred: {error["message"]}</Alert>
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
                {/* TODO(andresmgot): CdsButton "type" property doesn't work, so we need to use a normal <button>
                  https://github.com/vmware/clarity/issues/5038
                  */}
                <button className="btn btn-info-outline" type="button" onClick={closeModal}>
                  Cancel
                </button>
                <button className="btn btn-danger" type="submit" onClick={onConfirm}>
                  {confirmationButtonText || "Delete"}
                </button>
              </CdsModalActions>
            </>
          )}
        </CdsModal>
      )}
    </>
  );
}

export default ConfirmDialog;
