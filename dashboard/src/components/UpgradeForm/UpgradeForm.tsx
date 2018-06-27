import * as React from "react";
import AceEditor from "react-ace";
import { RouterAction } from "react-router-redux";

import { IServiceBinding } from "../../shared/ServiceBinding";
import { IChartState, IChartVersion } from "../../shared/types";
import DeploymentBinding from "../DeploymentForm/DeploymentBinding";
import DeploymentErrors from "../DeploymentForm/DeploymentErrors";

import "brace/mode/yaml";
import "brace/theme/xcode";

interface IDeploymentFormProps {
  appCurrentVersion: string;
  bindings: IServiceBinding[];
  chartName: string;
  namespace: string;
  releaseName: string;
  repo: string;
  error: Error | undefined;
  selected: IChartState["selected"];
  upgradeApp: (
    version: IChartVersion,
    releaseName: string,
    namespace: string,
    values?: string,
  ) => Promise<boolean>;
  push: (location: string) => RouterAction;
  fetchChartVersions: (id: string) => Promise<{}>;
  getBindings: (ns: string) => Promise<IServiceBinding[]>;
  getChartVersion: (id: string, chartVersion: string) => Promise<{}>;
  getChartValues: (id: string, chartVersion: string) => Promise<any>;
  clearRepo: () => any;
}

interface IDeploymentFormState {
  isDeploying: boolean;
  // deployment options
  appValues?: string;
  valuesModified: boolean;
}

class UpgradeForm extends React.Component<IDeploymentFormProps, IDeploymentFormState> {
  public state: IDeploymentFormState = {
    appValues: undefined,
    isDeploying: false,
    valuesModified: false,
  };

  public componentDidMount() {
    const {
      appCurrentVersion,
      chartName,
      fetchChartVersions,
      getChartValues,
      getChartVersion,
      repo,
    } = this.props;
    const chartID = `${repo}/${chartName}`;
    fetchChartVersions(chartID);
    getChartVersion(chartID, appCurrentVersion);
    getChartValues(chartID, appCurrentVersion);
  }

  public componentDidUpdate(prevProps: IDeploymentFormProps) {
    const { selected } = this.props;
    if (selected.values && !this.state.appValues && !this.state.valuesModified) {
      // First load, set initial values
      this.setState({ appValues: selected.values });
    }
    if (selected.values && this.state.appValues && selected.values !== this.state.appValues) {
      // Values has been modified either because the user has edit them
      // or because the selected version is now different
      if (!this.state.valuesModified) {
        // Only update the default values if the user has not modify them
        this.setState({ appValues: selected.values });
      }
    }
  }

  public render() {
    const { selected, bindings } = this.props;
    const { version, versions } = selected;
    const { appValues } = this.state;
    if (!version || !versions || !versions.length || this.state.isDeploying) {
      return <div> Loading </div>;
    }
    return (
      <div>
        <form className="container padding-b-bigger" onSubmit={this.handleDeploy}>
          <div className="row">
            <div className="col-8">{this.props.error && <DeploymentErrors {...this.props} />}</div>
            <div className="col-12">
              <h2>
                {this.props.releaseName} ({this.props.chartName})
              </h2>
            </div>
            <div className="col-8">
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
                      {v.attributes.version === this.props.appCurrentVersion ? "(current)" : ""}
                    </option>
                  ))}
                </select>
              </div>
              <div style={{ marginBottom: "1em" }}>
                <label htmlFor="values">Values (YAML)</label>
                <AceEditor
                  mode="yaml"
                  theme="xcode"
                  name="values"
                  width="100%"
                  onChange={this.handleValuesChange}
                  setOptions={{ showPrintMargin: false }}
                  editorProps={{ $blockScrolling: Infinity }}
                  value={appValues}
                />
              </div>
              <div>
                <button className="button button-primary" type="submit">
                  Submit
                </button>
                <button className="button" onClick={this.handleReselectChartRepo}>
                  Select Chart repo
                </button>
              </div>
            </div>
            <div className="col-4">
              {bindings.length > 0 && <DeploymentBinding {...this.props} />}
            </div>
          </div>
        </form>
      </div>
    );
  }

  public handleDeploy = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const { releaseName, namespace, selected, upgradeApp, push } = this.props;
    this.setState({ isDeploying: true });
    const { appValues } = this.state;
    if (selected.version) {
      const deployed = await upgradeApp(selected.version, releaseName, namespace, appValues);
      this.setState({ isDeploying: false });
      if (deployed) {
        push(`/apps/ns/${namespace}/${releaseName}`);
      }
    }
  };

  public handleChartVersionChange = (e: React.FormEvent<HTMLSelectElement>) => {
    const { repo, chartName, getChartVersion, getChartValues } = this.props;
    const chartID = `${repo}/${chartName}`;
    getChartVersion(chartID, e.currentTarget.value);
    if (!this.state.valuesModified) {
      // Only update the default values if the user has not modify them
      getChartValues(chartID, e.currentTarget.value);
    }
  };

  public handleValuesChange = (value: string) => {
    this.setState({ appValues: value, valuesModified: true });
  };

  public handleReselectChartRepo = () => {
    this.props.clearRepo();
  };
}

export default UpgradeForm;
