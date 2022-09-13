// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { InstalledPackageStatus_StatusReason } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { PackageRepositoryAuth_PackageRepositoryAuthType } from "gen/kubeappsapis/core/packages/v1alpha1/repositories";
import carvelIcon from "icons/carvel.svg";
import fluxIcon from "icons/flux.svg";
import helmIcon from "icons/helm.svg";
import olmIcon from "icons/olm-icon.svg";
import placeholder from "icons/placeholder.svg";
import { IConfig } from "./Config";
import { PluginNames, RepositoryStorageTypes } from "./types";
import {
  escapeRegExp,
  getAppStatusLabel,
  getPluginByName,
  getPluginIcon,
  getPluginName,
  getPluginPackageName,
  getPluginsRequiringSA,
  getPluginsSupportingRollback,
  getSupportedPackageRepositoryAuthTypes,
  getValueFromEvent,
  isGlobalNamespace,
  MAX_DESC_LENGTH,
  trimDescription,
} from "./utils";

it("escapeRegExp", () => {
  const res = escapeRegExp("^dog");
  expect(res).toBe("\\^dog");
});

it("getValueFromEvent", () => {
  const trueCheckbox = {
    currentTarget: {
      value: "true",
      type: "checkbox",
    },
  } as React.FormEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>;
  const falseCheckbox = {
    currentTarget: {
      value: "false",
      type: "checkbox",
    },
  } as React.FormEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>;
  const number = {
    currentTarget: {
      value: "10",
      type: "number",
    },
  } as React.FormEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>;
  const negativeNumber = {
    currentTarget: {
      value: "-10",
      type: "number",
    },
  } as React.FormEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>;

  const trueCheckboxValue = getValueFromEvent(trueCheckbox);
  expect(trueCheckboxValue).toBe(true);
  const falseCheckboxValue = getValueFromEvent(falseCheckbox);
  expect(falseCheckboxValue).toBe(false);
  const numberValue = getValueFromEvent(number);
  expect(numberValue).toBe(10);
  const negativeNumberValue = getValueFromEvent(negativeNumber);
  expect(negativeNumberValue).toBe(-10);
});

it("trimDescription", () => {
  const shortString = "short string";
  const trimmedShortString = trimDescription(shortString);
  expect(trimmedShortString.length).toBe(shortString.length);

  const longString =
    "Bacon ipsum dolor amet pork loin ham pork filet mignon fatback meatball pork belly ball tip leberkas tail short loin pork chop swine. Picanha swine t-bone, leberkas brisket pig";
  const trimmedlongString = trimDescription(longString);
  expect(trimmedlongString.length).toBe(MAX_DESC_LENGTH);
});

it("getPluginIcon", () => {
  expect(getPluginIcon("chart")).toBe(helmIcon);
  expect(getPluginIcon("helm")).toBe(helmIcon);
  expect(getPluginIcon("operator")).toBe(olmIcon);
  expect(getPluginIcon("fluflu")).toBe(placeholder);
  expect(getPluginIcon({ name: PluginNames.PACKAGES_HELM, version: "" })).toBe(helmIcon);
  expect(getPluginIcon({ name: PluginNames.PACKAGES_FLUX, version: "" })).toBe(fluxIcon);
  expect(getPluginIcon({ name: PluginNames.PACKAGES_KAPP, version: "" })).toBe(carvelIcon);
});

it("getPluginName", () => {
  expect(getPluginName("chart")).toBe("Helm");
  expect(getPluginName("helm")).toBe("Helm");
  expect(getPluginName("operator")).toBe("Operator");
  expect(getPluginName("fluflu")).toBe("unknown plugin");
  expect(getPluginName({ name: PluginNames.PACKAGES_HELM, version: "" })).toBe("Helm");
  expect(getPluginName({ name: PluginNames.PACKAGES_FLUX, version: "" })).toBe("Flux");
  expect(getPluginName({ name: PluginNames.PACKAGES_KAPP, version: "" })).toBe("Carvel");
});

