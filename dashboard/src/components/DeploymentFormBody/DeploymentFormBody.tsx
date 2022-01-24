// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import { CdsIcon } from "@cds/react/icon";
import Alert from "components/js/Alert";
import Tabs from "components/Tabs";
import { isEqual } from "lodash";
import { useEffect, useState } from "react";
import { parseValues, retrieveBasicFormParams, setValue } from "../../shared/schema";
import { DeploymentEvent, IBasicFormParam, IPackageState } from "../../shared/types";
import { getValueFromEvent } from "../../shared/utils";
import ConfirmDialog from "../ConfirmDialog/ConfirmDialog";
import LoadingWrapper from "../LoadingWrapper/LoadingWrapper";
import AdvancedDeploymentForm from "./AdvancedDeploymentForm";
import BasicDeploymentForm from "./BasicDeploymentForm/BasicDeploymentForm";
import DifferentialSelector from "./DifferentialSelector";
import DifferentialTab from "./DifferentialTab";

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
}

function DeploymentFormBody({
  deploymentEvent,
  packageId,
  packageVersion,
  deployedValues,
  packagesIsFetching,
  selected,
  appValues,
  setValues,
  setValuesModified,
}: IDeploymentFormBodyProps) {
  const [basicFormParameters, setBasicFormParameters] = useState([] as IBasicFormParam[]);
  const [restoreModalIsOpen, setRestoreModalOpen] = useState(false);
  const [defaultValues, setDefaultValues] = useState("");

  const { availablePackageDetail, versions, schema, values, pkgVersion, error } = selected;

  useEffect(() => {
    const params = retrieveBasicFormParams(appValues, schema);
    if (!isEqual(params, basicFormParameters)) {
      setBasicFormParameters(params);
    }
  }, [setBasicFormParameters, schema, appValues, basicFormParameters]);

  useEffect(() => {
    setDefaultValues(values || "");
  }, [values]);

  const handleValuesChange = (value: string) => {
    setValues(value);
    setValuesModified();
  };
  const refreshBasicParameters = () => {
    setBasicFormParameters(retrieveBasicFormParams(appValues, schema));
  };

  const handleBasicFormParamChange = (param: IBasicFormParam) => {
    const parsedDefaultValues = parseValues(defaultValues);
    return (e: React.FormEvent<any>) => {
      setValuesModified();
      if (parsedDefaultValues !== defaultValues) {
        setDefaultValues(parsedDefaultValues);
      }
      const value = getValueFromEvent(e);
      setBasicFormParameters(
        basicFormParameters.map(p => (p.path === param.path ? { ...param, value } : p)),
      );
      // Change raw values
      setValues(setValue(appValues, param.path, value));
    };
  };

  // The basic form should be rendered if there are params to show
  const shouldRenderBasicForm = () => {
    return Object.keys(basicFormParameters).length > 0;
  };

  const closeRestoreDefaultValuesModal = () => {
    setRestoreModalOpen(false);
  };

  const openRestoreDefaultValuesModal = () => {
    setRestoreModalOpen(true);
  };

  const restoreDefaultValues = () => {
    if (values) {
      setValues(values);
      setBasicFormParameters(retrieveBasicFormParams(values, schema));
    }
    setRestoreModalOpen(false);
  };
  if (error) {
    return (
      <Alert theme="danger">
        Unable to fetch package "{packageId}" ({packageVersion}): Got {error.message}
      </Alert>
    );
  }
  if (packagesIsFetching || !availablePackageDetail || !versions.length) {
    return (
      <LoadingWrapper
        className="margin-t-xxl"
        loadingText={`Fetching ${decodeURIComponent(packageId)}...`}
      />
    );
  }
  const tabColumns = [
    "YAML",
    <DifferentialTab
      key="differential-selector"
      deploymentEvent={deploymentEvent}
      defaultValues={defaultValues}
      deployedValues={deployedValues || ""}
      appValues={appValues}
    />,
  ] as Array<string | JSX.Element | JSX.Element[]>;
  const tabData = [
    <AdvancedDeploymentForm
      appValues={appValues}
      handleValuesChange={handleValuesChange}
      key="advanced-deployment-form"
    >
      <p>
        <b>Note:</b> Only comments from the original package values will be preserved.
      </p>
    </AdvancedDeploymentForm>,
    <DifferentialSelector
      key="differential-selector"
      deploymentEvent={deploymentEvent}
      defaultValues={defaultValues}
      deployedValues={deployedValues || ""}
      appValues={appValues}
    />,
  ];
  if (shouldRenderBasicForm()) {
    tabColumns.unshift(
      <span role="presentation" onClick={refreshBasicParameters}>
        Form
      </span>,
    );
    tabData.unshift(
      <BasicDeploymentForm
        deploymentEvent={deploymentEvent}
        params={basicFormParameters}
        handleBasicFormParamChange={handleBasicFormParamChange}
        appValues={appValues}
        handleValuesChange={handleValuesChange}
      />,
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
      </div>
      <div className="deployment-form-control-buttons">
        <CdsButton status="primary" type="submit">
          <CdsIcon shape="deploy" /> Deploy {pkgVersion}
        </CdsButton>
        {/* TODO(andresmgot): CdsButton "type" property doesn't work, so we need to use a normal <button>
            https://github.com/vmware/clarity/issues/5038
          */}
        <span className="color-icon-info">
          <button
            className="btn btn-info-outline"
            type="button"
            onClick={openRestoreDefaultValuesModal}
            style={{ marginTop: "-22px" }}
          >
            <CdsIcon shape="backup-restore" /> Restore Defaults
          </button>
        </span>
      </div>
    </div>
  );
}

export default DeploymentFormBody;
