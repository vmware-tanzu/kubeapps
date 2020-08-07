import { RouterAction } from "connected-react-router";
import * as yaml from "js-yaml";
import * as React from "react";

import OperatorNotSupported from "components/OperatorList/OperatorsNotSupported";
import * as url from "shared/url";
import { IClusterServiceVersion, IResource } from "../../shared/types";
import NotFoundErrorPage from "../ErrorAlert/NotFoundErrorAlert";
import OperatorInstanceFormBody from "../OperatorInstanceFormBody";
import PageHeader from "../PageHeader";

export interface IOperatorInstanceUpgradeFormProps {
  csvName: string;
  crdName: string;
  isFetching: boolean;
  cluster: string;
  namespace: string;
  resourceName: string;
  getResource: (
    namespace: string,
    csvName: string,
    crdName: string,
    resourceName: string,
  ) => Promise<void>;
  updateResource: (
    namespace: string,
    apiVersion: string,
    resource: string,
    resourceName: string,
    body: object,
  ) => Promise<boolean>;
  push: (location: string) => RouterAction;
  csv?: IClusterServiceVersion;
  resource?: IResource;
  errors: {
    fetch?: Error;
    create?: Error;
  };
}

export interface IOperatorInstanceUpgradeFormBodyState {
  defaultValues: string;
}

class DeploymentFormBody extends React.Component<
  IOperatorInstanceUpgradeFormProps,
  IOperatorInstanceUpgradeFormBodyState
> {
  public state: IOperatorInstanceUpgradeFormBodyState = {
    defaultValues: "",
  };

  public componentDidMount() {
    const { csvName, crdName, resourceName, namespace, getResource } = this.props;
    getResource(namespace, csvName, crdName, resourceName);
  }

  public componentDidUpdate(prevProps: IOperatorInstanceUpgradeFormProps) {
    const { resource } = this.props;
    if (resource !== prevProps.resource && resource) {
      this.setState({
        defaultValues: yaml.safeDump(resource),
      });
    }
  }

  public render() {
    const { isFetching, errors, resourceName, cluster, namespace, resource, csvName } = this.props;
    const { defaultValues } = this.state;

    if (cluster !== "default") {
      return <OperatorNotSupported namespace={namespace} />;
    }

    if (!errors.fetch && !isFetching && !resource) {
      return <NotFoundErrorPage resource={resourceName} namespace={namespace} />;
    }
    return (
      <>
        <PageHeader>
          <h1>Update {resourceName}</h1>
        </PageHeader>
        <main>
          <OperatorInstanceFormBody
            csvName={csvName}
            isFetching={isFetching}
            namespace={namespace}
            handleDeploy={this.handleDeploy}
            errors={errors}
            defaultValues={defaultValues}
          />
        </main>
      </>
    );
  }

  private handleDeploy = async (resource: IResource) => {
    const { updateResource, crdName, resourceName, cluster, namespace, push, csvName } = this.props;

    const created = await updateResource(
      namespace,
      resource.apiVersion,
      crdName.split(".")[0],
      resourceName,
      resource,
    );
    if (created) {
      push(url.app.operatorInstances.view(cluster, namespace, csvName, crdName, resourceName));
    }
  };
}

export default DeploymentFormBody;
