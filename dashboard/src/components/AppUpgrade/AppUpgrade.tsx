import * as React from "react";

import { RouterAction } from "react-router-redux";
import { hapi } from "../../shared/hapi/release";
import { IServiceBindingWithSecret } from "../../shared/ServiceBinding";
import { IAppRepository, IChartState, IChartVersion } from "../../shared/types";
import DeploymentErrors from "../DeploymentForm/DeploymentErrors";
import LoadingWrapper from "../LoadingWrapper";
import UpgradeForm from "../UpgradeForm";
import SelectRepoForm from "../UpgradeForm/SelectRepoForm";

interface IAppUpgradeProps {
  app: hapi.release.Release;
  bindingsWithSecrets: IServiceBindingWithSecret[];
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
  getApp: (releaseName: string, namespace: string) => void;
  getBindings: (ns: string) => void;
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
    const { releaseName, getApp, namespace, fetchRepositories } = this.props;
    getApp(releaseName, namespace);
    fetchRepositories();
  }

  public render() {
    const { app, repos, error, repo } = this.props;
    if (
      !repos ||
      !app ||
      !app.chart ||
      !app.chart.metadata ||
      !app.chart.metadata.name ||
      !app.chart.metadata.version
    ) {
      if (!error) {
        return <LoadingWrapper />;
      } else {
        return (
          <DeploymentErrors
            {...this.props}
            chartName={(app.chart && app.chart.metadata && app.chart.metadata.name) || ""}
            repo={repo.metadata.name}
          />
        );
      }
    }
    if (!this.props.repo.metadata) {
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
          repo={this.props.repo.metadata.name}
        />
      </div>
    );
  }
}

export default AppUpgrade;
