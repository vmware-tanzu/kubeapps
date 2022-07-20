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

interface IStartButtonProps {
  installedPackageRef: InstalledPackageReference;
  releaseStatus: InstalledPackageStatus | undefined | null;
  disabled?: boolean;
}

export default function StartButton({
  installedPackageRef,
  releaseStatus,
  disabled,
}: IStartButtonProps) {
  const [modalIsOpen, setModal] = useState(false);
  const [starting, setStarting] = useState(false);
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();
  const error = useSelector((state: IStoreState) => state.apps.error);

  const openModal = () => setModal(true);
  const closeModal = () => setModal(false);
  const handleStartClick = async () => {
    setStarting(true);
    const started = await dispatch(
      actions.installedpackages.startInstalledPackage(installedPackageRef),
    );
    setStarting(false);
    if (started) {
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
        id="start-button"
        disabled={disabled}
      >
        <CdsIcon shape="fast-forward" /> Start
      </StatusAwareButton>
      <ConfirmDialog
        modalIsOpen={modalIsOpen}
        loading={starting}
        onConfirm={handleStartClick}
        closeModal={closeModal}
        headerText={"Start application"}
        confirmationText="Are you sure you want to start the application?"
        confirmationButtonText="Start"
        error={error}
      />
    </>
  );
}
