import { CdsButton } from "components/Clarity/clarity";
import Alert from "components/js/Alert";
import React, { useState } from "react";
import LoadingWrapper from "../../../LoadingWrapper/LoadingWrapper.v2";
import "./RollbackDialog.v2.css";

interface IRollbackDialogProps {
  loading: boolean;
  currentRevision: number;
  onConfirm: (revision: number) => Promise<any>;
  closeModal: () => void;
  error?: Error;
}

function RollbackDialog({
  loading,
  currentRevision,
  error,
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
  return (
    <div className="rollback-menu">
      {error && <Alert theme="danger">Found error: {error.message}</Alert>}
      {loading && <p className="rollback-menu-text">Loading, please wait.</p>}
      <LoadingWrapper loaded={!loading}>
        {disableRollback ? (
          <span className="rollback-menu-text">
            The application has not been upgraded, it's not possible to rollback.
          </span>
        ) : (
          <>
            <label htmlFor="revision-selector" className="rollback-menu-label">
              Select the revision to which you want to rollback (current: {currentRevision})
            </label>
            <div className="clr-select-wrapper">
              <select
                id="revision-selector"
                onChange={selectRevision}
                className="clr-page-size-select"
              >
                {options.map(o => (
                  <option key={o} value={o}>
                    {o}
                  </option>
                ))}
              </select>
            </div>
          </>
        )}
        <div className="rollback-menu-buttons">
          <CdsButton action="outline" type="button" onClick={closeModal}>
            Cancel
          </CdsButton>
          <CdsButton status="danger" type="submit" onClick={onClick} disabled={disableRollback}>
            Rollback
          </CdsButton>
        </div>
      </LoadingWrapper>
    </div>
  );
}

export default RollbackDialog;
