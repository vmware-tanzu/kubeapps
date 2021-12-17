import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import carvelIcon from "../icons/carvel.svg";
import fluxIcon from "../icons/flux.svg";
import helmIcon from "../icons/helm.svg";
import olmIcon from "../icons/olm-icon.svg";
import placeholder from "../placeholder.png";

export enum PluginNames {
  PACKAGES_HELM = "helm.packages",
  PACKAGES_FLUX = "fluxv2.packages",
  PACKAGES_KAPP = "kapp_controller.packages",
}

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

// Perhaps the logic of these functions should be provided by each plugin itself, namely:
// i) return its icon; ii) return its user-faced name; iii) return its user-faced package name
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
      case PluginNames.PACKAGES_HELM:
        return helmIcon;
      case PluginNames.PACKAGES_FLUX:
        return fluxIcon;
      case PluginNames.PACKAGES_KAPP:
        return carvelIcon;
      default:
        return placeholder;
    }
  }
}

export function getPluginName(plugin?: Plugin | string) {
  // Temporary case while operators are not supported as kubeapps apis plugin
  if (typeof plugin === "string") {
    switch (plugin) {
      case "chart":
        return "Helm";
      case "helm":
        return "Helm";
      case "operator":
        return "Operator";
      default:
        return "unknown plugin";
    }
  } else {
    switch (plugin?.name) {
      case PluginNames.PACKAGES_HELM:
        return "Helm";
      case PluginNames.PACKAGES_FLUX:
        return "Flux";
      case PluginNames.PACKAGES_KAPP:
        return "Carvel";
      default:
        return plugin?.name;
    }
  }
}

export function getPluginPackageName(plugin?: Plugin | string) {
  // Temporary case while operators are not supported as kubeapps apis plugin
  if (typeof plugin === "string") {
    switch (plugin) {
      case "chart":
        return "Helm Chart";
      case "helm":
        return "Helm Chart";
      case "operator":
        return "Operator";
      default:
        return "unknown plugin package";
    }
  } else {
    switch (plugin?.name) {
      case PluginNames.PACKAGES_HELM:
        return "Helm Chart";
      case PluginNames.PACKAGES_FLUX:
        return "Helm Chart via Flux";
      case PluginNames.PACKAGES_KAPP:
        return "Carvel Package";
      default:
        return `${plugin?.name} package`;
    }
  }
}

export function getPluginsRequiringSA(): string[] {
  return [PluginNames.PACKAGES_FLUX, PluginNames.PACKAGES_KAPP];
}

export function getPluginsSupportingRollback(): string[] {
  return [PluginNames.PACKAGES_HELM];
}