it("getPluginPackageName", () => {
  expect(getPluginPackageName("chart")).toBe("Helm Chart");
  expect(getPluginPackageName("helm")).toBe("Helm Chart");
  expect(getPluginPackageName("operator")).toBe("Operator");
  expect(getPluginPackageName("fluflu")).toBe("unknown plugin package");
  expect(getPluginPackageName({ name: PluginNames.PACKAGES_HELM, version: "" })).toBe("Helm Chart");
  expect(getPluginPackageName({ name: PluginNames.PACKAGES_FLUX, version: "" })).toBe(
    "Helm Chart via Flux",
  );
  expect(getPluginPackageName({ name: PluginNames.PACKAGES_KAPP, version: "" })).toBe(
    "Carvel Package",
  );
  expect(getPluginPackageName("chart", true)).toBe("Helm Charts");
  expect(getPluginPackageName("helm", true)).toBe("Helm Charts");
  expect(getPluginPackageName("operator", true)).toBe("Operators");
  expect(getPluginPackageName("fluflu", true)).toBe("unknown plugin packages");
  expect(getPluginPackageName({ name: PluginNames.PACKAGES_HELM, version: "" }, true)).toBe(
    "Helm Charts",
  );
  expect(getPluginPackageName({ name: PluginNames.PACKAGES_FLUX, version: "" }, true)).toBe(
    "Helm Charts via Flux",
  );
  expect(getPluginPackageName({ name: PluginNames.PACKAGES_KAPP, version: "" }, true)).toBe(
    "Carvel Packages",
  );
});

it("getPluginByName", () => {
  expect(getPluginByName(PluginNames.PACKAGES_HELM)).toStrictEqual({
    name: PluginNames.PACKAGES_HELM,
    version: "v1alpha1",
  });
  expect(getPluginByName(PluginNames.PACKAGES_FLUX)).toStrictEqual({
    name: PluginNames.PACKAGES_FLUX,
    version: "v1alpha1",
  });
  expect(getPluginByName(PluginNames.PACKAGES_KAPP)).toStrictEqual({
    name: PluginNames.PACKAGES_KAPP,
    version: "v1alpha1",
  });
  expect(getPluginByName("fluflu")).toStrictEqual({
    name: "",
    version: "",
  });
});

it("getPluginsRequiringSA", () => {
  expect(getPluginsRequiringSA()).toStrictEqual([
    PluginNames.PACKAGES_FLUX,
    PluginNames.PACKAGES_KAPP,
  ]);
});

it("getPluginsSupportingRollback", () => {
  expect(getPluginsSupportingRollback()).toStrictEqual([PluginNames.PACKAGES_HELM]);
});

it("getAppStatusLabel", () => {
  expect(getAppStatusLabel(InstalledPackageStatus_StatusReason.STATUS_REASON_UNSPECIFIED)).toBe(
    "unspecified",
  );
  expect(getAppStatusLabel(InstalledPackageStatus_StatusReason.STATUS_REASON_FAILED)).toBe(
    "failed",
  );
  expect(getAppStatusLabel(InstalledPackageStatus_StatusReason.STATUS_REASON_INSTALLED)).toBe(
    "installed",
  );
  expect(getAppStatusLabel(InstalledPackageStatus_StatusReason.STATUS_REASON_PENDING)).toBe(
    "pending",
  );
  expect(getAppStatusLabel(InstalledPackageStatus_StatusReason.STATUS_REASON_UNINSTALLED)).toBe(
    "uninstalled",
  );
  expect(getAppStatusLabel(InstalledPackageStatus_StatusReason.UNRECOGNIZED)).toBe("unrecognized");
});

