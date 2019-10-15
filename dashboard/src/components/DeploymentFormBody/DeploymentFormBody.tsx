import { RouterAction } from "connected-react-router";
import * as React from "react";
import { Tab, TabList, TabPanel, Tabs } from "react-tabs";

import { retrieveBasicFormParams, setValue } from "../../shared/schema";
import { IBasicFormParam, IChartState } from "../../shared/types";
import { getValueFromEvent } from "../../shared/utils";
import { ErrorSelector } from "../ErrorAlert";
import LoadingWrapper from "../LoadingWrapper";
import AdvancedDeploymentForm from "./AdvancedDeploymentForm";
import BasicDeploymentForm from "./BasicDeploymentForm";

import "react-tabs/style/react-tabs.css";
import "./Tabs.css";

export interface IDeploymentFormBodyProps {
  chartID: string;
  chartVersion: string;
  originalValues?: string;
  namespace: string;
  releaseName?: string;
  selected: IChartState["selected"];
  appValues: string;
  valuesModified: boolean;
  push: (location: string) => RouterAction;
  goBack?: () => RouterAction;
  fetchChartVersions: (id: string) => void;
  getChartVersion: (id: string, chartVersion: string) => void;
  setValues: (values: string) => void;
  setValuesModified: () => void;
  enableBasicForm: boolean;
}

export interface IDeploymentFormBodyState {
  basicFormParameters: { [key: string]: IBasicFormParam };
}

class DeploymentFormBody extends React.Component<
  IDeploymentFormBodyProps,
  IDeploymentFormBodyState
> {
  public state: IDeploymentFormBodyState = {
    basicFormParameters: {},
  };

  public componentDidMount() {
    const { chartID, fetchChartVersions, getChartVersion, chartVersion } = this.props;
    fetchChartVersions(chartID);
    getChartVersion(chartID, chartVersion);
  }

  public componentWillReceiveProps = (nextProps: IDeploymentFormBodyProps) => {
    const { chartID, chartVersion, getChartVersion } = this.props;

    if (chartVersion !== nextProps.chartVersion) {
      // New version detected
      getChartVersion(chartID, nextProps.chartVersion);
      return;
    }

    if (nextProps.selected !== this.props.selected) {
      // The values or the schema has changed
      let values = "";
      if (!this.props.valuesModified) {
        // If the version is the current one, reuse original params
        // (this only applies to the upgrade form that has originalValues defined)
        if (
          nextProps.selected.version &&
          nextProps.selected.version.attributes.version === this.props.chartVersion &&
          this.props.originalValues
        ) {
          values = this.props.originalValues || "";
        } else {
          // In other case, use the default values for the selected version
          values = nextProps.selected.values || "";
        }
        this.props.setValues(values);
      } else {
        // If the user has modified the values, use the ones defined
        values = this.props.appValues;
      }
      if (nextProps.selected.schema) {
        this.setState({
          basicFormParameters: retrieveBasicFormParams(values, nextProps.selected.schema),
        });
      }
      return;
    }
  };

  public render() {
    const { selected, chartID, chartVersion, goBack, appValues } = this.props;
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
                {this.props.releaseName && v.attributes.version === this.props.chartVersion
                  ? "(current)"
                  : ""}
              </option>
            ))}
          </select>
        </div>
        {this.shouldRenderBasicForm() ? (
          this.renderTabs()
        ) : (
          <AdvancedDeploymentForm
            appValues={appValues}
            handleValuesChange={this.handleValuesChange}
          />
        )}
        <div>
          <button className="button button-primary margin-t-big" type="submit">
            Submit
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
    const isUpgradeForm = !!this.props.releaseName;

    if (isUpgradeForm) {
      const { chartID, getChartVersion } = this.props;
      getChartVersion(chartID, e.currentTarget.value);
    } else {
      this.props.push(
        `/apps/ns/${this.props.namespace}/new/${this.props.chartID}/versions/${
          e.currentTarget.value
        }`,
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
            <Tab onClick={this.refreshBasicParameters}>Basic</Tab>
            <Tab>Advanced</Tab>
          </TabList>
          <TabPanel>
            <BasicDeploymentForm
              params={this.state.basicFormParameters}
              handleBasicFormParamChange={this.handleBasicFormParamChange}
              appValues={this.props.appValues}
              handleValuesChange={this.handleValuesChange}
            />
          </TabPanel>
          <TabPanel>
            <AdvancedDeploymentForm
              appValues={this.props.appValues}
              handleValuesChange={this.handleValuesChange}
            />
          </TabPanel>
        </Tabs>
      </div>
    );
  };

  private handleBasicFormParamChange = (name: string, param: IBasicFormParam) => {
    return (e: React.FormEvent<HTMLInputElement>) => {
      this.props.setValuesModified();
      const value = getValueFromEvent(e);
      this.setState({
        // Change param definition
        basicFormParameters: {
          ...this.state.basicFormParameters,
          [name]: {
            ...param,
            value,
          },
        },
      });
      // Change raw values
      this.props.setValues(setValue(this.props.appValues, param.path, value));
    };
  };

  // The basic form should be rendered both if it's enabled and if there are params to show
  private shouldRenderBasicForm = () => {
    return this.props.enableBasicForm && Object.keys(this.state.basicFormParameters).length > 0;
  };
}

export default DeploymentFormBody;
