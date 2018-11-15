import * as React from "react";

import placeholder from "../../placeholder.png";
import { IClusterServiceClass } from "../../shared/ClusterServiceClass";
import { IServicePlan } from "../../shared/ServiceCatalog";
import { IServiceInstance } from "../../shared/ServiceInstance";
import Card, { CardContent, CardGrid, CardIcon } from "../Card";

interface IServiceInstanceInfoProps {
  instance: IServiceInstance;
  svcClass?: IClusterServiceClass;
  plan?: IServicePlan;
}

const ServiceInstanceInfo: React.SFC<IServiceInstanceInfoProps> = props => {
  const { instance, svcClass, plan } = props;
  const name = instance.metadata.name;
  const externalMetadata = svcClass && svcClass.spec.externalMetadata;
  const imageUrl = (externalMetadata && externalMetadata.imageUrl) || placeholder;

  return (
    <CardGrid className="ServiceInstanceInfo">
      <Card>
        <CardIcon icon={imageUrl} />
        <CardContent>
          <h5>{name}</h5>
          {svcClass && renderSvcClassInfo(svcClass)}
          {plan && renderPlanInfo(plan)}
        </CardContent>
      </Card>
    </CardGrid>
  );
};

const renderSvcClassInfo: React.SFC<IClusterServiceClass> = svcClass => {
  const { spec } = svcClass;
  const { externalMetadata } = spec;
  const svcName = externalMetadata ? externalMetadata.displayName : spec.externalName;
  const description = externalMetadata ? externalMetadata.longDescription : spec.description;

  return (
    <React.Fragment>
      <strong>Class:</strong> {svcName}
      <p>{description}</p>
    </React.Fragment>
  );
};

const renderPlanInfo: React.SFC<IServicePlan> = svcPlan => {
  const { spec } = svcPlan;
  const { externalMetadata } = spec;
  const planName = externalMetadata ? externalMetadata.displayName : spec.externalName;
  const description =
    externalMetadata && externalMetadata.bullets ? externalMetadata.bullets : [spec.description];
  const free = svcPlan.spec.free ? <span>Free ✓</span> : undefined;
  const bullets = (
    <div>
      <ul>
        {description.map(bullet => (
          <li key={bullet}>{bullet}</li>
        ))}
      </ul>
    </div>
  );

  return (
    <React.Fragment>
      <strong>Plan:</strong> {planName}
      {free && <p className="type-small margin-reset margin-b-big type-color-light-blue">{free}</p>}
      {bullets}
    </React.Fragment>
  );
};

export default ServiceInstanceInfo;
