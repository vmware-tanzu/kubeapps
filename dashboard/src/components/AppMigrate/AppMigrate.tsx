import * as React from "react";

import { RouterAction } from "react-router-redux";
import { IApp, IChartState, IChartVersion } from "../../shared/types";
import { IAppRepository } from "../../shared/types";

import MigrateForm from "../../components/MigrateForm";

interface IAppMigrateProps {
  app: IApp;
  error: Error | undefined;
  namespace: string;
  releaseName: string;
  repos: IAppRepository[];
  selected: IChartState["selected"];
  migrateApp: (
    version: IChartVersion,
    releaseName: string,
    namespace: string,
    values?: string,
  ) => Promise<boolean>;
  getApp: (releaseName: string, namespace: string) => Promise<void>;
  push: (location: string) => RouterAction;
  fetchRepositories: () => Promise<void>;
}

class AppMigrate extends React.Component<IAppMigrateProps> {
  public componentDidMount() {
    const { fetchRepositories, releaseName, getApp, namespace } = this.props;
    getApp(releaseName, namespace);
    fetchRepositories();
  }

  public componentWillReceiveProps(nextProps: IAppMigrateProps) {
    const { releaseName, getApp, namespace } = this.props;
    if (nextProps.namespace !== namespace) {
      getApp(releaseName, nextProps.namespace);
    }
  }

  public render() {
    const { app, repos } = this.props;
    if (
      !repos ||
      !app ||
      !app.data ||
      !app.data.chart ||
      !app.data.chart.metadata ||
      !app.data.chart.metadata.version ||
      !app.data.chart.values
    ) {
      return <div>Loading</div>;
    }
    return (
      <div>
        <MigrateForm
          {...this.props}
          chartID={app.data.name}
          chartVersion={app.data.chart.metadata.version}
          chartValues={app.data.chart.values.raw}
          chartName={app.data.chart.metadata.name || ""}
          chartRepoName=""
          chartRepoURL=""
          chartRepoAuth=""
        />
      </div>
    );
  }
}

export default AppMigrate;
