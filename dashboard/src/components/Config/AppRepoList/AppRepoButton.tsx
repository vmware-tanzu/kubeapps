import React, { useState } from "react";

import { CdsButton } from "@clr/react/button";
import { CdsIcon } from "@clr/react/icon";
import actions from "actions";
import Modal from "components/js/Modal/Modal";
import { useDispatch } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { IAppRepository, ISecret, IStoreState } from "../../../shared/types";
import "./AppRepoButton.css";
import { AppRepoForm } from "./AppRepoForm";

interface IAppRepoAddButtonProps {
  namespace: string;
  kubeappsNamespace: string;
  text?: string;
  primary?: boolean;
  repo?: IAppRepository;
  secret?: ISecret;
  disabled?: boolean;
  title?: string;
}

export function AppRepoAddButton({
  text,
  namespace,
  kubeappsNamespace,
  repo,
  secret,
  primary = true,
  title,
  disabled,
}: IAppRepoAddButtonProps) {
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();
  const [modalIsOpen, setModalOpen] = useState(false);
  const openModal = () => setModalOpen(true);
  const closeModal = () => setModalOpen(false);
  const onSubmit = (
    name: string,
    url: string,
    type: string,
    authHeader: string,
    customCA: string,
    syncJobPodTemplate: string,
    registrySecrets: string[],
    ociRepositories: string[],
  ) => {
    if (repo) {
      return dispatch(
        actions.repos.updateRepo(
          name,
          namespace,
          url,
          type,
          authHeader,
          customCA,
          syncJobPodTemplate,
          registrySecrets,
          ociRepositories,
        ),
      );
    } else {
      return dispatch(
        actions.repos.installRepo(
          name,
          namespace,
          url,
          type,
          authHeader,
          customCA,
          syncJobPodTemplate,
          registrySecrets,
          ociRepositories,
        ),
      );
    }
  };

  return (
    <>
      <CdsButton
        onClick={openModal}
        action={primary ? "solid" : "outline"}
        disabled={disabled}
        title={title}
      >
        {primary ? <CdsIcon shape="plus-circle" inverse={true} /> : <></>}{" "}
        {text || "Add App Repository"}
      </CdsButton>
      <Modal
        staticBackdrop={false}
        showModal={modalIsOpen}
        onModalClose={closeModal}
        modalSize="lg"
      >
        <div className="modal-close" onClick={closeModal}>
          <CdsIcon shape="times-circle" size="md" solid={true} />
        </div>
        <AppRepoForm
          onSubmit={onSubmit}
          onAfterInstall={closeModal}
          repo={repo}
          secret={secret}
          namespace={namespace}
          kubeappsNamespace={kubeappsNamespace}
        />
      </Modal>
    </>
  );
}
