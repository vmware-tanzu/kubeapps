import { RouterAction } from "connected-react-router";
import * as yaml from "js-yaml";
import * as Moniker from "moniker-native";
import * as React from "react";
import { Tab, TabList, TabPanel, Tabs } from "react-tabs";

import { IClusterServiceVersion, IClusterServiceVersionCRD, IResource } from "../../shared/types";
import ConfirmDialog from "../ConfirmDialog";
import AdvancedDeploymentForm from "../DeploymentFormBody/AdvancedDeploymentForm";
import Differential from "../DeploymentFormBody/Differential";
import { ErrorSelector } from "../ErrorAlert";
import NotFoundErrorPage from "../ErrorAlert/NotFoundErrorAlert";
import LoadingWrapper from "../LoadingWrapper";
import PageHeader from "../PageHeader";

export interface IOperatorInstanceFormProps {
  csvName: string;
  crdName: string;
  isFetching: boolean;
  namespace: string;
  getCSV: (namespace: string, csvName: string) => void;
  createResource: (
    namespace: string,
    apiVersion: string,
    resource: string,
    body: object,
  ) => Promise<boolean>;
  push: (location: string) => RouterAction;
  csv?: IClusterServiceVersion;
  errors: {
    fetch?: Error;
    create?: Error;
  };
}

export interface IOperatorInstanceFormBodyState {
  name: string;
  values: string;
  defaultValues: string;
  restoreDefaultValuesModalIsOpen: boolean;
  submittedResourceName: string;
  crd?: IClusterServiceVersionCRD;
}

class DeploymentFormBody extends React.Component<
  IOperatorInstanceFormProps,
  IOperatorInstanceFormBodyState
> {
  public state: IOperatorInstanceFormBodyState = {
    name: Moniker.choose(),
    values: "",
    defaultValues: "",
    submittedResourceName: "",
    restoreDefaultValuesModalIsOpen: false,
  };

  public componentDidMount() {
    const { getCSV, csvName, namespace } = this.props;
    getCSV(namespace, csvName);
  }

  public componentDidUpdate(prevProps: IOperatorInstanceFormProps) {
    const { csv, crdName } = this.props;
    if (csv && csv !== prevProps.csv) {
      csv.spec.customresourcedefinitions.owned.forEach(ownedCRD => {
        if (ownedCRD.name === crdName) {
          // Got the target CRD, extract the example
          const kind = ownedCRD.kind;
          const rawExamples = csv.metadata.annotations["alm-examples"];
          const examples = JSON.parse(rawExamples) as IResource[];
          examples.forEach(example => {
            if (example.kind === kind) {
              // Found the example, set the default values
              const yamlValues = yaml.safeDump(example);
              this.setState({
                values: yamlValues,
                defaultValues: yamlValues,
                crd: ownedCRD,
              });
            }
          });
        }
      });
    }
  }

  public render() {
    const { isFetching, errors, csvName, crdName } = this.props;
    const { crd, submittedResourceName } = this.state;

    if (errors.fetch) {
      return (
        <ErrorSelector
          error={errors.fetch}
          resource={`Operator Instance "${csvName}" (${crd?.name})`}
        />
      );
    }
    if (isFetching) {
      return <LoadingWrapper />;
    }
    if (!crd) {
      return (
        <NotFoundErrorPage
          header={
            <span>
              {crdName} not found in the definition of {csvName}
            </span>
          }
        />
      );
    }
    return (
      <>
        <PageHeader>
          <h1>Create {crd.kind}</h1>
        </PageHeader>
        <main>
          {errors.create && (
            <ErrorSelector
              error={errors.create}
              resource={`Operator Instance "${submittedResourceName}" (${crd?.name})`}
            />
          )}
          <p>{crd.description}</p>
          <ConfirmDialog
            modalIsOpen={this.state.restoreDefaultValuesModalIsOpen}
            loading={false}
            confirmationText={"Are you sure you want to restore the default instance values?"}
            confirmationButtonText={"Restore"}
            onConfirm={this.restoreDefaultValues}
            closeModal={this.closeRestoreDefaultValuesModal}
          />
          {this.renderTabs()}
          <form className="container padding-b-bigger" onSubmit={this.handleDeploy}>
            <div className="margin-t-big">
              <button className="button button-primary" type="submit">
                Submit
              </button>
              <button className="button" type="button" onClick={this.openRestoreDefaultValuesModal}>
                Restore Defaults
              </button>
            </div>
          </form>
        </main>
      </>
    );
  }

  private handleValuesChange = (values: string) => {
    this.setState({ values });
  };

  private renderTabs = () => {
    return (
      <div className="margin-t-normal row">
        <Tabs className="col-8">
          <TabList>
            <Tab>YAML</Tab>
            <Tab>Changes</Tab>
          </TabList>
          <TabPanel>
            <AdvancedDeploymentForm
              appValues={this.state.values}
              handleValuesChange={this.handleValuesChange}
            />
          </TabPanel>
          <TabPanel>
            <Differential
              title="Difference from example defaults"
              oldValues={this.state.defaultValues}
              newValues={this.state.values}
              emptyDiffText="No changes detected from example defaults"
            />
          </TabPanel>
        </Tabs>
      </div>
    );
  };

  private closeRestoreDefaultValuesModal = () => {
    this.setState({ restoreDefaultValuesModalIsOpen: false });
  };

  private openRestoreDefaultValuesModal = () => {
    this.setState({ restoreDefaultValuesModalIsOpen: true });
  };

  private restoreDefaultValues = () => {
    this.setState({ values: this.state.defaultValues, restoreDefaultValuesModalIsOpen: false });
  };

  private handleDeploy = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const { createResource, push, namespace } = this.props;
    const { values, crd } = this.state;
    const resourceType = crd!.name.split(".")[0];
    // TODO: Catch errors
    const resource: IResource = yaml.safeLoad(values);
    this.setState({ submittedResourceName: resource.metadata.name });
    const created = await createResource(namespace, resource.apiVersion, resourceType, resource);
    if (created) {
      push(`/operators-instances/ns/${namespace}/${resource.metadata.name}`);
    }
  };
}

export default DeploymentFormBody;
