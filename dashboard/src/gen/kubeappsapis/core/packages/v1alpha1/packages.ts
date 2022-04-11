/* eslint-disable */
import Long from "long";
import { grpc } from "@improbable-eng/grpc-web";
import _m0 from "protobufjs/minimal";
import { Any } from "../../../../google/protobuf/any";
import { Plugin } from "../../../../kubeappsapis/core/plugins/v1alpha1/plugins";
import { BrowserHeaders } from "browser-headers";

export const protobufPackage = "kubeappsapis.core.packages.v1alpha1";

/**
 * GetAvailablePackageSummariesRequest
 *
 * Request for GetAvailablePackageSummaries
 */
export interface GetAvailablePackageSummariesRequest {
  /** The context (cluster/namespace) for the request */
  context?: Context;
  /** The filters used for the request */
  filterOptions?: FilterOptions;
  /** Pagination options specifying where to start and how many results to include. */
  paginationOptions?: PaginationOptions;
}

/**
 * GetAvailablePackageDetailRequest
 *
 * Request for GetAvailablePackageDetail
 */
export interface GetAvailablePackageDetailRequest {
  /**
   * The information required to uniquely
   * identify an available package
   */
  availablePackageRef?: AvailablePackageReference;
  /**
   * Optional specific version (or version reference) to request.
   * By default the latest version (or latest version matching the reference)
   * will be returned.
   */
  pkgVersion: string;
}

/**
 * GetAvailablePackageVersionsRequest
 *
 * Request for GetAvailablePackageVersions
 */
export interface GetAvailablePackageVersionsRequest {
  /**
   * The information required to uniquely
   * identify an available package
   */
  availablePackageRef?: AvailablePackageReference;
  /**
   * Optional version reference for which full version history is required.  By
   * default a summary of versions is returned as outlined in the response.
   * Plugins can choose not to implement this and provide the summary only, it
   * is provided for completeness only.
   */
  pkgVersion: string;
}

/**
 * GetInstalledPackageSummariesRequest
 *
 * Request for GetInstalledPackageSummaries
 */
export interface GetInstalledPackageSummariesRequest {
  /** The context (cluster/namespace) for the request. */
  context?: Context;
  /** Pagination options specifying where to start and how many results to include. */
  paginationOptions?: PaginationOptions;
}

/**
 * GetInstalledPackageDetailRequest
 *
 * Request for GetInstalledPackageDetail
 */
export interface GetInstalledPackageDetailRequest {
  /**
   * The information required to uniquely
   * identify an installed package
   */
  installedPackageRef?: InstalledPackageReference;
}

/**
 * CreateInstalledPackageRequest
 *
 * Request for CreateInstalledPackage
 */
export interface CreateInstalledPackageRequest {
  /** A reference uniquely identifying the package available for installation. */
  availablePackageRef?: AvailablePackageReference;
  /** The target context where the package is intended to be installed. */
  targetContext?: Context;
  /** A user-provided name for the installed package (eg. project-x-db) */
  name: string;
  /**
   * For helm this will be the exact version in VersionReference.version
   * For other plugins we can extend the VersionReference as needed.
   */
  pkgVersionReference?: VersionReference;
  /**
   * An optional serialized values string to be included when templating a package
   * in the format expected by the plugin. Included when the backend format doesn't
   * use secrets or configmaps for values or supports both. These values are layered
   * on top of any values refs above, when relevant.
   */
  values: string;
  /**
   * An optional field for specifying data common to systems that reconcile
   * the package on the cluster.
   */
  reconciliationOptions?: ReconciliationOptions;
}

/**
 * UpdateInstalledPackageRequest
 *
 * Request for UpdateInstalledPackage. The intent is to reach the desired state specified
 * by the fields in the request, while leaving other fields intact. This is a whole
 * object "Update" semantics rather than "Patch" semantics. The caller will provide the
 * values for the fields below, which will replace, or be overlayed onto, the
 * corresponding fields in the existing resource. For example, with the
 * UpdateInstalledPackageRequest, it is not possible to change just the 'package version
 * reference' without also specifying 'values' field. As a side effect, not specifying the
 * 'values' field in the request means there are no values specified in the desired state.
 * So the meaning of each field value is describing the desired state of the corresponding
 * field in the resource after the update operation has completed the renconciliation.
 */
export interface UpdateInstalledPackageRequest {
  /**
   * A reference uniquely identifying the installed package being updated.
   * Required
   */
  installedPackageRef?: InstalledPackageReference;
  /**
   * For helm this will be the exact version in VersionReference.version
   * For fluxv2 this could be any semver constraint expression
   * For other plugins we can extend the VersionReference as needed. Optional
   */
  pkgVersionReference?: VersionReference;
  /**
   * An optional serialized values string to be included when templating a
   * package in the format expected by the plugin. Included when the backend
   * format doesn't use secrets or configmaps for values or supports both.
   * These values are layered on top of any values refs above, when
   * relevant.
   */
  values: string;
  /**
   * An optional field for specifying data common to systems that reconcile
   * the package on the cluster.
   */
  reconciliationOptions?: ReconciliationOptions;
}

/**
 * DeleteInstalledPackageRequest
 *
 * Request for DeleteInstalledPackage
 */
export interface DeleteInstalledPackageRequest {
  /** A reference to uniquely identify the installed package to be deleted. */
  installedPackageRef?: InstalledPackageReference;
}

/**
 * GetInstalledPackageResourceRefsRequest
 *
 * Request for GetInstalledPackageResourceRefs
 */
export interface GetInstalledPackageResourceRefsRequest {
  installedPackageRef?: InstalledPackageReference;
}

/**
 * GetAvailablePackageSummariesResponse
 *
 * Response for GetAvailablePackageSummaries
 */
export interface GetAvailablePackageSummariesResponse {
  /**
   * Available packages summaries
   *
   * List of AvailablePackageSummary
   */
  availablePackageSummaries: AvailablePackageSummary[];
  /**
   * Next page token
   *
   * This field represents the pagination token to retrieve the next page of
   * results. If the value is "", it means no further results for the request.
   */
  nextPageToken: string;
  /**
   * Categories
   *
   * This optional field contains the distinct category names considering the FilterOptions.
   */
  categories: string[];
}

/**
 * GetAvailablePackageDetailResponse
 *
 * Response for GetAvailablePackageDetail
 */
export interface GetAvailablePackageDetailResponse {
  /**
   * Available package detail
   *
   * The requested AvailablePackageDetail
   */
  availablePackageDetail?: AvailablePackageDetail;
}

/**
 * GetAvailablePackageVersionsResponse
 *
 * Response for GetAvailablePackageVersions
 */
export interface GetAvailablePackageVersionsResponse {
  /**
   * Package app versions
   *
   * By default (when version_query is empty or ignored) the response
   * should contain an ordered summary of versions including the most recent three
   * patch versions of the most recent three minor versions of the most recent three
   * major versions when available, something like:
   * [
   *   { pkg_version: "10.3.19", app_version: "2.16.8" },
   *   { pkg_version: "10.3.18", app_version: "2.16.8" },
   *   { pkg_version: "10.3.17", app_version: "2.16.7" },
   *   { pkg_version: "10.2.6", app_version: "2.15.3" },
   *   { pkg_version: "10.2.5", app_version: "2.15.2" },
   *   { pkg_version: "10.2.4", app_version: "2.15.2" },
   *   { pkg_version: "10.1.8", app_version: "2.13.5" },
   *   { pkg_version: "10.1.7", app_version: "2.13.5" },
   *   { pkg_version: "10.1.6", app_version: "2.13.5" },
   *   { pkg_version: "9.5.4", app_version: "2.8.9" },
   *   ...
   *   { pkg_version: "8.2.5", app_version: "1.19.5" },
   *   ...
   * ]
   * If a version_query is present and the plugin chooses to support it,
   * the full history of versions matching the version query should be returned.
   */
  packageAppVersions: PackageAppVersion[];
}

/**
 * GetInstalledPackageSummariesResponse
 *
 * Response for GetInstalledPackageSummaries
 */
export interface GetInstalledPackageSummariesResponse {
  /**
   * Installed packages summaries
   *
   * List of InstalledPackageSummary
   */
  installedPackageSummaries: InstalledPackageSummary[];
  /**
   * Next page token
   *
   * This field represents the pagination token to retrieve the next page of
   * results. If the value is "", it means no further results for the request.
   */
  nextPageToken: string;
}

/**
 * GetInstalledPackageDetailResponse
 *
 * Response for GetInstalledPackageDetail
 */
export interface GetInstalledPackageDetailResponse {
  /**
   * InstalledPackageDetail
   *
   * The requested InstalledPackageDetail
   */
  installedPackageDetail?: InstalledPackageDetail;
}

/**
 * CreateInstalledPackageResponse
 *
 * Response for CreateInstalledPackage
 */
export interface CreateInstalledPackageResponse {
  installedPackageRef?: InstalledPackageReference;
}

/**
 * UpdateInstalledPackageResponse
 *
 * Response for UpdateInstalledPackage
 */
export interface UpdateInstalledPackageResponse {
  installedPackageRef?: InstalledPackageReference;
}

/**
 * DeleteInstalledPackageResponse
 *
 * Response for DeleteInstalledPackage
 */
export interface DeleteInstalledPackageResponse {}

/**
 * GetInstalledPackageResourceRefsResponse
 *
 * Response for GetInstalledPackageResourceRefs
 */
export interface GetInstalledPackageResourceRefsResponse {
  context?: Context;
  resourceRefs: ResourceRef[];
}

/**
 * AvailablePackageSummary
 *
 * An AvailablePackageSummary provides a summary of a package available for installation
 * useful when aggregating many available packages.
 */
export interface AvailablePackageSummary {
  /**
   * Available package reference
   *
   * A reference uniquely identifying the package.
   */
  availablePackageRef?: AvailablePackageReference;
  /**
   * Available package name
   *
   * The name of the available package
   */
  name: string;
  /**
   * Latest available version
   *
   * The latest version available for this package. Often expected when viewing
   * a summary of many available packages.
   */
  latestVersion?: PackageAppVersion;
  /**
   * Available package Icon URL
   *
   * A url for an icon.
   */
  iconUrl: string;
  /**
   * Available package display name
   *
   * A name as displayed to users
   */
  displayName: string;
  /**
   * Available package short description
   *
   * A short description of the app provided by the package
   */
  shortDescription: string;
  /**
   * Available package categories
   *
   * A user-facing list of category names useful for creating richer user interfaces.
   * Plugins can choose not to implement this
   */
  categories: string[];
}

/**
 * AvailablePackageDetail
 *
 * An AvailablePackageDetail provides additional details required when
 * inspecting an individual package.
 */
export interface AvailablePackageDetail {
  /**
   * Available package reference
   *
   * A reference uniquely identifying the package.
   */
  availablePackageRef?: AvailablePackageReference;
  /**
   * Available package name
   *
   * The name of the available package
   */
  name: string;
  /**
   * Available version
   *
   * The version of the package and application.
   */
  version?: PackageAppVersion;
  /** the url of the package repository that contains this package */
  repoUrl: string;
  /** the url of the “home” for the package */
  homeUrl: string;
  /**
   * Available package icon URL
   *
   * A url for an icon.
   */
  iconUrl: string;
  /**
   * Available package display name
   *
   * A name as displayed to users
   */
  displayName: string;
  /**
   * Available package short description
   *
   * A short description of the app provided by the package
   */
  shortDescription: string;
  /**
   * Available package long description
   *
   * A longer description of the package, a few sentences.
   */
  longDescription: string;
  /**
   * Available package readme
   *
   * A longer README with potentially pages of formatted Markdown.
   */
  readme: string;
  /**
   * Available package default values
   *
   * An example of default values used during package templating that can serve
   * as documentation or a starting point for user customization.
   */
  defaultValues: string;
  valuesSchema: string;
  /** source urls for the package */
  sourceUrls: string[];
  /**
   * Available package maintainers
   *
   * List of Maintainer
   */
  maintainers: Maintainer[];
  /**
   * Available package categories
   *
   * A user-facing list of category names useful for creating richer user interfaces.
   * Plugins can choose not to implement this
   */
  categories: string[];
  /**
   * Custom data added by the plugin
   *
   * A plugin can define custom details for data which is not yet, or never will
   * be specified in the core.packaging.CreateInstalledPackageRequest fields. The use
   * of an `Any` field means that each plugin can define the structure of this
   * message as required, while still satisfying the core interface.
   * See https://developers.google.com/protocol-buffers/docs/proto3#any
   */
  customDetail?: Any;
}

/**
 * InstalledPackageSummary
 *
 * An InstalledPackageSummary provides a summary of an installed package
 * useful when aggregating many installed packages.
 */
export interface InstalledPackageSummary {
  /**
   * InstalledPackageReference
   *
   * A reference uniquely identifying the package.
   */
  installedPackageRef?: InstalledPackageReference;
  /**
   * Name
   *
   * A name given to the installation of the package (eg. "my-postgresql-for-testing").
   */
  name: string;
  /**
   * PkgVersionReference
   *
   * The package version reference defines a version or constraint limiting
   * matching package versions.
   */
  pkgVersionReference?: VersionReference;
  /**
   * CurrentVersion
   *
   * The current version of the package being reconciled, which may be
   * in one of these states:
   *  - has been successfully installed/upgraded or
   *  - is currently being installed/upgraded or
   *  - has failed to install/upgrade
   */
  currentVersion?: PackageAppVersion;
  /**
   * Installed package icon URL
   *
   * A url for an icon.
   */
  iconUrl: string;
  /**
   * PackageDisplayName
   *
   * The package name as displayed to users (provided by the package, eg. "PostgreSQL")
   */
  pkgDisplayName: string;
  /**
   * ShortDescription
   *
   * A short description of the package (provided by the package)
   */
  shortDescription: string;
  /**
   * LatestMatchingVersion
   *
   * Only non-empty if an available upgrade matches the specified pkg_version_reference.
   * For example, if the pkg_version_reference is ">10.3.0 < 10.4.0" and 10.3.1
   * is installed, then:
   *   * if 10.3.2 is available, latest_matching_version should be 10.3.2, but
   *   * if 10.4 is available while >10.3.1 is not, this should remain empty.
   */
  latestMatchingVersion?: PackageAppVersion;
  /**
   * LatestVersion
   *
   * The latest version available for this package, regardless of the pkg_version_reference.
   */
  latestVersion?: PackageAppVersion;
  /**
   * Status
   *
   * The current status of the installed package.
   */
  status?: InstalledPackageStatus;
}

/**
 * InstalledPackageDetail
 *
 * An InstalledPackageDetail includes details about the installed package that are
 * typically useful when presenting a single installed package.
 */
export interface InstalledPackageDetail {
  /**
   * InstalledPackageReference
   *
   * A reference uniquely identifying the installed package.
   */
  installedPackageRef?: InstalledPackageReference;
  /**
   * PkgVersionReference
   *
   * The package version reference defines a version or constraint limiting
   * matching package versions.
   */
  pkgVersionReference?: VersionReference;
  /**
   * Installed package name
   *
   * The name given to the installed package
   */
  name: string;
  /**
   * CurrentVersion
   *
   * The version of the package which is currently installed.
   */
  currentVersion?: PackageAppVersion;
  /**
   * ValuesApplied
   *
   * The values applied currently for the installed package.
   */
  valuesApplied: string;
  /**
   * ReconciliationOptions
   *
   * An optional field specifying data common to systems that reconcile
   * the package installation on the cluster asynchronously. In particular,
   * this specifies the service account used to perform the reconcilliation.
   */
  reconciliationOptions?: ReconciliationOptions;
  /**
   * Status
   *
   * The current status of the installed package.
   */
  status?: InstalledPackageStatus;
  /**
   * PostInstallationNotes
   *
   * Optional notes generated by package and intended for the user post installation.
   */
  postInstallationNotes: string;
  /**
   * Available package reference
   *
   * A reference to the available package for this installation.
   * Useful to lookup the package display name, icon and other info.
   */
  availablePackageRef?: AvailablePackageReference;
  /**
   * LatestMatchingVersion
   *
   * Only non-empty if an available upgrade matches the specified pkg_version_reference.
   * For example, if the pkg_version_reference is ">10.3.0 < 10.4.0" and 10.3.1
   * is installed, then:
   *   * if 10.3.2 is available, latest_matching_version should be 10.3.2, but
   *   * if 10.4 is available while >10.3.1 is not, this should remain empty.
   */
  latestMatchingVersion?: PackageAppVersion;
  /**
   * LatestVersion
   *
   * The latest version available for this package, regardless of the pkg_version_reference.
   */
  latestVersion?: PackageAppVersion;
  /**
   * Custom data added by the plugin
   *
   * A plugin can define custom details for data which is not yet, or never will
   * be specified in the core.packaging.CreateInstalledPackageRequest fields. The use
   * of an `Any` field means that each plugin can define the structure of this
   * message as required, while still satisfying the core interface.
   * See https://developers.google.com/protocol-buffers/docs/proto3#any
   */
  customDetail?: Any;
}

/**
 * Context
 *
 * A Context specifies the context of the message
 */
