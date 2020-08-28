import Tabs from "components/Tabs";
import React, { useEffect, useState } from "react";

import { CdsButton, CdsIcon } from "components/Clarity/clarity";
import Alert from "components/js/Alert";
import { retrieveBasicFormParams, setValue } from "../../shared/schema";
import { DeploymentEvent, IBasicFormParam, IChartState } from "../../shared/types";
import { getValueFromEvent } from "../../shared/utils";
import ConfirmDialog from "../ConfirmDialog/ConfirmDialog.v2";
import LoadingWrapper from "../LoadingWrapper/LoadingWrapper.v2";
import AdvancedDeploymentForm from "./AdvancedDeploymentForm.v2";
import BasicDeploymentForm from "./BasicDeploymentForm/BasicDeploymentForm.v2";
import DifferentialSelector from "./DifferentialSelector";

export interface IDeploymentFormBodyProps {
  deploymentEvent: DeploymentEvent;
  chartID: string;
  chartVersion: string;
  deployedValues?: string;
  chartsIsFetching: boolean;
  selected: IChartState["selected"];
  appValues: string;
  setValues: (values: string) => void;
  setValuesModified: () => void;
}

function DeploymentFormBody({
  deploymentEvent,
  chartID,
  chartVersion,
  deployedValues,
  chartsIsFetching,
  selected,
  appValues,
  setValues,
  setValuesModified,
}: IDeploymentFormBodyProps) {
  const [basicFormParameters, setBasicFormParameters] = useState([] as IBasicFormParam[]);
  const [restoreModalIsOpen, setRestoreModalOpen] = useState(false);
  const [formParamsPopulated, setFormParamsPopulated] = useState(false);
  const { version, versions, schema } = selected;

  useEffect(() => {
    if (appValues !== "" && !formParamsPopulated) {
      setBasicFormParameters(retrieveBasicFormParams(appValues, schema));
      setFormParamsPopulated(true);
    }
  }, [setBasicFormParameters, schema, appValues, formParamsPopulated]);

  const handleValuesChange = (value: string) => {
    setValues(value);
    setValuesModified();
  };
  const refreshBasicParameters = () => {
    setBasicFormParameters(retrieveBasicFormParams(appValues, schema));
  };

  const handleBasicFormParamChange = (param: IBasicFormParam) => {
    return (e: React.FormEvent<any>) => {
      setValuesModified();
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
    if (selected.values) {
      setValues(selected.values);
      setBasicFormParameters(retrieveBasicFormParams(selected.values, selected.schema));
    }
    setRestoreModalOpen(false);
  };
  if (selected.error) {
    return (
      <Alert theme="danger">
        Unable to fetch chart "{chartID}" ({chartVersion}): Got {selected.error.message}
      </Alert>
    );
  }
  if (chartsIsFetching || !version || !versions.length) {
    return <LoadingWrapper />;
  }
  const tabColumns = ["YAML", "Changes"] as Array<string | JSX.Element | JSX.Element[]>;
  const tabData = [
    <AdvancedDeploymentForm
      appValues={appValues}
      handleValuesChange={handleValuesChange}
      key="advanced-deployment-form"
    >
      <p>
        <b>Note:</b> Only comments from the original chart values will be preserved.
      </p>
    </AdvancedDeploymentForm>,
    <DifferentialSelector
      key="differential-selector"
      deploymentEvent={deploymentEvent}
      defaultValues={selected.values || ""}
      deployedValues={deployedValues || ""}
      appValues={appValues}
    />,
  ];
  if (shouldRenderBasicForm()) {
    tabColumns.unshift(<span onClick={refreshBasicParameters}>Form</span>);
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
        confirmationText={"Are you sure you want to restore the default chart values?"}
        confirmationButtonText={"Restore"}
        onConfirm={restoreDefaultValues}
        closeModal={closeRestoreDefaultValuesModal}
      />
      <div className="deployment-form-tabs">
        <Tabs columns={tabColumns} data={tabData} id="deployment-form-body-tabs" />
      </div>
      <div className="deployment-form-control-buttons">
        <CdsButton status="primary" type="submit">
          <CdsIcon shape="deploy" inverse={true} /> Deploy v{version.attributes.version}
        </CdsButton>
        <CdsButton action="outline" type="button" onClick={openRestoreDefaultValuesModal}>
          <CdsIcon shape="backup-restore" inverse={true} /> Restore Defaults
        </CdsButton>
      </div>
    </div>
  );
}

export default DeploymentFormBody;
