import * as React from "react";

import { IServiceBroker, IServiceClass } from "../../shared/ServiceCatalog";
import { Card, CardContainer } from "../Card";

export interface IClassListProps {
  broker: IServiceBroker | undefined;
  classes: IServiceClass[];
  getBrokers: () => Promise<IServiceBroker[]>;
  getClasses: () => Promise<IServiceClass[]>;
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
        <p>Types of services available from this broker</p>
        <CardContainer>
          {/* {classes.map(svcClass => { */}
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
                  linkTo={`${location.pathname}/${svcClass.spec.externalName}`}
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
