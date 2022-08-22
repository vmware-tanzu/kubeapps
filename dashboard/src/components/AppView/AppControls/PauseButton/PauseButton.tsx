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

interface IPauseButtonProps {
  installedPackageRef: InstalledPackageReference;
  releaseStatus: InstalledPackageStatus | undefined | null;
  disabled?: boolean;
}

export default function PauseButton({
  installedPackageRef,
  releaseStatus,
  disabled,
}: IPauseButtonProps) {
  const [modalIsOpen, setModal] = useState(false);
  const [pausing, setPausing] = useState(false);
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();
  const error = useSelector((state: IStoreState) => state.apps.error);

  const openModal = () => setModal(true);
  const closeModal = () => setModal(false);
  const handlePauseClick = async () => {
    setPausing(true);
    const paused = await dispatch(
      actions.installedpackages.pauseInstalledPackage(installedPackageRef),
    );
    setPausing(false);
    if (paused) {
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
        status="primary"
        onClick={openModal}
        releaseStatus={releaseStatus}
        id="pause-button"
        disabled={disabled}
      >
        <CdsIcon shape="stop" /> Pause
      </StatusAwareButton>
      <ConfirmDialog
        modalIsOpen={modalIsOpen}
        loading={pausing}
        onConfirm={handlePauseClick}
        closeModal={closeModal}
        headerText={"Pause application"}
        confirmationText="Are you sure you want to pause the application?"
        confirmationButtonText="Pause"
        error={error}
      />
    </>
  );
}
