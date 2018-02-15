import * as React from "react";
import AceEditor from "react-ace";
import { RouterAction } from "react-router-redux";
import { IChartState, IChartVersion } from "../../shared/types";

import "brace/mode/yaml";
import "brace/theme/xcode";

interface IAppNewProps {
  chartID: string;
  deployChart: (
    version: IChartVersion,
    releaseName: string,
    namespace: string,
    values?: string,
  ) => Promise<{}>;
  selected: IChartState["selected"];
  chartVersion: string;
  push: (location: string) => RouterAction;
  getChartVersion: (id: string, chartVersion: string) => Promise<{}>;
  getChartValues: (id: string, chartVersion: string) => Promise<{}>;
}

interface IAppNewState {
  isDeploying: boolean;
  // deployment options
  releaseName: string;
  namespace: string;
  appValues?: string;
  valuesModified: boolean;
  error?: string;
}

class AppNew extends React.Component<IAppNewProps, IAppNewState> {
  public state: IAppNewState = {
    appValues: undefined,
    error: undefined,
    isDeploying: false,
    namespace: "default",
    releaseName: "",
    valuesModified: false,
  };
  public componentDidMount() {
    const { chartID, getChartVersion, getChartValues, chartVersion } = this.props;
    getChartVersion(chartID, chartVersion);
    getChartValues(chartID, chartVersion);
  }

  public componentWillReceiveProps(nextProps: IAppNewProps) {
    const { version, values } = nextProps.selected;
    if (version && values && !this.state.valuesModified) {
      this.setState({ appValues: values });
    }
  }

  public render() {
    if (!this.props.selected.version && !this.state.appValues) {
      return <div>Loading</div>;
    }
    return (
      <div>
        {this.state.error && (
          <div className="container padding-v-bigger bg-action">{this.state.error}</div>
        )}
        <form onSubmit={this.handleDeploy}>
          <div>
            <label htmlFor="releaseName">Name</label>
            <input
              id="releaseName"
              onChange={this.handleReleaseNameChange}
              value={this.state.releaseName}
              required={true}
            />
          </div>
          <div>
            <label htmlFor="namespace">Namespace</label>
            <input
              name="namespace"
              onChange={this.handleNamespaceChange}
              value={this.state.namespace}
            />
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
              value={this.state.appValues}
            />
          </div>
          <div>
            <button className="button button-primary" type="submit">
              Submit
            </button>
          </div>
        </form>
      </div>
    );
  }

  public handleDeploy = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const { selected, deployChart, push } = this.props;
    this.setState({ isDeploying: true });
    const { releaseName, namespace, appValues } = this.state;
    if (selected.version && appValues) {
      deployChart(selected.version, releaseName, namespace, appValues)
        .then(() => push(`/apps/${namespace}/${namespace}-${releaseName}`))
        .catch(err => this.setState({ isDeploying: false, error: err.toString() }));
    }
  };

  public handleReleaseNameChange = (e: React.FormEvent<HTMLInputElement>) => {
    this.setState({ releaseName: e.currentTarget.value });
  };
  public handleNamespaceChange = (e: React.FormEvent<HTMLInputElement>) => {
    this.setState({ namespace: e.currentTarget.value });
  };
  public handleValuesChange = (value: string) => {
    this.setState({ appValues: value, valuesModified: true });
  };
}

export default AppNew;
