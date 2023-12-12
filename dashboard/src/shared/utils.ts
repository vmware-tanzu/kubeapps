// Copyright 2018-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { Code } from "@connectrpc/connect";
import { proto3 } from "@bufbuild/protobuf";
import { InstalledPackageStatus_StatusReason } from "gen/kubeappsapis/core/packages/v1alpha1/packages_pb";
import { PackageRepositoryAuth_PackageRepositoryAuthType } from "gen/kubeappsapis/core/packages/v1alpha1/repositories_pb";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins_pb";
import carvelIcon from "icons/carvel.svg";
import fluxIcon from "icons/flux.svg";
import helmIcon from "icons/helm.svg";
import olmIcon from "icons/olm-icon.svg";
import placeholder from "icons/placeholder.svg";
import { toNumber } from "lodash";
import { IConfig } from "./Config";
import {
  BadRequestNetworkError,
  ConflictNetworkError,
  CustomError,
  ForbiddenNetworkError,
  GatewayTimeoutNetworkError,
  InternalServerNetworkError,
  NotFoundNetworkError,
  NotImplementedNetworkError,
  PluginNames,
  RepositoryStorageTypes,
  RequestTimeoutNetworkError,
  ServerUnavailableNetworkError,
  TooManyRequestsNetworkError,
  UnauthorizedNetworkError,
} from "./types";

export const k8sObjectNameRegex = "[a-z0-9]([-a-z0-9]*[a-z0-9])?(.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*";
export const basicFormsDebounceTime = 500;

export function escapeRegExp(str: string) {
  return str.replace(/[-[\]/{}()*+?.\\^$|]/g, "\\$&");
}

export function getStringValue(value: any, type?: string) {
  const usedType = type || typeof value;
  let result = value?.toString();
  if (["array", "object"].includes(usedType)) {
    try {
      result = JSON.stringify(value);
    } catch (e) {
      result = value?.toString();
    }
  }
  return result || "";
}

export function getValueFromString(value: string, type?: string) {
  const usedType = type || typeof value;
  let result = value?.toString();
  if (["array", "object"].includes(usedType)) {
    try {
      result = JSON.parse(value);
      if (usedType === "object" && typeof result !== "object") {
        result = value?.toString();
      }
    } catch (e) {
      result = value?.toString();
    }
  }
  return result;
}

export function getValueFromEvent(
  e: React.FormEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>,
) {
  const value: any = e.currentTarget.value;
  switch (e.currentTarget.type) {
    case "checkbox":
      // value is a boolean
      return value === "true";
    case "number":
    case "range":
      // value is a number
      return toNumber(value);
    case "array":
      return getValueFromString(value, "array");
    case "object":
      return getValueFromString(value, "object");
    default:
      return getValueFromString(value);
  }
}

