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
  /**
   * PaginationOptions
   *
   * Pagination options specifying where to start and how many results to include.
   */
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
  availablePackagesSummaries: AvailablePackageSummary[];
  /**
   * Next page token
   *
   * This field represents the pagination token to retrieve the next page of
   * results. If the value is "", it means no further results for the request.
   */
  nextPageToken: string;
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
   * List of AvailablePackageDetail
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
  packageAppVersions: GetAvailablePackageVersionsResponse_PackageAppVersion[];
}

/** PackageAppVersion conveys both the package version and the packaged app version. */
export interface GetAvailablePackageVersionsResponse_PackageAppVersion {
  pkgVersion: string;
  appVersion: string;
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
   * Latest available package version
   *
   * The latest version available for this package. Often expected when viewing
   * a summary of many available packages.
   */
  latestPkgVersion: string;
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
   * Available package version
   *
   * The version of the package (eg. chart version)
   */
  pkgVersion: string;
  /**
   * Available package app version
   *
   * The version of the packaged application (eg. wordpress version or whatever).
   */
  appVersion: string;
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
  /**
   * Available package maintainers
   *
   * List of Maintainer
   */
  maintainers: Maintainer[];
  /**
   * Custom data added by the plugin
   *
   * Some additional information added by the plugin
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

const baseGetAvailablePackageSummariesResponse: object = { nextPageToken: "" };

export const GetAvailablePackageSummariesResponse = {
  encode(
    message: GetAvailablePackageSummariesResponse,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    for (const v of message.availablePackagesSummaries) {
      AvailablePackageSummary.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    if (message.nextPageToken !== "") {
      writer.uint32(18).string(message.nextPageToken);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetAvailablePackageSummariesResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseGetAvailablePackageSummariesResponse,
    } as GetAvailablePackageSummariesResponse;
    message.availablePackagesSummaries = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.availablePackagesSummaries.push(
            AvailablePackageSummary.decode(reader, reader.uint32()),
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

  fromJSON(object: any): GetAvailablePackageSummariesResponse {
    const message = {
      ...baseGetAvailablePackageSummariesResponse,
    } as GetAvailablePackageSummariesResponse;
    message.availablePackagesSummaries = [];
    if (
      object.availablePackagesSummaries !== undefined &&
      object.availablePackagesSummaries !== null
    ) {
      for (const e of object.availablePackagesSummaries) {
        message.availablePackagesSummaries.push(AvailablePackageSummary.fromJSON(e));
      }
    }
    if (object.nextPageToken !== undefined && object.nextPageToken !== null) {
      message.nextPageToken = String(object.nextPageToken);
    } else {
      message.nextPageToken = "";
    }
    return message;
  },

  toJSON(message: GetAvailablePackageSummariesResponse): unknown {
    const obj: any = {};
    if (message.availablePackagesSummaries) {
      obj.availablePackagesSummaries = message.availablePackagesSummaries.map(e =>
        e ? AvailablePackageSummary.toJSON(e) : undefined,
      );
    } else {
      obj.availablePackagesSummaries = [];
    }
    message.nextPageToken !== undefined && (obj.nextPageToken = message.nextPageToken);
    return obj;
  },

  fromPartial(
    object: DeepPartial<GetAvailablePackageSummariesResponse>,
  ): GetAvailablePackageSummariesResponse {
    const message = {
      ...baseGetAvailablePackageSummariesResponse,
    } as GetAvailablePackageSummariesResponse;
    message.availablePackagesSummaries = [];
    if (
      object.availablePackagesSummaries !== undefined &&
      object.availablePackagesSummaries !== null
    ) {
      for (const e of object.availablePackagesSummaries) {
        message.availablePackagesSummaries.push(AvailablePackageSummary.fromPartial(e));
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
      GetAvailablePackageVersionsResponse_PackageAppVersion.encode(
        v!,
        writer.uint32(10).fork(),
      ).ldelim();
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
          message.packageAppVersions.push(
            GetAvailablePackageVersionsResponse_PackageAppVersion.decode(reader, reader.uint32()),
          );
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
        message.packageAppVersions.push(
          GetAvailablePackageVersionsResponse_PackageAppVersion.fromJSON(e),
        );
      }
    }
    return message;
  },

  toJSON(message: GetAvailablePackageVersionsResponse): unknown {
    const obj: any = {};
    if (message.packageAppVersions) {
      obj.packageAppVersions = message.packageAppVersions.map(e =>
        e ? GetAvailablePackageVersionsResponse_PackageAppVersion.toJSON(e) : undefined,
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
        message.packageAppVersions.push(
          GetAvailablePackageVersionsResponse_PackageAppVersion.fromPartial(e),
        );
      }
    }
    return message;
  },
};

const baseGetAvailablePackageVersionsResponse_PackageAppVersion: object = {
  pkgVersion: "",
  appVersion: "",
};

export const GetAvailablePackageVersionsResponse_PackageAppVersion = {
  encode(
    message: GetAvailablePackageVersionsResponse_PackageAppVersion,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.pkgVersion !== "") {
      writer.uint32(10).string(message.pkgVersion);
    }
    if (message.appVersion !== "") {
      writer.uint32(18).string(message.appVersion);
    }
    return writer;
  },

  decode(
    input: _m0.Reader | Uint8Array,
    length?: number,
  ): GetAvailablePackageVersionsResponse_PackageAppVersion {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseGetAvailablePackageVersionsResponse_PackageAppVersion,
    } as GetAvailablePackageVersionsResponse_PackageAppVersion;
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

  fromJSON(object: any): GetAvailablePackageVersionsResponse_PackageAppVersion {
    const message = {
      ...baseGetAvailablePackageVersionsResponse_PackageAppVersion,
    } as GetAvailablePackageVersionsResponse_PackageAppVersion;
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

  toJSON(message: GetAvailablePackageVersionsResponse_PackageAppVersion): unknown {
    const obj: any = {};
    message.pkgVersion !== undefined && (obj.pkgVersion = message.pkgVersion);
    message.appVersion !== undefined && (obj.appVersion = message.appVersion);
    return obj;
  },

  fromPartial(
    object: DeepPartial<GetAvailablePackageVersionsResponse_PackageAppVersion>,
  ): GetAvailablePackageVersionsResponse_PackageAppVersion {
    const message = {
      ...baseGetAvailablePackageVersionsResponse_PackageAppVersion,
    } as GetAvailablePackageVersionsResponse_PackageAppVersion;
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

const baseAvailablePackageSummary: object = {
  name: "",
  latestPkgVersion: "",
  iconUrl: "",
  displayName: "",
  shortDescription: "",
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
    if (message.latestPkgVersion !== "") {
      writer.uint32(26).string(message.latestPkgVersion);
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
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): AvailablePackageSummary {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseAvailablePackageSummary,
    } as AvailablePackageSummary;
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
          message.latestPkgVersion = reader.string();
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
    if (object.latestPkgVersion !== undefined && object.latestPkgVersion !== null) {
      message.latestPkgVersion = String(object.latestPkgVersion);
    } else {
      message.latestPkgVersion = "";
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
    return message;
  },

  toJSON(message: AvailablePackageSummary): unknown {
    const obj: any = {};
    message.availablePackageRef !== undefined &&
      (obj.availablePackageRef = message.availablePackageRef
        ? AvailablePackageReference.toJSON(message.availablePackageRef)
        : undefined);
    message.name !== undefined && (obj.name = message.name);
    message.latestPkgVersion !== undefined && (obj.latestPkgVersion = message.latestPkgVersion);
    message.iconUrl !== undefined && (obj.iconUrl = message.iconUrl);
    message.displayName !== undefined && (obj.displayName = message.displayName);
    message.shortDescription !== undefined && (obj.shortDescription = message.shortDescription);
    return obj;
  },

  fromPartial(object: DeepPartial<AvailablePackageSummary>): AvailablePackageSummary {
    const message = {
      ...baseAvailablePackageSummary,
    } as AvailablePackageSummary;
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
    if (object.latestPkgVersion !== undefined && object.latestPkgVersion !== null) {
      message.latestPkgVersion = object.latestPkgVersion;
    } else {
      message.latestPkgVersion = "";
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
    return message;
  },
};

const baseAvailablePackageDetail: object = {
  name: "",
  pkgVersion: "",
  appVersion: "",
  iconUrl: "",
  displayName: "",
  shortDescription: "",
  longDescription: "",
  readme: "",
  defaultValues: "",
  valuesSchema: "",
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
    if (message.pkgVersion !== "") {
      writer.uint32(26).string(message.pkgVersion);
    }
    if (message.appVersion !== "") {
      writer.uint32(34).string(message.appVersion);
    }
    if (message.iconUrl !== "") {
      writer.uint32(42).string(message.iconUrl);
    }
    if (message.displayName !== "") {
      writer.uint32(50).string(message.displayName);
    }
    if (message.shortDescription !== "") {
      writer.uint32(58).string(message.shortDescription);
    }
    if (message.longDescription !== "") {
      writer.uint32(66).string(message.longDescription);
    }
    if (message.readme !== "") {
      writer.uint32(74).string(message.readme);
    }
    if (message.defaultValues !== "") {
      writer.uint32(82).string(message.defaultValues);
    }
    if (message.valuesSchema !== "") {
      writer.uint32(90).string(message.valuesSchema);
    }
    for (const v of message.maintainers) {
      Maintainer.encode(v!, writer.uint32(98).fork()).ldelim();
    }
    if (message.customDetail !== undefined) {
      Any.encode(message.customDetail, writer.uint32(106).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): AvailablePackageDetail {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseAvailablePackageDetail } as AvailablePackageDetail;
    message.maintainers = [];
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
          message.pkgVersion = reader.string();
          break;
        case 4:
          message.appVersion = reader.string();
          break;
        case 5:
          message.iconUrl = reader.string();
          break;
        case 6:
          message.displayName = reader.string();
          break;
        case 7:
          message.shortDescription = reader.string();
          break;
        case 8:
          message.longDescription = reader.string();
          break;
        case 9:
          message.readme = reader.string();
          break;
        case 10:
          message.defaultValues = reader.string();
          break;
        case 11:
          message.valuesSchema = reader.string();
          break;
        case 12:
          message.maintainers.push(Maintainer.decode(reader, reader.uint32()));
          break;
        case 13:
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
    message.maintainers = [];
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
    if (object.maintainers !== undefined && object.maintainers !== null) {
      for (const e of object.maintainers) {
        message.maintainers.push(Maintainer.fromJSON(e));
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
    message.pkgVersion !== undefined && (obj.pkgVersion = message.pkgVersion);
    message.appVersion !== undefined && (obj.appVersion = message.appVersion);
    message.iconUrl !== undefined && (obj.iconUrl = message.iconUrl);
    message.displayName !== undefined && (obj.displayName = message.displayName);
    message.shortDescription !== undefined && (obj.shortDescription = message.shortDescription);
    message.longDescription !== undefined && (obj.longDescription = message.longDescription);
    message.readme !== undefined && (obj.readme = message.readme);
    message.defaultValues !== undefined && (obj.defaultValues = message.defaultValues);
    message.valuesSchema !== undefined && (obj.valuesSchema = message.valuesSchema);
    if (message.maintainers) {
      obj.maintainers = message.maintainers.map(e => (e ? Maintainer.toJSON(e) : undefined));
    } else {
      obj.maintainers = [];
    }
    message.customDetail !== undefined &&
      (obj.customDetail = message.customDetail ? Any.toJSON(message.customDetail) : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<AvailablePackageDetail>): AvailablePackageDetail {
    const message = { ...baseAvailablePackageDetail } as AvailablePackageDetail;
    message.maintainers = [];
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
    if (object.maintainers !== undefined && object.maintainers !== null) {
      for (const e of object.maintainers) {
        message.maintainers.push(Maintainer.fromPartial(e));
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
}

export class PackagesServiceClientImpl implements PackagesService {
  private readonly rpc: Rpc;

  constructor(rpc: Rpc) {
    this.rpc = rpc;
    this.GetAvailablePackageSummaries = this.GetAvailablePackageSummaries.bind(this);
    this.GetAvailablePackageDetail = this.GetAvailablePackageDetail.bind(this);
    this.GetAvailablePackageVersions = this.GetAvailablePackageVersions.bind(this);
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
