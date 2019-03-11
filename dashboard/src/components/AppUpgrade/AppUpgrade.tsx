import * as React from "react";

import { RouterAction } from "connected-react-router";
import {
  IAppRepository,
  IChartState,
  IChartVersion,
  IRBACRole,
  IRelease,
} from "../../shared/types";
import { ErrorSelector } from "../ErrorAlert";
import LoadingWrapper from "../LoadingWrapper";
import UpgradeForm from "../UpgradeForm";
import SelectRepoForm from "../UpgradeForm/SelectRepoForm";

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
  upgradeApp: (
    version: IChartVersion,
    releaseName: string,
    namespace: string,
    values?: string,
  ) => Promise<boolean>;
  clearRepo: () => void;
  checkChart: (repo: string, chartName: string) => void;
  fetchChartVersions: (id: string) => Promise<IChartVersion[]>;
  getAppWithUpdateInfo: (releaseName: string, namespace: string) => void;
  getChartVersion: (id: string, chartVersion: string) => void;
  getChartValues: (id: string, chartVersion: string) => void;
  push: (location: string) => RouterAction;
  fetchRepositories: () => void;
}

interface IAppUpgradeState {
  selectRepoForm: JSX.Element;
}

class AppUpgrade extends React.Component<IAppUpgradeProps, IAppUpgradeState> {
  public componentDidMount() {
    const { releaseName, getAppWithUpdateInfo, namespace, fetchRepositories } = this.props;
    getAppWithUpdateInfo(releaseName, namespace);
    fetchRepositories();
  }

  public render() {
    const { app, repos, error, namespace, releaseName, repoError, isFetching } = this.props;
    let { repo } = this.props;
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
          <ErrorSelector
            error={new Error("Unable to any repo to upgrade from")}
            resource={releaseName}
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
    if (app.updateInfo) {
      const repoWithLatest = repos.find(
        r => r.metadata.name === (app.updateInfo && app.updateInfo.repository.name),
      );
      if (repoWithLatest) {
        repo = repoWithLatest;
      }
    }
    if (!repo.metadata) {
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
