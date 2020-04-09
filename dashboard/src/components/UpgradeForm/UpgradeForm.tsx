import { RouterAction } from "connected-react-router";
import * as jsonpatch from "fast-json-patch";
import { JSONSchema4 } from "json-schema";
import * as React from "react";
import * as YAML from "yaml";

import { deleteValue, setValue } from "../../shared/schema";
import { IChartState, IChartVersion } from "../../shared/types";
import DeploymentFormBody from "../DeploymentFormBody/DeploymentFormBody";
import { ErrorSelector } from "../ErrorAlert";
import LoadingWrapper from "../LoadingWrapper";

export interface IUpgradeFormProps {
  appCurrentVersion: string;
  appCurrentValues?: string;
  chartName: string;
  namespace: string;
  releaseName: string;
  repo: string;
  repoNamespace: string;
  error?: Error;
  selected: IChartState["selected"];
  deployed: IChartState["deployed"];
  upgradeApp: (
    version: IChartVersion,
    chartNamespace: string,
    releaseName: string,
    namespace: string,
    values?: string,
    schema?: JSONSchema4,
  ) => Promise<boolean>;
  push: (location: string) => RouterAction;
  goBack: () => RouterAction;
  fetchChartVersions: (namespace: string, id: string) => Promise<IChartVersion[]>;
  getChartVersion: (namespace: string, id: string, chartVersion: string) => void;
}

interface IUpgradeFormState {
  appValues: string;
  valuesModified: boolean;
  isDeploying: boolean;
  modifications?: jsonpatch.Operation[];
}

class UpgradeForm extends React.Component<IUpgradeFormProps, IUpgradeFormState> {
  public state: IUpgradeFormState = {
    appValues: this.props.appCurrentValues || "",
    isDeploying: false,
    valuesModified: false,
  };

  public componentDidMount() {
    const chartID = `${this.props.repo}/${this.props.chartName}`;
    this.props.fetchChartVersions(this.props.repoNamespace, chartID);
  }

  public componentDidUpdate = (prevProps: IUpgradeFormProps) => {
    let modifications = this.state.modifications;
    if (this.props.deployed.values && !modifications) {
      // Calculate modifications from the default values
      const defaultValuesObj = YAML.parse(this.props.deployed.values);
      const deployedValuesObj = YAML.parse(this.props.appCurrentValues || "");
      modifications = jsonpatch.compare(defaultValuesObj, deployedValuesObj);
      this.setState({ modifications });
      this.setState({ appValues: this.applyModifications(modifications, this.state.appValues) });
    }

    if (prevProps.selected.version !== this.props.selected.version && !this.state.valuesModified) {
      // Apply modifications to the new selected version
      const appValues = modifications
        ? this.applyModifications(modifications, this.props.selected.values || "")
        : this.props.selected.values || "";
      this.setState({ appValues });
    }
  };

  public render() {
    const { namespace, releaseName, error, selected } = this.props;
    if (error) {
      return (
        <ErrorSelector error={error} namespace={namespace} action="update" resource={releaseName} />
      );
    }
    if (selected.versions.length === 0) {
      return <LoadingWrapper />;
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
              chartNamespace={this.props.repoNamespace}
              chartID={chartID}
              chartVersion={this.props.appCurrentVersion}
              deployedValues={this.applyModifications(
                this.state.modifications || [],
                this.props.deployed.values || "",
              )}
              namespace={this.props.namespace}
              releaseVersion={this.props.appCurrentVersion}
              selected={this.props.selected}
              push={this.props.push}
              goBack={this.props.goBack}
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

  public setValuesModified = () => {
    this.setState({ valuesModified: true });
  };

  public handleValuesChange = (value: string) => {
    this.setState({ appValues: value });
  };

  public handleDeploy = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const { selected, push, upgradeApp, releaseName, namespace, repoNamespace } = this.props;
    const { appValues } = this.state;

    this.setState({ isDeploying: true });
    if (selected.version) {
      const deployed = await upgradeApp(
        selected.version,
        repoNamespace,
        releaseName,
        namespace,
        appValues,
        selected.schema,
      );
      this.setState({ isDeploying: false });
      if (deployed) {
        push(`/ns/${namespace}/apps/${releaseName}`);
      }
    }
  };

  private applyModifications(modifications: jsonpatch.Operation[], appValues: string) {
    // And we add any possible change made to the original version
    if (modifications.length) {
      modifications.forEach(modification => {
        // Transform the JSON Path to the format expected by setValue
        // /a/b/c => a.b.c
        const path = modification.path.replace(/^\//, "").replace(/\//g, ".");
        if (modification.op === "remove") {
          appValues = deleteValue(appValues, path);
        } else {
          // Transform the modification as a ReplaceOperation to read its value
          const value = (modification as jsonpatch.ReplaceOperation<any>).value;
          appValues = setValue(appValues, path, value);
        }
      });
    }
    return appValues;
  }
}

export default UpgradeForm;
