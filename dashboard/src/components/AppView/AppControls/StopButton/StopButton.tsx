// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsIcon } from "@cds/react/icon";
import actions from "actions";
import ConfirmDialog from "components/ConfirmDialog/ConfirmDialog";
import { push } from "connected-react-router";
import {
  InstalledPackageReference,
  InstalledPackageStatus,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { IStoreState } from "shared/types";
import { app } from "shared/url";
import StatusAwareButton from "../StatusAwareButton/StatusAwareButton";

interface IStopButtonProps {
  installedPackageRef: InstalledPackageReference;
  releaseStatus: InstalledPackageStatus | undefined | null;
  disabled?: boolean;
}

export default function StopButton({
  installedPackageRef,
  releaseStatus,
  disabled,
}: IStopButtonProps) {
  const [modalIsOpen, setModal] = useState(false);
  const [stopping, setStopping] = useState(false);
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();
  const error = useSelector((state: IStoreState) => state.apps.error);

  const openModal = () => setModal(true);
  const closeModal = () => setModal(false);
  const handleStopClick = async () => {
    setStopping(true);
    const stopped = await dispatch(
      actions.installedpackages.stopInstalledPackage(installedPackageRef),
    );
    setStopping(false);
    if (stopped) {
      dispatch(
        push(
          app.apps.list(
            installedPackageRef.context?.cluster,
            installedPackageRef.context?.namespace,
          ),
        ),
      );
    }
  };

  return (
    <>
      <StatusAwareButton
        status="danger"
        onClick={openModal}
        releaseStatus={releaseStatus}
        id="stop-button"
        disabled={disabled}
      >
        <CdsIcon shape="stop" /> Stop
      </StatusAwareButton>
      <ConfirmDialog
        modalIsOpen={modalIsOpen}
        loading={stopping}
        onConfirm={handleStopClick}
        closeModal={closeModal}
        headerText={"Stop application"}
        confirmationText="Are you sure you want to stop the application?"
        error={error}
      />
    </>
  );
}
