import * as React from "react";

import { RouterAction } from "react-router-redux";
import { IServiceBinding } from "../../shared/ServiceBinding";
import { IApp, IChartState, IChartVersion } from "../../shared/types";

import DeploymentForm from "../../components/DeploymentForm";

interface IAppEditProps {
  app: IApp;
  bindings: IServiceBinding[];
  error: Error | undefined;
  helmCRDReleaseName: string;
  namespace: string;
  tillerReleaseName: string;
  selected: IChartState["selected"];
  deployChart: (
    helmCRDReleaseName: string,
    version: IChartVersion,
    tillerReleaseName: string,
    namespace: string,
    values?: string,
    resourceVersion?: string,
  ) => Promise<boolean>;
  fetchChartVersions: (id: string) => Promise<{}>;
  getApp: (
    helmCRDReleaseName: string,
    tillerReleaseName: string,
    namespace: string,
  ) => Promise<void>;
  getBindings: () => Promise<IServiceBinding[]>;
  getChartVersion: (id: string, chartVersion: string) => Promise<{}>;
  getChartValues: (id: string, chartVersion: string) => Promise<any>;
  push: (location: string) => RouterAction;
}

class AppEdit extends React.Component<IAppEditProps> {
  public componentDidMount() {
    const { helmCRDReleaseName, tillerReleaseName, getApp, namespace } = this.props;
    getApp(helmCRDReleaseName, tillerReleaseName, namespace);
  }

  public componentWillReceiveProps(nextProps: IAppEditProps) {
    const { helmCRDReleaseName, tillerReleaseName, getApp, namespace } = this.props;
    if (nextProps.namespace !== namespace) {
      getApp(helmCRDReleaseName, tillerReleaseName, nextProps.namespace);
    }
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
          chartVersion={app.hr.spec.version}
        />
      </div>
    );
  }
}

export default AppEdit;
