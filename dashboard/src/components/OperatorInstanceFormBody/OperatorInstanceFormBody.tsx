import * as yaml from "js-yaml";
import { get } from "lodash";
import * as React from "react";
import { Tab, TabList, TabPanel, Tabs } from "react-tabs";

import { definedNamespaces } from "../../shared/Namespace";
import { IClusterServiceVersionCRD, IResource, UnprocessableEntity } from "../../shared/types";
import ConfirmDialog from "../ConfirmDialog";
import AdvancedDeploymentForm from "../DeploymentFormBody/AdvancedDeploymentForm";
import Differential from "../DeploymentFormBody/Differential";
import { ErrorSelector } from "../ErrorAlert";
import UnexpectedErrorPage from "../ErrorAlert/UnexpectedErrorAlert";
import LoadingWrapper from "../LoadingWrapper";

export interface IOperatorInstanceFormProps {
  csvName: string;
  isFetching: boolean;
  namespace: string;
  handleDeploy: (resource: IResource) => void;
  defaultValues: string;
  crd?: IClusterServiceVersionCRD;
  errors: {
    fetch?: Error;
    create?: Error;
    update?: Error;
  };
}

export interface IOperatorInstanceFormBodyState {
  values: string;
  restoreDefaultValuesModalIsOpen: boolean;
  submittedResourceName: string;
  error?: Error;
}

class DeploymentFormBody extends React.Component<
  IOperatorInstanceFormProps,
  IOperatorInstanceFormBodyState
> {
  public state: IOperatorInstanceFormBodyState = {
    values: this.props.defaultValues,
    submittedResourceName: "",
    restoreDefaultValuesModalIsOpen: false,
  };

  public componentDidUpdate(prevProps: IOperatorInstanceFormProps) {
    const { defaultValues } = this.props;
    if (prevProps.defaultValues !== defaultValues) {
      this.setState({ values: defaultValues });
    }
  }

  public render() {
    const { isFetching, errors, csvName, namespace, crd } = this.props;
    const { submittedResourceName, error } = this.state;

    if (errors.fetch) {
      return (
        <ErrorSelector
          error={errors.fetch}
          resource={`Operator Instance "${csvName}" (${crd?.name})`}
        />
      );
    }
    if (namespace === definedNamespaces.all) {
      return <UnexpectedErrorPage title="Select a namespace before creating a new instance." />;
    }
    if (isFetching) {
      return <LoadingWrapper />;
    }
    return (
      <>
        {errors.create && (
          <ErrorSelector
            error={errors.create}
            resource={`Operator Instance "${submittedResourceName}"`}
          />
        )}
        {errors.update && (
          <ErrorSelector
            error={errors.update}
            resource={`Operator Instance "${submittedResourceName}"`}
          />
        )}
        {error && <ErrorSelector error={error} resource={"Operator Instance"} />}
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
              oldValues={this.props.defaultValues}
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
    this.setState({ values: this.props.defaultValues, restoreDefaultValuesModalIsOpen: false });
  };

  private handleDeploy = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    // Clean possible previous errors
    this.setState({ error: undefined });
    const { values } = this.state;
    let resource: IResource = {} as any;
    try {
      resource = yaml.safeLoad(values);
    } catch (e) {
      this.setState({
        error: new UnprocessableEntity(`Unable to parse the given YAML. Got: ${e.message}`),
      });
      return;
    }
    if (!resource.apiVersion) {
      this.setState({
        error: new UnprocessableEntity(
          "Unable parse the resource. Make sure it contains a valid apiVersion",
        ),
      });
      return;
    }
    const resourceName = get(resource, "metadata.name");
    this.setState({ submittedResourceName: resourceName });
    this.props.handleDeploy(resource);
  };
}

export default DeploymentFormBody;
