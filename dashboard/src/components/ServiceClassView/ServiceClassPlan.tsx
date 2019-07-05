import { RouterAction } from "connected-react-router";
import * as React from "react";
import { IClusterServiceClass } from "shared/ClusterServiceClass";
import { IServicePlan } from "shared/ServiceCatalog";
import ProvisionButton from "./ProvisionButton";

interface IServiceClassPlanProps {
  svcClass: IClusterServiceClass;
  svcPlan: IServicePlan;
  createError?: Error;
  provision: (
    instanceName: string,
    namespace: string,
    className: string,
    planName: string,
    parameters: {},
  ) => Promise<boolean>;
  push: (location: string) => RouterAction;
  namespace: string;
}

const ServiceClassPlan: React.SFC<IServiceClassPlanProps> = props => {
  const { svcClass, svcPlan, createError, provision, push, namespace } = props;
  const { spec } = svcPlan;
  const { externalMetadata } = spec;
  const name = externalMetadata ? externalMetadata.displayName : spec.externalName;
  const description =
    externalMetadata && externalMetadata.bullets ? externalMetadata.bullets : [spec.description];
  const bullets = (
    <div>
      <ul className="margin-reset">
        {description.map(bullet => (
          <li key={bullet}>{bullet}</li>
        ))}
      </ul>
    </div>
  );

  return (
    <tr key={svcPlan.spec.externalID}>
      <td>
        <b>{name}</b>
      </td>
      <td>{bullets}</td>
      <td className="text-r">
        <ProvisionButton
          selectedClass={svcClass}
          selectedPlan={svcPlan}
          provision={provision}
          push={push}
          namespace={namespace}
          error={createError}
        />
      </td>
    </tr>
  );
};

export default ServiceClassPlan;