export interface Context {
  /**
   * Cluster
   *
   * A cluster name can be provided to target a specific cluster if multiple
   * clusters are configured, otherwise all clusters will be assumed.
   */
  cluster: string;
  /**
   * Namespace
   *
   * A namespace must be provided if the context of the operation is for a resource
   * or resources in a particular namespace.
   * For requests to list items, not including a namespace here implies that the context
   * for the request is everything the requesting user can read, though the result can
   * be filtered by any filtering options of the request. Plugins may choose to return
   * Unimplemented for some queries for which we do not yet have a need.
   */
  namespace: string;
}

/**
 * AvailablePackageReference
 *
 * An AvailablePackageReference has the minimum information required to uniquely
 * identify an available package. This is re-used on the summary and details of an
 * available package.
 */
export interface AvailablePackageReference {
  /**
   * Available package context
   *
   * The context (cluster/namespace) for the package.
   */
  context?: Context;
  /**
   * Available package identifier
   *
   * The fully qualified identifier for the available package
   * (ie. a unique name for the context). For some packaging systems
   * (particularly those where an available package is backed by a CR) this
   * will just be the name, but for others such as those where an available
   * package is not backed by a CR (eg. standard helm) it may be necessary
   * to include the repository in the name or even the repo namespace
   * to ensure this is unique.
   * For example two helm repositories can define
   * an "apache" chart that is available globally, the names would need to
   * encode that to be unique (ie. "repoA:apache" and "repoB:apache").
   */
  identifier: string;
  /**
   * Plugin for the available package
   *
   * The plugin used to interact with this available package.
   * This field should be omitted when the request is in the context of a specific plugin.
   */
  plugin?: Plugin;
}

/**
 * Maintainer
 *
 * Maintainers for the package.
 */
export interface Maintainer {
  /**
   * Maintainer name
   *
   * A maintainer name
   */
  name: string;
  /**
   * Maintainer email
   *
   * A maintainer email
   */
  email: string;
}

/**
 * FilterOptions
 *
 * FilterOptions available when requesting summaries
 */
export interface FilterOptions {
  /**
   * Text query
   *
   * Text query for the request
   */
  query: string;
  /**
   * Categories
   *
   * Collection of categories for the request
   */
  categories: string[];
  /**
   * Repositories
   *
   * Collection of repositories where the packages belong to
   */
  repositories: string[];
  /**
   * Package version
   *
   * Package version for the request
   */
  pkgVersion: string;
  /**
   * App version
   *
   * Packaged app version for the request
   */
  appVersion: string;
}

/**
 * PaginationOptions
 *
 * The PaginationOptions based on the example proto at:
 * https://cloud.google.com/apis/design/design_patterns#list_pagination
 * just encapsulated in a message so it can be reused on different request messages.
 */
export interface PaginationOptions {
  /**
   * Page token
   *
   * The client uses this field to request a specific page of the list results.
   */
  pageToken: string;
  /**
   * Page size
   *
   * Clients use this field to specify the maximum number of results to be
   * returned by the server. The server may further constrain the maximum number
   * of results returned in a single page. If the page_size is 0, the server
   * will decide the number of results to be returned.
   */
  pageSize: number;
}

/**
 * InstalledPackageReference
 *
 * An InstalledPackageReference has the minimum information required to uniquely
 * identify an installed package.
 */
export interface InstalledPackageReference {
  /**
   * Installed package context
   *
   * The context (cluster/namespace) for the package.
   */
  context?: Context;
  /**
   * The fully qualified identifier for the installed package
   * (ie. a unique name for the context).
   */
  identifier: string;
  /**
   * The plugin used to identify and interact with the installed package.
   * This field can be omitted when the request is in the context of a specific plugin.
   */
  plugin?: Plugin;
}

/**
 * VersionReference
 *
 * A VersionReference defines a version or constraint limiting matching versions.
 * The reason it is a separate message is so that in the future we can add other
 * fields as necessary (such as something similar to Carvel's `prereleases` option
 * to its versionSelection).
 */
export interface VersionReference {
  /**
   * Version
   *
   * The format of the version constraint depends on the backend. For example,
   * for a flux v2 and Carvel it’s a semver expression, such as ">=10.3 < 10.4"
   */
  version: string;
}

/**
 * InstalledPackageStatus
 *
 * An InstalledPackageStatus reports on the current status of the installation.
 */
export interface InstalledPackageStatus {
  /**
   * Ready
   *
   * An indication of whether the installation is ready or not
   */
  ready: boolean;
  /**
   * Reason
   *
   * An enum indicating the reason for the current status.
   */
  reason: InstalledPackageStatus_StatusReason;
  /**
   * UserReason
   *
   * Optional text to return for user context, which may be plugin specific.
   */
  userReason: string;
}

/**
 * StatusReason
 *
 * Generic reasons why an installed package may be ready or not.
 * These should make sense across different packaging plugins.
 */
export enum InstalledPackageStatus_StatusReason {
  STATUS_REASON_UNSPECIFIED = 0,
  STATUS_REASON_INSTALLED = 1,
  STATUS_REASON_UNINSTALLED = 2,
  STATUS_REASON_FAILED = 3,
  STATUS_REASON_PENDING = 4,
  UNRECOGNIZED = -1,
}

export function installedPackageStatus_StatusReasonFromJSON(
  object: any,
): InstalledPackageStatus_StatusReason {
  switch (object) {
    case 0:
    case "STATUS_REASON_UNSPECIFIED":
      return InstalledPackageStatus_StatusReason.STATUS_REASON_UNSPECIFIED;
    case 1:
    case "STATUS_REASON_INSTALLED":
      return InstalledPackageStatus_StatusReason.STATUS_REASON_INSTALLED;
    case 2:
    case "STATUS_REASON_UNINSTALLED":
      return InstalledPackageStatus_StatusReason.STATUS_REASON_UNINSTALLED;
    case 3:
    case "STATUS_REASON_FAILED":
      return InstalledPackageStatus_StatusReason.STATUS_REASON_FAILED;
    case 4:
    case "STATUS_REASON_PENDING":
      return InstalledPackageStatus_StatusReason.STATUS_REASON_PENDING;
    case -1:
    case "UNRECOGNIZED":
    default:
      return InstalledPackageStatus_StatusReason.UNRECOGNIZED;
  }
}

export function installedPackageStatus_StatusReasonToJSON(
  object: InstalledPackageStatus_StatusReason,
): string {
  switch (object) {
    case InstalledPackageStatus_StatusReason.STATUS_REASON_UNSPECIFIED:
      return "STATUS_REASON_UNSPECIFIED";
    case InstalledPackageStatus_StatusReason.STATUS_REASON_INSTALLED:
      return "STATUS_REASON_INSTALLED";
    case InstalledPackageStatus_StatusReason.STATUS_REASON_UNINSTALLED:
      return "STATUS_REASON_UNINSTALLED";
    case InstalledPackageStatus_StatusReason.STATUS_REASON_FAILED:
      return "STATUS_REASON_FAILED";
    case InstalledPackageStatus_StatusReason.STATUS_REASON_PENDING:
      return "STATUS_REASON_PENDING";
    default:
      return "UNKNOWN";
  }
}

/**
 * ReconciliationOptions
 *
 * ReconciliationOptions enable specifying standard fields for backends that continuously
 * reconcile a package install as new matching versions are released. Most of the naming
 * is from the flux HelmReleaseSpec though it maps directly to equivalent fields on Carvel's
 * InstalledPackage.
 */
export interface ReconciliationOptions {
  /**
   * Reconciliation Interval
   *
   * The interval with which the package is checked for reconciliation (in seconds)
   */
  interval: number;
  /**
   * Suspend
   *
   * Whether reconciliation should be suspended until otherwise enabled.
   * This can be utilized to e.g. temporarily ignore chart changes, and
   * prevent a Helm release from getting upgraded
   */
  suspend: boolean;
  /**
   * ServiceAccountName
   *
   * A name for a service account in the same namespace which should be used
   * to perform the reconciliation.
   */
  serviceAccountName: string;
}

/**
 * Package AppVersion
 *
 * PackageAppVersion conveys both the package version and the packaged app version.
 */
export interface PackageAppVersion {
  /**
   * Package version
   *
   * Version of the package itself
   */
  pkgVersion: string;
  /**
   * Application version
   *
   * Version of the packaged application
   */
  appVersion: string;
}

/**
 * Resource reference
 *
 * A reference to a Kubernetes resource related to a specific installed package.
 * The context (cluster) for each resource is that of the related
 * installed package.
 */
export interface ResourceRef {
  /**
   * The APIVersion directly from the resource has the group and version, eg. "apps/v1"
   * or just the version for core resources.
   */
  apiVersion: string;
  /**
   * The Kind directly from the templated manifest. Together with the APIVersion this
   * forms the GroupVersionKind.
   */
  kind: string;
  /** The name of the specific resource in the context of the installed package. */
  name: string;
  /**
   * The namespace of the specific resource in the context of the installed
   * package. In most cases this will be identical to the namespace of the
   * installed package. Exceptions will be non-namespaced resources and packages
   * that install resources in other namespaces for special reasons.
   */
  namespace: string;
}

const baseGetAvailablePackageSummariesRequest: object = {};

export const GetAvailablePackageSummariesRequest = {
  encode(
    message: GetAvailablePackageSummariesRequest,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.context !== undefined) {
      Context.encode(message.context, writer.uint32(10).fork()).ldelim();
    }
    if (message.filterOptions !== undefined) {
      FilterOptions.encode(message.filterOptions, writer.uint32(18).fork()).ldelim();
    }
    if (message.paginationOptions !== undefined) {
      PaginationOptions.encode(message.paginationOptions, writer.uint32(26).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetAvailablePackageSummariesRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseGetAvailablePackageSummariesRequest,
    } as GetAvailablePackageSummariesRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.context = Context.decode(reader, reader.uint32());
          break;
        case 2:
          message.filterOptions = FilterOptions.decode(reader, reader.uint32());
          break;
        case 3:
          message.paginationOptions = PaginationOptions.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): GetAvailablePackageSummariesRequest {
    const message = {
      ...baseGetAvailablePackageSummariesRequest,
    } as GetAvailablePackageSummariesRequest;
    if (object.context !== undefined && object.context !== null) {
      message.context = Context.fromJSON(object.context);
    } else {
      message.context = undefined;
    }
    if (object.filterOptions !== undefined && object.filterOptions !== null) {
      message.filterOptions = FilterOptions.fromJSON(object.filterOptions);
    } else {
      message.filterOptions = undefined;
    }
    if (object.paginationOptions !== undefined && object.paginationOptions !== null) {
      message.paginationOptions = PaginationOptions.fromJSON(object.paginationOptions);
    } else {
      message.paginationOptions = undefined;
    }
    return message;
  },

  toJSON(message: GetAvailablePackageSummariesRequest): unknown {
    const obj: any = {};
    message.context !== undefined &&
      (obj.context = message.context ? Context.toJSON(message.context) : undefined);
    message.filterOptions !== undefined &&
      (obj.filterOptions = message.filterOptions
        ? FilterOptions.toJSON(message.filterOptions)
        : undefined);
    message.paginationOptions !== undefined &&
      (obj.paginationOptions = message.paginationOptions
        ? PaginationOptions.toJSON(message.paginationOptions)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<GetAvailablePackageSummariesRequest>,
  ): GetAvailablePackageSummariesRequest {
    const message = {
      ...baseGetAvailablePackageSummariesRequest,
    } as GetAvailablePackageSummariesRequest;
    if (object.context !== undefined && object.context !== null) {
      message.context = Context.fromPartial(object.context);
    } else {
      message.context = undefined;
    }
    if (object.filterOptions !== undefined && object.filterOptions !== null) {
      message.filterOptions = FilterOptions.fromPartial(object.filterOptions);
    } else {
      message.filterOptions = undefined;
    }
    if (object.paginationOptions !== undefined && object.paginationOptions !== null) {
      message.paginationOptions = PaginationOptions.fromPartial(object.paginationOptions);
    } else {
      message.paginationOptions = undefined;
    }
    return message;
  },
};

const baseGetAvailablePackageDetailRequest: object = { pkgVersion: "" };

export const GetAvailablePackageDetailRequest = {
  encode(
    message: GetAvailablePackageDetailRequest,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.availablePackageRef !== undefined) {
      AvailablePackageReference.encode(
        message.availablePackageRef,
        writer.uint32(10).fork(),
      ).ldelim();
    }
    if (message.pkgVersion !== "") {
      writer.uint32(18).string(message.pkgVersion);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetAvailablePackageDetailRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseGetAvailablePackageDetailRequest,
    } as GetAvailablePackageDetailRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.availablePackageRef = AvailablePackageReference.decode(reader, reader.uint32());
          break;
        case 2:
          message.pkgVersion = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): GetAvailablePackageDetailRequest {
    const message = {
      ...baseGetAvailablePackageDetailRequest,
    } as GetAvailablePackageDetailRequest;
    if (object.availablePackageRef !== undefined && object.availablePackageRef !== null) {
      message.availablePackageRef = AvailablePackageReference.fromJSON(object.availablePackageRef);
    } else {
      message.availablePackageRef = undefined;
    }
    if (object.pkgVersion !== undefined && object.pkgVersion !== null) {
      message.pkgVersion = String(object.pkgVersion);
    } else {
      message.pkgVersion = "";
    }
    return message;
  },

  toJSON(message: GetAvailablePackageDetailRequest): unknown {
    const obj: any = {};
    message.availablePackageRef !== undefined &&
      (obj.availablePackageRef = message.availablePackageRef
        ? AvailablePackageReference.toJSON(message.availablePackageRef)
        : undefined);
    message.pkgVersion !== undefined && (obj.pkgVersion = message.pkgVersion);
    return obj;
  },

  fromPartial(
    object: DeepPartial<GetAvailablePackageDetailRequest>,
  ): GetAvailablePackageDetailRequest {
    const message = {
      ...baseGetAvailablePackageDetailRequest,
    } as GetAvailablePackageDetailRequest;
    if (object.availablePackageRef !== undefined && object.availablePackageRef !== null) {
      message.availablePackageRef = AvailablePackageReference.fromPartial(
        object.availablePackageRef,
      );
    } else {
      message.availablePackageRef = undefined;
    }
    if (object.pkgVersion !== undefined && object.pkgVersion !== null) {
      message.pkgVersion = object.pkgVersion;
    } else {
      message.pkgVersion = "";
    }
    return message;
  },
};

const baseGetAvailablePackageVersionsRequest: object = { pkgVersion: "" };

export const GetAvailablePackageVersionsRequest = {
  encode(
    message: GetAvailablePackageVersionsRequest,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.availablePackageRef !== undefined) {
      AvailablePackageReference.encode(
        message.availablePackageRef,
        writer.uint32(10).fork(),
      ).ldelim();
    }
    if (message.pkgVersion !== "") {
      writer.uint32(18).string(message.pkgVersion);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetAvailablePackageVersionsRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseGetAvailablePackageVersionsRequest,
    } as GetAvailablePackageVersionsRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.availablePackageRef = AvailablePackageReference.decode(reader, reader.uint32());
          break;
        case 2:
          message.pkgVersion = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): GetAvailablePackageVersionsRequest {
    const message = {
      ...baseGetAvailablePackageVersionsRequest,
    } as GetAvailablePackageVersionsRequest;
    if (object.availablePackageRef !== undefined && object.availablePackageRef !== null) {
      message.availablePackageRef = AvailablePackageReference.fromJSON(object.availablePackageRef);
    } else {
      message.availablePackageRef = undefined;
    }
    if (object.pkgVersion !== undefined && object.pkgVersion !== null) {
      message.pkgVersion = String(object.pkgVersion);
    } else {
      message.pkgVersion = "";
    }
    return message;
  },

  toJSON(message: GetAvailablePackageVersionsRequest): unknown {
    const obj: any = {};
    message.availablePackageRef !== undefined &&
      (obj.availablePackageRef = message.availablePackageRef
        ? AvailablePackageReference.toJSON(message.availablePackageRef)
        : undefined);
    message.pkgVersion !== undefined && (obj.pkgVersion = message.pkgVersion);
    return obj;
  },

  fromPartial(
    object: DeepPartial<GetAvailablePackageVersionsRequest>,
  ): GetAvailablePackageVersionsRequest {
    const message = {
      ...baseGetAvailablePackageVersionsRequest,
    } as GetAvailablePackageVersionsRequest;
    if (object.availablePackageRef !== undefined && object.availablePackageRef !== null) {
      message.availablePackageRef = AvailablePackageReference.fromPartial(
        object.availablePackageRef,
      );
    } else {
      message.availablePackageRef = undefined;
    }
    if (object.pkgVersion !== undefined && object.pkgVersion !== null) {
      message.pkgVersion = object.pkgVersion;
    } else {
      message.pkgVersion = "";
    }
    return message;
  },
};

const baseGetInstalledPackageSummariesRequest: object = {};

