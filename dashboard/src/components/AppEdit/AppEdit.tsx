import * as React from "react";

import { RouterAction } from "react-router-redux";
import { IServiceBinding } from "../../shared/ServiceBinding";
import { IApp, IChartState, IChartVersion } from "../../shared/types";

import DeploymentForm from "../../components/DeploymentForm";

interface IAppEditProps {
  app: IApp;
  bindings: IServiceBinding[];
  namespace: string;
  releaseName: string;
  selected: IChartState["selected"];
  version: string;
  deployChart: (
    version: IChartVersion,
    releaseName: string,
    namespace: string,
    values?: string,
    resourceVersion?: string,
  ) => Promise<{}>;
  fetchChartVersions: (id: string) => Promise<{}>;
  getApp: (releaseName: string, namespace: string) => Promise<void>;
  getBindings: () => Promise<IServiceBinding[]>;
  getChartVersion: (id: string, chartVersion: string) => Promise<{}>;
  getChartValues: (id: string, chartVersion: string) => Promise<any>;
  push: (location: string) => RouterAction;
  selectChartVersionAndGetFiles: (version: IChartVersion) => Promise<{}>;
}

class AppEdit extends React.Component<IAppEditProps> {
  public componentDidMount() {
    const { releaseName, getApp, namespace } = this.props;
    getApp(releaseName, namespace);
  }

  public render() {
    const { app } = this.props;

    if (!app || !app.hr || !app.chart) {
      return <div>Loading</div>;
    }

    return (
      <div>
        <DeploymentForm
          {...this.props}
          hr={app.hr}
          chartID={app.chart.id}
          chartVersion={this.props.version ? this.props.version : app.hr.spec.version}
        />
      </div>
    );
  }
}

export default AppEdit;
