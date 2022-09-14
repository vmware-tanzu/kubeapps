// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import { CdsIcon } from "@cds/react/icon";
import { CdsModal, CdsModalContent, CdsModalHeader } from "@cds/react/modal";
import actions from "actions";
import { PackageRepositoryReference } from "gen/kubeappsapis/core/packages/v1alpha1/repositories";
import { useState } from "react";
import { useDispatch } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { IPkgRepoFormData, IStoreState } from "shared/types";
import { PkgRepoForm } from "./PkgRepoForm";

export interface IPkgRepoAddButtonProps {
  namespace: string;
  helmGlobalNamespace: string;
  carvelGlobalNamespace: string;
  text?: string;
  primary?: boolean;
  packageRepoRef?: PackageRepositoryReference;
  disabled?: boolean;
  title?: string;
}

export function PkgRepoAddButton({
  text,
  namespace,
  helmGlobalNamespace,
  carvelGlobalNamespace,
  packageRepoRef,
  primary = true,
  title,
  disabled,
}: IPkgRepoAddButtonProps) {
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();
  const [modalIsOpen, setModalOpen] = useState(false);
  const openModal = () => {
    dispatch(actions.repos.requestRepoDetail());
    setModalOpen(true);
  };
  const closeModal = () => setModalOpen(false);
  const onSubmit = (request: IPkgRepoFormData) => {
    // decide whether to create or update the repository
    if (packageRepoRef) {
      return dispatch(actions.repos.updateRepo(request));
    } else {
      return dispatch(actions.repos.addRepo(request));
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
            <PkgRepoForm
              onSubmit={onSubmit}
              onAfterInstall={closeModal}
              packageRepoRef={packageRepoRef}
              namespace={namespace}
              helmGlobalNamespace={helmGlobalNamespace}
              carvelGlobalNamespace={carvelGlobalNamespace}
            />
          </CdsModalContent>
        </CdsModal>
      )}
    </>
  );
}