export function getOptionalMin(num1?: number, num2?: number) {
  if (num1 && num2) {
    return Math.min(num1, num2);
  }
  return num1 || num2 || undefined;
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

export function getPluginsAllowingSA(): string[] {
  return [PluginNames.PACKAGES_FLUX, PluginNames.PACKAGES_KAPP];
}

// getPluginsRequiringSA should return a subset of getPluginsAllowingSA
export function getPluginsRequiringSA(): string[] {
  return [PluginNames.PACKAGES_KAPP];
}

export function getPluginsSupportingRollback(): string[] {
  return [PluginNames.PACKAGES_HELM];
}

export function getAppStatusLabel(
  statusReason: InstalledPackageStatus_StatusReason = InstalledPackageStatus_StatusReason.UNSPECIFIED,
): string {
  const statusReasonName = proto3
    .getEnumType(InstalledPackageStatus_StatusReason)
    .findNumber(statusReason)!.name;
  return statusReasonName.toString().replace("STATUS_REASON_", "").toLowerCase();
}

export function getSupportedPackageRepositoryAuthTypes(
  plugin: Plugin,
  type?: string,
): PackageRepositoryAuth_PackageRepositoryAuthType[] {
  switch (plugin.name) {
    case PluginNames.PACKAGES_HELM:
      return [
        PackageRepositoryAuth_PackageRepositoryAuthType.AUTHORIZATION_HEADER,
        PackageRepositoryAuth_PackageRepositoryAuthType.BASIC_AUTH,
        PackageRepositoryAuth_PackageRepositoryAuthType.BEARER,
        PackageRepositoryAuth_PackageRepositoryAuthType.DOCKER_CONFIG_JSON,
      ];
    case PluginNames.PACKAGES_FLUX:
      switch (type) {
        case RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_HELM:
          return [
            PackageRepositoryAuth_PackageRepositoryAuthType.BASIC_AUTH,
            PackageRepositoryAuth_PackageRepositoryAuthType.TLS,
          ];
        case RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_OCI:
          return [
            PackageRepositoryAuth_PackageRepositoryAuthType.BASIC_AUTH,
            PackageRepositoryAuth_PackageRepositoryAuthType.DOCKER_CONFIG_JSON,
          ];
        default:
          return [
            PackageRepositoryAuth_PackageRepositoryAuthType.BASIC_AUTH,
            PackageRepositoryAuth_PackageRepositoryAuthType.TLS,
          ];
      }
    case PluginNames.PACKAGES_KAPP:
      // the available auth options in Carvel are type-specific
      // extracted from https://github.com/vmware-tanzu/carvel-kapp-controller/blob/v0.40.0/pkg/apis/kappctrl/v1alpha1/types_fetch.go
      // by looking for "Secret may include one"
      // also see https://carvel.dev/kapp-controller/docs/v0.43.2/app-overview/#specfetch
      switch (type) {
        // "Secret with auth details. allowed keys: ssh-privatekey, ssh-knownhosts, username, password"
        case RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_CARVEL_GIT:
          return [
            PackageRepositoryAuth_PackageRepositoryAuthType.BASIC_AUTH,
            PackageRepositoryAuth_PackageRepositoryAuthType.SSH,
          ];
        // "Secret may include one or more keys: username, password"
        case RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_CARVEL_HTTP:
          return [PackageRepositoryAuth_PackageRepositoryAuthType.BASIC_AUTH];
        // "Secret may include one or more keys: username, password, token"
        case RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_CARVEL_IMAGE:
          return [
            PackageRepositoryAuth_PackageRepositoryAuthType.BASIC_AUTH,
            PackageRepositoryAuth_PackageRepositoryAuthType.BEARER,
            PackageRepositoryAuth_PackageRepositoryAuthType.DOCKER_CONFIG_JSON,
          ];
        // "Secret may include one or more keys: username, password, token"
        case RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_CARVEL_IMGPKGBUNDLE:
          return [
            PackageRepositoryAuth_PackageRepositoryAuthType.BASIC_AUTH,
            PackageRepositoryAuth_PackageRepositoryAuthType.BEARER,
            PackageRepositoryAuth_PackageRepositoryAuthType.DOCKER_CONFIG_JSON,
          ];
        // inline is not supported for write
        case RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_CARVEL_INLINE:
          return [];
        default:
          return [];
      }
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
      return namespace === kubeappsConfig.helmGlobalNamespace;
    case PluginNames.PACKAGES_KAPP:
      return namespace === kubeappsConfig.carvelGlobalNamespace;
    // Currently, Flux doesn't support global repositories
    case PluginNames.PACKAGES_FLUX:
      return false;
    default:
      return false;
  }
}

export function getGlobalNamespaceOrNamespace(
  namespace: string,
  pluginName: string,
  kubeappsConfig: IConfig,
) {
  switch (pluginName) {
    case PluginNames.PACKAGES_HELM:
      return kubeappsConfig.helmGlobalNamespace;
    case PluginNames.PACKAGES_KAPP:
      return kubeappsConfig.carvelGlobalNamespace;
    // Currently, Flux doesn't support global repositories, so returning the namespace so we have a value
    case PluginNames.PACKAGES_FLUX:
      return namespace;
    default:
      return "unknown";
  }
}

// Using the mapping from GRPC and grpc-gateway
// See https://github.com/googleapis/googleapis/blob/master/google/rpc/code.proto
// See https://github.com/grpc-ecosystem/grpc-gateway/blob/master/runtime/errors.go
export function convertGrpcAuthError(e: any): CustomError | any {
  const msg = (e?.metadata?.get("grpc-message") || "").toString();
  switch (e?.code) {
    case Code.Unauthenticated:
      return new UnauthorizedNetworkError(msg);
    case Code.FailedPrecondition:
      // Use `FAILED_PRECONDITION` if the client should not retry until the system state has been explicitly fixed.
      //TODO(agamez): this code shouldn't be returned by the API, but it is
      if (["credentials", "unauthorized"].some(p => msg?.toLowerCase()?.includes(p))) {
        return new UnauthorizedNetworkError(msg);
      } else {
        return new BadRequestNetworkError(msg);
      }
    case Code.Internal:
      //TODO(agamez): this code shouldn't be returned by the API, but it is
      if (["credentials", "unauthorized"].some(p => msg?.toLowerCase()?.includes(p))) {
        return new UnauthorizedNetworkError(msg);
      } else {
        return new InternalServerNetworkError(msg);
      }
    case Code.PermissionDenied:
      return new ForbiddenNetworkError(msg);
    case Code.NotFound:
      return new NotFoundNetworkError(msg);
    case Code.AlreadyExists:
      return new ConflictNetworkError(msg);
    case Code.InvalidArgument:
      return new BadRequestNetworkError(msg);
    case Code.DeadlineExceeded:
      return new GatewayTimeoutNetworkError(msg);
    case Code.ResourceExhausted:
      return new TooManyRequestsNetworkError(msg);
    case Code.Aborted:
      //  Use `ABORTED` if the client should retry at a higher level
      return new ConflictNetworkError(msg);
    case Code.Unimplemented:
      return new NotImplementedNetworkError(msg);
    case Code.OutOfRange:
      return new BadRequestNetworkError(msg);
    case Code.Unavailable:
      // Use `UNAVAILABLE` if the client can retry just the failing call.
      return new ServerUnavailableNetworkError(msg);
    case Code.DataLoss:
      return new InternalServerNetworkError(msg);
    case Code.Unknown:
      return new InternalServerNetworkError(msg);
    case Code.Canceled:
      return new RequestTimeoutNetworkError(msg);
    default:
      return e;
  }
}
