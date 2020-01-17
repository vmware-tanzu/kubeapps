import * as React from "react";

import { RouterAction } from "connected-react-router";
import { JSONSchema4 } from "json-schema";
import { Link } from "react-router-dom";
import {
  IAppRepository,
  IChartState,
  IChartVersion,
  IRBACRole,
  IRelease,
} from "../../shared/types";
import { ErrorSelector, MessageAlert } from "../ErrorAlert";
import LoadingWrapper from "../LoadingWrapper";
import SelectRepoForm from "../SelectRepoForm";
import UpgradeForm from "../UpgradeForm";

interface IAppUpgradeProps {
  app: IRelease;
  isFetching: boolean;
  error: Error | undefined;
  repoError: Error | undefined;
  kubeappsNamespace: string;
  namespace: string;
  releaseName: string;
  version: string;
  repos: IAppRepository[];
  repo: IAppRepository;
  selected: IChartState["selected"];
  deployed: IChartState["deployed"];
  upgradeApp: (
    version: IChartVersion,
    releaseName: string,
    namespace: string,
    values?: string,
    schema?: JSONSchema4,
  ) => Promise<boolean>;
  clearRepo: () => void;
  checkChart: (repo: string, chartName: string) => void;
  fetchChartVersions: (id: string) => Promise<IChartVersion[]>;
  getAppWithUpdateInfo: (releaseName: string, namespace: string) => void;
  getChartVersion: (id: string, chartVersion: string) => void;
  getDeployedChartVersion: (id: string, chartVersion: string) => void;
  push: (location: string) => RouterAction;
  goBack: () => RouterAction;
  fetchRepositories: () => void;
}

interface IAppUpgradeState {
  repo?: IAppRepository;
}

class AppUpgrade extends React.Component<IAppUpgradeProps, IAppUpgradeState> {
  public state: IAppUpgradeState = {};

  public componentDidMount() {
    const { releaseName, getAppWithUpdateInfo, namespace, fetchRepositories } = this.props;
    getAppWithUpdateInfo(releaseName, namespace);
    fetchRepositories();
  }

  public componentDidUpdate(prevProps: IAppUpgradeProps) {
    const { repos, app } = this.props;
    let repo = this.state.repo;
    // Retrieve the current repo
    if (!repo) {
      repo = this.props.repo;
      if (repo && repo.metadata) {
        // If the repository comes from the properties, use it
        this.setState({ repo });
      } else {
        // If there is update info about the app we can automatically chose the repository
        // with the latest version
        if (app && app.updateInfo) {
          const repoWithLatest = repos.find(
            r => r.metadata.name === (app.updateInfo && app.updateInfo.repository.name),
          );
          if (repoWithLatest) {
            this.setState({ repo: repoWithLatest });
            repo = repoWithLatest;
          }
        }
      }
    }

    if (app) {
      const { chart } = app;
      if (
        chart &&
        chart.metadata &&
        chart.metadata.name &&
        chart.metadata.version &&
        repo &&
        repo.metadata &&
        repo.metadata.name &&
        (prevProps.app !== app || this.state.repo !== repo)
      ) {
        const chartID = `${repo.metadata.name}/${chart.metadata.name}`;
        this.props.getDeployedChartVersion(chartID, chart.metadata.version);
      }
    }
  }

  public render() {
    const { app, repos, error, namespace, releaseName, repoError, isFetching } = this.props;
    const { repo } = this.state;
    if (isFetching) {
      return <LoadingWrapper />;
    }
    if (
      !repos ||
      repos.length === 0 ||
      !app ||
      !app.chart ||
      !app.chart.metadata ||
      !app.chart.metadata.name ||
      !app.chart.metadata.version
    ) {
      if (error) {
        return (
          <ErrorSelector
            error={error}
            namespace={namespace}
            action="update"
            defaultRequiredRBACRoles={{ update: this.requiredRBACRoles() }}
            resource={releaseName}
          />
        );
      } else if (repoError) {
        return (
          <ErrorSelector
            error={repoError}
            namespace={this.props.kubeappsNamespace}
            action="view"
            defaultRequiredRBACRoles={{ view: this.requiredRBACRoles() }}
            resource="App Repositories"
          />
        );
      } else if (repos.length === 0) {
        return (
          <MessageAlert
            level={"warning"}
            children={
              <div>
                <h5>Chart repositories not found.</h5>
                Manage your Helm chart repositories in Kubeapps by visiting the{" "}
                <Link to={"/config/repos"}>App repositories configuration</Link> page.
              </div>
            }
          />
        );
      } else {
        return (
          <ErrorSelector
            error={new Error("Unable to obtain the required information to upgrade")}
            resource={releaseName}
          />
        );
      }
    }
    if (!repo || !repo.metadata) {
      return (
        <div>
          <SelectRepoForm
            {...this.props}
            error={this.props.repoError}
            chartName={app.chart.metadata.name}
          />
        </div>
      );
    }
    return (
      <div>
        <UpgradeForm
          {...this.props}
          appCurrentVersion={app.chart.metadata.version}
          appCurrentValues={(app.config && app.config.raw) || ""}
          chartName={app.chart.metadata.name}
          repo={repo.metadata.name}
        />
      </div>
    );
  }

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

export default AppUpgrade;
