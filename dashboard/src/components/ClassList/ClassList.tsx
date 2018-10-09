import * as React from "react";
import { Link } from "react-router-dom";

import { IClusterServiceClass } from "../../shared/ClusterServiceClass";
import { IRBACRole } from "../../shared/types";
import Card, { CardContent, CardFooter, CardGrid, CardIcon } from "../Card";
import { ErrorSelector, MessageAlert } from "../ErrorAlert";
import LoadingWrapper from "../LoadingWrapper";
import PageHeader from "../PageHeader";

export interface IClassListProps {
  error: Error | undefined;
  classes: IClusterServiceClass[];
  isFetching: boolean;
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

export default class ClassList extends React.Component<IClassListProps> {
  public componentDidMount() {
    this.props.getClasses();
  }

  public render() {
    const { error, classes, isFetching } = this.props;
    return (
      <div>
        <PageHeader>
          <h1>Classes</h1>
        </PageHeader>
        <main>
          <LoadingWrapper loaded={!isFetching}>
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
            {classes.length === 0 ? (
              <MessageAlert level={"warning"}>
                <h5>Unable to find any class.</h5>
              </MessageAlert>
            ) : (
              <CardGrid>
                {classes
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
