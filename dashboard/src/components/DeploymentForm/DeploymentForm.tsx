import { RouterAction } from "connected-react-router";
import * as Moniker from "moniker-native";
import * as React from "react";

import { JSONSchema4 } from "json-schema";
import {
  ForbiddenError,
  IChartState,
  IChartVersion,
  InternalServerError,
} from "../../shared/types";
import * as url from "../../shared/url";
import DeploymentFormBody from "../DeploymentFormBody/DeploymentFormBody";
import { ErrorSelector, UnexpectedErrorAlert } from "../ErrorAlert";
import LoadingWrapper from "../LoadingWrapper";

import "react-tabs/style/react-tabs.css";

export interface IDeploymentFormProps {
  kubeappsNamespace: string;
  chartNamespace: string;
  cluster: string;
  chartID: string;
  chartVersion: string;
  error: Error | undefined;
  chartsIsFetching: boolean;
  selected: IChartState["selected"];
  deployChart: (
    targetCluster: string,
    targetNamespace: string,
    version: IChartVersion,
    chartNamespace: string,
    releaseName: string,
    values?: string,
    schema?: JSONSchema4,
  ) => Promise<boolean>;
  push: (location: string) => RouterAction;
  fetchChartVersions: (namespace: string, id: string) => Promise<IChartVersion[]>;
  getChartVersion: (namespace: string, id: string, chartVersion: string) => void;
  namespace: string;
}

export interface IDeploymentFormState {
  isDeploying: boolean;
  releaseName: string;
  // Name of the release that was submitted for creation
  // This is different than releaseName since it is also used in the error banner
  // and we do not want to use releaseName since it is controller by the form field.
  latestSubmittedReleaseName: string;
  appValues: string;
  valuesModified: boolean;
}

class DeploymentForm extends React.Component<IDeploymentFormProps, IDeploymentFormState> {
  public state: IDeploymentFormState = {
    releaseName: Moniker.choose(),
    appValues: this.props.selected.values || "",
    isDeploying: false,
    latestSubmittedReleaseName: "",
    valuesModified: false,
  };

  public componentDidMount() {
    this.props.fetchChartVersions(this.props.chartNamespace, this.props.chartID);
  }

  public componentDidUpdate(prevProps: IDeploymentFormProps) {
    if (prevProps.selected.version !== this.props.selected.version && !this.state.valuesModified) {
      this.setState({ appValues: this.props.selected.values || "" });
    }
  }

  public render() {
    const { namespace, error } = this.props;
    if (error) {
      if (error.constructor === ForbiddenError) {
        // Only if the error is a ForbiddenError use the error selector
        // to parse the required roles
        return (
          <ErrorSelector
            error={error}
            namespace={namespace}
            action="create"
            resource={this.state.latestSubmittedReleaseName}
          />
        );
      }
      return (
        <UnexpectedErrorAlert
          title={`Sorry! The installation of ${this.state.latestSubmittedReleaseName} failed`}
          text={error.message}
          raw={true}
          showGenericMessage={error.constructor === InternalServerError}
        >
          {error.constructor === InternalServerError ? (
            <span>
              The server returned an internal error. If the problem persists, please contant
              Kubeapps maintainers.
            </span>
          ) : (
            <span>
              If you are unable to install the application, contact the chart maintainers or if you
              think the issue is related to Kubeapps, please open an{" "}
              <a
                href="https://github.com/kubeapps/kubeapps/issues/new"
                target="_blank"
                rel="noopener noreferrer"
              >
                issue in GitHub
              </a>
              .
            </span>
          )}
        </UnexpectedErrorAlert>
      );
    }
    if (this.state.isDeploying) {
      return <LoadingWrapper />;
    }
    return (
      <form className="container padding-b-bigger" onSubmit={this.handleDeploy}>
        <div className="row">
          <div className="col-12">
            <h2>{this.props.chartID}</h2>
          </div>
          <div className="col-8">
            <div>
              <label htmlFor="releaseName">Name</label>
              <input
                id="releaseName"
                pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
                title="Use lower case alphanumeric characters, '-' or '.'"
                onChange={this.handleReleaseNameChange}
                value={this.state.releaseName}
                required={true}
              />
            </div>
            <DeploymentFormBody
              deploymentEvent="install"
              chartNamespace={this.props.chartNamespace}
              cluster={this.props.cluster}
              chartID={this.props.chartID}
              chartVersion={this.props.chartVersion}
              chartsIsFetching={this.props.chartsIsFetching}
              namespace={this.props.namespace}
              selected={this.props.selected}
              push={this.props.push}
              getChartVersion={this.props.getChartVersion}
              setValues={this.handleValuesChange}
              appValues={this.state.appValues}
              setValuesModified={this.setValuesModified}
            />
          </div>
        </div>
      </form>
    );
  }

  public handleValuesChange = (value: string) => {
    this.setState({ appValues: value });
  };

  public setValuesModified = () => {
    this.setState({ valuesModified: true });
  };

  public handleDeploy = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const { chartNamespace, cluster, selected, deployChart, push, namespace } = this.props;
    const { releaseName, appValues } = this.state;

    this.setState({ isDeploying: true, latestSubmittedReleaseName: releaseName });
    if (selected.version) {
      const deployed = await deployChart(
        cluster,
        namespace,
        selected.version,
        chartNamespace,
        releaseName,
        appValues,
        selected.schema,
      );
      this.setState({ isDeploying: false });
      if (deployed) {
        push(url.app.apps.get(cluster, namespace, releaseName));
      }
    }
  };

  public handleReleaseNameChange = (e: React.FormEvent<HTMLInputElement>) => {
    this.setState({ releaseName: e.currentTarget.value });
  };
}

export default DeploymentForm;
