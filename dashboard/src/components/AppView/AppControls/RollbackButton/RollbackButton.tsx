// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsIcon } from "@cds/react/icon";
import actions from "actions";
import {
  InstalledPackageReference,
  InstalledPackageStatus,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { IStoreState } from "shared/types";
import StatusAwareButton from "../StatusAwareButton/StatusAwareButton";
import RollbackDialog from "./RollbackDialog";

export interface IRollbackButtonProps {
  installedPackageRef: InstalledPackageReference;
  revision: number;
  releaseStatus: InstalledPackageStatus | undefined | null;
  disabled?: boolean;
}

function RollbackButton({
  installedPackageRef,
  revision,
  releaseStatus,
  disabled,
}: IRollbackButtonProps) {
  const [modalIsOpen, setModalOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();
  const error = useSelector((state: IStoreState) => state.apps.error);
  const openModal = () => setModalOpen(true);
  const closeModal = () => setModalOpen(false);
  const handleRollback = async (r: number) => {
    setLoading(true);
    const success = await dispatch(
      actions.installedpackages.rollbackInstalledPackage(installedPackageRef, r),
    );
    setLoading(false);
    if (success) {
      closeModal();
    }
  };
  const button = (
    <StatusAwareButton
      status="primary"
      onClick={openModal}
      releaseStatus={releaseStatus}
      id="rollback-button"
      disabled={disabled}
    >
      <CdsIcon shape="rewind" /> Rollback
    </StatusAwareButton>
  );
  return disabled ? (
    <>{button}</>
  ) : (
    <>
      <RollbackDialog
        modalIsOpen={modalIsOpen}
        onConfirm={handleRollback}
        loading={loading}
        closeModal={closeModal}
        currentRevision={revision}
        error={error}
      />
      {button}
    </>
  );
}

export default RollbackButton;
