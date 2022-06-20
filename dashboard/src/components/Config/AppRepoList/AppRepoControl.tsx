// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import { useState } from "react";
import { useDispatch } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { IAppRepository, IStoreState } from "shared/types";
import actions from "../../../actions";
import ConfirmDialog from "../../ConfirmDialog/ConfirmDialog";
import { AppRepoAddButton } from "./AppRepoButton";
import "./AppRepoControl.css";

interface IAppRepoListItemProps {
  repo: IAppRepository;
  kubeappsNamespace: string;
  refetchRepos: () => void;
}

export function AppRepoControl({ repo, kubeappsNamespace, refetchRepos }: IAppRepoListItemProps) {
  const [modalIsOpen, setModalOpen] = useState(false);
  const openModal = () => setModalOpen(true);
  const closeModal = () => setModalOpen(false);
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();

  const handleDeleteClick = (repoName: string, repoNamespace: string) => {
    return async () => {
      await dispatch(actions.repos.deleteRepo(repoName, repoNamespace));
      refetchRepos();
      closeModal();
    };
  };

  return (
    <div className="apprepo-control-buttons">
      <ConfirmDialog
        onConfirm={handleDeleteClick(repo.metadata.name, repo.metadata.namespace)}
        modalIsOpen={modalIsOpen}
        loading={false}
        closeModal={closeModal}
        headerText={"Delete repository"}
        confirmationText={`Are you sure you want to delete the repository ${repo.metadata.name}?`}
      />

      <AppRepoAddButton
        title={`Edit the '${repo.metadata.name}' Package Repository`}
        namespace={repo.metadata.namespace}
        kubeappsNamespace={kubeappsNamespace}
        text="Edit"
        packageRepoRef={repo}
        primary={false}
      />
      <CdsButton
        id={`delete-repo-${repo.metadata.name}`}
        status="danger"
        onClick={openModal}
        action="outline"
      >
        Delete
      </CdsButton>
    </div>
  );
}
