// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { AvailablePackageSummary } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import OperatorCatalogItem from "./OperatorCatalogItem";
import PackageCatalogItem from "./PackageCatalogItem";

// TODO: this file should be refactored after the operators have been integrated in a plugin

export interface ICatalogItem {
  name: string;
  cluster: string;
  namespace: string;
}

export interface IPackageCatalogItem extends ICatalogItem {
  availablePackageSummary: AvailablePackageSummary;
}

export interface IOperatorCatalogItem extends ICatalogItem {
  id: string;
  csv: string;
  version: string;
  description: string;
  icon?: string;
}

export interface ICatalogItemProps {
  type: string;
  id: string;
  item: IPackageCatalogItem | IOperatorCatalogItem;
}

function CatalogItem(props: ICatalogItemProps) {
  if (props.type === "operator") {
    const item = props.item as IOperatorCatalogItem;
    return <OperatorCatalogItem {...item} />;
  } else {
    const item = props.item as IPackageCatalogItem;
    return <PackageCatalogItem {...item} />;
  }
}

export default CatalogItem;
