import * as React from "react";
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
  chartRepoName: string;
  chartRepoURL: string;
  chartRepoAuth: {};
}

class MigrateForm extends React.Component<IMigrationFormProps, IMigrationtFormState> {
  public state: IMigrationtFormState = {
    chartRepoAuth: {},
    chartRepoName: this.props.chartRepoName,
    chartRepoURL: "",
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
                In order to be able to manage {this.props.releaseName} select the repository it can
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
                  * If the repository containing {this.props.chartName} is not in the list add it{" "}
                  <a href="/config/repos"> here </a>.{" "}
                </p>
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
    }
  };

  public handleChartRepoNameChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    let repoURL = "";
    let auth = {};
    this.props.repos.forEach(r => {
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

  private renderError() {
    const { error, namespace, releaseName } = this.props;
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
