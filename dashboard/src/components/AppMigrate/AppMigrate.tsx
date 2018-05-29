import * as React from "react";

import { RouterAction } from "react-router-redux";
import { IApp, IChartState, IChartVersion } from "../../shared/types";
import { IAppRepository } from "../../shared/types";

import MigrateForm from "../../components/MigrateForm";

interface IAppMigrateProps {
  app: IApp;
  error: Error | undefined;
  namespace: string;
  helmCRDReleaseName: string;
  tillerReleaseName: string;
  repos: IAppRepository[];
  selected: IChartState["selected"];
  deployChart: (
    helmCRDReleaseName: string,
    version: IChartVersion,
    tillerReleaseName: string,
    namespace: string,
    values?: string,
    resourceVersion?: string,
  ) => Promise<boolean>;
  getApp: (tillerReleaseName: string, namespace: string) => Promise<void>;
  push: (location: string) => RouterAction;
  fetchRepositories: () => Promise<void>;
}

class AppMigrate extends React.Component<IAppMigrateProps> {
  public componentDidMount() {
    const { fetchRepositories, tillerReleaseName, getApp, namespace } = this.props;
    getApp(tillerReleaseName, namespace);
    fetchRepositories();
  }

  public componentWillReceiveProps(nextProps: IAppMigrateProps) {
    const { tillerReleaseName, getApp, namespace } = this.props;
    if (nextProps.namespace !== namespace) {
      getApp(tillerReleaseName, nextProps.namespace);
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