export const GetInstalledPackageSummariesRequest = {
  encode(
    message: GetInstalledPackageSummariesRequest,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.context !== undefined) {
      Context.encode(message.context, writer.uint32(10).fork()).ldelim();
    }
    if (message.paginationOptions !== undefined) {
      PaginationOptions.encode(message.paginationOptions, writer.uint32(18).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetInstalledPackageSummariesRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseGetInstalledPackageSummariesRequest,
    } as GetInstalledPackageSummariesRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.context = Context.decode(reader, reader.uint32());
          break;
        case 2:
          message.paginationOptions = PaginationOptions.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): GetInstalledPackageSummariesRequest {
    const message = {
      ...baseGetInstalledPackageSummariesRequest,
    } as GetInstalledPackageSummariesRequest;
    if (object.context !== undefined && object.context !== null) {
      message.context = Context.fromJSON(object.context);
    } else {
      message.context = undefined;
    }
    if (object.paginationOptions !== undefined && object.paginationOptions !== null) {
      message.paginationOptions = PaginationOptions.fromJSON(object.paginationOptions);
    } else {
      message.paginationOptions = undefined;
    }
    return message;
  },

  toJSON(message: GetInstalledPackageSummariesRequest): unknown {
    const obj: any = {};
    message.context !== undefined &&
      (obj.context = message.context ? Context.toJSON(message.context) : undefined);
    message.paginationOptions !== undefined &&
      (obj.paginationOptions = message.paginationOptions
        ? PaginationOptions.toJSON(message.paginationOptions)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<GetInstalledPackageSummariesRequest>,
  ): GetInstalledPackageSummariesRequest {
    const message = {
      ...baseGetInstalledPackageSummariesRequest,
    } as GetInstalledPackageSummariesRequest;
    if (object.context !== undefined && object.context !== null) {
      message.context = Context.fromPartial(object.context);
    } else {
      message.context = undefined;
    }
    if (object.paginationOptions !== undefined && object.paginationOptions !== null) {
      message.paginationOptions = PaginationOptions.fromPartial(object.paginationOptions);
    } else {
      message.paginationOptions = undefined;
    }
    return message;
  },
};

const baseGetInstalledPackageDetailRequest: object = {};

export const GetInstalledPackageDetailRequest = {
  encode(
    message: GetInstalledPackageDetailRequest,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.installedPackageRef !== undefined) {
      InstalledPackageReference.encode(
        message.installedPackageRef,
        writer.uint32(10).fork(),
      ).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetInstalledPackageDetailRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseGetInstalledPackageDetailRequest,
    } as GetInstalledPackageDetailRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.installedPackageRef = InstalledPackageReference.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): GetInstalledPackageDetailRequest {
    const message = {
      ...baseGetInstalledPackageDetailRequest,
    } as GetInstalledPackageDetailRequest;
    if (object.installedPackageRef !== undefined && object.installedPackageRef !== null) {
      message.installedPackageRef = InstalledPackageReference.fromJSON(object.installedPackageRef);
    } else {
      message.installedPackageRef = undefined;
    }
    return message;
  },

  toJSON(message: GetInstalledPackageDetailRequest): unknown {
    const obj: any = {};
    message.installedPackageRef !== undefined &&
      (obj.installedPackageRef = message.installedPackageRef
        ? InstalledPackageReference.toJSON(message.installedPackageRef)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<GetInstalledPackageDetailRequest>,
  ): GetInstalledPackageDetailRequest {
    const message = {
      ...baseGetInstalledPackageDetailRequest,
    } as GetInstalledPackageDetailRequest;
    if (object.installedPackageRef !== undefined && object.installedPackageRef !== null) {
      message.installedPackageRef = InstalledPackageReference.fromPartial(
        object.installedPackageRef,
      );
    } else {
      message.installedPackageRef = undefined;
    }
    return message;
  },
};

const baseCreateInstalledPackageRequest: object = { name: "", values: "" };

export const CreateInstalledPackageRequest = {
  encode(
    message: CreateInstalledPackageRequest,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.availablePackageRef !== undefined) {
      AvailablePackageReference.encode(
        message.availablePackageRef,
        writer.uint32(10).fork(),
      ).ldelim();
    }
    if (message.targetContext !== undefined) {
      Context.encode(message.targetContext, writer.uint32(18).fork()).ldelim();
    }
    if (message.name !== "") {
      writer.uint32(26).string(message.name);
    }
    if (message.pkgVersionReference !== undefined) {
      VersionReference.encode(message.pkgVersionReference, writer.uint32(34).fork()).ldelim();
    }
    if (message.values !== "") {
      writer.uint32(42).string(message.values);
    }
    if (message.reconciliationOptions !== undefined) {
      ReconciliationOptions.encode(
        message.reconciliationOptions,
        writer.uint32(50).fork(),
      ).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): CreateInstalledPackageRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseCreateInstalledPackageRequest,
    } as CreateInstalledPackageRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.availablePackageRef = AvailablePackageReference.decode(reader, reader.uint32());
          break;
        case 2:
          message.targetContext = Context.decode(reader, reader.uint32());
          break;
        case 3:
          message.name = reader.string();
          break;
        case 4:
          message.pkgVersionReference = VersionReference.decode(reader, reader.uint32());
          break;
        case 5:
          message.values = reader.string();
          break;
        case 6:
          message.reconciliationOptions = ReconciliationOptions.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): CreateInstalledPackageRequest {
    const message = {
      ...baseCreateInstalledPackageRequest,
    } as CreateInstalledPackageRequest;
    if (object.availablePackageRef !== undefined && object.availablePackageRef !== null) {
      message.availablePackageRef = AvailablePackageReference.fromJSON(object.availablePackageRef);
    } else {
      message.availablePackageRef = undefined;
    }
    if (object.targetContext !== undefined && object.targetContext !== null) {
      message.targetContext = Context.fromJSON(object.targetContext);
    } else {
      message.targetContext = undefined;
    }
    if (object.name !== undefined && object.name !== null) {
      message.name = String(object.name);
    } else {
      message.name = "";
    }
    if (object.pkgVersionReference !== undefined && object.pkgVersionReference !== null) {
      message.pkgVersionReference = VersionReference.fromJSON(object.pkgVersionReference);
    } else {
      message.pkgVersionReference = undefined;
    }
    if (object.values !== undefined && object.values !== null) {
      message.values = String(object.values);
    } else {
      message.values = "";
    }
    if (object.reconciliationOptions !== undefined && object.reconciliationOptions !== null) {
      message.reconciliationOptions = ReconciliationOptions.fromJSON(object.reconciliationOptions);
    } else {
      message.reconciliationOptions = undefined;
    }
    return message;
  },

  toJSON(message: CreateInstalledPackageRequest): unknown {
    const obj: any = {};
    message.availablePackageRef !== undefined &&
      (obj.availablePackageRef = message.availablePackageRef
        ? AvailablePackageReference.toJSON(message.availablePackageRef)
        : undefined);
    message.targetContext !== undefined &&
      (obj.targetContext = message.targetContext
        ? Context.toJSON(message.targetContext)
        : undefined);
    message.name !== undefined && (obj.name = message.name);
    message.pkgVersionReference !== undefined &&
      (obj.pkgVersionReference = message.pkgVersionReference
        ? VersionReference.toJSON(message.pkgVersionReference)
        : undefined);
    message.values !== undefined && (obj.values = message.values);
    message.reconciliationOptions !== undefined &&
      (obj.reconciliationOptions = message.reconciliationOptions
        ? ReconciliationOptions.toJSON(message.reconciliationOptions)
        : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<CreateInstalledPackageRequest>): CreateInstalledPackageRequest {
    const message = {
      ...baseCreateInstalledPackageRequest,
    } as CreateInstalledPackageRequest;
    if (object.availablePackageRef !== undefined && object.availablePackageRef !== null) {
      message.availablePackageRef = AvailablePackageReference.fromPartial(
        object.availablePackageRef,
      );
    } else {
      message.availablePackageRef = undefined;
    }
    if (object.targetContext !== undefined && object.targetContext !== null) {
      message.targetContext = Context.fromPartial(object.targetContext);
    } else {
      message.targetContext = undefined;
    }
    if (object.name !== undefined && object.name !== null) {
      message.name = object.name;
    } else {
      message.name = "";
    }
    if (object.pkgVersionReference !== undefined && object.pkgVersionReference !== null) {
      message.pkgVersionReference = VersionReference.fromPartial(object.pkgVersionReference);
    } else {
      message.pkgVersionReference = undefined;
    }
    if (object.values !== undefined && object.values !== null) {
      message.values = object.values;
    } else {
      message.values = "";
    }
    if (object.reconciliationOptions !== undefined && object.reconciliationOptions !== null) {
      message.reconciliationOptions = ReconciliationOptions.fromPartial(
        object.reconciliationOptions,
      );
    } else {
      message.reconciliationOptions = undefined;
    }
    return message;
  },
};

const baseUpdateInstalledPackageRequest: object = { values: "" };

export const UpdateInstalledPackageRequest = {
  encode(
    message: UpdateInstalledPackageRequest,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.installedPackageRef !== undefined) {
      InstalledPackageReference.encode(
        message.installedPackageRef,
        writer.uint32(10).fork(),
      ).ldelim();
    }
    if (message.pkgVersionReference !== undefined) {
      VersionReference.encode(message.pkgVersionReference, writer.uint32(18).fork()).ldelim();
    }
    if (message.values !== "") {
      writer.uint32(26).string(message.values);
    }
    if (message.reconciliationOptions !== undefined) {
      ReconciliationOptions.encode(
        message.reconciliationOptions,
        writer.uint32(34).fork(),
      ).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): UpdateInstalledPackageRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseUpdateInstalledPackageRequest,
    } as UpdateInstalledPackageRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.installedPackageRef = InstalledPackageReference.decode(reader, reader.uint32());
          break;
        case 2:
          message.pkgVersionReference = VersionReference.decode(reader, reader.uint32());
          break;
        case 3:
          message.values = reader.string();
          break;
        case 4:
          message.reconciliationOptions = ReconciliationOptions.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): UpdateInstalledPackageRequest {
    const message = {
      ...baseUpdateInstalledPackageRequest,
    } as UpdateInstalledPackageRequest;
    if (object.installedPackageRef !== undefined && object.installedPackageRef !== null) {
      message.installedPackageRef = InstalledPackageReference.fromJSON(object.installedPackageRef);
    } else {
      message.installedPackageRef = undefined;
    }
    if (object.pkgVersionReference !== undefined && object.pkgVersionReference !== null) {
      message.pkgVersionReference = VersionReference.fromJSON(object.pkgVersionReference);
    } else {
      message.pkgVersionReference = undefined;
    }
    if (object.values !== undefined && object.values !== null) {
      message.values = String(object.values);
    } else {
      message.values = "";
    }
    if (object.reconciliationOptions !== undefined && object.reconciliationOptions !== null) {
      message.reconciliationOptions = ReconciliationOptions.fromJSON(object.reconciliationOptions);
    } else {
      message.reconciliationOptions = undefined;
    }
    return message;
  },

  toJSON(message: UpdateInstalledPackageRequest): unknown {
    const obj: any = {};
    message.installedPackageRef !== undefined &&
      (obj.installedPackageRef = message.installedPackageRef
        ? InstalledPackageReference.toJSON(message.installedPackageRef)
        : undefined);
    message.pkgVersionReference !== undefined &&
      (obj.pkgVersionReference = message.pkgVersionReference
        ? VersionReference.toJSON(message.pkgVersionReference)
        : undefined);
    message.values !== undefined && (obj.values = message.values);
    message.reconciliationOptions !== undefined &&
      (obj.reconciliationOptions = message.reconciliationOptions
        ? ReconciliationOptions.toJSON(message.reconciliationOptions)
        : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<UpdateInstalledPackageRequest>): UpdateInstalledPackageRequest {
    const message = {
      ...baseUpdateInstalledPackageRequest,
    } as UpdateInstalledPackageRequest;
    if (object.installedPackageRef !== undefined && object.installedPackageRef !== null) {
      message.installedPackageRef = InstalledPackageReference.fromPartial(
        object.installedPackageRef,
      );
    } else {
      message.installedPackageRef = undefined;
    }
    if (object.pkgVersionReference !== undefined && object.pkgVersionReference !== null) {
      message.pkgVersionReference = VersionReference.fromPartial(object.pkgVersionReference);
    } else {
      message.pkgVersionReference = undefined;
    }
    if (object.values !== undefined && object.values !== null) {
      message.values = object.values;
    } else {
      message.values = "";
    }
    if (object.reconciliationOptions !== undefined && object.reconciliationOptions !== null) {
      message.reconciliationOptions = ReconciliationOptions.fromPartial(
        object.reconciliationOptions,
      );
    } else {
      message.reconciliationOptions = undefined;
    }
    return message;
  },
};

const baseDeleteInstalledPackageRequest: object = {};

export const DeleteInstalledPackageRequest = {
  encode(
    message: DeleteInstalledPackageRequest,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.installedPackageRef !== undefined) {
      InstalledPackageReference.encode(
        message.installedPackageRef,
        writer.uint32(10).fork(),
      ).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): DeleteInstalledPackageRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseDeleteInstalledPackageRequest,
    } as DeleteInstalledPackageRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.installedPackageRef = InstalledPackageReference.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): DeleteInstalledPackageRequest {
    const message = {
      ...baseDeleteInstalledPackageRequest,
    } as DeleteInstalledPackageRequest;
    if (object.installedPackageRef !== undefined && object.installedPackageRef !== null) {
      message.installedPackageRef = InstalledPackageReference.fromJSON(object.installedPackageRef);
    } else {
      message.installedPackageRef = undefined;
    }
    return message;
  },

  toJSON(message: DeleteInstalledPackageRequest): unknown {
    const obj: any = {};
    message.installedPackageRef !== undefined &&
      (obj.installedPackageRef = message.installedPackageRef
        ? InstalledPackageReference.toJSON(message.installedPackageRef)
        : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<DeleteInstalledPackageRequest>): DeleteInstalledPackageRequest {
    const message = {
      ...baseDeleteInstalledPackageRequest,
    } as DeleteInstalledPackageRequest;
    if (object.installedPackageRef !== undefined && object.installedPackageRef !== null) {
      message.installedPackageRef = InstalledPackageReference.fromPartial(
        object.installedPackageRef,
      );
    } else {
      message.installedPackageRef = undefined;
    }
    return message;
  },
};

const baseGetInstalledPackageResourceRefsRequest: object = {};

export const GetInstalledPackageResourceRefsRequest = {
  encode(
    message: GetInstalledPackageResourceRefsRequest,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.installedPackageRef !== undefined) {
      InstalledPackageReference.encode(
        message.installedPackageRef,
        writer.uint32(10).fork(),
      ).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetInstalledPackageResourceRefsRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseGetInstalledPackageResourceRefsRequest,
    } as GetInstalledPackageResourceRefsRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.installedPackageRef = InstalledPackageReference.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): GetInstalledPackageResourceRefsRequest {
    const message = {
      ...baseGetInstalledPackageResourceRefsRequest,
    } as GetInstalledPackageResourceRefsRequest;
    if (object.installedPackageRef !== undefined && object.installedPackageRef !== null) {
      message.installedPackageRef = InstalledPackageReference.fromJSON(object.installedPackageRef);
    } else {
      message.installedPackageRef = undefined;
    }
    return message;
  },

  toJSON(message: GetInstalledPackageResourceRefsRequest): unknown {
    const obj: any = {};
    message.installedPackageRef !== undefined &&
      (obj.installedPackageRef = message.installedPackageRef
        ? InstalledPackageReference.toJSON(message.installedPackageRef)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<GetInstalledPackageResourceRefsRequest>,
  ): GetInstalledPackageResourceRefsRequest {
    const message = {
      ...baseGetInstalledPackageResourceRefsRequest,
    } as GetInstalledPackageResourceRefsRequest;
    if (object.installedPackageRef !== undefined && object.installedPackageRef !== null) {
      message.installedPackageRef = InstalledPackageReference.fromPartial(
        object.installedPackageRef,
      );
    } else {
      message.installedPackageRef = undefined;
    }
    return message;
  },
};

const baseGetAvailablePackageSummariesResponse: object = {
  nextPageToken: "",
  categories: "",
};

