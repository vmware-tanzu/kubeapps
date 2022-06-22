// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import {
  PackageRepositoryReference,
  PackageRepositorySummary,
} from "gen/kubeappsapis/core/packages/v1alpha1/repositories";
import { useState } from "react";
import { useDispatch } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { IStoreState } from "shared/types";
import actions from "../../../actions";
import ConfirmDialog from "../../ConfirmDialog/ConfirmDialog";
import { AppRepoAddButton } from "./AppRepoButton";
import "./AppRepoControl.css";

interface IAppRepoListItemProps {
  repo: PackageRepositorySummary;
  kubeappsNamespace: string;
  refetchRepos: () => void;
}

export function AppRepoControl({ repo, kubeappsNamespace, refetchRepos }: IAppRepoListItemProps) {
  const [modalIsOpen, setModalOpen] = useState(false);
  const openModal = () => setModalOpen(true);
  const closeModal = () => setModalOpen(false);
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();

  const handleDeleteClick = (packageRepoRef: PackageRepositoryReference) => {
    return async () => {
      await dispatch(actions.repos.deleteRepo(packageRepoRef));
      refetchRepos();
      closeModal();
    };
  };

  return (
    <div className="apprepo-control-buttons">
      <ConfirmDialog
        onConfirm={handleDeleteClick(repo.packageRepoRef!)}
        modalIsOpen={modalIsOpen}
        loading={false}
        closeModal={closeModal}
        headerText={"Delete repository"}
        confirmationText={`Are you sure you want to delete the repository ${repo.name}?`}
      />

      <AppRepoAddButton
        title={`Edit the '${repo.name}' Package Repository`}
        namespace={repo.packageRepoRef?.context?.namespace || ""}
        kubeappsNamespace={kubeappsNamespace}
        text="Edit"
        packageRepoRef={repo.packageRepoRef}
        primary={false}
      />
      <CdsButton
        id={`delete-repo-${repo.name}`}
        status="danger"
        onClick={openModal}
        action="outline"
      >
        Delete
      </CdsButton>
    </div>
  );
}
