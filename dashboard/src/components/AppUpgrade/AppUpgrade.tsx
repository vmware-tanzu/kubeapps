import * as React from "react";

import { RouterAction } from "react-router-redux";
import { hapi } from "../../shared/hapi/release";
import { IServiceBinding } from "../../shared/ServiceBinding";
import { IAppRepository, IChartState, IChartVersion } from "../../shared/types";
import * as errors from "../DeploymentForm/errors";
import SelectRepoForm from "../SelectRepoForm";
import UpgradeForm from "../UpgradeForm";

interface IAppUpgradeProps {
  app: hapi.release.Release;
  bindings: IServiceBinding[];
  error: Error | undefined;
  repoError: Error | undefined;
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
  clearRepo: () => any;
  checkChart: (repo: string, chartName: string) => any;
  fetchChartVersions: (id: string) => Promise<{}>;
  getApp: (releaseName: string, namespace: string) => Promise<void>;
  getBindings: () => Promise<IServiceBinding[]>;
  getChartVersion: (id: string, chartVersion: string) => Promise<{}>;
  getChartValues: (id: string, chartVersion: string) => Promise<any>;
  push: (location: string) => RouterAction;
  fetchRepositories: () => Promise<void>;
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
    const { app, repos, error, releaseName, namespace } = this.props;
    if (
      !repos ||
      !app ||
      !app.chart ||
      !app.chart.metadata ||
      !app.chart.metadata.name ||
      !app.chart.metadata.version
    ) {
      if (!error) {
        return <div>Loading</div>;
      } else {
        return <div className="col-8">{errors.render(error, releaseName, namespace)}</div>;
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
          chartName={app.chart.metadata.name}
          repo={this.props.repo.metadata.name}
        />
      </div>
    );
  }
}

export default AppUpgrade;
