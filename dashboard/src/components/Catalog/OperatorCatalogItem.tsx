// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { app } from "shared/url";
import { getPluginIcon, trimDescription } from "shared/utils";
import placeholder from "icons/placeholder.svg";
import InfoCard from "../InfoCard/InfoCard";
import { IOperatorCatalogItem } from "./CatalogItem";

export default function OperatorCatalogItem(props: IOperatorCatalogItem) {
  const { icon, name, csv, version, description, cluster, namespace, id } = props;
  const iconSrc = icon || placeholder;
  // Cosmetic change, remove the version from the csv name
  const csvName = props.csv.split(".")[0];
  const link = app.operatorInstances.new(cluster, namespace, csv, id);
  return (
    <InfoCard
      key={id}
      title={name}
      link={link}
      info={version || "-"}
      icon={iconSrc}
      description={trimDescription(description)}
      bgIcon={getPluginIcon("operator")}
      tag1Content={csvName}
      tag2Content={"operator"}
      tag2Class={"label-info-secondary"}
    />
  );
}
