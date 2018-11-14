import * as React from "react";
import { Link } from "react-router-dom";

import { IClusterServiceClass } from "../../shared/ClusterServiceClass";
import { IRBACRole } from "../../shared/types";
import Card, { CardContent, CardFooter, CardGrid, CardIcon } from "../Card";
import { ErrorSelector, MessageAlert } from "../ErrorAlert";
import LoadingWrapper from "../LoadingWrapper";
import PageHeader from "../PageHeader";

export interface IServiceClassListProps {
  error: Error | undefined;
  classes: {
    isFetching: boolean;
    list: IClusterServiceClass[];
  };
  getClasses: () => void;
}

const RequiredRBACRoles: IRBACRole[] = [
  {
    apiGroup: "servicecatalog.k8s.io",
    clusterWide: true,
    resource: "clusterserviceclasses",
    verbs: ["list"],
  },
];

class ServiceClassList extends React.Component<IServiceClassListProps> {
  public componentDidMount() {
    this.props.getClasses();
  }

  public render() {
    const { error, classes } = this.props;
    return (
      <div>
        <PageHeader>
          <h1>Classes</h1>
        </PageHeader>
        <main>
          <LoadingWrapper loaded={!classes.isFetching}>
            {error ? (
              <ErrorSelector
                error={error}
                action="list"
                defaultRequiredRBACRoles={{ list: RequiredRBACRoles }}
                resource="Service Classes"
              />
            ) : (
              <p>Types of services available from all brokers</p>
            )}
            {classes.list.length === 0 ? (
              <MessageAlert level={"warning"}>
                <div>
                  <h5>Service Classes not found.</h5>
                  The Service Catalog server may have failed to populate them. You can find more
                  information about how the Service Catalog works{" "}
                  <a
                    target="_blank"
                    href={
                      "https://kubernetes.io/docs/concepts/extend-kubernetes/service-catalog/#usage"
                    }
                  >
                    here
                  </a>
                  .
                </div>
              </MessageAlert>
            ) : (
              <CardGrid>
                {classes.list
                  .sort((a, b) => a.spec.externalName.localeCompare(b.spec.externalName))
                  .map(svcClass => {
                    const { spec } = svcClass;
                    const { externalMetadata } = spec;
                    const name = externalMetadata
                      ? externalMetadata.displayName
                      : spec.externalName;
                    const description = externalMetadata
                      ? externalMetadata.longDescription
                      : spec.description;
                    const imageUrl = externalMetadata && externalMetadata.imageUrl;
                    const link = `/services/brokers/${
                      svcClass.spec.clusterServiceBrokerName
                    }/classes/${svcClass.spec.externalName}`;

                    const card = (
                      <Card key={svcClass.metadata.uid} responsive={true} responsiveColumns={3}>
                        <CardIcon icon={imageUrl} />
                        <CardContent>
                          <h5>{name}</h5>
                          <p className="margin-b-reset">{description}</p>
                        </CardContent>
                        <CardFooter className="text-c">
                          <Link className="button button-accent" to={link}>
                            Select a plan
                          </Link>
                        </CardFooter>
                      </Card>
                    );
                    return card;
                  })}
              </CardGrid>
            )}
          </LoadingWrapper>
        </main>
      </div>
    );
  }
}

export default ServiceClassList;
