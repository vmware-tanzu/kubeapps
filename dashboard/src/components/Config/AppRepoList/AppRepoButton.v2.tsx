import React, { useState } from "react";

import actions from "actions";
import { CdsButton, CdsIcon } from "components/Clarity/clarity";
import Modal from "components/js/Modal/Modal";
import { useDispatch } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { IAppRepository, ISecret, IStoreState } from "../../../shared/types";
import "./AppRepoButton.v2.css";
import { AppRepoForm } from "./AppRepoForm.v2";

interface IAppRepoAddButtonProps {
  namespace: string;
  kubeappsNamespace: string;
  text?: string;
  primary?: boolean;
  repo?: IAppRepository;
  secret?: ISecret;
}

export function AppRepoAddButton({
  text,
  namespace,
  kubeappsNamespace,
  repo,
  secret,
  primary = true,
}: IAppRepoAddButtonProps) {
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();
  const [modalIsOpen, setModalOpen] = useState(false);
  const openModal = () => setModalOpen(true);
  const closeModal = () => setModalOpen(false);
  const onSubmit = (
    name: string,
    url: string,
    authHeader: string,
    customCA: string,
    syncJobPodTemplate: string,
    registrySecrets: string[],
  ) => {
    if (repo) {
      return dispatch(
        actions.repos.updateRepo(
          name,
          namespace,
          url,
          authHeader,
          customCA,
          syncJobPodTemplate,
          registrySecrets,
        ),
      );
    } else {
      return dispatch(
        actions.repos.installRepo(
          name,
          namespace,
          url,
          authHeader,
          customCA,
          syncJobPodTemplate,
          registrySecrets,
        ),
      );
    }
  };

  return (
    <>
      <CdsButton onClick={openModal} action={primary ? "solid" : "outline"}>
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
