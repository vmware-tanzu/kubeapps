import InfoCard from "components/InfoCard/InfoCard.v2";
import React from "react";
import { Operators } from "shared/Operators";
import { IResource } from "shared/types";
import { api, app } from "shared/url";
import { trimDescription } from "shared/utils";

interface ICatalogItemsProps {
  operators: IResource[];
  cluster: string;
  namespace: string;
}

export default function OperatorItems({ operators, cluster, namespace }: ICatalogItemsProps) {
  if (operators.length === 0) {
    return <p>No operator matches the current filter.</p>;
  }
  return (
    <>
      {operators.map(operator => {
        const channel = Operators.getDefaultChannel(operator);
        return (
          <InfoCard
            key={operator.metadata.name}
            link={app.operators.view(cluster, namespace, operator.metadata.name)}
            title={operator.metadata.name}
            icon={api.operators.operatorIcon(namespace, operator.metadata.name)}
            info={`v${channel?.currentCSVDesc.version}`}
            tag1Content={operator.status.provider.name}
            description={trimDescription(channel?.currentCSVDesc.annotations.description || "")}
          />
        );
      })}
    </>
  );
}
