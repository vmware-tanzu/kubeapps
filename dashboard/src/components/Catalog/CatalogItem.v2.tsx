import { IRepo } from "../../shared/types";
import ChartCatalogItem from "./ChartCatalogItem";
import OperatorCatalogItem from "./OperatorCatalogItem";

export interface ICatalogItem {
  id: string;
  name: string;
  version: string;
  description: string;
  namespace: string;
  icon?: string;
}

export interface IChartCatalogItem extends ICatalogItem {
  repo: IRepo;
}

export interface IOperatorCatalogItem extends ICatalogItem {
  csv: string;
}

export interface ICatalogItemProps {
  type: string;
  key: string;
  item: IChartCatalogItem | IOperatorCatalogItem;
}

function CatalogItem(props: ICatalogItemProps) {
  if (props.type === "operator") {
    const item = props.item as IOperatorCatalogItem;
    return OperatorCatalogItem(item);
  } else {
    const item = props.item as IChartCatalogItem;
    return ChartCatalogItem(item);
  }
}

export default CatalogItem;
