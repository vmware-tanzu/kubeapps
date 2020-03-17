import { RouterAction } from "connected-react-router";
import * as jsonpatch from "fast-json-patch";
import * as React from "react";
import { Tab, TabList, TabPanel, Tabs } from "react-tabs";

import { retrieveBasicFormParams, setValue } from "../../shared/schema";
import { IBasicFormParam, IChartState } from "../../shared/types";
import { getValueFromEvent } from "../../shared/utils";
import ConfirmDialog from "../ConfirmDialog";
import { ErrorSelector } from "../ErrorAlert";
import Hint from "../Hint";
import LoadingWrapper from "../LoadingWrapper";
import AdvancedDeploymentForm from "./AdvancedDeploymentForm";
import BasicDeploymentForm from "./BasicDeploymentForm";

import "react-tabs/style/react-tabs.css";
import Differential from "./Differential";
import "./Tabs.css";

export interface IDeploymentFormBodyProps {
  chartID: string;
  chartVersion: string;
  deployedValues?: string;
  namespace: string;
  releaseVersion?: string;
  selected: IChartState["selected"];
  appValues: string;
  push: (location: string) => RouterAction;
  goBack?: () => RouterAction;
  getChartVersion: (id: string, chartVersion: string) => void;
  setValues: (values: string) => void;
  setValuesModified: () => void;
}

export interface IDeploymentFormBodyState {
  basicFormParameters: IBasicFormParam[];
  restoreDefaultValuesModalIsOpen: boolean;
  modifications?: jsonpatch.Operation[];
}

class DeploymentFormBody extends React.Component<
  IDeploymentFormBodyProps,
  IDeploymentFormBodyState
