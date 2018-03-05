import * as React from "react";
import AceEditor from "react-ace";
import { RouterAction } from "react-router-redux";

import { IServiceBinding } from "../../shared/ServiceBinding";
import { IChartState, IChartVersion } from "../../shared/types";

import "brace/mode/yaml";
import "brace/theme/xcode";

interface IAppNewProps {
  bindings: IServiceBinding[];
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
  getBindings: () => Promise<IServiceBinding[]>;
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
  selectedBinding: IServiceBinding | undefined;
}

class AppNew extends React.Component<IAppNewProps, IAppNewState> {
  public state: IAppNewState = {
    appValues: undefined,
    error: undefined,
    isDeploying: false,
    namespace: "default",
    releaseName: "",
    selectedBinding: undefined,
    valuesModified: false,
  };
  public componentDidMount() {
    const { chartID, getBindings, getChartVersion, getChartValues, chartVersion } = this.props;
    getBindings();
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
    const { bindings } = this.props;
    const { selectedBinding } = this.state;
    let bindingDetail = <div />;
    if (selectedBinding) {
      const {
        instanceRef,
        secretName,
        secretDatabase,
        secretHost,
        secretPassword,
        secretPort,
        secretUsername,
      } = selectedBinding.spec;

      const statuses: Array<[string, string | undefined]> = [
        ["Instance", instanceRef.name],
        ["Secret", secretName],
        ["Database", secretDatabase],
        ["Host", secretHost],
        ["Password", secretPassword],
        ["Port", secretPort],
        ["Username", secretUsername],
      ];

      bindingDetail = (
        <dl className="container margin-normal">
          {statuses.map(statusPair => {
            const [key, value] = statusPair;
            return [
              <dt key={key}>{key}</dt>,
              <dd key={value}>
                <code>{value}</code>
              </dd>,
            ];
          })}
        </dl>
      );
    }

    return (
      <div>
        {this.state.error && (
          <div className="container padding-v-bigger bg-action">{this.state.error}</div>
        )}
        <form className="container" onSubmit={this.handleDeploy}>
          <div className="row">
            <div className="col-8">
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
            </div>
            <div className="col-4">
              {bindings.length > 0 && (
                <div>
                  <p>[Optional] Select a service binding for your new app</p>
                  <label htmlFor="bindings">Bindings</label>
                  <select onChange={this.onBindingChange}>
                    <option key="none" value="none">
                      {" "}
                      -- Select one --
                    </option>
                    {bindings.map(b => (
                      <option
                        key={b.metadata.name}
                        selected={
                          b.metadata.name === (selectedBinding && selectedBinding.metadata.name)
                        }
                        value={b.metadata.name}
                      >
                        {b.metadata.name}
                      </option>
                    ))}
                  </select>
                  {bindingDetail}
                </div>
              )}
            </div>
          </div>
        </form>
      </div>
    );
  }

  public onBindingChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    this.setState({
      selectedBinding:
        this.props.bindings.find(binding => binding.metadata.name === e.target.value) || undefined,
    });
  };

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
