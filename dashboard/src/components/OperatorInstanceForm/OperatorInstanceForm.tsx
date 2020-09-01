import { RouterAction } from "connected-react-router";
import * as yaml from "js-yaml";
import { get } from "lodash";
import * as React from "react";

import OperatorNotSupported from "components/OperatorList/OperatorsNotSupported";
import * as url from "shared/url";
import { IClusterServiceVersion, IClusterServiceVersionCRD, IResource } from "../../shared/types";
import NotFoundErrorPage from "../ErrorAlert/NotFoundErrorAlert";
import OperatorInstanceFormBody from "../OperatorInstanceFormBody";
import PageHeader from "../PageHeader";

export interface IOperatorInstanceFormProps {
  csvName: string;
  crdName: string;
  isFetching: boolean;
  cluster: string;
  namespace: string;
  kubeappsCluster: string;
  getCSV: (cluster: string, namespace: string, csvName: string) => void;
  createResource: (
    cluster: string,
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
  defaultValues: string;
  crd?: IClusterServiceVersionCRD;
}

class DeploymentFormBody extends React.Component<
  IOperatorInstanceFormProps,
  IOperatorInstanceFormBodyState
> {
  public state: IOperatorInstanceFormBodyState = {
    defaultValues: "",
  };

  public componentDidMount() {
    const { cluster, getCSV, csvName, namespace } = this.props;
    getCSV(cluster, namespace, csvName);
  }

  public componentDidUpdate(prevProps: IOperatorInstanceFormProps) {
    const { csv, crdName } = this.props;
    if (csv && csv !== prevProps.csv) {
      csv.spec.customresourcedefinitions.owned.forEach(ownedCRD => {
        if (ownedCRD.name === crdName) {
          // Got the target CRD, extract the example
          const kind = ownedCRD.kind;
          const rawExamples = get(csv, 'metadata.annotations["alm-examples"]', "[]");
          const examples = JSON.parse(rawExamples) as IResource[];
          let defaultValues = "";
          examples.forEach(example => {
            if (example.kind === kind) {
              // Found the example, set the default values
              defaultValues = yaml.safeDump(example);
            }
          });
          this.setState({
            defaultValues,
            crd: ownedCRD,
          });
        }
      });
    }
  }

  public render() {
    const {
      isFetching,
      errors,
      csvName,
      crdName,
      cluster,
      namespace,
      kubeappsCluster,
    } = this.props;
    const { crd, defaultValues } = this.state;
    if (cluster !== kubeappsCluster) {
      return <OperatorNotSupported kubeappsCluster={kubeappsCluster} namespace={namespace} />;
    }
    if (!errors.fetch && !isFetching && !crd) {
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
          <h1>Create {crd?.kind}</h1>
        </PageHeader>
        <main>
          <p>{crd?.description}</p>
          <OperatorInstanceFormBody
            csvName={csvName}
            isFetching={isFetching}
            namespace={namespace}
            handleDeploy={this.handleDeploy}
            crd={crd}
            errors={errors}
            defaultValues={defaultValues}
          />
        </main>
      </>
    );
  }

  private handleDeploy = async (resource: IResource) => {
    const { createResource, push, cluster, namespace, csv } = this.props;
    const { crd } = this.state;
    if (!crd || !csv) {
      // Unexpected error, CRD and CSV should have been previously populated
      throw new Error(`Missing CRD (${JSON.stringify(crd)}) or CSV (${JSON.stringify(csv)})`);
    }
    const resourceType = crd.name.split(".")[0];
    const created = await createResource(
      cluster,
      namespace,
      resource.apiVersion,
      resourceType,
      resource,
    );
    if (created) {
      push(
        url.app.operatorInstances.view(
          cluster,
          namespace,
          csv.metadata.name,
          crd.name,
          resource.metadata.name,
        ),
      );
    }
  };
}

export default DeploymentFormBody;
