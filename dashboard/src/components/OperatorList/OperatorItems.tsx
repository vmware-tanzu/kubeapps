// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import InfoCard from "components/InfoCard/InfoCard";
import { Operators } from "shared/Operators";
import { IResource } from "shared/types";
import { api, app } from "shared/url";
import { trimDescription } from "shared/utils";

interface ICatalogItemsProps {
  operators: IResource[];
  cluster: string;
}

export default function OperatorItems({ operators, cluster }: ICatalogItemsProps) {
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
            link={app.operators.view(cluster, operator.metadata.namespace, operator.metadata.name)}
            title={operator.metadata.name}
            icon={api.operators.operatorIcon(
              cluster,
              operator.metadata.namespace,
              operator.metadata.name,
            )}
            info={`v${channel?.currentCSVDesc.version}`}
            tag1Content={operator.status.provider.name}
            description={trimDescription(channel?.currentCSVDesc.annotations.description || "")}
          />
        );
      })}
    </>
  );
}
