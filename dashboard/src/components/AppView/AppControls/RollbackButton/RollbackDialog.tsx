import { CdsButton } from "@cds/react/button";
import { CdsControlMessage } from "@cds/react/forms";
import { CdsModal, CdsModalActions, CdsModalContent, CdsModalHeader } from "@cds/react/modal";
import { CdsSelect } from "@cds/react/select";
import Alert from "components/js/Alert";
import { useState } from "react";
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
  const [targetRevision, setTargetRevision] = useState(currentRevision - 1);
  const options: number[] = [];
  // If there are no revisions to rollback to, disable
  const disableRollback = currentRevision === 1;
  const selectRevision = (e: React.FormEvent<HTMLSelectElement>) => {
    setTargetRevision(Number(e.currentTarget.value));
  };
  const onClick = () => {
    onConfirm(targetRevision);
  };
  // Use as options the number of versions without the latest
  for (let i = currentRevision - 1; i > 0; i--) {
    options.push(i);
  }

  /* eslint-disable jsx-a11y/label-has-associated-control */
  return (
    <>
      {modalIsOpen && (
        <CdsModal closable={false} onCloseChange={closeModal}>
          <CdsModalHeader>Rollback application</CdsModalHeader>
          <CdsModalContent>
            {error && <Alert theme="danger">An error occurred: {error.message}</Alert>}
            <LoadingWrapper className="center" loadingText="Loading, please wait" loaded={!loading}>
              {disableRollback ? (
                <p>The application has not been upgraded, it's not possible to rollback.</p>
              ) : (
                <>
                  <CdsSelect layout="horizontal" id="revision-selector" onChange={selectRevision}>
                    <label>Select the revision to which you want to rollback</label>
                    <select>
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