export const GetAvailablePackageSummariesResponse = {
  encode(
    message: GetAvailablePackageSummariesResponse,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    for (const v of message.availablePackageSummaries) {
      AvailablePackageSummary.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    if (message.nextPageToken !== "") {
      writer.uint32(18).string(message.nextPageToken);
    }
    for (const v of message.categories) {
      writer.uint32(26).string(v!);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetAvailablePackageSummariesResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseGetAvailablePackageSummariesResponse,
    } as GetAvailablePackageSummariesResponse;
    message.availablePackageSummaries = [];
    message.categories = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.availablePackageSummaries.push(
            AvailablePackageSummary.decode(reader, reader.uint32()),
          );
          break;
        case 2:
          message.nextPageToken = reader.string();
          break;
        case 3:
          message.categories.push(reader.string());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): GetAvailablePackageSummariesResponse {
    const message = {
      ...baseGetAvailablePackageSummariesResponse,
    } as GetAvailablePackageSummariesResponse;
    message.availablePackageSummaries = [];
    message.categories = [];
    if (
      object.availablePackageSummaries !== undefined &&
      object.availablePackageSummaries !== null
    ) {
      for (const e of object.availablePackageSummaries) {
        message.availablePackageSummaries.push(AvailablePackageSummary.fromJSON(e));
      }
    }
    if (object.nextPageToken !== undefined && object.nextPageToken !== null) {
      message.nextPageToken = String(object.nextPageToken);
    } else {
      message.nextPageToken = "";
    }
    if (object.categories !== undefined && object.categories !== null) {
      for (const e of object.categories) {
        message.categories.push(String(e));
      }
    }
    return message;
  },

  toJSON(message: GetAvailablePackageSummariesResponse): unknown {
    const obj: any = {};
    if (message.availablePackageSummaries) {
      obj.availablePackageSummaries = message.availablePackageSummaries.map(e =>
        e ? AvailablePackageSummary.toJSON(e) : undefined,
      );
    } else {
      obj.availablePackageSummaries = [];
    }
    message.nextPageToken !== undefined && (obj.nextPageToken = message.nextPageToken);
    if (message.categories) {
      obj.categories = message.categories.map(e => e);
    } else {
      obj.categories = [];
    }
    return obj;
  },

  fromPartial(
    object: DeepPartial<GetAvailablePackageSummariesResponse>,
  ): GetAvailablePackageSummariesResponse {
    const message = {
      ...baseGetAvailablePackageSummariesResponse,
    } as GetAvailablePackageSummariesResponse;
    message.availablePackageSummaries = [];
    message.categories = [];
    if (
      object.availablePackageSummaries !== undefined &&
      object.availablePackageSummaries !== null
    ) {
      for (const e of object.availablePackageSummaries) {
        message.availablePackageSummaries.push(AvailablePackageSummary.fromPartial(e));
      }
    }
    if (object.nextPageToken !== undefined && object.nextPageToken !== null) {
      message.nextPageToken = object.nextPageToken;
    } else {
      message.nextPageToken = "";
    }
    if (object.categories !== undefined && object.categories !== null) {
      for (const e of object.categories) {
        message.categories.push(e);
      }
    }
    return message;
  },
};

const baseGetAvailablePackageDetailResponse: object = {};

export const GetAvailablePackageDetailResponse = {
  encode(
    message: GetAvailablePackageDetailResponse,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.availablePackageDetail !== undefined) {
      AvailablePackageDetail.encode(
        message.availablePackageDetail,
        writer.uint32(10).fork(),
      ).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetAvailablePackageDetailResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseGetAvailablePackageDetailResponse,
    } as GetAvailablePackageDetailResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.availablePackageDetail = AvailablePackageDetail.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): GetAvailablePackageDetailResponse {
    const message = {
      ...baseGetAvailablePackageDetailResponse,
    } as GetAvailablePackageDetailResponse;
    if (object.availablePackageDetail !== undefined && object.availablePackageDetail !== null) {
      message.availablePackageDetail = AvailablePackageDetail.fromJSON(
        object.availablePackageDetail,
      );
    } else {
      message.availablePackageDetail = undefined;
    }
    return message;
  },

  toJSON(message: GetAvailablePackageDetailResponse): unknown {
    const obj: any = {};
    message.availablePackageDetail !== undefined &&
      (obj.availablePackageDetail = message.availablePackageDetail
        ? AvailablePackageDetail.toJSON(message.availablePackageDetail)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<GetAvailablePackageDetailResponse>,
  ): GetAvailablePackageDetailResponse {
    const message = {
      ...baseGetAvailablePackageDetailResponse,
    } as GetAvailablePackageDetailResponse;
    if (object.availablePackageDetail !== undefined && object.availablePackageDetail !== null) {
      message.availablePackageDetail = AvailablePackageDetail.fromPartial(
        object.availablePackageDetail,
      );
    } else {
      message.availablePackageDetail = undefined;
    }
    return message;
  },
};

const baseGetAvailablePackageVersionsResponse: object = {};

export const GetAvailablePackageVersionsResponse = {
  encode(
    message: GetAvailablePackageVersionsResponse,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    for (const v of message.packageAppVersions) {
      PackageAppVersion.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetAvailablePackageVersionsResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseGetAvailablePackageVersionsResponse,
    } as GetAvailablePackageVersionsResponse;
    message.packageAppVersions = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.packageAppVersions.push(PackageAppVersion.decode(reader, reader.uint32()));
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): GetAvailablePackageVersionsResponse {
    const message = {
      ...baseGetAvailablePackageVersionsResponse,
    } as GetAvailablePackageVersionsResponse;
    message.packageAppVersions = [];
    if (object.packageAppVersions !== undefined && object.packageAppVersions !== null) {
      for (const e of object.packageAppVersions) {
        message.packageAppVersions.push(PackageAppVersion.fromJSON(e));
      }
    }
    return message;
  },

  toJSON(message: GetAvailablePackageVersionsResponse): unknown {
    const obj: any = {};
    if (message.packageAppVersions) {
      obj.packageAppVersions = message.packageAppVersions.map(e =>
        e ? PackageAppVersion.toJSON(e) : undefined,
      );
    } else {
      obj.packageAppVersions = [];
    }
    return obj;
  },

  fromPartial(
    object: DeepPartial<GetAvailablePackageVersionsResponse>,
  ): GetAvailablePackageVersionsResponse {
    const message = {
      ...baseGetAvailablePackageVersionsResponse,
    } as GetAvailablePackageVersionsResponse;
    message.packageAppVersions = [];
    if (object.packageAppVersions !== undefined && object.packageAppVersions !== null) {
      for (const e of object.packageAppVersions) {
        message.packageAppVersions.push(PackageAppVersion.fromPartial(e));
      }
    }
    return message;
  },
};

const baseGetInstalledPackageSummariesResponse: object = { nextPageToken: "" };

export const GetInstalledPackageSummariesResponse = {
  encode(
    message: GetInstalledPackageSummariesResponse,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    for (const v of message.installedPackageSummaries) {
      InstalledPackageSummary.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    if (message.nextPageToken !== "") {
      writer.uint32(18).string(message.nextPageToken);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetInstalledPackageSummariesResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseGetInstalledPackageSummariesResponse,
    } as GetInstalledPackageSummariesResponse;
    message.installedPackageSummaries = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.installedPackageSummaries.push(
            InstalledPackageSummary.decode(reader, reader.uint32()),
          );
          break;
        case 2:
          message.nextPageToken = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): GetInstalledPackageSummariesResponse {
    const message = {
      ...baseGetInstalledPackageSummariesResponse,
    } as GetInstalledPackageSummariesResponse;
    message.installedPackageSummaries = [];
    if (
      object.installedPackageSummaries !== undefined &&
      object.installedPackageSummaries !== null
    ) {
      for (const e of object.installedPackageSummaries) {
        message.installedPackageSummaries.push(InstalledPackageSummary.fromJSON(e));
      }
    }
    if (object.nextPageToken !== undefined && object.nextPageToken !== null) {
      message.nextPageToken = String(object.nextPageToken);
    } else {
      message.nextPageToken = "";
    }
    return message;
  },

  toJSON(message: GetInstalledPackageSummariesResponse): unknown {
    const obj: any = {};
    if (message.installedPackageSummaries) {
      obj.installedPackageSummaries = message.installedPackageSummaries.map(e =>
        e ? InstalledPackageSummary.toJSON(e) : undefined,
      );
    } else {
      obj.installedPackageSummaries = [];
    }
    message.nextPageToken !== undefined && (obj.nextPageToken = message.nextPageToken);
    return obj;
  },

  fromPartial(
    object: DeepPartial<GetInstalledPackageSummariesResponse>,
  ): GetInstalledPackageSummariesResponse {
    const message = {
      ...baseGetInstalledPackageSummariesResponse,
    } as GetInstalledPackageSummariesResponse;
    message.installedPackageSummaries = [];
    if (
      object.installedPackageSummaries !== undefined &&
      object.installedPackageSummaries !== null
    ) {
      for (const e of object.installedPackageSummaries) {
        message.installedPackageSummaries.push(InstalledPackageSummary.fromPartial(e));
      }
    }
    if (object.nextPageToken !== undefined && object.nextPageToken !== null) {
      message.nextPageToken = object.nextPageToken;
    } else {
      message.nextPageToken = "";
    }
    return message;
  },
};

const baseGetInstalledPackageDetailResponse: object = {};

export const GetInstalledPackageDetailResponse = {
  encode(
    message: GetInstalledPackageDetailResponse,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.installedPackageDetail !== undefined) {
      InstalledPackageDetail.encode(
        message.installedPackageDetail,
        writer.uint32(10).fork(),
      ).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetInstalledPackageDetailResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseGetInstalledPackageDetailResponse,
    } as GetInstalledPackageDetailResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.installedPackageDetail = InstalledPackageDetail.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): GetInstalledPackageDetailResponse {
    const message = {
      ...baseGetInstalledPackageDetailResponse,
    } as GetInstalledPackageDetailResponse;
    if (object.installedPackageDetail !== undefined && object.installedPackageDetail !== null) {
      message.installedPackageDetail = InstalledPackageDetail.fromJSON(
        object.installedPackageDetail,
      );
    } else {
      message.installedPackageDetail = undefined;
    }
    return message;
  },

  toJSON(message: GetInstalledPackageDetailResponse): unknown {
    const obj: any = {};
    message.installedPackageDetail !== undefined &&
      (obj.installedPackageDetail = message.installedPackageDetail
        ? InstalledPackageDetail.toJSON(message.installedPackageDetail)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<GetInstalledPackageDetailResponse>,
  ): GetInstalledPackageDetailResponse {
    const message = {
      ...baseGetInstalledPackageDetailResponse,
    } as GetInstalledPackageDetailResponse;
    if (object.installedPackageDetail !== undefined && object.installedPackageDetail !== null) {
      message.installedPackageDetail = InstalledPackageDetail.fromPartial(
        object.installedPackageDetail,
      );
    } else {
      message.installedPackageDetail = undefined;
    }
    return message;
  },
};

const baseCreateInstalledPackageResponse: object = {};

export const CreateInstalledPackageResponse = {
  encode(
    message: CreateInstalledPackageResponse,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.installedPackageRef !== undefined) {
      InstalledPackageReference.encode(
        message.installedPackageRef,
        writer.uint32(10).fork(),
      ).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): CreateInstalledPackageResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseCreateInstalledPackageResponse,
    } as CreateInstalledPackageResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.installedPackageRef = InstalledPackageReference.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): CreateInstalledPackageResponse {
    const message = {
      ...baseCreateInstalledPackageResponse,
    } as CreateInstalledPackageResponse;
    if (object.installedPackageRef !== undefined && object.installedPackageRef !== null) {
      message.installedPackageRef = InstalledPackageReference.fromJSON(object.installedPackageRef);
    } else {
      message.installedPackageRef = undefined;
    }
    return message;
  },

  toJSON(message: CreateInstalledPackageResponse): unknown {
    const obj: any = {};
    message.installedPackageRef !== undefined &&
      (obj.installedPackageRef = message.installedPackageRef
        ? InstalledPackageReference.toJSON(message.installedPackageRef)
        : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<CreateInstalledPackageResponse>): CreateInstalledPackageResponse {
    const message = {
      ...baseCreateInstalledPackageResponse,
    } as CreateInstalledPackageResponse;
    if (object.installedPackageRef !== undefined && object.installedPackageRef !== null) {
      message.installedPackageRef = InstalledPackageReference.fromPartial(
        object.installedPackageRef,
      );
    } else {
      message.installedPackageRef = undefined;
    }
    return message;
  },
};

const baseUpdateInstalledPackageResponse: object = {};

export const UpdateInstalledPackageResponse = {
  encode(
    message: UpdateInstalledPackageResponse,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.installedPackageRef !== undefined) {
      InstalledPackageReference.encode(
        message.installedPackageRef,
        writer.uint32(10).fork(),
      ).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): UpdateInstalledPackageResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseUpdateInstalledPackageResponse,
    } as UpdateInstalledPackageResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.installedPackageRef = InstalledPackageReference.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): UpdateInstalledPackageResponse {
    const message = {
      ...baseUpdateInstalledPackageResponse,
    } as UpdateInstalledPackageResponse;
    if (object.installedPackageRef !== undefined && object.installedPackageRef !== null) {
      message.installedPackageRef = InstalledPackageReference.fromJSON(object.installedPackageRef);
    } else {
      message.installedPackageRef = undefined;
    }
    return message;
  },

  toJSON(message: UpdateInstalledPackageResponse): unknown {
    const obj: any = {};
    message.installedPackageRef !== undefined &&
      (obj.installedPackageRef = message.installedPackageRef
        ? InstalledPackageReference.toJSON(message.installedPackageRef)
        : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<UpdateInstalledPackageResponse>): UpdateInstalledPackageResponse {
    const message = {
      ...baseUpdateInstalledPackageResponse,
    } as UpdateInstalledPackageResponse;
    if (object.installedPackageRef !== undefined && object.installedPackageRef !== null) {
      message.installedPackageRef = InstalledPackageReference.fromPartial(
        object.installedPackageRef,
      );
    } else {
      message.installedPackageRef = undefined;
    }
    return message;
  },
};

const baseDeleteInstalledPackageResponse: object = {};

export const DeleteInstalledPackageResponse = {
  encode(_: DeleteInstalledPackageResponse, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): DeleteInstalledPackageResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseDeleteInstalledPackageResponse,
    } as DeleteInstalledPackageResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(_: any): DeleteInstalledPackageResponse {
    const message = {
      ...baseDeleteInstalledPackageResponse,
    } as DeleteInstalledPackageResponse;
    return message;
  },

  toJSON(_: DeleteInstalledPackageResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(_: DeepPartial<DeleteInstalledPackageResponse>): DeleteInstalledPackageResponse {
    const message = {
      ...baseDeleteInstalledPackageResponse,
    } as DeleteInstalledPackageResponse;
    return message;
  },
};

const baseGetInstalledPackageResourceRefsResponse: object = {};

export const GetInstalledPackageResourceRefsResponse = {
  encode(
    message: GetInstalledPackageResourceRefsResponse,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.context !== undefined) {
      Context.encode(message.context, writer.uint32(10).fork()).ldelim();
    }
    for (const v of message.resourceRefs) {
      ResourceRef.encode(v!, writer.uint32(18).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetInstalledPackageResourceRefsResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseGetInstalledPackageResourceRefsResponse,
    } as GetInstalledPackageResourceRefsResponse;
    message.resourceRefs = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.context = Context.decode(reader, reader.uint32());
          break;
        case 2:
          message.resourceRefs.push(ResourceRef.decode(reader, reader.uint32()));
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): GetInstalledPackageResourceRefsResponse {
    const message = {
      ...baseGetInstalledPackageResourceRefsResponse,
    } as GetInstalledPackageResourceRefsResponse;
    message.resourceRefs = [];
    if (object.context !== undefined && object.context !== null) {
      message.context = Context.fromJSON(object.context);
    } else {
      message.context = undefined;
    }
    if (object.resourceRefs !== undefined && object.resourceRefs !== null) {
      for (const e of object.resourceRefs) {
        message.resourceRefs.push(ResourceRef.fromJSON(e));
      }
    }
    return message;
  },

  toJSON(message: GetInstalledPackageResourceRefsResponse): unknown {
    const obj: any = {};
    message.context !== undefined &&
      (obj.context = message.context ? Context.toJSON(message.context) : undefined);
    if (message.resourceRefs) {
      obj.resourceRefs = message.resourceRefs.map(e => (e ? ResourceRef.toJSON(e) : undefined));
    } else {
      obj.resourceRefs = [];
    }
    return obj;
  },

  fromPartial(
    object: DeepPartial<GetInstalledPackageResourceRefsResponse>,
  ): GetInstalledPackageResourceRefsResponse {
    const message = {
      ...baseGetInstalledPackageResourceRefsResponse,
    } as GetInstalledPackageResourceRefsResponse;
    message.resourceRefs = [];
    if (object.context !== undefined && object.context !== null) {
      message.context = Context.fromPartial(object.context);
    } else {
      message.context = undefined;
    }
    if (object.resourceRefs !== undefined && object.resourceRefs !== null) {
      for (const e of object.resourceRefs) {
        message.resourceRefs.push(ResourceRef.fromPartial(e));
      }
    }
    return message;
  },
};

const baseAvailablePackageSummary: object = {
  name: "",
  iconUrl: "",
  displayName: "",
  shortDescription: "",
  categories: "",
};

