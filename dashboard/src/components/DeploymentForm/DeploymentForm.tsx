import { RouterAction } from "connected-react-router";
import * as Moniker from "moniker-native";
import * as React from "react";
import AceEditor from "react-ace";

import { IChartState, IChartVersion } from "../../shared/types";
import { ErrorSelector } from "../ErrorAlert";
import LoadingWrapper from "../LoadingWrapper";

import "brace/mode/yaml";
import "brace/theme/xcode";
import "./DeploymentForm.css";

export interface IDeploymentFormProps {
  kubeappsNamespace: string;
  chartID: string;
  chartVersion: string;
  error: Error | undefined;
  selected: IChartState["selected"];
  deployChart: (
    version: IChartVersion,
    releaseName: string,
    namespace: string,
    values?: string,
  ) => Promise<boolean>;
  push: (location: string) => RouterAction;
  fetchChartVersions: (id: string) => void;
  getChartVersion: (id: string, chartVersion: string) => void;
  getChartValues: (id: string, chartVersion: string) => void;
  namespace: string;
  enableBasicForm: boolean;
}

export interface IDeploymentFormState {
  isDeploying: boolean;
  // deployment options
  releaseName: string;
  // Name of the release that was submitted for creation
  // This is different than releaseName since it is also used in the error banner
  // and we do not want to use releaseName since it is controller by the form field.
  latestSubmittedReleaseName: string;
  namespace: string;
  appValues?: string;
  valuesModified: boolean;
  showBasicForm: boolean;
}

class DeploymentForm extends React.Component<IDeploymentFormProps, IDeploymentFormState> {
  public state: IDeploymentFormState = {
    appValues: undefined,
    isDeploying: false,
    namespace: this.props.namespace,
    releaseName: Moniker.choose(),
    latestSubmittedReleaseName: "",
    valuesModified: false,
    // Use the basic form by default if supported
    showBasicForm: this.props.enableBasicForm,
  };

  public componentDidMount() {
    const { chartID, fetchChartVersions, getChartVersion, chartVersion } = this.props;
    fetchChartVersions(chartID);
    getChartVersion(chartID, chartVersion);
  }

  public componentWillReceiveProps(nextProps: IDeploymentFormProps) {
    const {
      chartID,
      chartVersion,
      getChartValues,
      getChartVersion,
      namespace,
      selected,
    } = this.props;
    const { version } = selected;

    if (nextProps.namespace !== namespace) {
      this.setState({ namespace: nextProps.namespace });
      return;
    }

    if (chartVersion !== nextProps.chartVersion) {
      getChartVersion(chartID, nextProps.chartVersion);
      return;
    }

    if (nextProps.selected.version && nextProps.selected.version !== this.props.selected.version) {
      getChartValues(chartID, nextProps.selected.version.attributes.version);
      return;
    }

    if (!this.state.valuesModified) {
      if (version) {
        this.setState({ appValues: nextProps.selected.values });
      }
    }
  }

  public render() {
    const { selected, chartID, chartVersion, namespace } = this.props;
    const { version, versions } = selected;
    const { latestSubmittedReleaseName } = this.state;
    if (selected.error) {
      return (
        <ErrorSelector error={selected.error} resource={`Chart "${chartID}" (${chartVersion})`} />
      );
    }
    if (!version || !versions.length || this.state.isDeploying) {
      return <LoadingWrapper />;
    }
    return (
      <div>
        <form className="container padding-b-bigger" onSubmit={this.handleDeploy}>
          {this.props.error && (
            <ErrorSelector
              error={this.props.error}
              namespace={namespace}
              action="create"
              resource={latestSubmittedReleaseName}
            />
          )}
          <div className="row">
            <div className="col-12">
              <h2>{this.props.chartID}</h2>
            </div>
            <div className="col-8">
              <div>
                <label htmlFor="releaseName">Name</label>
                <input
                  id="releaseName"
                  pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
                  title="Use lower case alphanumeric characters, '-' or '.'"
                  onChange={this.handleReleaseNameChange}
                  value={this.state.releaseName}
                  required={true}
                />
              </div>
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
                    </option>
                  ))}
                </select>
              </div>
              {this.props.enableBasicForm && this.renderTabs()}
              {this.state.showBasicForm ? this.renderBasicForm() : this.renderAdvancedForm()}
              <div>
                <button className="button button-primary margin-t-big" type="submit">
                  Submit
                </button>
              </div>
            </div>
          </div>
        </form>
      </div>
    );
  }

  public handleDeploy = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const { selected, deployChart, push } = this.props;
    const { releaseName, namespace, appValues } = this.state;

    this.setState({ isDeploying: true, latestSubmittedReleaseName: releaseName });
    if (selected.version) {
      const deployed = await deployChart(selected.version, releaseName, namespace, appValues);
      if (deployed) {
        push(`/apps/ns/${namespace}/${releaseName}`);
      } else {
        this.setState({ isDeploying: false });
      }
    }
  };

  public handleReleaseNameChange = (e: React.FormEvent<HTMLInputElement>) => {
    this.setState({ releaseName: e.currentTarget.value });
  };

  public handleChartVersionChange = (e: React.FormEvent<HTMLSelectElement>) => {
    this.props.push(
      `/apps/ns/${this.props.namespace}/new/${this.props.chartID}/versions/${
        e.currentTarget.value
      }`,
    );
  };

  public handleValuesChange = (value: string) => {
    this.setState({ appValues: value, valuesModified: true });
  };

  private renderBasicForm = () => {
    return <div>Basic Form!</div>;
  };

  private renderAdvancedForm = () => {
    return (
      <div>
        <label htmlFor="values">Values (YAML)</label>
        <AceEditor
          mode="yaml"
          theme="xcode"
          name="values"
          width="100%"
          onChange={this.handleValuesChange}
          setOptions={{ showPrintMargin: false }}
          editorProps={{ $blockScrolling: Infinity }}
          value={this.state.appValues}
        />
      </div>
    );
  };

  private setBasicForm = (enable: boolean) => {
    // OnClick requires to return a function
    return () => {
      this.setState({ showBasicForm: enable });
    };
  };

  private renderTabs = () => {
    return (
      <div className="margin-t-normal">
        <div className="Tabs">
          <div className={`Tabs__Tab ${this.state.showBasicForm ? "Tabs__Tab-active" : ""}`}>
            <button type="button" onClick={this.setBasicForm(true)}>
              Basic
            </button>
          </div>
          <div className={`Tabs__Tab ${this.state.showBasicForm ? "" : "Tabs__Tab-active"}`}>
            <button type="button" onClick={this.setBasicForm(false)}>
              Advanced
            </button>
          </div>
        </div>
      </div>
    );
  };
}

export default DeploymentForm;
