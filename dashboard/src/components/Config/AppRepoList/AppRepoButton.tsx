// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import { CdsIcon } from "@cds/react/icon";
import { CdsModal, CdsModalContent, CdsModalHeader } from "@cds/react/modal";
import actions from "actions";
import {
  PackageRepositoryAuth_PackageRepositoryAuthType,
  PackageRepositoryReference,
} from "gen/kubeappsapis/core/packages/v1alpha1/repositories";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import { useState } from "react";
import { useDispatch } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { IAppRepositoryFilter, IStoreState } from "shared/types";
import { AppRepoForm } from "./AppRepoForm";

interface IAppRepoAddButtonProps {
  namespace: string;
  kubeappsNamespace: string;
  text?: string;
  primary?: boolean;
  packageRepoRef?: PackageRepositoryReference;
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
  const onSubmit = (
    name: string,
    plugin: Plugin,
    url: string,
    type: string,
    description: string,
    authHeader: string,
    authRegCreds: string,
    customCA: string,
    syncJobPodTemplate: string,
    registrySecrets: string[],
    ociRepositories: string[],
    skipTLS: boolean,
    passCredentials: boolean,
    authMethod: PackageRepositoryAuth_PackageRepositoryAuthType,
    filter?: IAppRepositoryFilter,
  ) => {
    if (packageRepoRef) {
      return dispatch(
        actions.repos.updateRepo(
          name,
          plugin,
          namespace,
          url,
          type,
          description,
          authHeader,
          authRegCreds,
          customCA,
          syncJobPodTemplate,
          registrySecrets,
          ociRepositories,
          skipTLS,
          passCredentials,
          authMethod,
          filter,
        ),
      );
    } else {
      return dispatch(
        actions.repos.installRepo(
          name,
          plugin,
          namespace,
          url,
          type,
          description,
          authHeader,
          authRegCreds,
          customCA,
          syncJobPodTemplate,
          registrySecrets,
          ociRepositories,
          skipTLS,
          passCredentials,
          authMethod,
          filter,
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
