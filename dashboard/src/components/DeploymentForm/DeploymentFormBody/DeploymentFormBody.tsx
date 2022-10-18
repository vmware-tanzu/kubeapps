// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import { CdsControlMessage } from "@cds/react/forms";
import { CdsIcon } from "@cds/react/icon";
import ConfirmDialog from "components/ConfirmDialog";
import Alert from "components/js/Alert";
import LoadingWrapper from "components/LoadingWrapper";
import Tabs from "components/Tabs";
import { isEmpty } from "lodash";
import { FormEvent, RefObject, useCallback, useEffect, useState } from "react";
import { retrieveBasicFormParams, updateCurrentConfigByKey } from "shared/schema";
import { DeploymentEvent, IBasicFormParam, IPackageState } from "shared/types";
import { getValueFromEvent } from "shared/utils";
import { parseToYamlNode, setPathValueInYamlNode, toStringYamlNode } from "shared/yamlUtils";
import YAML from "yaml";
import AdvancedDeploymentForm from "./AdvancedDeploymentForm";
import BasicDeploymentForm from "./BasicDeploymentForm/BasicDeploymentForm";

export interface IDeploymentFormBodyProps {
  deploymentEvent: DeploymentEvent;
  packageId: string;
  packageVersion: string;
  deployedValues?: string;
  packagesIsFetching: boolean;
  selected: IPackageState["selected"];
  appValues: string;
  setValues: (values: string) => void;
  setValuesModified: () => void;
  formRef: RefObject<HTMLFormElement>;
}

