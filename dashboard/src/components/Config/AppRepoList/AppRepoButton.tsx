import { CdsButton } from "@cds/react/button";
import { CdsIcon } from "@cds/react/icon";
import { CdsModal, CdsModalContent, CdsModalHeader } from "@cds/react/modal";
import actions from "actions";
import { useState } from "react";
import { useDispatch } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { IAppRepository, IAppRepositoryFilter, IStoreState } from "shared/types";
import { AppRepoForm } from "./AppRepoForm";

interface IAppRepoAddButtonProps {
  namespace: string;
  kubeappsNamespace: string;
  text?: string;
  primary?: boolean;
  repo?: IAppRepository;
  disabled?: boolean;
  title?: string;
}

export function AppRepoAddButton({
  text,
  namespace,
  kubeappsNamespace,
  repo,
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
    description: string,
    authHeader: string,
    authRegCreds: string,
    customCA: string,
    syncJobPodTemplate: string,
    registrySecrets: string[],
    ociRepositories: string[],
    skipTLS: boolean,
    passCredentials: boolean,
    filter?: IAppRepositoryFilter,
  ) => {
    if (repo) {
      return dispatch(
        actions.repos.updateRepo(
          name,
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
          filter,
        ),
      );
    } else {
      return dispatch(
        actions.repos.installRepo(
          name,
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
        {primary ? <CdsIcon shape="plus-circle" /> : <></>} {text || "Add App Repository"}
      </CdsButton>
      {modalIsOpen && (
        <CdsModal size={"lg"} onCloseChange={closeModal}>
          <CdsModalHeader>{title}</CdsModalHeader>
          <CdsModalContent>
            <AppRepoForm
              onSubmit={onSubmit}
              onAfterInstall={closeModal}
              repo={repo}
              namespace={namespace}
              kubeappsNamespace={kubeappsNamespace}
            />
          </CdsModalContent>
        </CdsModal>
      )}
    </>
  );
}