export const AvailablePackageSummary = {
  encode(message: AvailablePackageSummary, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.availablePackageRef !== undefined) {
      AvailablePackageReference.encode(
        message.availablePackageRef,
        writer.uint32(10).fork(),
      ).ldelim();
    }
    if (message.name !== "") {
      writer.uint32(18).string(message.name);
    }
    if (message.latestVersion !== undefined) {
      PackageAppVersion.encode(message.latestVersion, writer.uint32(26).fork()).ldelim();
    }
    if (message.iconUrl !== "") {
      writer.uint32(34).string(message.iconUrl);
    }
    if (message.displayName !== "") {
      writer.uint32(42).string(message.displayName);
    }
    if (message.shortDescription !== "") {
      writer.uint32(50).string(message.shortDescription);
    }
    for (const v of message.categories) {
      writer.uint32(58).string(v!);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): AvailablePackageSummary {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseAvailablePackageSummary,
    } as AvailablePackageSummary;
    message.categories = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.availablePackageRef = AvailablePackageReference.decode(reader, reader.uint32());
          break;
        case 2:
          message.name = reader.string();
          break;
        case 3:
          message.latestVersion = PackageAppVersion.decode(reader, reader.uint32());
          break;
        case 4:
          message.iconUrl = reader.string();
          break;
        case 5:
          message.displayName = reader.string();
          break;
        case 6:
          message.shortDescription = reader.string();
          break;
        case 7:
          message.categories.push(reader.string());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): AvailablePackageSummary {
    const message = {
      ...baseAvailablePackageSummary,
    } as AvailablePackageSummary;
    message.categories = [];
    if (object.availablePackageRef !== undefined && object.availablePackageRef !== null) {
      message.availablePackageRef = AvailablePackageReference.fromJSON(object.availablePackageRef);
    } else {
      message.availablePackageRef = undefined;
    }
    if (object.name !== undefined && object.name !== null) {
      message.name = String(object.name);
    } else {
      message.name = "";
    }
    if (object.latestVersion !== undefined && object.latestVersion !== null) {
      message.latestVersion = PackageAppVersion.fromJSON(object.latestVersion);
    } else {
      message.latestVersion = undefined;
    }
    if (object.iconUrl !== undefined && object.iconUrl !== null) {
      message.iconUrl = String(object.iconUrl);
    } else {
      message.iconUrl = "";
    }
    if (object.displayName !== undefined && object.displayName !== null) {
      message.displayName = String(object.displayName);
    } else {
      message.displayName = "";
    }
    if (object.shortDescription !== undefined && object.shortDescription !== null) {
      message.shortDescription = String(object.shortDescription);
    } else {
      message.shortDescription = "";
    }
    if (object.categories !== undefined && object.categories !== null) {
      for (const e of object.categories) {
        message.categories.push(String(e));
      }
    }
    return message;
  },

  toJSON(message: AvailablePackageSummary): unknown {
    const obj: any = {};
    message.availablePackageRef !== undefined &&
      (obj.availablePackageRef = message.availablePackageRef
        ? AvailablePackageReference.toJSON(message.availablePackageRef)
        : undefined);
    message.name !== undefined && (obj.name = message.name);
    message.latestVersion !== undefined &&
      (obj.latestVersion = message.latestVersion
        ? PackageAppVersion.toJSON(message.latestVersion)
        : undefined);
    message.iconUrl !== undefined && (obj.iconUrl = message.iconUrl);
    message.displayName !== undefined && (obj.displayName = message.displayName);
    message.shortDescription !== undefined && (obj.shortDescription = message.shortDescription);
    if (message.categories) {
      obj.categories = message.categories.map(e => e);
    } else {
      obj.categories = [];
    }
    return obj;
  },

  fromPartial(object: DeepPartial<AvailablePackageSummary>): AvailablePackageSummary {
    const message = {
      ...baseAvailablePackageSummary,
    } as AvailablePackageSummary;
    message.categories = [];
    if (object.availablePackageRef !== undefined && object.availablePackageRef !== null) {
      message.availablePackageRef = AvailablePackageReference.fromPartial(
        object.availablePackageRef,
      );
    } else {
      message.availablePackageRef = undefined;
    }
    if (object.name !== undefined && object.name !== null) {
      message.name = object.name;
    } else {
      message.name = "";
    }
    if (object.latestVersion !== undefined && object.latestVersion !== null) {
      message.latestVersion = PackageAppVersion.fromPartial(object.latestVersion);
    } else {
      message.latestVersion = undefined;
    }
    if (object.iconUrl !== undefined && object.iconUrl !== null) {
      message.iconUrl = object.iconUrl;
    } else {
      message.iconUrl = "";
    }
    if (object.displayName !== undefined && object.displayName !== null) {
      message.displayName = object.displayName;
    } else {
      message.displayName = "";
    }
    if (object.shortDescription !== undefined && object.shortDescription !== null) {
      message.shortDescription = object.shortDescription;
    } else {
      message.shortDescription = "";
    }
    if (object.categories !== undefined && object.categories !== null) {
      for (const e of object.categories) {
        message.categories.push(e);
      }
    }
    return message;
  },
};

const baseAvailablePackageDetail: object = {
  name: "",
  repoUrl: "",
  homeUrl: "",
  iconUrl: "",
  displayName: "",
  shortDescription: "",
  longDescription: "",
  readme: "",
  defaultValues: "",
  valuesSchema: "",
  sourceUrls: "",
  categories: "",
};

