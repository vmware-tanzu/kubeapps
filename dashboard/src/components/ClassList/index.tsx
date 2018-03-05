import * as React from "react";
import { Link } from "react-router-dom";

import { IClusterServiceClass } from "../../shared/ClusterServiceClass";
import { IServiceBroker } from "../../shared/ServiceCatalog";
import Card, { CardContent, CardFooter, CardGrid, CardIcon } from "../Card";

export interface IClassListProps {
  classes: IClusterServiceClass[];
  getBrokers: () => Promise<IServiceBroker[]>;
  getClasses: () => Promise<IClusterServiceClass[]>;
}

export class ClassList extends React.Component<IClassListProps> {
  public componentDidMount() {
    this.props.getBrokers();
    this.props.getClasses();
  }

  public render() {
    const { classes } = this.props;
    return (
      <div>
        <h2>Classes</h2>
        <p>Types of services available from all brokers</p>
        <CardGrid>
          {classes
            .sort((a, b) => a.spec.externalName.localeCompare(b.spec.externalName))
            .map(svcClass => {
              const { spec } = svcClass;
              const { externalMetadata } = spec;
              const name = externalMetadata ? externalMetadata.displayName : spec.externalName;
              const description = externalMetadata
                ? externalMetadata.longDescription
                : spec.description;
              const imageUrl = externalMetadata && externalMetadata.imageUrl;
              const link = `/services/brokers/${svcClass.spec.clusterServiceBrokerName}/classes/${
                svcClass.spec.externalName
              }`;

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
      </div>
    );
  }
}
