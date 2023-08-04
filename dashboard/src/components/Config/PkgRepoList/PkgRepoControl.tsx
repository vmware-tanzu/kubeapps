// Copyright 2020-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import actions from "actions";
import ConfirmDialog from "components/ConfirmDialog";
import {
  PackageRepositoryReference,
  PackageRepositorySummary,
} from "gen/kubeappsapis/core/packages/v1alpha1/repositories_pb";
import { useState } from "react";
import { useDispatch } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { IStoreState } from "shared/types";
import { PkgRepoAddButton } from "./PkgRepoButton";
import "./PkgRepoControl.css";

export interface IPkgRepoListItemProps {
  repo: PackageRepositorySummary;
  helmGlobalNamespace: string;
  carvelGlobalNamespace: string;
  refetchRepos: () => void;
}

export function PkgRepoControl({
  repo,
  helmGlobalNamespace,
  carvelGlobalNamespace,
  refetchRepos,
}: IPkgRepoListItemProps) {
  const [modalIsOpen, setModalOpen] = useState(false);
  const openModal = () => setModalOpen(true);
  const closeModal = () => setModalOpen(false);
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();

  const handleDeleteClick = (packageRepoRef: PackageRepositoryReference) => {
    return async () => {
      const deleted = await dispatch(actions.repos.deleteRepo(packageRepoRef));
      if (deleted) {
        refetchRepos();
      }
      closeModal();
    };
  };

  return (
    <div className="pkgrepo-control-buttons">
      <ConfirmDialog
        onConfirm={handleDeleteClick(repo.packageRepoRef!)}
        modalIsOpen={modalIsOpen}
        loading={false}
        closeModal={closeModal}
        headerText={"Delete repository"}
        confirmationText={`Are you sure you want to delete the repository ${repo.name}?`}
      />

      <PkgRepoAddButton
        title={`Edit the '${repo.name}' Package Repository`}
        namespace={repo.packageRepoRef?.context?.namespace || ""}
        helmGlobalNamespace={helmGlobalNamespace}
        carvelGlobalNamespace={carvelGlobalNamespace}
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
