// Copyright 2019-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import { CdsControlMessage } from "@cds/react/forms";
import { CdsIcon } from "@cds/react/icon";
import { JSONSchemaType } from "ajv";
import AlertGroup from "components/AlertGroup";
import ConfirmDialog from "components/ConfirmDialog";
import LoadingWrapper from "components/LoadingWrapper";
import Tabs from "components/Tabs";
import { isEmpty } from "lodash";
import { FormEvent, RefObject, useCallback, useEffect, useState } from "react";
import { useSelector } from "react-redux";
import {
  retrieveBasicFormParams,
  schemaToObject,
  schemaToString,
  updateCurrentConfigByKey,
} from "shared/schema";
import { DeploymentEvent, IBasicFormParam, IPackageState, IStoreState } from "shared/types";
import { getValueFromEvent } from "shared/utils";
import { parseToYamlNode, setPathValueInYamlNode, toStringYamlNode } from "shared/yamlUtils";
import YAML from "yaml";
import AdvancedDeploymentForm from "./AdvancedDeploymentForm";
import BasicDeploymentForm from "./BasicDeploymentForm/BasicDeploymentForm";
import SchemaEditorForm from "./SchemaEditorForm";

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
  const {
    config: { featureFlags },
  } = useSelector((state: IStoreState) => state);

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

  const [schemaFromTheAvailablePackageString, setSchemaFromTheAvailablePackageString] =
    useState("");
  const [schemaFromTheParentContainerString, setSchemaFromTheParentContainerString] = useState("");
  const [schemaFromTheParentContainerParsed, setSchemaFromTheParentContainerParsed] = useState(
    {} as JSONSchemaType<any>,
  );

  const [restoreModalIsOpen, setRestoreModalOpen] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [unsavedChangesMap] = useState(new Map<string, any>());
  const [shouldSubmitForm, setShouldSubmitForm] = useState(false);

  // whenever the parsed values change (for instance, when a new pkg version is selected),
  // we need to force a new extraction of the params from the schema
  useEffect(() => {
    if (
      !isLoading &&
      !isEmpty(valuesFromTheParentContainerNodes) &&
      shouldRenderBasicForm(schemaFromTheParentContainerParsed)
    ) {
      const initialParamsFromContainer = retrieveBasicFormParams(
        valuesFromTheParentContainerNodes,
        valuesFromTheAvailablePackageNodes,
        schemaFromTheParentContainerParsed,
        deploymentEvent,
        valuesFromTheDeployedPackageNodes,
      );
      setParamsFromComponentState(initialParamsFromContainer);
      setIsLoading(false);
    }
  }, [
    deploymentEvent,
    isLoading,
    schemaFromTheParentContainerParsed,
    valuesFromTheAvailablePackageNodes,
    valuesFromTheDeployedPackageNodes,
    valuesFromTheParentContainerNodes,
  ]);

  // parse and store in the component state the values from the available package
  useEffect(() => {
    if (valuesFromTheAvailablePackage) {
      setIsLoading(true);
      setValuesFromTheAvailablePackageNodes(parseToYamlNode(valuesFromTheAvailablePackage));
      setIsLoading(false);
    }
  }, [valuesFromTheAvailablePackage]);

  // parse and store in the component state the current values (which come from the parent container)
  useEffect(() => {
    if (valuesFromTheParentContainer) {
      setIsLoading(true);
      setValuesFromTheParentContainerNodes(parseToYamlNode(valuesFromTheParentContainer));
      setIsLoading(false);
    }
  }, [valuesFromTheParentContainer]);

  // parse and store in the component state the values from the deployed package
  useEffect(() => {
    if (valuesFromTheDeployedPackage) {
      setIsLoading(true);
      setValuesFromTheDeployedPackageNodes(parseToYamlNode(valuesFromTheDeployedPackage));
      setIsLoading(false);
    }
  }, [valuesFromTheDeployedPackage]);

  // initialize the schema string in component state with the package's schema
  useEffect(() => {
    setIsLoading(true);
    const schemaString = schemaToString(schemaFromTheAvailablePackage);
    setSchemaFromTheAvailablePackageString(schemaString);
    setSchemaFromTheParentContainerString(schemaString);
    setSchemaFromTheParentContainerParsed(schemaFromTheAvailablePackage as JSONSchemaType<any>);
    setIsLoading(false);
  }, [schemaFromTheAvailablePackage]);

  // parse and store in the component state the current schema
  useEffect(() => {
    setIsLoading(true);
    const schemaObject = schemaToObject(schemaFromTheParentContainerString);
    setSchemaFromTheParentContainerParsed(schemaObject);
    setIsLoading(false);
  }, [schemaFromTheParentContainerString]);

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

  const handleSchemaEditorChange = (value: string) => {
    setSchemaFromTheParentContainerString(value);
    setValuesModified();
  };

  // re-build the table based on the new YAML
  const refreshBasicParameters = () => {
    if (shouldRenderBasicForm(schemaFromTheParentContainerParsed)) {
      setIsLoading(true);
      setParamsFromComponentState(
        retrieveBasicFormParams(
          valuesFromTheParentContainerNodes,
          valuesFromTheAvailablePackageNodes,
          schemaFromTheParentContainerParsed,
          deploymentEvent,
          valuesFromTheDeployedPackageNodes,
        ),
      );
      setIsLoading(false);
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
    setIsLoading(true);
    setValuesFromTheParentContainer(valuesFromTheAvailablePackage || "");
    setSchemaFromTheParentContainerParsed(schemaFromTheAvailablePackage as JSONSchemaType<any>);
    setSchemaFromTheParentContainerString(schemaToString(schemaFromTheAvailablePackage));
    if (shouldRenderBasicForm(schemaFromTheParentContainerParsed)) {
      setParamsFromComponentState(
        retrieveBasicFormParams(
          valuesFromTheAvailablePackageNodes,
          valuesFromTheAvailablePackageNodes,
          schemaFromTheParentContainerParsed,
          deploymentEvent,
          valuesFromTheDeployedPackageNodes,
        ),
      );
    }
    setRestoreModalOpen(false);
    setIsLoading(false);
  };

  // early return if error
  if (error) {
    return (
      <AlertGroup status="danger">
        Unable to fetch the package "{decodeURIComponent(packageId)} ({packageVersion})":{" "}
        {error.message}.
      </AlertGroup>
    );
  }

  // early return if loading
  if (
    isLoading ||
    packagesIsFetching ||
    !availablePackageDetail ||
    (!versions.length &&
      shouldRenderBasicForm(schemaFromTheParentContainerParsed) &&
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
  const tabColumns: [JSX.Element, () => void][] = [];
  const tabData: JSX.Element[] = [];

  // Basic form creation
  if (shouldRenderBasicForm(schemaFromTheParentContainerParsed)) {
    tabColumns.push([
      // This is a tuple, not an array requiring a key.
      // eslint-disable-next-line react/jsx-key
      <div>
        <span>Visual editor</span>
      </div>,
      refreshBasicParameters,
    ]);
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
  tabColumns.push([
    // This is a tuple, not an array requiring a key.
    // eslint-disable-next-line react/jsx-key
    <div>
      <span>YAML editor</span>
    </div>,
    saveAllChanges,
  ]);
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
  if (featureFlags?.schemaEditor?.enabled) {
    // Schema editor creation, if the feature flag is enabled
    tabColumns.push([
      // This is a tuple, not an array requiring a key.
      // eslint-disable-next-line react/jsx-key
      <div>
        <span>Schema editor (advanced)</span>
      </div>,
      saveAllChanges,
    ]);
    tabData.push(
      <SchemaEditorForm
        schemaFromTheParentContainer={schemaFromTheParentContainerString}
        schemaFromTheAvailablePackage={schemaFromTheAvailablePackageString}
        handleValuesChange={handleSchemaEditorChange}
        key="schema-editor-form"
      ></SchemaEditorForm>,
    );
  }

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
