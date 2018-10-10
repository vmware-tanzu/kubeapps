import * as React from "react";

import { definedNamespaces } from "../../../shared/Namespace";
import { IServiceBroker } from "../../../shared/ServiceCatalog";
import { ForbiddenError, IRBACRole } from "../../../shared/types";
import Card, { CardContent, CardFooter, CardGrid } from "../../Card";
import {
  PermissionsErrorAlert,
  ServiceBrokersNotFoundAlert,
  UnexpectedErrorAlert,
} from "../../ErrorAlert";
import SyncButton from "./SyncButton";

import "./ServiceBrokerList.css";

interface IServiceBrokerListProps {
  errors: {
    fetch?: Error;
    update?: Error;
  };
  brokers: IServiceBroker[];
  sync: (broker: IServiceBroker) => Promise<any>;
}

export const RequiredRBACRoles: { [s: string]: IRBACRole[] } = {
  resync: [
    {
      apiGroup: "servicecatalog.k8s.io",
      clusterWide: true,
      resource: "clusterservicebrokers",
      verbs: ["patch"],
    },
  ],
  view: [
    {
      apiGroup: "servicecatalog.k8s.io",
      clusterWide: true,
      resource: "clusterservicebrokers",
      verbs: ["list"],
    },
  ],
};

class ServiceBrokerList extends React.Component<IServiceBrokerListProps> {
  public render() {
    const { brokers, errors, sync } = this.props;
    return (
      <div>
        <h1>Brokers</h1>
        <hr />
        {errors.fetch ? (
          this.renderError(errors.fetch)
        ) : brokers.length > 0 ? (
          <div>
            {errors.update && this.renderError(errors.update, "resync")}
            <CardGrid className="BrokerList">
              {brokers.map(broker => (
                <Card key={broker.metadata.uid} responsive={true} responsiveColumns={3}>
                  <CardContent>
                    <h2 className="margin-reset">{broker.metadata.name}</h2>
                    <p className="type-small margin-reset margin-b-big BrokerList__url">
                      {broker.spec.url}
                    </p>
                    <p className="margin-b-reset">
                      Last updated {broker.status.lastCatalogRetrievalTime}
                    </p>
                  </CardContent>
                  <CardFooter className="text-c">
                    <SyncButton sync={sync} broker={broker} />
                  </CardFooter>
                </Card>
              ))}
            </CardGrid>
          </div>
        ) : (
          <ServiceBrokersNotFoundAlert />
        )}
      </div>
    );
  }

  // TODO: Replace with ErrorSelector
  private renderError(error: Error, action = "view") {
    return error instanceof ForbiddenError ? (
      <PermissionsErrorAlert
        action={`${action} Service Brokers`}
        namespace={definedNamespaces.all}
        roles={RequiredRBACRoles[action]}
      />
    ) : (
      <UnexpectedErrorAlert />
    );
  }
}

export default ServiceBrokerList;