> {
  public state: IDeploymentFormBodyState = {
    basicFormParameters: [],
    restoreDefaultValuesModalIsOpen: false,
  };

  public componentDidMount() {
    const { chartID, getChartVersion, chartVersion } = this.props;
    getChartVersion(chartID, chartVersion);
  }

  public componentDidUpdate = (prevProps: IDeploymentFormBodyProps) => {
    const { chartID, chartVersion, getChartVersion, selected, appValues } = this.props;

    if (chartVersion !== prevProps.chartVersion) {
      // New version detected
      getChartVersion(chartID, chartVersion);
      return;
    }

    if (
      // If the selected schema changes
      prevProps.selected.schema !== selected.schema ||
      // or the selected values (because it's a different version)
      prevProps.selected.values !== selected.values ||
      // or the current values (and we hadn't process those values yet)
      (prevProps.appValues !== appValues && prevProps.appValues.length === 0)
    ) {
      // Parse the basic parameters
      this.setState({
        basicFormParameters: retrieveBasicFormParams(appValues, selected.schema),
      });
    }
  };

  public render() {
    const { selected, chartID, chartVersion, goBack } = this.props;
    const { version, versions } = selected;
    if (selected.error) {
      return (
        <ErrorSelector error={selected.error} resource={`Chart "${chartID}" (${chartVersion})`} />
      );
    }
    if (!version || !versions.length) {
      return <LoadingWrapper />;
    }
    return (
      <div>
        <ConfirmDialog
          modalIsOpen={this.state.restoreDefaultValuesModalIsOpen}
          loading={false}
          confirmationText={"Are you sure you want to restore the default chart values?"}
          confirmationButtonText={"Restore"}
          onConfirm={this.restoreDefaultValues}
          closeModal={this.closeRestoreDefaultValuesModal}
        />
        <div>
          <label htmlFor="chartVersion">Version</label>
          <select
            id="chartVersion"
            onChange={this.handleChartVersionChange}
            value={version.attributes.version}
            required={true}
          >
            {versions.map(v => (
              <option key={v.id} value={v.attributes.version}>
                {v.attributes.version}{" "}
                {this.props.releaseVersion && v.attributes.version === this.props.releaseVersion
                  ? "(current)"
                  : ""}
              </option>
            ))}
          </select>
        </div>
        {this.renderTabs()}
        <div className="margin-t-big">
          <button className="button button-primary" type="submit">
            Submit
          </button>
          <button className="button" type="button" onClick={this.openRestoreDefaultValuesModal}>
            Restore Chart Defaults
          </button>
          {goBack && (
            <button className="button" type="button" onClick={goBack}>
              Back
            </button>
          )}
        </div>
      </div>
    );
  }

  private handleChartVersionChange = (e: React.FormEvent<HTMLSelectElement>) => {
    // TODO(andres): This requires refactoring. Currently, the deploy and upgrade
    // forms behave differently. In the deployment form, a change in the version
    // changes the route but in the case of the upgrade it only changes the state
    const isUpgradeForm = !!this.props.releaseVersion;

    if (isUpgradeForm) {
      const { chartID, getChartVersion } = this.props;
      getChartVersion(chartID, e.currentTarget.value);
    } else {
      this.props.push(
        `/apps/ns/${this.props.namespace}/new/${this.props.chartID}/versions/${e.currentTarget.value}`,
      );
    }
  };

  private handleValuesChange = (value: string) => {
    this.props.setValues(value);
    this.props.setValuesModified();
  };

  private refreshBasicParameters = () => {
    this.setState({
      basicFormParameters: retrieveBasicFormParams(
        this.props.appValues,
        this.props.selected.schema,
      ),
    });
  };

  private renderTabs = () => {
    return (
      <div className="margin-t-normal">
        <Tabs>
          <TabList>
            {this.shouldRenderBasicForm() && (
              <Tab onClick={this.refreshBasicParameters}>
                Form{" "}
                <Hint reactTooltipOpts={{ delayHide: 100 }} id="basicFormHelp">
                  <span>
                    This form has been automatically generated based on the chart schema.
                    <br />
                    This feature is currently in a beta state. If you find an issue please report it{" "}
                    <a target="_blank" href="https://github.com/kubeapps/kubeapps/issues/new">
                      here.
                    </a>
                  </span>
                </Hint>
              </Tab>
            )}
            <Tab>Values (YAML)</Tab>
            <Tab>Changes</Tab>
          </TabList>
          {this.shouldRenderBasicForm() && (
            <TabPanel>
              <BasicDeploymentForm
                params={this.state.basicFormParameters}
                handleBasicFormParamChange={this.handleBasicFormParamChange}
                appValues={this.props.appValues}
                handleValuesChange={this.handleValuesChange}
              />
            </TabPanel>
          )}
          <TabPanel>
            <AdvancedDeploymentForm
              appValues={this.props.appValues}
              handleValuesChange={this.handleValuesChange}
            >
              <p>
                <b>Note:</b> Only comments from the original chart values will be preserved.
              </p>
            </AdvancedDeploymentForm>
          </TabPanel>
          <TabPanel>{this.renderDiff()}</TabPanel>
        </Tabs>
      </div>
    );
  };

  private handleBasicFormParamChange = (param: IBasicFormParam) => {
    return (e: React.FormEvent<HTMLInputElement>) => {
      this.props.setValuesModified();
      const value = getValueFromEvent(e);
      this.setState({
        // Change param definition
        basicFormParameters: this.state.basicFormParameters.map(p =>
          p.path === param.path ? { ...param, value } : p,
        ),
      });
      // Change raw values
      this.props.setValues(setValue(this.props.appValues, param.path, value));
    };
  };

  // The basic form should be rendered if there are params to show
  private shouldRenderBasicForm = () => {
    return Object.keys(this.state.basicFormParameters).length > 0;
  };

  private closeRestoreDefaultValuesModal = () => {
    this.setState({ restoreDefaultValuesModalIsOpen: false });
  };

  private openRestoreDefaultValuesModal = () => {
    this.setState({ restoreDefaultValuesModalIsOpen: true });
  };

  private restoreDefaultValues = () => {
    if (this.props.selected.values) {
      this.props.setValues(this.props.selected.values);
      this.setState({
        basicFormParameters: retrieveBasicFormParams(
          this.props.selected.values,
          this.props.selected.schema,
        ),
      });
    }
    this.setState({ restoreDefaultValuesModalIsOpen: false });
  };

  private renderDiff = () => {
    let oldValues = "";
    let title = "";
    let emptyDiffText = "";
    if (this.props.deployedValues) {
      // If there are already some deployed values (upgrade scenario)
      // We compare the values from the old release and the new one
      oldValues = this.props.deployedValues;
      title = "Difference from deployed version";
      emptyDiffText = "The values for the new release are identical to the deployed version.";
    } else {
      // If it's a new deployment, we show the different from the default
      // values for the selected version
      oldValues = this.props.selected.values || "";
      title = "Difference from chart defaults";
      emptyDiffText = "No changes detected from chart defaults.";
    }
    return (
      <Differential
        title={title}
        oldValues={oldValues}
        newValues={this.props.appValues}
        emptyDiffText={emptyDiffText}
      />
    );
  };
}

export default DeploymentFormBody;
