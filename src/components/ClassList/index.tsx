import * as React from "react";

import { IClusterServiceClass } from "../../shared/ClusterServiceClass";
import { IServiceBroker } from "../../shared/ServiceCatalog";
import { Card, CardContainer } from "../Card";

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
        <CardContainer>
          {classes
            .sort((a, b) => a.spec.externalName.localeCompare(b.spec.externalName))
            .map(svcClass => {
              const tags = svcClass.spec.tags.reduce<string>((joined, tag) => {
                return `${joined} ${tag},`;
              }, "");
              const { spec } = svcClass;
              const { externalMetadata } = spec;
              const name = externalMetadata ? externalMetadata.displayName : spec.externalName;
              const description = externalMetadata
                ? externalMetadata.longDescription
                : spec.description;
              const imageUrl = externalMetadata && externalMetadata.imageUrl;

              const card = (
                <Card
                  key={svcClass.metadata.uid}
                  header={name}
                  icon={imageUrl}
                  body={description}
                  buttonText="View Plans"
                  linkTo={`/services/brokers/${svcClass.spec.clusterServiceBrokerName}/classes/${
                    svcClass.spec.externalName
                  }`}
                  notes={
                    <span style={{ fontSize: "small" }}>
                      <strong>Tags:</strong> {tags}
                    </span>
                  }
                />
              );
              return card;
            })}
        </CardContainer>
      </div>
    );
  }
}
