import * as React from "react";
import AceEditor from "react-ace";
import { RouterAction } from "react-router-redux";
import { IAppRepository } from "../../shared/types";

import {
  ForbiddenError,
  IChartAttributes,
  IChartVersion,
  IRBACRole,
  IRepo,
  MissingChart,
  NotFoundError,
} from "../../shared/types";
import { NotFoundErrorAlert, PermissionsErrorAlert, UnexpectedErrorAlert } from "../ErrorAlert";

import "brace/mode/yaml";
import "brace/theme/xcode";

const RequiredRBACRoles: IRBACRole[] = [
  {
    apiGroup: "helm.bitnami.com",
    resource: "helmreleases",
    verbs: ["create", "patch"],
  },
  {
    apiGroup: "kubeapps.com",
    namespace: "kubeapps",
    resource: "apprepositories",
    verbs: ["get"],
  },
];

interface IMigrationFormProps {
  chartID: string;
  chartVersion: string;
  error: Error | undefined;
  migrateApp: (
    version: IChartVersion,
    releaseName: string,
    namespace: string,
    values?: string,
  ) => Promise<boolean>;
  push: (location: string) => RouterAction;
  namespace: string;
  releaseName: string;
  chartValues: string | null | undefined;
  chartName: string;
  chartRepoAuth: {};
  chartRepoName: string;
  chartRepoURL: string;
  repos: IAppRepository[];
}

interface IMigrationtFormState {
  isDeploying: boolean;
  releaseName: string;
  chartValues: string;
  chartVersion: string;
  namespace: string;
  chartName: string;
  chartRepoName: string;
  chartRepoURL: string;
  chartRepoAuth: {};
  repos: IAppRepository[];
}

class MigrateForm extends React.Component<IMigrationFormProps, IMigrationtFormState> {
  public state: IMigrationtFormState = {
    chartName: this.props.chartName,
    chartRepoAuth: {},
    chartRepoName: this.props.chartRepoName,
    chartRepoURL: "",
    chartValues: this.props.chartValues || "",
    chartVersion: this.props.chartVersion,
    isDeploying: false,
    namespace: this.props.namespace,
    releaseName: this.props.releaseName,
    repos: this.props.repos,
  };

  public render() {
    return (
      <div>
        <form className="container padding-b-bigger" onSubmit={this.handleDeploy}>
          <div className="row">
            <div className="col-8">{this.props.error && this.renderError()}</div>
            <div className="col-12">
              <h2>{this.props.chartID}</h2>
            </div>
            <div className="col-12">
              <p>
                In order to be able to manage {this.state.releaseName} select the repository it can
                be retrieved from.
              </p>
            </div>
            <div className="col-8">
              <div>
                <label htmlFor="chartRepoName">Chart Repository Name *</label>
                <select
                  id="chartRepoName"
                  onChange={this.handleChartRepoNameChange}
                  value={this.state.chartRepoName}
                  required={true}
                >
                  {this.state.chartRepoName === "" && <option key="" value="" />}
                  {this.props.repos.map(r => (
                    <option key={r.metadata.name} value={r.metadata.name}>
                      {r.metadata.name}
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label htmlFor="chartRepoURL">Chart Repository URL</label>
                <input
                  id="chartRepoURL"
                  value={
                    this.state.chartRepoURL === ""
                      ? "(Select a repository name)"
                      : this.state.chartRepoURL
                  }
                  required={true}
                  disabled={true}
                />
              </div>
              <div>
                <p>
                  {" "}
                  * If the repository containing this chart is not in the list add it{" "}
                  <a href="/config/repos"> here </a>{" "}
                </p>
              </div>
              <div>
                <label htmlFor="releaseName">Release Name</label>
                <input
                  id="releaseName"
                  value={this.state.releaseName}
                  required={true}
                  disabled={true}
                />
              </div>
              <div>
                <label htmlFor="chartName">Chart Name</label>
                <input
                  id="chartName"
                  value={this.state.chartName}
                  required={true}
                  disabled={true}
                />
                <div>
                  <label htmlFor="chartVersion">Chart Version</label>
                  <input
                    id="chartVersion"
                    value={this.state.chartVersion}
                    required={true}
                    disabled={true}
                  />
                </div>
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
                  value={this.state.chartValues}
                />
              </div>
              <div>
                <button className="button button-primary" type="submit">
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
    const chartRepo = {
      auth: this.state.chartRepoAuth,
      name: this.state.chartRepoName,
      url: this.state.chartRepoURL,
    } as IRepo;
    const chartData = {
      name: this.props.chartName,
      repo: chartRepo,
    } as IChartAttributes;
    const version = {
      attributes: {
        version: this.props.chartVersion,
      },
      id: this.props.chartVersion,
      relationships: {
        chart: {
          data: chartData,
        },
      },
    } as IChartVersion;
    const { releaseName, namespace } = this.props;
    const deployed = await this.props.migrateApp(
      version,
      releaseName,
      namespace,
      this.props.chartValues || "",
    );
    if (deployed) {
      this.props.push(`/apps/ns/${namespace}/${releaseName}`);
    } else {
      this.setState({ isDeploying: false });
    }
  };

  public handleChartRepoNameChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    let repoURL = "";
    let auth = {};
    this.state.repos.forEach(r => {
      if (r.metadata.name === e.currentTarget.value && r.spec) {
        repoURL = r.spec.url;
        auth = r.spec.auth;
      }
    });
    this.setState({
      chartRepoAuth: auth,
      chartRepoName: e.currentTarget.value,
      chartRepoURL: repoURL,
    });
  };
  public handleChartRepoURLChange = (e: React.FormEvent<HTMLInputElement>) => {
    this.setState({ chartRepoURL: e.currentTarget.value });
  };
  public handleValuesChange = (value: string) => {
    this.setState({ chartValues: value });
  };

  private renderError() {
    const { error, namespace } = this.props;
    const { releaseName } = this.state;
    const roles = RequiredRBACRoles;
    roles[0].verbs = ["create"];
    switch (error && error.constructor) {
      case MissingChart:
        return (
          <NotFoundErrorAlert
            header={`Chart not found in the given repository. Please choose the repository that contains the chart.`}
          />
        );
      case ForbiddenError:
        return (
          <PermissionsErrorAlert
            namespace={namespace}
            roles={roles}
            action={`Create Application "${releaseName}"`}
          />
        );
      case NotFoundError:
        return (
          <NotFoundErrorAlert resource={`Application "${releaseName}"`} namespace={namespace} />
        );
      default:
        return <UnexpectedErrorAlert />;
    }
  }
}

export default MigrateForm;