it("getSupportedPackageRepositoryAuthTypes", () => {
  expect(
    getSupportedPackageRepositoryAuthTypes({
      name: PluginNames.PACKAGES_HELM,
      version: "",
    }).toString(),
  ).toBe(
    [
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_AUTHORIZATION_HEADER,
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BEARER,
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON,
    ].toString(),
  );
  expect(
    getSupportedPackageRepositoryAuthTypes(
      { name: PluginNames.PACKAGES_HELM, version: "" },
      RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_HELM,
    ).toString(),
  ).toBe(
    [
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_AUTHORIZATION_HEADER,
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BEARER,
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON,
    ].toString(),
  );
  expect(
    getSupportedPackageRepositoryAuthTypes(
      { name: PluginNames.PACKAGES_HELM, version: "" },
      RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_OCI,
    ).toString(),
  ).toBe(
    [
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_AUTHORIZATION_HEADER,
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BEARER,
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON,
    ].toString(),
  );
  expect(
    getSupportedPackageRepositoryAuthTypes({
      name: PluginNames.PACKAGES_FLUX,
      version: "",
    }).toString(),
  ).toBe(
    [
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_OPAQUE,
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_SSH,
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_TLS,
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON,
    ].toString(),
  );
  expect(
    getSupportedPackageRepositoryAuthTypes(
      { name: PluginNames.PACKAGES_FLUX, version: "" },
      RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_HELM,
    ).toString(),
  ).toBe(
    [
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_OPAQUE,
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_SSH,
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_TLS,
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON,
    ].toString(),
  );
  expect(
    getSupportedPackageRepositoryAuthTypes(
      { name: PluginNames.PACKAGES_FLUX, version: "" },
      RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_OCI,
    ).toString(),
  ).toBe(
    [
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_OPAQUE,
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_SSH,
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_TLS,
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON,
    ].toString(),
  );
  expect(
    getSupportedPackageRepositoryAuthTypes({
      name: PluginNames.PACKAGES_KAPP,
      version: "",
    }).toString(),
  ).toBe([].toString());
  expect(
    getSupportedPackageRepositoryAuthTypes(
      { name: PluginNames.PACKAGES_KAPP, version: "" },
      RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_CARVEL_GIT,
    ).toString(),
  ).toBe(
    [
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_SSH,
    ].toString(),
  );
  expect(
    getSupportedPackageRepositoryAuthTypes(
      { name: PluginNames.PACKAGES_KAPP, version: "" },
      RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_CARVEL_HTTP,
    ).toString(),
  ).toBe(
    [
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
    ].toString(),
  );
  expect(
    getSupportedPackageRepositoryAuthTypes(
      { name: PluginNames.PACKAGES_KAPP, version: "" },
      RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_CARVEL_IMAGE,
    ).toString(),
  ).toBe(
    [
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BEARER,
    ].toString(),
  );
  expect(
    getSupportedPackageRepositoryAuthTypes(
      { name: PluginNames.PACKAGES_KAPP, version: "" },
      RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_CARVEL_IMGPKGBUNDLE,
    ).toString(),
  ).toBe(
    [
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BEARER,
    ].toString(),
  );
  expect(
    getSupportedPackageRepositoryAuthTypes(
      { name: PluginNames.PACKAGES_KAPP, version: "" },
      RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_CARVEL_INLINE,
    ).toString(),
  ).toBe([].toString());
  expect(getSupportedPackageRepositoryAuthTypes({ name: "foo", version: "" }).toString()).toBe(
    [].toString(),
  );
});

it("isGlobalNamespace", () => {
  const kubeappsConfig = {
    helmGlobalNamespace: "helm-global",
    carvelGlobalNamespace: "carvel-global",
  } as IConfig;
  expect(isGlobalNamespace("helm-global", PluginNames.PACKAGES_HELM, kubeappsConfig)).toBe(true);
  expect(isGlobalNamespace("helm-global", PluginNames.PACKAGES_KAPP, kubeappsConfig)).toBe(false);
  expect(isGlobalNamespace("helm-global", PluginNames.PACKAGES_FLUX, kubeappsConfig)).toBe(true);
  expect(isGlobalNamespace("carvel-global", PluginNames.PACKAGES_HELM, kubeappsConfig)).toBe(false);
  expect(isGlobalNamespace("carvel-global", PluginNames.PACKAGES_KAPP, kubeappsConfig)).toBe(true);
  expect(isGlobalNamespace("carvel-global", PluginNames.PACKAGES_FLUX, kubeappsConfig)).toBe(true);
});
