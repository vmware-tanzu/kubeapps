import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import { IRepo } from "shared/types";
import PackageCatalogItem from "./PackageCatalogItem";
import OperatorCatalogItem from "./OperatorCatalogItem";

export interface ICatalogItem {
  id: string;
  name: string;
  version: string;
  description: string;
  cluster: string;
  namespace: string;
  icon?: string;
}

export interface IPackageCatalogItem extends ICatalogItem {
  id: string;
  repo: IRepo;
  plugin: Plugin;
}

export interface IOperatorCatalogItem extends ICatalogItem {
  id: string;
  csv: string;
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
