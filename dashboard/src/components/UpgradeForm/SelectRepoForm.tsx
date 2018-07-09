import * as React from "react";

import { ForbiddenError, IAppRepository, IRBACRole, MissingChart } from "../../shared/types";

import { NotFoundErrorAlert, PermissionsErrorAlert, UnexpectedErrorAlert } from "../ErrorAlert";

import "brace/mode/yaml";
import "brace/theme/xcode";

const RequiredRBACRoles: IRBACRole = {
  apiGroup: "kubeapps.com",
  namespace: "kubeapps",
  resource: "apprepositories",
  verbs: ["get"],
};

interface ISelectRepoFormProps {
  error: Error | undefined;
  repo: IAppRepository;
  repos: IAppRepository[];
  chartName: string;
  checkChart: (repo: string, chartName: string) => any;
}

interface ISelectRepoFormState {
  repo: string;
}

class SelectRepoForm extends React.Component<ISelectRepoFormProps, ISelectRepoFormState> {
  public state: ISelectRepoFormState = {
    repo:
      this.props.repo.metadata && this.props.repo.metadata.name
        ? this.props.repo.metadata.name
        : "",
  };

  public render() {
    if (!this.props.repos) {
      return <div> Loading </div>;
    }
    return (
      <div className="container margin-normal">
        <div className="col-8">{this.props.error && this.renderError()}</div>
        <div className="col-12">
          <h2>Select the source repository of {this.props.chartName}</h2>
        </div>
        <div className="col-8">
          <label htmlFor="chartRepoName">Chart Repository Name *</label>
          <select
            id="chartRepoName"
            onChange={this.handleChartRepoNameChange}
            value={this.state.repo}
            required={true}
          >
            {!this.state.repo && <option key="" value="" />}
            {this.props.repos.map(r => (
              <option key={r.metadata.name} value={r.metadata.name}>
                {r.metadata.name} ({this.getRepoURL(r.metadata.name)})
              </option>
            ))}
          </select>
        </div>
        <div>
          <p>
            {" "}
            * If the repository containing {this.props.chartName} is not in the list add it{" "}
            <a href="/config/repos"> here </a>.{" "}
          </p>
        </div>
      </div>
    );
  }

  public handleChartRepoNameChange = async (e: React.ChangeEvent<HTMLSelectElement>) => {
    this.props.checkChart(e.currentTarget.value, this.props.chartName);
    this.setState({ repo: e.currentTarget.value });
  };

  private getRepoURL = (name: string) => {
    let res = "";
    this.props.repos.forEach(r => {
      if (r.metadata.name === name && r.spec) {
        res = r.spec.url;
      }
    });
    return res;
  };

  private renderError() {
    if (this.props.error) {
      switch (this.props.error.constructor) {
        case MissingChart:
          return <NotFoundErrorAlert header={this.props.error.message} />;
        case ForbiddenError:
          return (
            <PermissionsErrorAlert
              namespace={RequiredRBACRoles.namespace || ""}
              roles={[RequiredRBACRoles]}
              action={`list app repositories`}
            />
          );
      }
    }
    return <UnexpectedErrorAlert />;
  }
}

export default SelectRepoForm;
