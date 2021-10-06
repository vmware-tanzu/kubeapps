import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import helmIcon from "../icons/helm.svg";
import olmIcon from "../icons/operator-framework.svg";
import placeholder from "../placeholder.png";
import ResourceRef from "./ResourceRef";
import { IK8sList, IKubeItem, IResource, ISecret } from "./types";

export function escapeRegExp(str: string) {
  return str.replace(/[-[\]/{}()*+?.\\^$|]/g, "\\$&");
}

export function getValueFromEvent(
  e: React.FormEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>,
) {
  let value: any = e.currentTarget.value;
  switch (e.currentTarget.type) {
    case "checkbox":
      // value is a boolean
      value = value === "true";
      break;
    case "number":
      // value is a number
      value = parseInt(value, 10);
      break;
  }
  return value;
}

// 3 lines description max
const MAX_DESC_LENGTH = 90;

export function trimDescription(desc: string): string {
  if (desc.length > MAX_DESC_LENGTH) {
    // Trim to the last word under the max length
    return desc.substr(0, desc.lastIndexOf(" ", MAX_DESC_LENGTH)).concat("...");
  }
  return desc;
}

export function flattenResources(
  refs: ResourceRef[],
  resources: { [s: string]: IKubeItem<IResource | IK8sList<IResource, {}>> },
) {
  const result: Array<IKubeItem<IResource | ISecret>> = [];
  refs.forEach(ref => {
    const kubeItem = resources[ref.getResourceURL()];
    if (kubeItem) {
      const itemList = kubeItem.item as IK8sList<IResource | ISecret, {}>;
      if (itemList && itemList.items) {
        itemList.items.forEach(i => {
          result.push({
            isFetching: kubeItem.isFetching,
            error: kubeItem.error,
            item: i,
          });
        });
      } else {
        result.push(kubeItem as IKubeItem<IResource | ISecret>);
      }
    }
  });
  return result;
}

export function getPluginIcon(plugin?: Plugin | string) {
  // Temporary case while operators are not supported as kubeapps apis plugin
  if (typeof plugin === "string") {
    switch (plugin) {
      case "chart":
        return helmIcon;
      case "helm":
        return helmIcon;
      case "operator":
        return olmIcon;
      default:
        return placeholder;
    }
  } else {
    switch (plugin?.name) {
      case "helm.packages":
        return helmIcon;
      default:
        return placeholder;
    }
  }
}
