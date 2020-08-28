import { CdsButton, CdsIcon } from "components/Clarity/clarity";
import Alert from "components/js/Alert";
import Tabs from "components/Tabs";
import * as yaml from "js-yaml";
import React, { useEffect, useState } from "react";
import { definedNamespaces } from "../../shared/Namespace";
import { IResource } from "../../shared/types";
import ConfirmDialog from "../ConfirmDialog/ConfirmDialog.v2";
import AdvancedDeploymentForm from "../DeploymentFormBody/AdvancedDeploymentForm.v2";
import Differential from "../DeploymentFormBody/Differential";
import LoadingWrapper from "../LoadingWrapper";

export interface IOperatorInstanceFormProps {
  isFetching: boolean;
  namespace: string;
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
  namespace,
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
    let resource: IResource = {} as any;
    try {
      resource = yaml.safeLoad(values);
    } catch (e) {
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

  if (namespace === definedNamespaces.all) {
    return <Alert theme="danger">Select a namespace before creating a new instance.</Alert>;
  }
  if (isFetching) {
    return <LoadingWrapper />;
  }
  return (
    <>
      <form onSubmit={parseAndDeploy}>
        {parseError && <Alert theme="danger">{parseError.message}</Alert>}
        <ConfirmDialog
          modalIsOpen={modalIsOpen}
          loading={false}
          confirmationText={"Are you sure you want to restore the default instance values?"}
          confirmationButtonText={"Restore"}
          onConfirm={restoreDefaultValues}
          closeModal={closeModal}
        />
        <div className="deployment-form-tabs">
          <Tabs
            id="deployment-form-body-tabs"
            columns={["YAML", "Changes"]}
            data={[
              <AdvancedDeploymentForm
                appValues={values}
                handleValuesChange={handleValuesChange}
                key="advanced-deployment-form"
              />,
              <Differential
                title={
                  deployedValues
                    ? "Difference from deployed values"
                    : "Difference from example defaults"
                }
                oldValues={deployedValues || defaultValues}
                newValues={values}
                emptyDiffText={
                  deployedValues
                    ? "No changes detected from deployed values"
                    : "No changes detected from example defaults"
                }
                key="differential-selector"
              />,
            ]}
          />
        </div>
        <div className="deployment-form-control-buttons">
          <CdsButton status="primary" type="submit">
            <CdsIcon shape="deploy" inverse={true} /> Deploy
          </CdsButton>
          <CdsButton action="outline" type="button" onClick={openModal}>
            <CdsIcon shape="backup-restore" inverse={true} /> Restore Defaults
          </CdsButton>
        </div>
      </form>
    </>
  );
}

export default DeploymentFormBody;