export const AvailablePackageDetail = {
  encode(message: AvailablePackageDetail, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.availablePackageRef !== undefined) {
      AvailablePackageReference.encode(
        message.availablePackageRef,
        writer.uint32(10).fork(),
      ).ldelim();
    }
    if (message.name !== "") {
      writer.uint32(18).string(message.name);
    }
    if (message.version !== undefined) {
      PackageAppVersion.encode(message.version, writer.uint32(26).fork()).ldelim();
    }
    if (message.repoUrl !== "") {
      writer.uint32(34).string(message.repoUrl);
    }
    if (message.homeUrl !== "") {
      writer.uint32(42).string(message.homeUrl);
    }
    if (message.iconUrl !== "") {
      writer.uint32(50).string(message.iconUrl);
    }
    if (message.displayName !== "") {
      writer.uint32(58).string(message.displayName);
    }
    if (message.shortDescription !== "") {
      writer.uint32(66).string(message.shortDescription);
    }
    if (message.longDescription !== "") {
      writer.uint32(74).string(message.longDescription);
    }
    if (message.readme !== "") {
      writer.uint32(82).string(message.readme);
    }
    if (message.defaultValues !== "") {
      writer.uint32(90).string(message.defaultValues);
    }
    if (message.valuesSchema !== "") {
      writer.uint32(98).string(message.valuesSchema);
    }
    for (const v of message.sourceUrls) {
      writer.uint32(106).string(v!);
    }
    for (const v of message.maintainers) {
      Maintainer.encode(v!, writer.uint32(114).fork()).ldelim();
    }
    for (const v of message.categories) {
      writer.uint32(122).string(v!);
    }
    if (message.customDetail !== undefined) {
      Any.encode(message.customDetail, writer.uint32(130).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): AvailablePackageDetail {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseAvailablePackageDetail } as AvailablePackageDetail;
    message.sourceUrls = [];
    message.maintainers = [];
    message.categories = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.availablePackageRef = AvailablePackageReference.decode(reader, reader.uint32());
          break;
        case 2:
          message.name = reader.string();
          break;
        case 3:
          message.version = PackageAppVersion.decode(reader, reader.uint32());
          break;
        case 4:
          message.repoUrl = reader.string();
          break;
        case 5:
          message.homeUrl = reader.string();
          break;
        case 6:
          message.iconUrl = reader.string();
          break;
        case 7:
          message.displayName = reader.string();
          break;
        case 8:
          message.shortDescription = reader.string();
          break;
        case 9:
          message.longDescription = reader.string();
          break;
        case 10:
          message.readme = reader.string();
          break;
        case 11:
          message.defaultValues = reader.string();
          break;
        case 12:
          message.valuesSchema = reader.string();
          break;
        case 13:
          message.sourceUrls.push(reader.string());
          break;
        case 14:
          message.maintainers.push(Maintainer.decode(reader, reader.uint32()));
          break;
        case 15:
          message.categories.push(reader.string());
          break;
        case 16:
          message.customDetail = Any.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): AvailablePackageDetail {
    const message = { ...baseAvailablePackageDetail } as AvailablePackageDetail;
    message.sourceUrls = [];
    message.maintainers = [];
    message.categories = [];
    if (object.availablePackageRef !== undefined && object.availablePackageRef !== null) {
      message.availablePackageRef = AvailablePackageReference.fromJSON(object.availablePackageRef);
    } else {
      message.availablePackageRef = undefined;
    }
    if (object.name !== undefined && object.name !== null) {
      message.name = String(object.name);
    } else {
      message.name = "";
    }
    if (object.version !== undefined && object.version !== null) {
      message.version = PackageAppVersion.fromJSON(object.version);
    } else {
      message.version = undefined;
    }
    if (object.repoUrl !== undefined && object.repoUrl !== null) {
      message.repoUrl = String(object.repoUrl);
    } else {
      message.repoUrl = "";
    }
    if (object.homeUrl !== undefined && object.homeUrl !== null) {
      message.homeUrl = String(object.homeUrl);
    } else {
      message.homeUrl = "";
    }
    if (object.iconUrl !== undefined && object.iconUrl !== null) {
      message.iconUrl = String(object.iconUrl);
    } else {
      message.iconUrl = "";
    }
    if (object.displayName !== undefined && object.displayName !== null) {
      message.displayName = String(object.displayName);
    } else {
      message.displayName = "";
    }
    if (object.shortDescription !== undefined && object.shortDescription !== null) {
      message.shortDescription = String(object.shortDescription);
    } else {
      message.shortDescription = "";
    }
    if (object.longDescription !== undefined && object.longDescription !== null) {
      message.longDescription = String(object.longDescription);
    } else {
      message.longDescription = "";
    }
    if (object.readme !== undefined && object.readme !== null) {
      message.readme = String(object.readme);
    } else {
      message.readme = "";
    }
    if (object.defaultValues !== undefined && object.defaultValues !== null) {
      message.defaultValues = String(object.defaultValues);
    } else {
      message.defaultValues = "";
    }
    if (object.valuesSchema !== undefined && object.valuesSchema !== null) {
      message.valuesSchema = String(object.valuesSchema);
    } else {
      message.valuesSchema = "";
    }
    if (object.sourceUrls !== undefined && object.sourceUrls !== null) {
      for (const e of object.sourceUrls) {
        message.sourceUrls.push(String(e));
      }
    }
    if (object.maintainers !== undefined && object.maintainers !== null) {
      for (const e of object.maintainers) {
        message.maintainers.push(Maintainer.fromJSON(e));
      }
    }
    if (object.categories !== undefined && object.categories !== null) {
      for (const e of object.categories) {
        message.categories.push(String(e));
      }
    }
    if (object.customDetail !== undefined && object.customDetail !== null) {
      message.customDetail = Any.fromJSON(object.customDetail);
    } else {
      message.customDetail = undefined;
    }
    return message;
  },

  toJSON(message: AvailablePackageDetail): unknown {
    const obj: any = {};
    message.availablePackageRef !== undefined &&
      (obj.availablePackageRef = message.availablePackageRef
        ? AvailablePackageReference.toJSON(message.availablePackageRef)
        : undefined);
    message.name !== undefined && (obj.name = message.name);
    message.version !== undefined &&
      (obj.version = message.version ? PackageAppVersion.toJSON(message.version) : undefined);
    message.repoUrl !== undefined && (obj.repoUrl = message.repoUrl);
    message.homeUrl !== undefined && (obj.homeUrl = message.homeUrl);
    message.iconUrl !== undefined && (obj.iconUrl = message.iconUrl);
    message.displayName !== undefined && (obj.displayName = message.displayName);
    message.shortDescription !== undefined && (obj.shortDescription = message.shortDescription);
    message.longDescription !== undefined && (obj.longDescription = message.longDescription);
    message.readme !== undefined && (obj.readme = message.readme);
    message.defaultValues !== undefined && (obj.defaultValues = message.defaultValues);
    message.valuesSchema !== undefined && (obj.valuesSchema = message.valuesSchema);
    if (message.sourceUrls) {
      obj.sourceUrls = message.sourceUrls.map(e => e);
    } else {
      obj.sourceUrls = [];
    }
    if (message.maintainers) {
      obj.maintainers = message.maintainers.map(e => (e ? Maintainer.toJSON(e) : undefined));
    } else {
      obj.maintainers = [];
    }
    if (message.categories) {
      obj.categories = message.categories.map(e => e);
    } else {
      obj.categories = [];
    }
    message.customDetail !== undefined &&
      (obj.customDetail = message.customDetail ? Any.toJSON(message.customDetail) : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<AvailablePackageDetail>): AvailablePackageDetail {
    const message = { ...baseAvailablePackageDetail } as AvailablePackageDetail;
    message.sourceUrls = [];
    message.maintainers = [];
    message.categories = [];
    if (object.availablePackageRef !== undefined && object.availablePackageRef !== null) {
      message.availablePackageRef = AvailablePackageReference.fromPartial(
        object.availablePackageRef,
      );
    } else {
      message.availablePackageRef = undefined;
    }
    if (object.name !== undefined && object.name !== null) {
      message.name = object.name;
    } else {
      message.name = "";
    }
    if (object.version !== undefined && object.version !== null) {
      message.version = PackageAppVersion.fromPartial(object.version);
    } else {
      message.version = undefined;
    }
    if (object.repoUrl !== undefined && object.repoUrl !== null) {
      message.repoUrl = object.repoUrl;
    } else {
      message.repoUrl = "";
    }
    if (object.homeUrl !== undefined && object.homeUrl !== null) {
      message.homeUrl = object.homeUrl;
    } else {
      message.homeUrl = "";
    }
    if (object.iconUrl !== undefined && object.iconUrl !== null) {
      message.iconUrl = object.iconUrl;
    } else {
      message.iconUrl = "";
    }
    if (object.displayName !== undefined && object.displayName !== null) {
      message.displayName = object.displayName;
    } else {
      message.displayName = "";
    }
    if (object.shortDescription !== undefined && object.shortDescription !== null) {
      message.shortDescription = object.shortDescription;
    } else {
      message.shortDescription = "";
    }
    if (object.longDescription !== undefined && object.longDescription !== null) {
      message.longDescription = object.longDescription;
    } else {
      message.longDescription = "";
    }
    if (object.readme !== undefined && object.readme !== null) {
      message.readme = object.readme;
    } else {
      message.readme = "";
    }
    if (object.defaultValues !== undefined && object.defaultValues !== null) {
      message.defaultValues = object.defaultValues;
    } else {
      message.defaultValues = "";
    }
    if (object.valuesSchema !== undefined && object.valuesSchema !== null) {
      message.valuesSchema = object.valuesSchema;
    } else {
      message.valuesSchema = "";
    }
    if (object.sourceUrls !== undefined && object.sourceUrls !== null) {
      for (const e of object.sourceUrls) {
        message.sourceUrls.push(e);
      }
    }
    if (object.maintainers !== undefined && object.maintainers !== null) {
      for (const e of object.maintainers) {
        message.maintainers.push(Maintainer.fromPartial(e));
      }
    }
    if (object.categories !== undefined && object.categories !== null) {
      for (const e of object.categories) {
        message.categories.push(e);
      }
    }
    if (object.customDetail !== undefined && object.customDetail !== null) {
      message.customDetail = Any.fromPartial(object.customDetail);
    } else {
      message.customDetail = undefined;
    }
    return message;
  },
};

const baseInstalledPackageSummary: object = {
  name: "",
  iconUrl: "",
  pkgDisplayName: "",
  shortDescription: "",
};

export const InstalledPackageSummary = {
  encode(message: InstalledPackageSummary, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.installedPackageRef !== undefined) {
      InstalledPackageReference.encode(
        message.installedPackageRef,
        writer.uint32(10).fork(),
      ).ldelim();
    }
    if (message.name !== "") {
      writer.uint32(18).string(message.name);
    }
    if (message.pkgVersionReference !== undefined) {
      VersionReference.encode(message.pkgVersionReference, writer.uint32(26).fork()).ldelim();
    }
    if (message.currentVersion !== undefined) {
      PackageAppVersion.encode(message.currentVersion, writer.uint32(34).fork()).ldelim();
    }
    if (message.iconUrl !== "") {
      writer.uint32(42).string(message.iconUrl);
    }
    if (message.pkgDisplayName !== "") {
      writer.uint32(50).string(message.pkgDisplayName);
    }
    if (message.shortDescription !== "") {
      writer.uint32(58).string(message.shortDescription);
    }
    if (message.latestMatchingVersion !== undefined) {
      PackageAppVersion.encode(message.latestMatchingVersion, writer.uint32(66).fork()).ldelim();
    }
    if (message.latestVersion !== undefined) {
      PackageAppVersion.encode(message.latestVersion, writer.uint32(74).fork()).ldelim();
    }
    if (message.status !== undefined) {
      InstalledPackageStatus.encode(message.status, writer.uint32(82).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): InstalledPackageSummary {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseInstalledPackageSummary,
    } as InstalledPackageSummary;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.installedPackageRef = InstalledPackageReference.decode(reader, reader.uint32());
          break;
        case 2:
          message.name = reader.string();
          break;
        case 3:
          message.pkgVersionReference = VersionReference.decode(reader, reader.uint32());
          break;
        case 4:
          message.currentVersion = PackageAppVersion.decode(reader, reader.uint32());
          break;
        case 5:
          message.iconUrl = reader.string();
          break;
        case 6:
          message.pkgDisplayName = reader.string();
          break;
        case 7:
          message.shortDescription = reader.string();
          break;
        case 8:
          message.latestMatchingVersion = PackageAppVersion.decode(reader, reader.uint32());
          break;
        case 9:
          message.latestVersion = PackageAppVersion.decode(reader, reader.uint32());
          break;
        case 10:
          message.status = InstalledPackageStatus.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): InstalledPackageSummary {
    const message = {
      ...baseInstalledPackageSummary,
    } as InstalledPackageSummary;
    if (object.installedPackageRef !== undefined && object.installedPackageRef !== null) {
      message.installedPackageRef = InstalledPackageReference.fromJSON(object.installedPackageRef);
    } else {
      message.installedPackageRef = undefined;
    }
    if (object.name !== undefined && object.name !== null) {
      message.name = String(object.name);
    } else {
      message.name = "";
    }
    if (object.pkgVersionReference !== undefined && object.pkgVersionReference !== null) {
      message.pkgVersionReference = VersionReference.fromJSON(object.pkgVersionReference);
    } else {
      message.pkgVersionReference = undefined;
    }
    if (object.currentVersion !== undefined && object.currentVersion !== null) {
      message.currentVersion = PackageAppVersion.fromJSON(object.currentVersion);
    } else {
      message.currentVersion = undefined;
    }
    if (object.iconUrl !== undefined && object.iconUrl !== null) {
      message.iconUrl = String(object.iconUrl);
    } else {
      message.iconUrl = "";
    }
    if (object.pkgDisplayName !== undefined && object.pkgDisplayName !== null) {
      message.pkgDisplayName = String(object.pkgDisplayName);
    } else {
      message.pkgDisplayName = "";
    }
    if (object.shortDescription !== undefined && object.shortDescription !== null) {
      message.shortDescription = String(object.shortDescription);
    } else {
      message.shortDescription = "";
    }
    if (object.latestMatchingVersion !== undefined && object.latestMatchingVersion !== null) {
      message.latestMatchingVersion = PackageAppVersion.fromJSON(object.latestMatchingVersion);
    } else {
      message.latestMatchingVersion = undefined;
    }
    if (object.latestVersion !== undefined && object.latestVersion !== null) {
      message.latestVersion = PackageAppVersion.fromJSON(object.latestVersion);
    } else {
      message.latestVersion = undefined;
    }
    if (object.status !== undefined && object.status !== null) {
      message.status = InstalledPackageStatus.fromJSON(object.status);
    } else {
      message.status = undefined;
    }
    return message;
  },

  toJSON(message: InstalledPackageSummary): unknown {
    const obj: any = {};
    message.installedPackageRef !== undefined &&
      (obj.installedPackageRef = message.installedPackageRef
        ? InstalledPackageReference.toJSON(message.installedPackageRef)
        : undefined);
    message.name !== undefined && (obj.name = message.name);
    message.pkgVersionReference !== undefined &&
      (obj.pkgVersionReference = message.pkgVersionReference
        ? VersionReference.toJSON(message.pkgVersionReference)
        : undefined);
    message.currentVersion !== undefined &&
      (obj.currentVersion = message.currentVersion
        ? PackageAppVersion.toJSON(message.currentVersion)
        : undefined);
    message.iconUrl !== undefined && (obj.iconUrl = message.iconUrl);
    message.pkgDisplayName !== undefined && (obj.pkgDisplayName = message.pkgDisplayName);
    message.shortDescription !== undefined && (obj.shortDescription = message.shortDescription);
    message.latestMatchingVersion !== undefined &&
      (obj.latestMatchingVersion = message.latestMatchingVersion
        ? PackageAppVersion.toJSON(message.latestMatchingVersion)
        : undefined);
    message.latestVersion !== undefined &&
      (obj.latestVersion = message.latestVersion
        ? PackageAppVersion.toJSON(message.latestVersion)
        : undefined);
    message.status !== undefined &&
      (obj.status = message.status ? InstalledPackageStatus.toJSON(message.status) : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<InstalledPackageSummary>): InstalledPackageSummary {
    const message = {
      ...baseInstalledPackageSummary,
    } as InstalledPackageSummary;
    if (object.installedPackageRef !== undefined && object.installedPackageRef !== null) {
      message.installedPackageRef = InstalledPackageReference.fromPartial(
        object.installedPackageRef,
      );
    } else {
      message.installedPackageRef = undefined;
    }
    if (object.name !== undefined && object.name !== null) {
      message.name = object.name;
    } else {
      message.name = "";
    }
    if (object.pkgVersionReference !== undefined && object.pkgVersionReference !== null) {
      message.pkgVersionReference = VersionReference.fromPartial(object.pkgVersionReference);
    } else {
      message.pkgVersionReference = undefined;
    }
    if (object.currentVersion !== undefined && object.currentVersion !== null) {
      message.currentVersion = PackageAppVersion.fromPartial(object.currentVersion);
    } else {
      message.currentVersion = undefined;
    }
    if (object.iconUrl !== undefined && object.iconUrl !== null) {
      message.iconUrl = object.iconUrl;
    } else {
      message.iconUrl = "";
    }
    if (object.pkgDisplayName !== undefined && object.pkgDisplayName !== null) {
      message.pkgDisplayName = object.pkgDisplayName;
    } else {
      message.pkgDisplayName = "";
    }
    if (object.shortDescription !== undefined && object.shortDescription !== null) {
      message.shortDescription = object.shortDescription;
    } else {
      message.shortDescription = "";
    }
    if (object.latestMatchingVersion !== undefined && object.latestMatchingVersion !== null) {
      message.latestMatchingVersion = PackageAppVersion.fromPartial(object.latestMatchingVersion);
    } else {
      message.latestMatchingVersion = undefined;
    }
    if (object.latestVersion !== undefined && object.latestVersion !== null) {
      message.latestVersion = PackageAppVersion.fromPartial(object.latestVersion);
    } else {
      message.latestVersion = undefined;
    }
    if (object.status !== undefined && object.status !== null) {
      message.status = InstalledPackageStatus.fromPartial(object.status);
    } else {
      message.status = undefined;
    }
    return message;
  },
};

const baseInstalledPackageDetail: object = {
  name: "",
  valuesApplied: "",
  postInstallationNotes: "",
};

export const InstalledPackageDetail = {
  encode(message: InstalledPackageDetail, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.installedPackageRef !== undefined) {
      InstalledPackageReference.encode(
        message.installedPackageRef,
        writer.uint32(10).fork(),
      ).ldelim();
    }
    if (message.pkgVersionReference !== undefined) {
      VersionReference.encode(message.pkgVersionReference, writer.uint32(18).fork()).ldelim();
    }
    if (message.name !== "") {
      writer.uint32(26).string(message.name);
    }
    if (message.currentVersion !== undefined) {
      PackageAppVersion.encode(message.currentVersion, writer.uint32(34).fork()).ldelim();
    }
    if (message.valuesApplied !== "") {
      writer.uint32(42).string(message.valuesApplied);
    }
    if (message.reconciliationOptions !== undefined) {
      ReconciliationOptions.encode(
        message.reconciliationOptions,
        writer.uint32(50).fork(),
      ).ldelim();
    }
    if (message.status !== undefined) {
      InstalledPackageStatus.encode(message.status, writer.uint32(58).fork()).ldelim();
    }
    if (message.postInstallationNotes !== "") {
      writer.uint32(66).string(message.postInstallationNotes);
    }
    if (message.availablePackageRef !== undefined) {
      AvailablePackageReference.encode(
        message.availablePackageRef,
        writer.uint32(74).fork(),
      ).ldelim();
    }
    if (message.latestMatchingVersion !== undefined) {
      PackageAppVersion.encode(message.latestMatchingVersion, writer.uint32(82).fork()).ldelim();
    }
    if (message.latestVersion !== undefined) {
      PackageAppVersion.encode(message.latestVersion, writer.uint32(90).fork()).ldelim();
    }
    if (message.customDetail !== undefined) {
      Any.encode(message.customDetail, writer.uint32(114).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): InstalledPackageDetail {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseInstalledPackageDetail } as InstalledPackageDetail;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.installedPackageRef = InstalledPackageReference.decode(reader, reader.uint32());
          break;
        case 2:
          message.pkgVersionReference = VersionReference.decode(reader, reader.uint32());
          break;
        case 3:
          message.name = reader.string();
          break;
        case 4:
          message.currentVersion = PackageAppVersion.decode(reader, reader.uint32());
          break;
        case 5:
          message.valuesApplied = reader.string();
          break;
        case 6:
          message.reconciliationOptions = ReconciliationOptions.decode(reader, reader.uint32());
          break;
        case 7:
          message.status = InstalledPackageStatus.decode(reader, reader.uint32());
          break;
        case 8:
          message.postInstallationNotes = reader.string();
          break;
        case 9:
          message.availablePackageRef = AvailablePackageReference.decode(reader, reader.uint32());
          break;
        case 10:
          message.latestMatchingVersion = PackageAppVersion.decode(reader, reader.uint32());
          break;
        case 11:
          message.latestVersion = PackageAppVersion.decode(reader, reader.uint32());
          break;
        case 14:
          message.customDetail = Any.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): InstalledPackageDetail {
    const message = { ...baseInstalledPackageDetail } as InstalledPackageDetail;
    if (object.installedPackageRef !== undefined && object.installedPackageRef !== null) {
      message.installedPackageRef = InstalledPackageReference.fromJSON(object.installedPackageRef);
    } else {
      message.installedPackageRef = undefined;
    }
    if (object.pkgVersionReference !== undefined && object.pkgVersionReference !== null) {
      message.pkgVersionReference = VersionReference.fromJSON(object.pkgVersionReference);
    } else {
      message.pkgVersionReference = undefined;
    }
    if (object.name !== undefined && object.name !== null) {
      message.name = String(object.name);
    } else {
      message.name = "";
    }
    if (object.currentVersion !== undefined && object.currentVersion !== null) {
      message.currentVersion = PackageAppVersion.fromJSON(object.currentVersion);
    } else {
      message.currentVersion = undefined;
    }
    if (object.valuesApplied !== undefined && object.valuesApplied !== null) {
      message.valuesApplied = String(object.valuesApplied);
    } else {
      message.valuesApplied = "";
    }
    if (object.reconciliationOptions !== undefined && object.reconciliationOptions !== null) {
      message.reconciliationOptions = ReconciliationOptions.fromJSON(object.reconciliationOptions);
    } else {
      message.reconciliationOptions = undefined;
    }
    if (object.status !== undefined && object.status !== null) {
      message.status = InstalledPackageStatus.fromJSON(object.status);
    } else {
      message.status = undefined;
    }
    if (object.postInstallationNotes !== undefined && object.postInstallationNotes !== null) {
      message.postInstallationNotes = String(object.postInstallationNotes);
    } else {
      message.postInstallationNotes = "";
    }
    if (object.availablePackageRef !== undefined && object.availablePackageRef !== null) {
      message.availablePackageRef = AvailablePackageReference.fromJSON(object.availablePackageRef);
    } else {
      message.availablePackageRef = undefined;
    }
    if (object.latestMatchingVersion !== undefined && object.latestMatchingVersion !== null) {
      message.latestMatchingVersion = PackageAppVersion.fromJSON(object.latestMatchingVersion);
    } else {
      message.latestMatchingVersion = undefined;
    }
    if (object.latestVersion !== undefined && object.latestVersion !== null) {
      message.latestVersion = PackageAppVersion.fromJSON(object.latestVersion);
    } else {
      message.latestVersion = undefined;
    }
    if (object.customDetail !== undefined && object.customDetail !== null) {
      message.customDetail = Any.fromJSON(object.customDetail);
    } else {
      message.customDetail = undefined;
    }
    return message;
  },

  toJSON(message: InstalledPackageDetail): unknown {
    const obj: any = {};
    message.installedPackageRef !== undefined &&
      (obj.installedPackageRef = message.installedPackageRef
        ? InstalledPackageReference.toJSON(message.installedPackageRef)
        : undefined);
    message.pkgVersionReference !== undefined &&
      (obj.pkgVersionReference = message.pkgVersionReference
        ? VersionReference.toJSON(message.pkgVersionReference)
        : undefined);
    message.name !== undefined && (obj.name = message.name);
    message.currentVersion !== undefined &&
      (obj.currentVersion = message.currentVersion
        ? PackageAppVersion.toJSON(message.currentVersion)
        : undefined);
    message.valuesApplied !== undefined && (obj.valuesApplied = message.valuesApplied);
    message.reconciliationOptions !== undefined &&
      (obj.reconciliationOptions = message.reconciliationOptions
        ? ReconciliationOptions.toJSON(message.reconciliationOptions)
        : undefined);
    message.status !== undefined &&
      (obj.status = message.status ? InstalledPackageStatus.toJSON(message.status) : undefined);
    message.postInstallationNotes !== undefined &&
      (obj.postInstallationNotes = message.postInstallationNotes);
    message.availablePackageRef !== undefined &&
      (obj.availablePackageRef = message.availablePackageRef
        ? AvailablePackageReference.toJSON(message.availablePackageRef)
        : undefined);
    message.latestMatchingVersion !== undefined &&
      (obj.latestMatchingVersion = message.latestMatchingVersion
        ? PackageAppVersion.toJSON(message.latestMatchingVersion)
        : undefined);
    message.latestVersion !== undefined &&
      (obj.latestVersion = message.latestVersion
        ? PackageAppVersion.toJSON(message.latestVersion)
        : undefined);
    message.customDetail !== undefined &&
      (obj.customDetail = message.customDetail ? Any.toJSON(message.customDetail) : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<InstalledPackageDetail>): InstalledPackageDetail {
    const message = { ...baseInstalledPackageDetail } as InstalledPackageDetail;
    if (object.installedPackageRef !== undefined && object.installedPackageRef !== null) {
      message.installedPackageRef = InstalledPackageReference.fromPartial(
        object.installedPackageRef,
      );
    } else {
      message.installedPackageRef = undefined;
    }
    if (object.pkgVersionReference !== undefined && object.pkgVersionReference !== null) {
      message.pkgVersionReference = VersionReference.fromPartial(object.pkgVersionReference);
    } else {
      message.pkgVersionReference = undefined;
    }
    if (object.name !== undefined && object.name !== null) {
      message.name = object.name;
    } else {
      message.name = "";
    }
    if (object.currentVersion !== undefined && object.currentVersion !== null) {
      message.currentVersion = PackageAppVersion.fromPartial(object.currentVersion);
    } else {
      message.currentVersion = undefined;
    }
    if (object.valuesApplied !== undefined && object.valuesApplied !== null) {
      message.valuesApplied = object.valuesApplied;
    } else {
      message.valuesApplied = "";
    }
    if (object.reconciliationOptions !== undefined && object.reconciliationOptions !== null) {
      message.reconciliationOptions = ReconciliationOptions.fromPartial(
        object.reconciliationOptions,
      );
    } else {
      message.reconciliationOptions = undefined;
    }
    if (object.status !== undefined && object.status !== null) {
      message.status = InstalledPackageStatus.fromPartial(object.status);
    } else {
      message.status = undefined;
    }
    if (object.postInstallationNotes !== undefined && object.postInstallationNotes !== null) {
      message.postInstallationNotes = object.postInstallationNotes;
    } else {
      message.postInstallationNotes = "";
    }
    if (object.availablePackageRef !== undefined && object.availablePackageRef !== null) {
      message.availablePackageRef = AvailablePackageReference.fromPartial(
        object.availablePackageRef,
      );
    } else {
      message.availablePackageRef = undefined;
    }
    if (object.latestMatchingVersion !== undefined && object.latestMatchingVersion !== null) {
      message.latestMatchingVersion = PackageAppVersion.fromPartial(object.latestMatchingVersion);
    } else {
      message.latestMatchingVersion = undefined;
    }
    if (object.latestVersion !== undefined && object.latestVersion !== null) {
      message.latestVersion = PackageAppVersion.fromPartial(object.latestVersion);
    } else {
      message.latestVersion = undefined;
    }
    if (object.customDetail !== undefined && object.customDetail !== null) {
      message.customDetail = Any.fromPartial(object.customDetail);
    } else {
      message.customDetail = undefined;
    }
    return message;
  },
};

const baseContext: object = { cluster: "", namespace: "" };

export const Context = {
  encode(message: Context, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.cluster !== "") {
      writer.uint32(10).string(message.cluster);
    }
    if (message.namespace !== "") {
      writer.uint32(18).string(message.namespace);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): Context {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseContext } as Context;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.cluster = reader.string();
          break;
        case 2:
          message.namespace = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): Context {
    const message = { ...baseContext } as Context;
    if (object.cluster !== undefined && object.cluster !== null) {
      message.cluster = String(object.cluster);
    } else {
      message.cluster = "";
    }
    if (object.namespace !== undefined && object.namespace !== null) {
      message.namespace = String(object.namespace);
    } else {
      message.namespace = "";
    }
    return message;
  },

  toJSON(message: Context): unknown {
    const obj: any = {};
    message.cluster !== undefined && (obj.cluster = message.cluster);
    message.namespace !== undefined && (obj.namespace = message.namespace);
    return obj;
  },

  fromPartial(object: DeepPartial<Context>): Context {
    const message = { ...baseContext } as Context;
    if (object.cluster !== undefined && object.cluster !== null) {
      message.cluster = object.cluster;
    } else {
      message.cluster = "";
    }
    if (object.namespace !== undefined && object.namespace !== null) {
      message.namespace = object.namespace;
    } else {
      message.namespace = "";
    }
    return message;
  },
};

const baseAvailablePackageReference: object = { identifier: "" };

export const AvailablePackageReference = {
  encode(message: AvailablePackageReference, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.context !== undefined) {
      Context.encode(message.context, writer.uint32(10).fork()).ldelim();
    }
    if (message.identifier !== "") {
      writer.uint32(18).string(message.identifier);
    }
    if (message.plugin !== undefined) {
      Plugin.encode(message.plugin, writer.uint32(26).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): AvailablePackageReference {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseAvailablePackageReference,
    } as AvailablePackageReference;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.context = Context.decode(reader, reader.uint32());
          break;
        case 2:
          message.identifier = reader.string();
          break;
        case 3:
          message.plugin = Plugin.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): AvailablePackageReference {
    const message = {
      ...baseAvailablePackageReference,
    } as AvailablePackageReference;
    if (object.context !== undefined && object.context !== null) {
      message.context = Context.fromJSON(object.context);
    } else {
      message.context = undefined;
    }
    if (object.identifier !== undefined && object.identifier !== null) {
      message.identifier = String(object.identifier);
    } else {
      message.identifier = "";
    }
    if (object.plugin !== undefined && object.plugin !== null) {
      message.plugin = Plugin.fromJSON(object.plugin);
    } else {
      message.plugin = undefined;
    }
    return message;
  },

  toJSON(message: AvailablePackageReference): unknown {
    const obj: any = {};
    message.context !== undefined &&
      (obj.context = message.context ? Context.toJSON(message.context) : undefined);
    message.identifier !== undefined && (obj.identifier = message.identifier);
    message.plugin !== undefined &&
      (obj.plugin = message.plugin ? Plugin.toJSON(message.plugin) : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<AvailablePackageReference>): AvailablePackageReference {
    const message = {
      ...baseAvailablePackageReference,
    } as AvailablePackageReference;
    if (object.context !== undefined && object.context !== null) {
      message.context = Context.fromPartial(object.context);
    } else {
      message.context = undefined;
    }
    if (object.identifier !== undefined && object.identifier !== null) {
      message.identifier = object.identifier;
    } else {
      message.identifier = "";
    }
    if (object.plugin !== undefined && object.plugin !== null) {
      message.plugin = Plugin.fromPartial(object.plugin);
    } else {
      message.plugin = undefined;
    }
    return message;
  },
};

const baseMaintainer: object = { name: "", email: "" };

export const Maintainer = {
  encode(message: Maintainer, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.name !== "") {
      writer.uint32(10).string(message.name);
    }
    if (message.email !== "") {
      writer.uint32(18).string(message.email);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): Maintainer {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMaintainer } as Maintainer;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.name = reader.string();
          break;
        case 2:
          message.email = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): Maintainer {
    const message = { ...baseMaintainer } as Maintainer;
    if (object.name !== undefined && object.name !== null) {
      message.name = String(object.name);
    } else {
      message.name = "";
    }
    if (object.email !== undefined && object.email !== null) {
      message.email = String(object.email);
    } else {
      message.email = "";
    }
    return message;
  },

  toJSON(message: Maintainer): unknown {
    const obj: any = {};
    message.name !== undefined && (obj.name = message.name);
    message.email !== undefined && (obj.email = message.email);
    return obj;
  },

  fromPartial(object: DeepPartial<Maintainer>): Maintainer {
    const message = { ...baseMaintainer } as Maintainer;
    if (object.name !== undefined && object.name !== null) {
      message.name = object.name;
    } else {
      message.name = "";
    }
    if (object.email !== undefined && object.email !== null) {
      message.email = object.email;
    } else {
      message.email = "";
    }
    return message;
  },
};

const baseFilterOptions: object = {
  query: "",
  categories: "",
  repositories: "",
  pkgVersion: "",
  appVersion: "",
};

export const FilterOptions = {
  encode(message: FilterOptions, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.query !== "") {
      writer.uint32(10).string(message.query);
    }
    for (const v of message.categories) {
      writer.uint32(18).string(v!);
    }
    for (const v of message.repositories) {
      writer.uint32(26).string(v!);
    }
    if (message.pkgVersion !== "") {
      writer.uint32(34).string(message.pkgVersion);
    }
    if (message.appVersion !== "") {
      writer.uint32(42).string(message.appVersion);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): FilterOptions {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseFilterOptions } as FilterOptions;
    message.categories = [];
    message.repositories = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.query = reader.string();
          break;
        case 2:
          message.categories.push(reader.string());
          break;
        case 3:
          message.repositories.push(reader.string());
          break;
        case 4:
          message.pkgVersion = reader.string();
          break;
        case 5:
          message.appVersion = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): FilterOptions {
    const message = { ...baseFilterOptions } as FilterOptions;
    message.categories = [];
    message.repositories = [];
    if (object.query !== undefined && object.query !== null) {
      message.query = String(object.query);
    } else {
      message.query = "";
    }
    if (object.categories !== undefined && object.categories !== null) {
      for (const e of object.categories) {
        message.categories.push(String(e));
      }
    }
    if (object.repositories !== undefined && object.repositories !== null) {
      for (const e of object.repositories) {
        message.repositories.push(String(e));
      }
    }
    if (object.pkgVersion !== undefined && object.pkgVersion !== null) {
      message.pkgVersion = String(object.pkgVersion);
    } else {
      message.pkgVersion = "";
    }
    if (object.appVersion !== undefined && object.appVersion !== null) {
      message.appVersion = String(object.appVersion);
    } else {
      message.appVersion = "";
    }
    return message;
  },

  toJSON(message: FilterOptions): unknown {
    const obj: any = {};
    message.query !== undefined && (obj.query = message.query);
    if (message.categories) {
      obj.categories = message.categories.map(e => e);
    } else {
      obj.categories = [];
    }
    if (message.repositories) {
      obj.repositories = message.repositories.map(e => e);
    } else {
      obj.repositories = [];
    }
    message.pkgVersion !== undefined && (obj.pkgVersion = message.pkgVersion);
    message.appVersion !== undefined && (obj.appVersion = message.appVersion);
    return obj;
  },

  fromPartial(object: DeepPartial<FilterOptions>): FilterOptions {
    const message = { ...baseFilterOptions } as FilterOptions;
    message.categories = [];
    message.repositories = [];
    if (object.query !== undefined && object.query !== null) {
      message.query = object.query;
    } else {
      message.query = "";
    }
    if (object.categories !== undefined && object.categories !== null) {
      for (const e of object.categories) {
        message.categories.push(e);
      }
    }
    if (object.repositories !== undefined && object.repositories !== null) {
      for (const e of object.repositories) {
        message.repositories.push(e);
      }
    }
    if (object.pkgVersion !== undefined && object.pkgVersion !== null) {
      message.pkgVersion = object.pkgVersion;
    } else {
      message.pkgVersion = "";
    }
    if (object.appVersion !== undefined && object.appVersion !== null) {
      message.appVersion = object.appVersion;
    } else {
      message.appVersion = "";
    }
    return message;
  },
};

const basePaginationOptions: object = { pageToken: "", pageSize: 0 };

export const PaginationOptions = {
  encode(message: PaginationOptions, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.pageToken !== "") {
      writer.uint32(10).string(message.pageToken);
    }
    if (message.pageSize !== 0) {
      writer.uint32(16).int32(message.pageSize);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): PaginationOptions {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...basePaginationOptions } as PaginationOptions;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.pageToken = reader.string();
          break;
        case 2:
          message.pageSize = reader.int32();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): PaginationOptions {
    const message = { ...basePaginationOptions } as PaginationOptions;
    if (object.pageToken !== undefined && object.pageToken !== null) {
      message.pageToken = String(object.pageToken);
    } else {
      message.pageToken = "";
    }
    if (object.pageSize !== undefined && object.pageSize !== null) {
      message.pageSize = Number(object.pageSize);
    } else {
      message.pageSize = 0;
    }
    return message;
  },

  toJSON(message: PaginationOptions): unknown {
    const obj: any = {};
    message.pageToken !== undefined && (obj.pageToken = message.pageToken);
    message.pageSize !== undefined && (obj.pageSize = message.pageSize);
    return obj;
  },

  fromPartial(object: DeepPartial<PaginationOptions>): PaginationOptions {
    const message = { ...basePaginationOptions } as PaginationOptions;
    if (object.pageToken !== undefined && object.pageToken !== null) {
      message.pageToken = object.pageToken;
    } else {
      message.pageToken = "";
    }
    if (object.pageSize !== undefined && object.pageSize !== null) {
      message.pageSize = object.pageSize;
    } else {
      message.pageSize = 0;
    }
    return message;
  },
};

