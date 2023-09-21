// Copyright 2020-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import { CdsIcon } from "@cds/react/icon";
import AlertGroup from "components/AlertGroup";
import ConfirmDialog from "components/ConfirmDialog";
import LoadingWrapper from "components/LoadingWrapper";
import Tabs from "components/Tabs";
import { useEffect, useState } from "react";
import { IResource } from "shared/types";
import { parseToJS } from "shared/yamlUtils";
import OperatorAdvancedDeploymentForm from "./OperatorAdvancedDeploymentForm/OperatorAdvancedDeploymentForm";

export interface IOperatorInstanceFormProps {
  isFetching: boolean;
  handleDeploy: (resource: IResource) => void;
  defaultValues: string;
  deployedValues?: string;
}

export interface IOperatorInstanceFormBodyState {
  values: string;
  restoreDefaultValuesModalIsOpen: boolean;
  submittedResourceName: string;
  error?: Error;
}

function DeploymentFormBody({
  defaultValues,
  isFetching,
  handleDeploy,
  deployedValues,
}: IOperatorInstanceFormProps) {
  const [values, setValues] = useState(defaultValues);
  const [modalIsOpen, setModalIsOpen] = useState(false);
  const [parseError, setParseError] = useState(undefined as Error | undefined);
  const closeModal = () => setModalIsOpen(false);
  const openModal = () => setModalIsOpen(true);

  const handleValuesChange = (newValues: string) => {
    setValues(newValues);
  };

  useEffect(() => {
    setValues(deployedValues || defaultValues);
  }, [defaultValues, deployedValues]);

  const restoreDefaultValues = () => {
    setValues(defaultValues);
    closeModal();
  };

  const parseAndDeploy = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    // Clean possible previous errors
    setParseError(undefined);
    let resource: any = {};
    try {
      resource = parseToJS(values);
    } catch (e: any) {
      setParseError(new Error(`Unable to parse the given YAML. Got: ${e.message}`));
      return;
    }
    if (!resource.apiVersion) {
      setParseError(
        new Error("Unable parse the resource. Make sure it contains a valid apiVersion"),
      );
      return;
    }
    handleDeploy(resource);
  };

  if (isFetching) {
    return (
      <LoadingWrapper
        className="margin-t-xxl"
        loadingText="Fetching application..."
        loaded={false}
      />
    );
  }
  return (
    <>
      <form onSubmit={parseAndDeploy}>
        {parseError && <AlertGroup status="danger">{parseError.message}.</AlertGroup>}
        <ConfirmDialog
          modalIsOpen={modalIsOpen}
          loading={false}
          headerText={"Restore defaults"}
          confirmationText={"Are you sure you want to restore the default instance values?"}
          confirmationButtonText={"Restore"}
          onConfirm={restoreDefaultValues}
          closeModal={closeModal}
        />
        <div className="deployment-form-tabs">
          <Tabs
            id="deployment-form-body-tabs"
            columns={[["YAML editor", () => {}]]}
            data={[
              <OperatorAdvancedDeploymentForm
                appValues={values}
                oldAppValues={deployedValues || defaultValues}
                handleValuesChange={handleValuesChange}
                key="advanced-deployment-form"
              />,
            ]}
          />
        </div>
        <div className="deployment-form-control-buttons">
          <CdsButton status="primary" type="submit">
            <CdsIcon shape="deploy" /> Deploy
          </CdsButton>
          <CdsButton type="button" status="primary" action="outline" onClick={openModal}>
            <CdsIcon shape="backup-restore" /> Restore Defaults
          </CdsButton>
        </div>
      </form>
    </>
  );
}

export default DeploymentFormBody;
