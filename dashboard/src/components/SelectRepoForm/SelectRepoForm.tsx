import * as React from "react";
import { Link } from "react-router-dom";

import { IAppRepository, IRBACRole } from "../../shared/types";
import LoadingWrapper from "../LoadingWrapper";

import { ErrorSelector, MessageAlert } from "../ErrorAlert";

interface ISelectRepoFormProps {
  isFetching: boolean;
  namespace: string;
  kubeappsNamespace: string;
  repoError?: Error;
  error?: Error;
  repo: IAppRepository;
  repos: IAppRepository[];
  chartName: string;
  checkChart: (repo: string, chartName: string) => any;
  fetchRepositories: (namespace: string) => void;
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

  public componentDidMount() {
    this.props.fetchRepositories(this.props.namespace);
  }

  public render() {
    if (this.props.isFetching) {
      return <LoadingWrapper />;
    }
    if (this.props.repoError) {
      return (
        <ErrorSelector
          error={this.props.repoError}
          namespace={this.props.kubeappsNamespace}
          action="view"
          defaultRequiredRBACRoles={{ view: this.requiredRBACRoles() }}
          resource="App Repositories"
        />
      );
    }
    if (this.props.repos.length === 0) {
      return (
        <MessageAlert
          level={"warning"}
          children={
            <div>
              <h5>Chart repositories not found.</h5>
              Manage your Helm chart repositories in Kubeapps by visiting the{" "}
              <Link to={`/config/ns/${this.props.namespace}/repos`}>
                App repositories configuration
              </Link>{" "}
              page.
            </div>
          }
        />
      );
    }
    return (
      <LoadingWrapper loaded={this.props.repos.length > 0}>
        <div className="container margin-normal">
          <div className="col-8">
            {this.props.error && (
              <ErrorSelector
                error={this.props.error}
                resource={`Chart ${this.state.repo}/${this.props.chartName}`}
              />
            )}
          </div>
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
      </LoadingWrapper>
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

  private requiredRBACRoles(): IRBACRole[] {
    return [
      {
        apiGroup: "kubeapps.com",
        namespace: this.props.kubeappsNamespace,
        resource: "apprepositories",
        verbs: ["get"],
      },
    ];
  }
}

export default SelectRepoForm;
