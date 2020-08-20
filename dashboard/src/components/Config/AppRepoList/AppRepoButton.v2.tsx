import React, { useState } from "react";

import actions from "actions";
import { CdsButton, CdsIcon } from "components/Clarity/clarity";
import Alert from "components/js/Alert";
import Modal from "components/js/Modal/Modal";
import { useDispatch, useSelector } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { IAppRepository, ISecret, IStoreState } from "../../../shared/types";
import "./AppRepo.css";
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
  const { errors } = useSelector((state: IStoreState) => state.repos);

  const onSubmit = (
    name: string,
    url: string,
    authHeader: string,
    customCA: string,
    syncJobPodTemplate: string,
    registrySecrets: string[],
  ) =>
    dispatch(
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

  return (
    <>
      <CdsButton onClick={openModal} action={primary ? "solid" : "outline"}>
        {primary ? <CdsIcon shape="plus-circle" inverse={true} /> : <></>}{" "}
        {text || "Add App Repository"}
      </CdsButton>
      <Modal showModal={modalIsOpen} onModalClose={closeModal}>
        {errors.create && (
          <Alert theme="danger">
            Found an error creating the repository: {errors.create.message}
          </Alert>
        )}
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
