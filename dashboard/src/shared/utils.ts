// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import {
  InstalledPackageStatus_StatusReason,
  installedPackageStatus_StatusReasonToJSON,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { PackageRepositoryAuth_PackageRepositoryAuthType } from "gen/kubeappsapis/core/packages/v1alpha1/repositories";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import carvelIcon from "icons/carvel.svg";
import fluxIcon from "icons/flux.svg";
import helmIcon from "icons/helm.svg";
import olmIcon from "icons/olm-icon.svg";
import placeholder from "icons/placeholder.svg";
import { IConfig } from "./Config";

export enum PluginNames {
  PACKAGES_HELM = "helm.packages",
  PACKAGES_FLUX = "fluxv2.packages",
  PACKAGES_KAPP = "kapp_controller.packages",
}

export const k8sObjectNameRegex = "[a-z0-9]([-a-z0-9]*[a-z0-9])?(.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*";

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
export const MAX_DESC_LENGTH = 90;

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

export function getPluginPackageName(plugin?: Plugin | PluginNames | string, plural = false) {
  // Temporary case while operators are not supported as kubeapps apis plugin
  if (typeof plugin === "string") {
    switch (plugin) {
      case "chart":
      case "helm":
      case PluginNames.PACKAGES_HELM:
        return plural ? "Helm Charts" : "Helm Chart";
      case PluginNames.PACKAGES_FLUX:
        return plural ? "Helm Charts via Flux" : "Helm Chart via Flux";
      case PluginNames.PACKAGES_KAPP:
        return plural ? "Carvel Packages" : "Carvel Package";
      case "operator":
        return plural ? "Operators" : "Operator";
      default:
        return `unknown plugin ${plural ? "packages" : "package"}`;
    }
  } else {
    switch (plugin?.name) {
      case PluginNames.PACKAGES_HELM:
        return plural ? "Helm Charts" : "Helm Chart";
      case PluginNames.PACKAGES_FLUX:
        return plural ? "Helm Charts via Flux" : "Helm Chart via Flux";
      case PluginNames.PACKAGES_KAPP:
        return plural ? "Carvel Packages" : "Carvel Package";
      default:
        return `${plugin?.name ? plugin.name : "unknown"} ${plural ? "packages" : "package"}`;
    }
  }
}

// TODO(agamez): replace with a proper call to the plugins server (see getPluginsServiceClientImpl)
export function getPluginByName(pluginName: PluginNames | string) {
  switch (pluginName) {
    case PluginNames.PACKAGES_HELM:
      return {
        name: PluginNames.PACKAGES_HELM,
        version: "v1alpha1",
      } as Plugin;
    case PluginNames.PACKAGES_FLUX:
      return {
        name: PluginNames.PACKAGES_FLUX,
        version: "v1alpha1",
      } as Plugin;
    case PluginNames.PACKAGES_KAPP:
      return {
        name: PluginNames.PACKAGES_KAPP,
        version: "v1alpha1",
      } as Plugin;
    default:
      return {
        name: "",
        version: "",
      } as Plugin;
  }
}

export function getPluginsRequiringSA(): string[] {
  return [PluginNames.PACKAGES_FLUX, PluginNames.PACKAGES_KAPP];
}

export function getPluginsSupportingRollback(): string[] {
  return [PluginNames.PACKAGES_HELM];
}

export function getAppStatusLabel(
  statusReason: InstalledPackageStatus_StatusReason = InstalledPackageStatus_StatusReason.STATUS_REASON_UNSPECIFIED,
): string {
  // The JSON versions of the reasons are forced to follow the standard
  // pattern STATUS_REASON_<reason> by buf.
  const jsonReason = installedPackageStatus_StatusReasonToJSON(statusReason);
  return jsonReason.replace("STATUS_REASON_", "").toLowerCase();
}

export function getSupportedPackageRepositoryAuthTypes(
  plugin: Plugin,
): PackageRepositoryAuth_PackageRepositoryAuthType[] {
  switch (plugin.name) {
    case PluginNames.PACKAGES_HELM:
      return [
        PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_AUTHORIZATION_HEADER,
        PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
        PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BEARER,
        PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON,
      ];
    case PluginNames.PACKAGES_FLUX:
      return [
        PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
        PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_OPAQUE,
        PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_SSH,
        PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_TLS,
        PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON,
      ];
    case PluginNames.PACKAGES_KAPP:
      return [
        PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
        PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BEARER,
        PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON,
        PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_SSH,
      ];
    default:
      return [];
  }
}

/**
 * Check if namespace is global depending on the plugin
 *
 * @param namespace Namespace to check
 * @param pluginName Plugin name
 * @param kubeappsConfig Kubeapps configuration
 * @returns true if namespace is global
 */
export function isGlobalNamespace(namespace: string, pluginName: string, kubeappsConfig: IConfig) {
  switch (pluginName) {
    case PluginNames.PACKAGES_HELM:
      return namespace === kubeappsConfig.globalReposNamespace;
    case PluginNames.PACKAGES_KAPP:
      return namespace === kubeappsConfig.carvelGlobalNamespace;
    default:
      // Currently, Flux doesn't support global namespaces
      return false;
  }
}