const baseInstalledPackageReference: object = { identifier: "" };

export const InstalledPackageReference = {
  encode(message: InstalledPackageReference, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.context !== undefined) {
      Context.encode(message.context, writer.uint32(10).fork()).ldelim();
    }
    if (message.identifier !== "") {
      writer.uint32(18).string(message.identifier);
    }
    if (message.plugin !== undefined) {
      Plugin.encode(message.plugin, writer.uint32(26).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): InstalledPackageReference {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseInstalledPackageReference,
    } as InstalledPackageReference;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.context = Context.decode(reader, reader.uint32());
          break;
        case 2:
          message.identifier = reader.string();
          break;
        case 3:
          message.plugin = Plugin.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): InstalledPackageReference {
    const message = {
      ...baseInstalledPackageReference,
    } as InstalledPackageReference;
    if (object.context !== undefined && object.context !== null) {
      message.context = Context.fromJSON(object.context);
    } else {
      message.context = undefined;
    }
    if (object.identifier !== undefined && object.identifier !== null) {
      message.identifier = String(object.identifier);
    } else {
      message.identifier = "";
    }
    if (object.plugin !== undefined && object.plugin !== null) {
      message.plugin = Plugin.fromJSON(object.plugin);
    } else {
      message.plugin = undefined;
    }
    return message;
  },

  toJSON(message: InstalledPackageReference): unknown {
    const obj: any = {};
    message.context !== undefined &&
      (obj.context = message.context ? Context.toJSON(message.context) : undefined);
    message.identifier !== undefined && (obj.identifier = message.identifier);
    message.plugin !== undefined &&
      (obj.plugin = message.plugin ? Plugin.toJSON(message.plugin) : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<InstalledPackageReference>): InstalledPackageReference {
    const message = {
      ...baseInstalledPackageReference,
    } as InstalledPackageReference;
    if (object.context !== undefined && object.context !== null) {
      message.context = Context.fromPartial(object.context);
    } else {
      message.context = undefined;
    }
    if (object.identifier !== undefined && object.identifier !== null) {
      message.identifier = object.identifier;
    } else {
      message.identifier = "";
    }
    if (object.plugin !== undefined && object.plugin !== null) {
      message.plugin = Plugin.fromPartial(object.plugin);
    } else {
      message.plugin = undefined;
    }
    return message;
  },
};

const baseVersionReference: object = { version: "" };

export const VersionReference = {
  encode(message: VersionReference, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.version !== "") {
      writer.uint32(10).string(message.version);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): VersionReference {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseVersionReference } as VersionReference;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.version = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): VersionReference {
    const message = { ...baseVersionReference } as VersionReference;
    if (object.version !== undefined && object.version !== null) {
      message.version = String(object.version);
    } else {
      message.version = "";
    }
    return message;
  },

  toJSON(message: VersionReference): unknown {
    const obj: any = {};
    message.version !== undefined && (obj.version = message.version);
    return obj;
  },

  fromPartial(object: DeepPartial<VersionReference>): VersionReference {
    const message = { ...baseVersionReference } as VersionReference;
    if (object.version !== undefined && object.version !== null) {
      message.version = object.version;
    } else {
      message.version = "";
    }
    return message;
  },
};

const baseInstalledPackageStatus: object = {
  ready: false,
  reason: 0,
  userReason: "",
};

export const InstalledPackageStatus = {
  encode(message: InstalledPackageStatus, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.ready === true) {
      writer.uint32(8).bool(message.ready);
    }
    if (message.reason !== 0) {
      writer.uint32(16).int32(message.reason);
    }
    if (message.userReason !== "") {
      writer.uint32(26).string(message.userReason);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): InstalledPackageStatus {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseInstalledPackageStatus } as InstalledPackageStatus;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.ready = reader.bool();
          break;
        case 2:
          message.reason = reader.int32() as any;
          break;
        case 3:
          message.userReason = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): InstalledPackageStatus {
    const message = { ...baseInstalledPackageStatus } as InstalledPackageStatus;
    if (object.ready !== undefined && object.ready !== null) {
      message.ready = Boolean(object.ready);
    } else {
      message.ready = false;
    }
    if (object.reason !== undefined && object.reason !== null) {
      message.reason = installedPackageStatus_StatusReasonFromJSON(object.reason);
    } else {
      message.reason = 0;
    }
    if (object.userReason !== undefined && object.userReason !== null) {
      message.userReason = String(object.userReason);
    } else {
      message.userReason = "";
    }
    return message;
  },

  toJSON(message: InstalledPackageStatus): unknown {
    const obj: any = {};
    message.ready !== undefined && (obj.ready = message.ready);
    message.reason !== undefined &&
      (obj.reason = installedPackageStatus_StatusReasonToJSON(message.reason));
    message.userReason !== undefined && (obj.userReason = message.userReason);
    return obj;
  },

  fromPartial(object: DeepPartial<InstalledPackageStatus>): InstalledPackageStatus {
    const message = { ...baseInstalledPackageStatus } as InstalledPackageStatus;
    if (object.ready !== undefined && object.ready !== null) {
      message.ready = object.ready;
    } else {
      message.ready = false;
    }
    if (object.reason !== undefined && object.reason !== null) {
      message.reason = object.reason;
    } else {
      message.reason = 0;
    }
    if (object.userReason !== undefined && object.userReason !== null) {
      message.userReason = object.userReason;
    } else {
      message.userReason = "";
    }
    return message;
  },
};

const baseReconciliationOptions: object = {
  interval: 0,
  suspend: false,
  serviceAccountName: "",
};

export const ReconciliationOptions = {
  encode(message: ReconciliationOptions, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.interval !== 0) {
      writer.uint32(8).int32(message.interval);
    }
    if (message.suspend === true) {
      writer.uint32(16).bool(message.suspend);
    }
    if (message.serviceAccountName !== "") {
      writer.uint32(26).string(message.serviceAccountName);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): ReconciliationOptions {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseReconciliationOptions } as ReconciliationOptions;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.interval = reader.int32();
          break;
        case 2:
          message.suspend = reader.bool();
          break;
        case 3:
          message.serviceAccountName = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): ReconciliationOptions {
    const message = { ...baseReconciliationOptions } as ReconciliationOptions;
    if (object.interval !== undefined && object.interval !== null) {
      message.interval = Number(object.interval);
    } else {
      message.interval = 0;
    }
    if (object.suspend !== undefined && object.suspend !== null) {
      message.suspend = Boolean(object.suspend);
    } else {
      message.suspend = false;
    }
    if (object.serviceAccountName !== undefined && object.serviceAccountName !== null) {
      message.serviceAccountName = String(object.serviceAccountName);
    } else {
      message.serviceAccountName = "";
    }
    return message;
  },

  toJSON(message: ReconciliationOptions): unknown {
    const obj: any = {};
    message.interval !== undefined && (obj.interval = message.interval);
    message.suspend !== undefined && (obj.suspend = message.suspend);
    message.serviceAccountName !== undefined &&
      (obj.serviceAccountName = message.serviceAccountName);
    return obj;
  },

  fromPartial(object: DeepPartial<ReconciliationOptions>): ReconciliationOptions {
    const message = { ...baseReconciliationOptions } as ReconciliationOptions;
    if (object.interval !== undefined && object.interval !== null) {
      message.interval = object.interval;
    } else {
      message.interval = 0;
    }
    if (object.suspend !== undefined && object.suspend !== null) {
      message.suspend = object.suspend;
    } else {
      message.suspend = false;
    }
    if (object.serviceAccountName !== undefined && object.serviceAccountName !== null) {
      message.serviceAccountName = object.serviceAccountName;
    } else {
      message.serviceAccountName = "";
    }
    return message;
  },
};

const basePackageAppVersion: object = { pkgVersion: "", appVersion: "" };

