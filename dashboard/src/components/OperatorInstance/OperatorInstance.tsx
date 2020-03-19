import * as yaml from "js-yaml";
import * as React from "react";

import placeholder from "../../placeholder.png";
import { IClusterServiceVersion, IResource } from "../../shared/types";
import AppNotes from "../AppView/AppNotes";
import AppValues from "../AppView/AppValues";
import Card, { CardContent, CardFooter, CardGrid, CardIcon } from "../Card";
import { ErrorSelector } from "../ErrorAlert";
import LoadingWrapper from "../LoadingWrapper";

export interface IOperatorInstanceProps {
  isFetching: boolean;
  namespace: string;
  csvName: string;
  crdName: string;
  instanceName: string;
  getResource: (
    namespace: string,
    csvName: string,
    crdName: string,
    resourceName: string,
  ) => Promise<void>;
  error?: Error;
  resource?: IResource;
  csv?: IClusterServiceVersion;
}

class OperatorInstance extends React.Component<IOperatorInstanceProps> {
  public componentDidMount() {
    const { csvName, crdName, instanceName, namespace, getResource } = this.props;
    getResource(namespace, csvName, crdName, instanceName);
  }

  public render() {
    const { isFetching, error, resource, csv, instanceName } = this.props;
    return (
      <section className="AppView padding-b-big">
        <main>
          <LoadingWrapper loaded={!isFetching}>
            {error && (
              <ErrorSelector
                error={error}
                action="get"
                resource={`Operator intance ${instanceName}`}
                namespace={this.props.namespace}
              />
            )}
            {resource && (
              <div className="row collapse-b-tablet">
                <div className="col-3">{csv && this.renderCSVInfo(resource, csv)}</div>
                <div className="col-9">
                  <div className="row padding-t-bigger">
                    <div className="col-4">
                      <h5>{instanceName}</h5>
                    </div>
                    <div className="col-8 text-r">
                      <button className="button button-danger">Delete</button>
                    </div>
                  </div>
                  <AppNotes title="Status" notes={yaml.safeDump(resource.status)} />
                  <AppValues values={yaml.safeDump(resource.spec)} />
                </div>
              </div>
            )}
          </LoadingWrapper>
        </main>
      </section>
    );
  }

  private renderCSVInfo = (resource: IResource, csv: IClusterServiceVersion) => {
    const { instanceName } = this.props;
    const icon = csv.spec.icon?.length
      ? `data:${csv.spec.icon[0].mediatype};base64,${csv.spec.icon[0].base64data}`
      : placeholder;
    const crd = csv.spec.customresourcedefinitions.owned.find(c => c.kind === resource.kind);
    return (
      <CardGrid className="ChartInfo">
        <Card>
          <CardIcon icon={icon} />
          <CardContent>
            <h5>{instanceName}</h5>
            <p className="margin-b-reset">{crd?.description}</p>
          </CardContent>
          <CardFooter>
            <div>
              <div>Cluster Service Version: v{csv.spec.version}</div>
              <div>Kind: {crd?.kind}</div>
            </div>
          </CardFooter>
        </Card>
      </CardGrid>
    );
  };
}

export default OperatorInstance;
