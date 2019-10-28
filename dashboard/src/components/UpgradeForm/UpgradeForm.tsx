import { RouterAction } from "connected-react-router";
import { JSONSchema4 } from "json-schema";
import * as React from "react";

import { IChartState, IChartVersion } from "../../shared/types";
import DeploymentFormBody from "../DeploymentFormBody/DeploymentFormBody";
import { ErrorSelector } from "../ErrorAlert";

export interface IUpgradeFormProps {
  appCurrentVersion: string;
  appCurrentValues?: string;
  chartName: string;
  namespace: string;
  releaseName: string;
  repo: string;
  error: Error | undefined;
  selected: IChartState["selected"];
  upgradeApp: (
    version: IChartVersion,
    releaseName: string,
    namespace: string,
    values?: string,
    schema?: JSONSchema4,
  ) => Promise<boolean>;
  push: (location: string) => RouterAction;
  goBack: () => RouterAction;
  fetchChartVersions: (id: string) => Promise<IChartVersion[]>;
  getChartVersion: (id: string, chartVersion: string) => void;
}

interface IUpgradeFormState {
  appValues: string;
  valuesModified: boolean;
  isDeploying: boolean;
}

class UpgradeForm extends React.Component<IUpgradeFormProps, IUpgradeFormState> {
  public state: IUpgradeFormState = {
    appValues: this.props.appCurrentValues || "",
    isDeploying: false,
    valuesModified: false,
  };

  public render() {
    const { namespace, releaseName, error } = this.props;
    if (error) {
      return (
        <ErrorSelector error={error} namespace={namespace} action="update" resource={releaseName} />
      );
    }

    const chartID = `${this.props.repo}/${this.props.chartName}`;
    return (
      <form className="container padding-b-bigger" onSubmit={this.handleDeploy}>
        <div className="row">
          <div className="col-12">
            <h2>{`${this.props.releaseName} (${chartID})`}</h2>
          </div>
          <div className="col-8">
            <DeploymentFormBody
              chartID={chartID}
              chartVersion={this.props.appCurrentVersion}
              originalValues={this.props.appCurrentValues}
              namespace={this.props.namespace}
              releaseName={this.props.releaseName}
              selected={this.props.selected}
              push={this.props.push}
              goBack={this.props.goBack}
              fetchChartVersions={this.props.fetchChartVersions}
              getChartVersion={this.props.getChartVersion}
              setValues={this.handleValuesChange}
              appValues={this.state.appValues}
              valuesModified={this.state.valuesModified}
              setValuesModified={this.setValuesModified}
            />
          </div>
        </div>
      </form>
    );
  }

  public setValuesModified = () => {
    this.setState({ valuesModified: true });
  };

  public handleValuesChange = (value: string) => {
    this.setState({ appValues: value });
  };

  public handleDeploy = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const { selected, push, upgradeApp, releaseName, namespace } = this.props;
    const { appValues } = this.state;

    this.setState({ isDeploying: true });
    if (selected.version) {
      const deployed = await upgradeApp(
        selected.version,
        releaseName,
        namespace,
        appValues,
        selected.schema,
      );
      this.setState({ isDeploying: false });
      if (deployed) {
        push(`/apps/ns/${namespace}/${releaseName}`);
      }
    }
  };
}

export default UpgradeForm;
