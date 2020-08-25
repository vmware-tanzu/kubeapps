import React, { useState } from "react";

import { CdsButton } from "components/Clarity/clarity";
import { useDispatch } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { IAppRepository, ISecret, IStoreState } from "shared/types";
import actions from "../../../actions";
import ConfirmDialog from "../../ConfirmDialog/ConfirmDialog.v2";
import { AppRepoAddButton } from "./AppRepoButton.v2";
import "./AppRepoControl.css";

interface IAppRepoListItemProps {
  repo: IAppRepository;
  namespace: string;
  kubeappsNamespace: string;
  secret?: ISecret;
}

export function AppRepoControl({
  namespace,
  repo,
  secret,
  kubeappsNamespace,
}: IAppRepoListItemProps) {
  const [modalIsOpen, setModalOpen] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const openModal = () => setModalOpen(true);
  const closeModal = () => setModalOpen(false);
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();

  const handleDeleteClick = (repoName: string, repoNamespace: string) => {
    return async () => {
      await dispatch(actions.repos.deleteRepo(repoName, repoNamespace));
      if (repoNamespace !== kubeappsNamespace) {
        // Re-fetch repos in both namespaces because otherwise, the state
        // will be updated only with the repos of repoNamespace and removing
        // the global ones.
        // TODO(andresmgot): This can be refactored once hex UI is dropped
        dispatch(actions.repos.fetchRepos(repoNamespace, kubeappsNamespace));
      } else {
        dispatch(actions.repos.fetchRepos(repoNamespace));
      }
      closeModal();
    };
  };

  const handleResyncClick = (repoName: string, repoNamespace: string) => {
    return () => {
      setRefreshing(true);
      dispatch(actions.repos.resyncRepo(repoName, repoNamespace));
      // Fake timeout to show progress
      // TODO(andresmgot): Ideally, we should show the progress of the sync but we don't
      // have that info yet: https://github.com/kubeapps/kubeapps/issues/153
      setTimeout(() => setRefreshing(false), 500);
    };
  };

  return (
    <div className="apprepo-control-buttons">
      <ConfirmDialog
        onConfirm={handleDeleteClick(repo.metadata.name, repo.metadata.namespace)}
        modalIsOpen={modalIsOpen}
        loading={false}
        closeModal={closeModal}
        confirmationText={`Are you sure you want to delete the repository ${repo.metadata.name}?`}
      />

      <AppRepoAddButton
        namespace={namespace}
        kubeappsNamespace={kubeappsNamespace}
        text="Edit"
        repo={repo}
        secret={secret}
        primary={false}
      />

      <CdsButton
        onClick={handleResyncClick(repo.metadata.name, repo.metadata.namespace)}
        action="outline"
        disabled={refreshing}
      >
        {refreshing ? "Refreshing" : "Refresh"}
      </CdsButton>
      <CdsButton status="danger" onClick={openModal} action="outline">
        Delete
      </CdsButton>
    </div>
  );
}
