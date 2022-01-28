// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import { CdsControlMessage } from "@cds/react/forms";
import { CdsModal, CdsModalActions, CdsModalContent, CdsModalHeader } from "@cds/react/modal";
import { CdsSelect } from "@cds/react/select";
import Alert from "components/js/Alert";
import { useEffect, useState } from "react";
import { DeleteError, FetchWarning } from "shared/types";
import LoadingWrapper from "../../../LoadingWrapper/LoadingWrapper";
import "./RollbackDialog.css";

interface IRollbackDialogProps {
  loading: boolean;
  currentRevision: number;
  onConfirm: (revision: number) => Promise<any>;
  closeModal: () => void;
  error?: Error;
  modalIsOpen: boolean;
}

function RollbackDialog({
  loading,
  currentRevision,
  error,
  modalIsOpen,
  onConfirm,
  closeModal,
}: IRollbackDialogProps) {
  const [targetRevision, setTargetRevision] = useState(currentRevision);
  const [hasUserChanges, setHasUserChanges] = useState(false);
  const options: number[] = [];
  const disableRollback = currentRevision === 1;
  const selectRevision = (e: React.ChangeEvent<HTMLSelectElement>) => {
    setHasUserChanges(true);
    setTargetRevision(Number(e.target.value));
  };
  const onClick = () => {
    onConfirm(Number(targetRevision));
  };
  // Use as options the number of versions without the latest
  for (let i = Number(currentRevision) - 1; i > 0; i--) {
    options.push(i);
  }

  useEffect(() => {
    if (!hasUserChanges) {
      setTargetRevision(currentRevision - 1);
    }
  }, [hasUserChanges, currentRevision]);

  /* eslint-disable jsx-a11y/label-has-associated-control */
  return (
    <>
      {modalIsOpen && (
        <CdsModal closable={false} onCloseChange={closeModal}>
          <CdsModalHeader>Rollback application</CdsModalHeader>
          <CdsModalContent>
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
            <LoadingWrapper className="center" loadingText="Loading, please wait" loaded={!loading}>
              {disableRollback ? (
                <p>The application has not been upgraded, it's not possible to rollback.</p>
              ) : (
                <>
                  <CdsSelect layout="horizontal" id="revision-selector">
                    <label>Select the revision to which you want to rollback</label>
                    <select value={targetRevision} onChange={selectRevision}>
                      {options.map(o => (
                        <option key={o} value={o}>
                          {o}
                        </option>
                      ))}
                    </select>
                    <CdsControlMessage>(current: {currentRevision})</CdsControlMessage>
                  </CdsSelect>
                </>
              )}
            </LoadingWrapper>
          </CdsModalContent>
          <CdsModalActions>
            <CdsButton action="outline" type="button" onClick={closeModal}>
              {" "}
              Cancel{" "}
            </CdsButton>
            <CdsButton status="danger" type="submit" onClick={onClick} disabled={disableRollback}>
              Rollback
            </CdsButton>
          </CdsModalActions>
        </CdsModal>
      )}
    </>
  );
}

export default RollbackDialog;
