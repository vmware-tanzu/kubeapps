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

interface IDeleteButtonProps {
  installedPackageRef: InstalledPackageReference;
  releaseStatus: InstalledPackageStatus | undefined | null;
  disabled?: boolean;
}

export default function DeleteButton({
  installedPackageRef,
  releaseStatus,
  disabled,
}: IDeleteButtonProps) {
  const [modalIsOpen, setModal] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();
  const error = useSelector((state: IStoreState) => state.apps.error);

  const openModal = () => setModal(true);
  const closeModal = () => setModal(false);
  const handleDeleteClick = async () => {
    setDeleting(true);
    const deleted = await dispatch(
      actions.installedpackages.deleteInstalledPackage(installedPackageRef),
    );
    setDeleting(false);
    if (deleted) {
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
        id="delete-button"
        disabled={disabled}
      >
        <CdsIcon shape="trash" /> Delete
      </StatusAwareButton>
      <ConfirmDialog
        modalIsOpen={modalIsOpen}
        loading={deleting}
        onConfirm={handleDeleteClick}
        closeModal={closeModal}
        headerText={"Delete application"}
        confirmationText="Are you sure you want to delete the application?"
        error={error}
      />
    </>
  );
}
