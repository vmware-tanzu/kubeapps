// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import { CdsIcon } from "@cds/react/icon";
import DifferentialTab from "components/DeploymentFormBody/DifferentialTab";
import Alert from "components/js/Alert";
import Tabs from "components/Tabs";
import * as yaml from "js-yaml";
import { useEffect, useState } from "react";
import { IResource } from "shared/types";
import ConfirmDialog from "../ConfirmDialog/ConfirmDialog";
import AdvancedDeploymentForm from "../DeploymentFormBody/AdvancedDeploymentForm";
import Differential from "../DeploymentFormBody/Differential";
import LoadingWrapper from "../LoadingWrapper";

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
      resource = yaml.load(values);
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
        {parseError && <Alert theme="danger">{parseError.message}</Alert>}
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
            columns={[
              "YAML",
              <DifferentialTab
                key="differential-selector"
                deploymentEvent={deployedValues ? "upgrade" : "install"}
                defaultValues={defaultValues}
                deployedValues={deployedValues || ""}
                appValues={values}
              />,
            ]}
            data={[
              <AdvancedDeploymentForm
                appValues={values}
                handleValuesChange={handleValuesChange}
                key="advanced-deployment-form"
              />,
              <Differential
                oldValues={deployedValues || defaultValues}
                newValues={values}
                emptyDiffElement={
                  deployedValues ? (
                    <span>No changes detected from deployed values</span>
                  ) : (
                    <span>No changes detected from example defaults</span>
                  )
                }
                key="differential-selector"
              />,
            ]}
          />
        </div>
        <div className="deployment-form-control-buttons">
          <CdsButton status="primary" type="submit">
            <CdsIcon shape="deploy" /> Deploy
          </CdsButton>
          {/* TODO(andresmgot): CdsButton "type" property doesn't work, so we need to use a normal <button>
            https://github.com/vmware/clarity/issues/5038
          */}
          <span className="color-icon-info">
            <button
              className="btn btn-info-outline"
              type="button"
              onClick={openModal}
              style={{ marginTop: "-22px" }}
            >
              <CdsIcon shape="backup-restore" /> Restore Defaults
            </button>
          </span>
        </div>
      </form>
    </>
  );
}

export default DeploymentFormBody;