export const PackageAppVersion = {
  encode(message: PackageAppVersion, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.pkgVersion !== "") {
      writer.uint32(10).string(message.pkgVersion);
    }
    if (message.appVersion !== "") {
      writer.uint32(18).string(message.appVersion);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): PackageAppVersion {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...basePackageAppVersion } as PackageAppVersion;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.pkgVersion = reader.string();
          break;
        case 2:
          message.appVersion = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): PackageAppVersion {
    const message = { ...basePackageAppVersion } as PackageAppVersion;
    if (object.pkgVersion !== undefined && object.pkgVersion !== null) {
      message.pkgVersion = String(object.pkgVersion);
    } else {
      message.pkgVersion = "";
    }
    if (object.appVersion !== undefined && object.appVersion !== null) {
      message.appVersion = String(object.appVersion);
    } else {
      message.appVersion = "";
    }
    return message;
  },

  toJSON(message: PackageAppVersion): unknown {
    const obj: any = {};
    message.pkgVersion !== undefined && (obj.pkgVersion = message.pkgVersion);
    message.appVersion !== undefined && (obj.appVersion = message.appVersion);
    return obj;
  },

  fromPartial(object: DeepPartial<PackageAppVersion>): PackageAppVersion {
    const message = { ...basePackageAppVersion } as PackageAppVersion;
    if (object.pkgVersion !== undefined && object.pkgVersion !== null) {
      message.pkgVersion = object.pkgVersion;
    } else {
      message.pkgVersion = "";
    }
    if (object.appVersion !== undefined && object.appVersion !== null) {
      message.appVersion = object.appVersion;
    } else {
      message.appVersion = "";
    }
    return message;
  },
};

const baseResourceRef: object = {
  apiVersion: "",
  kind: "",
  name: "",
  namespace: "",
};

export const ResourceRef = {
  encode(message: ResourceRef, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.apiVersion !== "") {
      writer.uint32(10).string(message.apiVersion);
    }
    if (message.kind !== "") {
      writer.uint32(18).string(message.kind);
    }
    if (message.name !== "") {
      writer.uint32(26).string(message.name);
    }
    if (message.namespace !== "") {
      writer.uint32(34).string(message.namespace);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): ResourceRef {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseResourceRef } as ResourceRef;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.apiVersion = reader.string();
          break;
        case 2:
          message.kind = reader.string();
          break;
        case 3:
          message.name = reader.string();
          break;
        case 4:
          message.namespace = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): ResourceRef {
    const message = { ...baseResourceRef } as ResourceRef;
    if (object.apiVersion !== undefined && object.apiVersion !== null) {
      message.apiVersion = String(object.apiVersion);
    } else {
      message.apiVersion = "";
    }
    if (object.kind !== undefined && object.kind !== null) {
      message.kind = String(object.kind);
    } else {
      message.kind = "";
    }
    if (object.name !== undefined && object.name !== null) {
      message.name = String(object.name);
    } else {
      message.name = "";
    }
    if (object.namespace !== undefined && object.namespace !== null) {
      message.namespace = String(object.namespace);
    } else {
      message.namespace = "";
    }
    return message;
  },

  toJSON(message: ResourceRef): unknown {
    const obj: any = {};
    message.apiVersion !== undefined && (obj.apiVersion = message.apiVersion);
    message.kind !== undefined && (obj.kind = message.kind);
    message.name !== undefined && (obj.name = message.name);
    message.namespace !== undefined && (obj.namespace = message.namespace);
    return obj;
  },

  fromPartial(object: DeepPartial<ResourceRef>): ResourceRef {
    const message = { ...baseResourceRef } as ResourceRef;
    if (object.apiVersion !== undefined && object.apiVersion !== null) {
      message.apiVersion = object.apiVersion;
    } else {
      message.apiVersion = "";
    }
    if (object.kind !== undefined && object.kind !== null) {
      message.kind = object.kind;
    } else {
      message.kind = "";
    }
    if (object.name !== undefined && object.name !== null) {
      message.name = object.name;
    } else {
      message.name = "";
    }
    if (object.namespace !== undefined && object.namespace !== null) {
      message.namespace = object.namespace;
    } else {
      message.namespace = "";
    }
    return message;
  },
};

/** Each packages v1alpha1 plugin must implement at least the following rpcs: */
export interface PackagesService {
  GetAvailablePackageSummaries(
    request: DeepPartial<GetAvailablePackageSummariesRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetAvailablePackageSummariesResponse>;
  GetAvailablePackageDetail(
    request: DeepPartial<GetAvailablePackageDetailRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetAvailablePackageDetailResponse>;
  GetAvailablePackageVersions(
    request: DeepPartial<GetAvailablePackageVersionsRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetAvailablePackageVersionsResponse>;
  GetInstalledPackageSummaries(
    request: DeepPartial<GetInstalledPackageSummariesRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetInstalledPackageSummariesResponse>;
  GetInstalledPackageDetail(
    request: DeepPartial<GetInstalledPackageDetailRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetInstalledPackageDetailResponse>;
  CreateInstalledPackage(
    request: DeepPartial<CreateInstalledPackageRequest>,
    metadata?: grpc.Metadata,
  ): Promise<CreateInstalledPackageResponse>;
  UpdateInstalledPackage(
    request: DeepPartial<UpdateInstalledPackageRequest>,
    metadata?: grpc.Metadata,
  ): Promise<UpdateInstalledPackageResponse>;
  DeleteInstalledPackage(
    request: DeepPartial<DeleteInstalledPackageRequest>,
    metadata?: grpc.Metadata,
  ): Promise<DeleteInstalledPackageResponse>;
  GetInstalledPackageResourceRefs(
    request: DeepPartial<GetInstalledPackageResourceRefsRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetInstalledPackageResourceRefsResponse>;
}

export class PackagesServiceClientImpl implements PackagesService {
  private readonly rpc: Rpc;

  constructor(rpc: Rpc) {
    this.rpc = rpc;
    this.GetAvailablePackageSummaries = this.GetAvailablePackageSummaries.bind(this);
    this.GetAvailablePackageDetail = this.GetAvailablePackageDetail.bind(this);
    this.GetAvailablePackageVersions = this.GetAvailablePackageVersions.bind(this);
    this.GetInstalledPackageSummaries = this.GetInstalledPackageSummaries.bind(this);
    this.GetInstalledPackageDetail = this.GetInstalledPackageDetail.bind(this);
    this.CreateInstalledPackage = this.CreateInstalledPackage.bind(this);
    this.UpdateInstalledPackage = this.UpdateInstalledPackage.bind(this);
    this.DeleteInstalledPackage = this.DeleteInstalledPackage.bind(this);
    this.GetInstalledPackageResourceRefs = this.GetInstalledPackageResourceRefs.bind(this);
  }

  GetAvailablePackageSummaries(
    request: DeepPartial<GetAvailablePackageSummariesRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetAvailablePackageSummariesResponse> {
    return this.rpc.unary(
      PackagesServiceGetAvailablePackageSummariesDesc,
      GetAvailablePackageSummariesRequest.fromPartial(request),
      metadata,
    );
  }

  GetAvailablePackageDetail(
    request: DeepPartial<GetAvailablePackageDetailRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetAvailablePackageDetailResponse> {
    return this.rpc.unary(
      PackagesServiceGetAvailablePackageDetailDesc,
      GetAvailablePackageDetailRequest.fromPartial(request),
      metadata,
    );
  }

  GetAvailablePackageVersions(
    request: DeepPartial<GetAvailablePackageVersionsRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetAvailablePackageVersionsResponse> {
    return this.rpc.unary(
      PackagesServiceGetAvailablePackageVersionsDesc,
      GetAvailablePackageVersionsRequest.fromPartial(request),
      metadata,
    );
  }

  GetInstalledPackageSummaries(
    request: DeepPartial<GetInstalledPackageSummariesRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetInstalledPackageSummariesResponse> {
    return this.rpc.unary(
      PackagesServiceGetInstalledPackageSummariesDesc,
      GetInstalledPackageSummariesRequest.fromPartial(request),
      metadata,
    );
  }

  GetInstalledPackageDetail(
    request: DeepPartial<GetInstalledPackageDetailRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetInstalledPackageDetailResponse> {
    return this.rpc.unary(
      PackagesServiceGetInstalledPackageDetailDesc,
      GetInstalledPackageDetailRequest.fromPartial(request),
      metadata,
    );
  }

  CreateInstalledPackage(
    request: DeepPartial<CreateInstalledPackageRequest>,
    metadata?: grpc.Metadata,
  ): Promise<CreateInstalledPackageResponse> {
    return this.rpc.unary(
      PackagesServiceCreateInstalledPackageDesc,
      CreateInstalledPackageRequest.fromPartial(request),
      metadata,
    );
  }

  UpdateInstalledPackage(
    request: DeepPartial<UpdateInstalledPackageRequest>,
    metadata?: grpc.Metadata,
  ): Promise<UpdateInstalledPackageResponse> {
    return this.rpc.unary(
      PackagesServiceUpdateInstalledPackageDesc,
      UpdateInstalledPackageRequest.fromPartial(request),
      metadata,
    );
  }

  DeleteInstalledPackage(
    request: DeepPartial<DeleteInstalledPackageRequest>,
    metadata?: grpc.Metadata,
  ): Promise<DeleteInstalledPackageResponse> {
    return this.rpc.unary(
      PackagesServiceDeleteInstalledPackageDesc,
      DeleteInstalledPackageRequest.fromPartial(request),
      metadata,
    );
  }

  GetInstalledPackageResourceRefs(
    request: DeepPartial<GetInstalledPackageResourceRefsRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetInstalledPackageResourceRefsResponse> {
    return this.rpc.unary(
      PackagesServiceGetInstalledPackageResourceRefsDesc,
      GetInstalledPackageResourceRefsRequest.fromPartial(request),
      metadata,
    );
  }
}

export const PackagesServiceDesc = {
  serviceName: "kubeappsapis.core.packages.v1alpha1.PackagesService",
};

export const PackagesServiceGetAvailablePackageSummariesDesc: UnaryMethodDefinitionish = {
  methodName: "GetAvailablePackageSummaries",
  service: PackagesServiceDesc,
  requestStream: false,
  responseStream: false,
  requestType: {
    serializeBinary() {
      return GetAvailablePackageSummariesRequest.encode(this).finish();
    },
  } as any,
  responseType: {
    deserializeBinary(data: Uint8Array) {
      return {
        ...GetAvailablePackageSummariesResponse.decode(data),
        toObject() {
          return this;
        },
      };
    },
  } as any,
};

export const PackagesServiceGetAvailablePackageDetailDesc: UnaryMethodDefinitionish = {
  methodName: "GetAvailablePackageDetail",
  service: PackagesServiceDesc,
  requestStream: false,
  responseStream: false,
  requestType: {
    serializeBinary() {
      return GetAvailablePackageDetailRequest.encode(this).finish();
    },
  } as any,
  responseType: {
    deserializeBinary(data: Uint8Array) {
      return {
        ...GetAvailablePackageDetailResponse.decode(data),
        toObject() {
          return this;
        },
      };
    },
  } as any,
};

export const PackagesServiceGetAvailablePackageVersionsDesc: UnaryMethodDefinitionish = {
  methodName: "GetAvailablePackageVersions",
  service: PackagesServiceDesc,
  requestStream: false,
  responseStream: false,
  requestType: {
    serializeBinary() {
      return GetAvailablePackageVersionsRequest.encode(this).finish();
    },
  } as any,
  responseType: {
    deserializeBinary(data: Uint8Array) {
      return {
        ...GetAvailablePackageVersionsResponse.decode(data),
        toObject() {
          return this;
        },
      };
    },
  } as any,
};

export const PackagesServiceGetInstalledPackageSummariesDesc: UnaryMethodDefinitionish = {
  methodName: "GetInstalledPackageSummaries",
  service: PackagesServiceDesc,
  requestStream: false,
  responseStream: false,
  requestType: {
    serializeBinary() {
      return GetInstalledPackageSummariesRequest.encode(this).finish();
    },
  } as any,
  responseType: {
    deserializeBinary(data: Uint8Array) {
      return {
        ...GetInstalledPackageSummariesResponse.decode(data),
        toObject() {
          return this;
        },
      };
    },
  } as any,
};

export const PackagesServiceGetInstalledPackageDetailDesc: UnaryMethodDefinitionish = {
  methodName: "GetInstalledPackageDetail",
  service: PackagesServiceDesc,
  requestStream: false,
  responseStream: false,
  requestType: {
    serializeBinary() {
      return GetInstalledPackageDetailRequest.encode(this).finish();
    },
  } as any,
  responseType: {
    deserializeBinary(data: Uint8Array) {
      return {
        ...GetInstalledPackageDetailResponse.decode(data),
        toObject() {
          return this;
        },
      };
    },
  } as any,
};

export const PackagesServiceCreateInstalledPackageDesc: UnaryMethodDefinitionish = {
  methodName: "CreateInstalledPackage",
  service: PackagesServiceDesc,
  requestStream: false,
  responseStream: false,
  requestType: {
    serializeBinary() {
      return CreateInstalledPackageRequest.encode(this).finish();
    },
  } as any,
  responseType: {
    deserializeBinary(data: Uint8Array) {
      return {
        ...CreateInstalledPackageResponse.decode(data),
        toObject() {
          return this;
        },
      };
    },
  } as any,
};

export const PackagesServiceUpdateInstalledPackageDesc: UnaryMethodDefinitionish = {
  methodName: "UpdateInstalledPackage",
  service: PackagesServiceDesc,
  requestStream: false,
  responseStream: false,
  requestType: {
    serializeBinary() {
      return UpdateInstalledPackageRequest.encode(this).finish();
    },
  } as any,
  responseType: {
    deserializeBinary(data: Uint8Array) {
      return {
        ...UpdateInstalledPackageResponse.decode(data),
        toObject() {
          return this;
        },
      };
    },
  } as any,
};

export const PackagesServiceDeleteInstalledPackageDesc: UnaryMethodDefinitionish = {
  methodName: "DeleteInstalledPackage",
  service: PackagesServiceDesc,
  requestStream: false,
  responseStream: false,
  requestType: {
    serializeBinary() {
      return DeleteInstalledPackageRequest.encode(this).finish();
    },
  } as any,
  responseType: {
    deserializeBinary(data: Uint8Array) {
      return {
        ...DeleteInstalledPackageResponse.decode(data),
        toObject() {
          return this;
        },
      };
    },
  } as any,
};

export const PackagesServiceGetInstalledPackageResourceRefsDesc: UnaryMethodDefinitionish = {
  methodName: "GetInstalledPackageResourceRefs",
  service: PackagesServiceDesc,
  requestStream: false,
  responseStream: false,
  requestType: {
    serializeBinary() {
      return GetInstalledPackageResourceRefsRequest.encode(this).finish();
    },
  } as any,
  responseType: {
    deserializeBinary(data: Uint8Array) {
      return {
        ...GetInstalledPackageResourceRefsResponse.decode(data),
        toObject() {
          return this;
        },
      };
    },
  } as any,
};

interface UnaryMethodDefinitionishR extends grpc.UnaryMethodDefinition<any, any> {
  requestStream: any;
  responseStream: any;
}

type UnaryMethodDefinitionish = UnaryMethodDefinitionishR;

interface Rpc {
  unary<T extends UnaryMethodDefinitionish>(
    methodDesc: T,
    request: any,
    metadata: grpc.Metadata | undefined,
  ): Promise<any>;
}

export class GrpcWebImpl {
  private host: string;
  private options: {
    transport?: grpc.TransportFactory;

    debug?: boolean;
    metadata?: grpc.Metadata;
  };

  constructor(
    host: string,
    options: {
      transport?: grpc.TransportFactory;

      debug?: boolean;
      metadata?: grpc.Metadata;
    },
  ) {
    this.host = host;
    this.options = options;
  }

  unary<T extends UnaryMethodDefinitionish>(
    methodDesc: T,
    _request: any,
    metadata: grpc.Metadata | undefined,
  ): Promise<any> {
    const request = { ..._request, ...methodDesc.requestType };
    const maybeCombinedMetadata =
      metadata && this.options.metadata
        ? new BrowserHeaders({
            ...this.options?.metadata.headersMap,
            ...metadata?.headersMap,
          })
        : metadata || this.options.metadata;
    return new Promise((resolve, reject) => {
      grpc.unary(methodDesc, {
        request,
        host: this.host,
        metadata: maybeCombinedMetadata,
        transport: this.options.transport,
        debug: this.options.debug,
        onEnd: function (response) {
          if (response.status === grpc.Code.OK) {
            resolve(response.message);
          } else {
            const err = new Error(response.statusMessage) as any;
            err.code = response.status;
            err.metadata = response.trailers;
            reject(err);
          }
        },
      });
    });
  }
}

type Builtin = Date | Function | Uint8Array | string | number | boolean | undefined;
export type DeepPartial<T> = T extends Builtin
  ? T
  : T extends Array<infer U>
  ? Array<DeepPartial<U>>
  : T extends ReadonlyArray<infer U>
  ? ReadonlyArray<DeepPartial<U>>
  : T extends {}
  ? { [K in keyof T]?: DeepPartial<T[K]> }
  : Partial<T>;

if (_m0.util.Long !== Long) {
  _m0.util.Long = Long as any;
  _m0.configure();
}