function DeploymentFormBody({
  deploymentEvent,
  packageId,
  packageVersion,
  deployedValues: valuesFromTheDeployedPackage,
  packagesIsFetching,
  selected,
  appValues: valuesFromTheParentContainer,
  setValues: setValuesFromTheParentContainer,
  setValuesModified,
  formRef,
}: IDeploymentFormBodyProps) {
  const {
    availablePackageDetail,
    versions,
    schema: schemaFromTheAvailablePackage,
    values: valuesFromTheAvailablePackage,
    pkgVersion,
    error,
  } = selected;

  // Component state
  const [paramsFromComponentState, setParamsFromComponentState] = useState([] as IBasicFormParam[]);
  const [valuesFromTheAvailablePackageNodes, setValuesFromTheAvailablePackageNodes] = useState(
    {} as YAML.Document.Parsed<YAML.ParsedNode>,
  );
  const [valuesFromTheDeployedPackageNodes, setValuesFromTheDeployedPackageNodes] = useState(
    {} as YAML.Document.Parsed<YAML.ParsedNode>,
  );
  const [valuesFromTheParentContainerNodes, setValuesFromTheParentContainerNodes] = useState(
    {} as YAML.Document.Parsed<YAML.ParsedNode>,
  );
  const [restoreModalIsOpen, setRestoreModalOpen] = useState(false);
  const [isLoaded, setIsLoaded] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [unsavedChangesMap] = useState(new Map<string, any>());
  const [shouldSubmitForm, setShouldSubmitForm] = useState(false);

  // whenever the parsed values change (for instance, when a new pkg version is selected),
  // we need to force a new extraction of the params from the schema
  useEffect(() => {
    if (!isLoaded && schemaFromTheAvailablePackage && !isEmpty(valuesFromTheParentContainerNodes)) {
      const initialParamsFromContainer = retrieveBasicFormParams(
        valuesFromTheParentContainerNodes,
        valuesFromTheAvailablePackageNodes,
        schemaFromTheAvailablePackage,
        deploymentEvent,
        valuesFromTheDeployedPackageNodes,
      );
      setParamsFromComponentState(initialParamsFromContainer);
      setIsLoaded(true);
      setIsLoading(false);
    }
  }, [
    deploymentEvent,
    isLoaded,
    paramsFromComponentState,
    schemaFromTheAvailablePackage,
    valuesFromTheAvailablePackageNodes,
    valuesFromTheDeployedPackageNodes,
    valuesFromTheParentContainerNodes,
  ]);

  // parse and store in the component state the values from the available package
  useEffect(() => {
    if (valuesFromTheAvailablePackage) {
      setValuesFromTheAvailablePackageNodes(parseToYamlNode(valuesFromTheAvailablePackage));
    }
  }, [isLoaded, valuesFromTheAvailablePackage]);

  // parse and store in the component state the current values (which come from the parent container)
  useEffect(() => {
    if (valuesFromTheParentContainer) {
      setValuesFromTheParentContainerNodes(parseToYamlNode(valuesFromTheParentContainer));
    }
  }, [isLoaded, valuesFromTheParentContainer]);

  // parse and store in the component state the values from the deployed package
  useEffect(() => {
    if (valuesFromTheDeployedPackage) {
      setValuesFromTheDeployedPackageNodes(parseToYamlNode(valuesFromTheDeployedPackage));
    }
  }, [isLoaded, valuesFromTheDeployedPackage, valuesFromTheParentContainer]);

  // when the shouldSubmitForm flag is enabled, the form will be submitted, but using a native
  // form submit event (to trigger the browser validations) instead of just calling its handler function
  useEffect(() => {
    if (shouldSubmitForm) {
      // the requestSubmit API was added recently,
      // but should replace the manual dispatch of a submit event with "bubbles:true" (react>=17)
      formRef?.current?.requestSubmit();
      setShouldSubmitForm(false);
    }
  }, [formRef, shouldSubmitForm]);

  // for each unsaved change in the component state, we need to update the values,
  // so that both the table and the yaml editor get the updated values
  const saveAllChanges = () => {
    let newValuesFromTheParentContainer, newParamsFromComponentState;
    unsavedChangesMap.forEach((value, key) => {
      setIsLoading(true);
      setValuesModified();
      newParamsFromComponentState = [
        ...updateCurrentConfigByKey(paramsFromComponentState, key, value),
      ];
      setParamsFromComponentState(newParamsFromComponentState);

      newValuesFromTheParentContainer = toStringYamlNode(
        setPathValueInYamlNode(valuesFromTheParentContainerNodes, key, value),
      );
      setValuesFromTheParentContainer(newValuesFromTheParentContainer);
    });
    unsavedChangesMap.clear();
    setIsLoading(false);
  };

  // save the pending changes and fire the submit event
  // via an useEffect, to actually get the most recent saved changes
  const handleDeployClick = () => {
    saveAllChanges();
    setShouldSubmitForm(true);
  };

  const handleYAMLEditorChange = (value: string) => {
    setValuesFromTheParentContainer(value);
    setValuesModified();
  };

  // re-build the table based on the new YAML
  const refreshBasicParameters = () => {
    if (schemaFromTheAvailablePackage && shouldRenderBasicForm(schemaFromTheAvailablePackage)) {
      setParamsFromComponentState(
        retrieveBasicFormParams(
          valuesFromTheParentContainerNodes,
          valuesFromTheAvailablePackageNodes,
          schemaFromTheAvailablePackage,
          deploymentEvent,
          valuesFromTheDeployedPackageNodes,
        ),
      );
    }
  };

  // a change in the table is just a new entry in the unsavedChangesMap for performance reasons
  // later on, the changes will be saved in bulk
  const handleTableChange = useCallback(
    (value: IBasicFormParam) => {
      return (e: FormEvent<any>) => {
        unsavedChangesMap.set(value.key, getValueFromEvent(e));
      };
    },
    [unsavedChangesMap],
  );

  // The basic form should be rendered if there are params to show
  const shouldRenderBasicForm = (schema: any) => {
    return !isEmpty(schema?.properties);
  };

  const closeRestoreDefaultValuesModal = () => {
    setRestoreModalOpen(false);
  };

  const openRestoreDefaultValuesModal = () => {
    setRestoreModalOpen(true);
  };

  const restoreDefaultValues = () => {
    setValuesFromTheParentContainer(valuesFromTheAvailablePackage || "");
    if (schemaFromTheAvailablePackage) {
      setParamsFromComponentState(
        retrieveBasicFormParams(
          valuesFromTheAvailablePackageNodes,
          valuesFromTheAvailablePackageNodes,
          schemaFromTheAvailablePackage,
          deploymentEvent,
          valuesFromTheDeployedPackageNodes,
        ),
      );
    }
    setRestoreModalOpen(false);
  };

  // early return if error
  if (error) {
    return (
      <Alert theme="danger">
        Unable to fetch package "{packageId}" ({packageVersion}): Got {error.message}
      </Alert>
    );
  }

  // early return if loading
  if (
    packagesIsFetching ||
    !availablePackageDetail ||
    (!versions.length &&
      shouldRenderBasicForm(schemaFromTheAvailablePackage) &&
      !isEmpty(paramsFromComponentState) &&
      !isEmpty(valuesFromTheAvailablePackageNodes))
  ) {
    return (
      <LoadingWrapper
        className="margin-t-xxl"
        loadingText={`Fetching ${decodeURIComponent(packageId)}...`}
      />
    );
  }

  // creation of the each tab + its content
  const tabColumns: JSX.Element[] = [];
  const tabData: JSX.Element[] = [];

  // Basic form creation
  if (shouldRenderBasicForm(schemaFromTheAvailablePackage)) {
    tabColumns.push(
      <div role="presentation" onClick={refreshBasicParameters}>
        <span>Visual editor</span>
      </div>,
    );
    tabData.push(
      <>
        <BasicDeploymentForm
          handleBasicFormParamChange={handleTableChange}
          deploymentEvent={deploymentEvent}
          paramsFromComponentState={paramsFromComponentState}
          isLoading={isLoading}
          saveAllChanges={saveAllChanges}
        />
      </>,
    );
  }

  // Text editor creation
  tabColumns.push(
    <div role="presentation" onClick={saveAllChanges}>
      <span>YAML editor</span>
    </div>,
  );
  tabData.push(
    <AdvancedDeploymentForm
      valuesFromTheParentContainer={valuesFromTheParentContainer}
      deploymentEvent={deploymentEvent}
      valuesFromTheAvailablePackage={valuesFromTheAvailablePackage || ""}
      valuesFromTheDeployedPackage={valuesFromTheDeployedPackage || ""}
      handleValuesChange={handleYAMLEditorChange}
      key="advanced-deployment-form"
    ></AdvancedDeploymentForm>,
  );

  return (
    <div>
      <ConfirmDialog
        modalIsOpen={restoreModalIsOpen}
        loading={false}
        headerText={"Restore defaults"}
        confirmationText={"Are you sure you want to restore the default package values?"}
        confirmationButtonText={"Restore"}
        onConfirm={restoreDefaultValues}
        closeModal={closeRestoreDefaultValuesModal}
      />
      <div className="deployment-form-tabs">
        <Tabs columns={tabColumns} data={tabData} id="deployment-form-body-tabs" />
        {/* eslint-disable jsx-a11y/anchor-is-valid */}
        <CdsControlMessage>
          The unsaved changes will automatically be applied before deploying or when visualizing the
          diff view. You can also{" "}
          <a
            id="table-manual-save"
            href="#"
            role="button"
            tabIndex={0}
            onClick={saveAllChanges}
            onKeyDown={saveAllChanges}
          >
            save the changes manually
          </a>
          .
        </CdsControlMessage>
      </div>
      <div className="deployment-form-control-buttons">
        <CdsButton status="primary" type="button" onClick={handleDeployClick}>
          <CdsIcon shape="deploy" /> Deploy {pkgVersion}
        </CdsButton>
        <CdsButton
          type="button"
          status="primary"
          action="outline"
          onClick={openRestoreDefaultValuesModal}
        >
          <CdsIcon shape="backup-restore" /> Restore Defaults
        </CdsButton>
      </div>
    </div>
  );
}

export default DeploymentFormBody;
