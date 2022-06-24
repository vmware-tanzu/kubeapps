// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import { CdsIcon } from "@cds/react/icon";
import { CdsModal, CdsModalContent, CdsModalHeader } from "@cds/react/modal";
import actions from "actions";
import { useState } from "react";
import { useDispatch } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { IAppRepository, IPkgRepoFormData, IStoreState } from "shared/types";
import { AppRepoForm } from "./AppRepoForm";

interface IAppRepoAddButtonProps {
  namespace: string;
  kubeappsNamespace: string;
  text?: string;
  primary?: boolean;
  packageRepoRef?: IAppRepository;
  disabled?: boolean;
  title?: string;
}

export function AppRepoAddButton({
  text,
  namespace,
  kubeappsNamespace,
  packageRepoRef,
  primary = true,
  title,
  disabled,
}: IAppRepoAddButtonProps) {
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();
  const [modalIsOpen, setModalOpen] = useState(false);
  const openModal = () => {
    dispatch(actions.repos.requestRepo());
    setModalOpen(true);
  };
  const closeModal = () => setModalOpen(false);
  const onSubmit = (request: IPkgRepoFormData) => {
    // decide whether to create or update the repository
    if (packageRepoRef) {
      return dispatch(actions.repos.updateRepo(namespace, request));
    } else {
      return dispatch(actions.repos.installRepo(namespace, request));
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
        {primary ? <CdsIcon shape="plus-circle" /> : <></>} {text || "Add Package Repository"}
      </CdsButton>
      {modalIsOpen && (
        <CdsModal size={"lg"} onCloseChange={closeModal}>
          <CdsModalHeader>{title}</CdsModalHeader>
          <CdsModalContent>
            <AppRepoForm
              onSubmit={onSubmit}
              onAfterInstall={closeModal}
              packageRepoRef={packageRepoRef}
              namespace={namespace}
              kubeappsNamespace={kubeappsNamespace}
            />
          </CdsModalContent>
        </CdsModal>
      )}
    </>
  );
}
