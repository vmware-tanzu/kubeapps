/* eslint-disable */
import Long from "long";
import { grpc } from "@improbable-eng/grpc-web";
import _m0 from "protobufjs/minimal";
import {
  InstalledPackageReference,
  ResourceRef,
  Context,
} from "../../../../kubeappsapis/core/packages/v1alpha1/packages";
import { Observable } from "rxjs";
import { BrowserHeaders } from "browser-headers";
import { share } from "rxjs/operators";

export const protobufPackage = "kubeappsapis.plugins.resources.v1alpha1";

/**
 * SecretType
 *
 * The type of secret. Currently Kubeapps itself only deals with OPAQUE
 * and docker config json secrets, but we define all so we can correctly
 * list the secret names with their types.
 * See https://kubernetes.io/docs/concepts/configuration/secret/#secret-types
 */
export enum SecretType {
  SECRET_TYPE_OPAQUE_UNSPECIFIED = 0,
  SECRET_TYPE_SERVICE_ACCOUNT_TOKEN = 1,
  SECRET_TYPE_DOCKER_CONFIG = 2,
  SECRET_TYPE_DOCKER_CONFIG_JSON = 3,
  SECRET_TYPE_BASIC_AUTH = 4,
  SECRET_TYPE_SSH_AUTH = 5,
  SECRET_TYPE_TLS = 6,
  SECRET_TYPE_BOOTSTRAP_TOKEN = 7,
  UNRECOGNIZED = -1,
}

export function secretTypeFromJSON(object: any): SecretType {
  switch (object) {
    case 0:
    case "SECRET_TYPE_OPAQUE_UNSPECIFIED":
      return SecretType.SECRET_TYPE_OPAQUE_UNSPECIFIED;
    case 1:
    case "SECRET_TYPE_SERVICE_ACCOUNT_TOKEN":
      return SecretType.SECRET_TYPE_SERVICE_ACCOUNT_TOKEN;
    case 2:
    case "SECRET_TYPE_DOCKER_CONFIG":
      return SecretType.SECRET_TYPE_DOCKER_CONFIG;
    case 3:
    case "SECRET_TYPE_DOCKER_CONFIG_JSON":
      return SecretType.SECRET_TYPE_DOCKER_CONFIG_JSON;
    case 4:
    case "SECRET_TYPE_BASIC_AUTH":
      return SecretType.SECRET_TYPE_BASIC_AUTH;
    case 5:
    case "SECRET_TYPE_SSH_AUTH":
      return SecretType.SECRET_TYPE_SSH_AUTH;
    case 6:
    case "SECRET_TYPE_TLS":
      return SecretType.SECRET_TYPE_TLS;
    case 7:
    case "SECRET_TYPE_BOOTSTRAP_TOKEN":
      return SecretType.SECRET_TYPE_BOOTSTRAP_TOKEN;
    case -1:
    case "UNRECOGNIZED":
    default:
      return SecretType.UNRECOGNIZED;
  }
}

export function secretTypeToJSON(object: SecretType): string {
  switch (object) {
    case SecretType.SECRET_TYPE_OPAQUE_UNSPECIFIED:
      return "SECRET_TYPE_OPAQUE_UNSPECIFIED";
    case SecretType.SECRET_TYPE_SERVICE_ACCOUNT_TOKEN:
      return "SECRET_TYPE_SERVICE_ACCOUNT_TOKEN";
    case SecretType.SECRET_TYPE_DOCKER_CONFIG:
      return "SECRET_TYPE_DOCKER_CONFIG";
    case SecretType.SECRET_TYPE_DOCKER_CONFIG_JSON:
      return "SECRET_TYPE_DOCKER_CONFIG_JSON";
    case SecretType.SECRET_TYPE_BASIC_AUTH:
      return "SECRET_TYPE_BASIC_AUTH";
    case SecretType.SECRET_TYPE_SSH_AUTH:
      return "SECRET_TYPE_SSH_AUTH";
    case SecretType.SECRET_TYPE_TLS:
      return "SECRET_TYPE_TLS";
    case SecretType.SECRET_TYPE_BOOTSTRAP_TOKEN:
      return "SECRET_TYPE_BOOTSTRAP_TOKEN";
    default:
      return "UNKNOWN";
  }
}

/**
 * GetResourcesRequest
 *
 * Request for GetResources that specifies the resource references to get or watch.
 */
export interface GetResourcesRequest {
  /**
   * InstalledPackageRef
   *
   * The installed package reference for which the resources are being fetched.
   */
  installedPackageRef?: InstalledPackageReference;
  /**
   * ResourceRefs
   *
   * The references to the resources that are to be fetched or watched.
   * If empty, all resources for the installed package are returned when only
   * getting the resources. It must be populated to watch resources to avoid
   * watching all resources unnecessarily.
   */
  resourceRefs: ResourceRef[];
  /**
   * Watch
   *
   * When true, this will cause the stream to remain open with updated
   * resources being sent as events are received from the Kubernetes API
   * server.
   */
  watch: boolean;
}

export interface GetResourcesResponse {
  /**
   * ResourceRef
   *
   * The resource reference for this single resource.
   */
  resourceRef?: ResourceRef;
  /**
   * Manifest
   *
   * The current manifest of the requested resource.  Initially the JSON
   * manifest will be returned a json-encoded string, enabling the existing
   * Kubeapps UI to replace its current direct api-server getting and watching
   * of resources, but we may in the future pull out further structured
   * metadata into this message as needed.
   */
  manifest: string;
}

/**
 * GetServiceAccountNamesRequest
 *
 * Request for GetServiceAccountNames
 */
export interface GetServiceAccountNamesRequest {
  /**
   * Context
   *
   * The context for which the service account names are being fetched.
   */
  context?: Context;
}

/**
 * GetServiceAccountNamesResponse
 *
 * Response for GetServiceAccountNames
 */
export interface GetServiceAccountNamesResponse {
  /**
   * ServiceAccountNames
   *
   * The list of Service Account names.
   */
  serviceaccountNames: string[];
}

/**
 * GetNamespaceNamesRequest
 *
 * Request for GetNamespaceNames
 */
export interface GetNamespaceNamesRequest {
  /**
   * Cluster
   *
   * The context for which the namespace names are being fetched.  The service
   * will attempt to list namespaces across the cluster, first with the users
   * credential, then with a configured service account if available.
   */
  cluster: string;
}

/**
 * GetNamespaceNamesResponse
 *
 * Response for GetNamespaceNames
 */
export interface GetNamespaceNamesResponse {
  /**
   * NamespaceNames
   *
   * The list of Namespace names.
   */
  namespaceNames: string[];
}

/**
 * CreateNamespaceRequest
 *
 * Request for CreateNamespace
 */
export interface CreateNamespaceRequest {
  /**
   * Context
   *
   * The context of the namespace being created.
   */
  context?: Context;
}

/**
 * CreateNamespaceResponse
 *
 * Response for CreateNamespace
 */
export interface CreateNamespaceResponse {}

/**
 * CheckNamespaceExistsRequest
 *
 * Request for CheckNamespaceExists
 */
export interface CheckNamespaceExistsRequest {
  /**
   * Context
   *
   * The context of the namespace being checked for existence.
   */
  context?: Context;
}

/**
 * CheckNamespaceExistsResponse
 *
 * Response for CheckNamespaceExists
 */
export interface CheckNamespaceExistsResponse {
  exists: boolean;
}

/**
 * CreateSecretRequest
 *
 * Request for CreateSecret
 */
export interface CreateSecretRequest {
  /**
   * Context
   *
   * The context of the secret being created.
   */
  context?: Context;
  /**
   * Type
   *
   * The type of the secret. Valid values are defined by the Type enumeration.
   */
  type: SecretType;
  /**
   * Name
   *
   * The name of the secret.
   */
  name: string;
  /**
   * StringData
   *
   * The map of keys and values. Note that we use StringData here so that
   * Kubernetes handles the base64 encoding of the key values for us.
   * See https://kubernetes.io/docs/concepts/configuration/secret/#overview-of-secrets
   */
  stringData: { [key: string]: string };
}

export interface CreateSecretRequest_StringDataEntry {
  key: string;
  value: string;
}

/**
 * CreateSecretResponse
 *
 * Response for CreateSecret
 */
export interface CreateSecretResponse {}

/**
 * GetSecretNamesRequest
 *
 * Request for GetSecretNames
 */
export interface GetSecretNamesRequest {
  /**
   * Context
   *
   * The context for which the secret names are being fetched.
   */
  context?: Context;
}

/**
 * GetSecretNamesResponse
 *
 * Response for GetSecretNames
 */
export interface GetSecretNamesResponse {
  /**
   * SecretNames
   *
   * The list of Service Account names.
   */
  secretNames: { [key: string]: SecretType };
}

export interface GetSecretNamesResponse_SecretNamesEntry {
  key: string;
  value: SecretType;
}

const baseGetResourcesRequest: object = { watch: false };

export const GetResourcesRequest = {
  encode(message: GetResourcesRequest, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.installedPackageRef !== undefined) {
      InstalledPackageReference.encode(
        message.installedPackageRef,
        writer.uint32(10).fork(),
      ).ldelim();
    }
    for (const v of message.resourceRefs) {
      ResourceRef.encode(v!, writer.uint32(18).fork()).ldelim();
    }
    if (message.watch === true) {
      writer.uint32(24).bool(message.watch);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetResourcesRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseGetResourcesRequest } as GetResourcesRequest;
    message.resourceRefs = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.installedPackageRef = InstalledPackageReference.decode(reader, reader.uint32());
          break;
        case 2:
          message.resourceRefs.push(ResourceRef.decode(reader, reader.uint32()));
          break;
        case 3:
          message.watch = reader.bool();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): GetResourcesRequest {
    const message = { ...baseGetResourcesRequest } as GetResourcesRequest;
    message.resourceRefs = [];
    if (object.installedPackageRef !== undefined && object.installedPackageRef !== null) {
      message.installedPackageRef = InstalledPackageReference.fromJSON(object.installedPackageRef);
    } else {
      message.installedPackageRef = undefined;
    }
    if (object.resourceRefs !== undefined && object.resourceRefs !== null) {
      for (const e of object.resourceRefs) {
        message.resourceRefs.push(ResourceRef.fromJSON(e));
      }
    }
    if (object.watch !== undefined && object.watch !== null) {
      message.watch = Boolean(object.watch);
    } else {
      message.watch = false;
    }
    return message;
  },

  toJSON(message: GetResourcesRequest): unknown {
    const obj: any = {};
    message.installedPackageRef !== undefined &&
      (obj.installedPackageRef = message.installedPackageRef
        ? InstalledPackageReference.toJSON(message.installedPackageRef)
        : undefined);
    if (message.resourceRefs) {
      obj.resourceRefs = message.resourceRefs.map(e => (e ? ResourceRef.toJSON(e) : undefined));
    } else {
      obj.resourceRefs = [];
    }
    message.watch !== undefined && (obj.watch = message.watch);
    return obj;
  },

  fromPartial(object: DeepPartial<GetResourcesRequest>): GetResourcesRequest {
    const message = { ...baseGetResourcesRequest } as GetResourcesRequest;
    message.resourceRefs = [];
    if (object.installedPackageRef !== undefined && object.installedPackageRef !== null) {
      message.installedPackageRef = InstalledPackageReference.fromPartial(
        object.installedPackageRef,
      );
    } else {
      message.installedPackageRef = undefined;
    }
    if (object.resourceRefs !== undefined && object.resourceRefs !== null) {
      for (const e of object.resourceRefs) {
        message.resourceRefs.push(ResourceRef.fromPartial(e));
      }
    }
    if (object.watch !== undefined && object.watch !== null) {
      message.watch = object.watch;
    } else {
      message.watch = false;
    }
    return message;
  },
};

const baseGetResourcesResponse: object = { manifest: "" };

export const GetResourcesResponse = {
  encode(message: GetResourcesResponse, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.resourceRef !== undefined) {
      ResourceRef.encode(message.resourceRef, writer.uint32(10).fork()).ldelim();
    }
    if (message.manifest !== "") {
      writer.uint32(18).string(message.manifest);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetResourcesResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseGetResourcesResponse } as GetResourcesResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.resourceRef = ResourceRef.decode(reader, reader.uint32());
          break;
        case 2:
          message.manifest = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): GetResourcesResponse {
    const message = { ...baseGetResourcesResponse } as GetResourcesResponse;
    if (object.resourceRef !== undefined && object.resourceRef !== null) {
      message.resourceRef = ResourceRef.fromJSON(object.resourceRef);
    } else {
      message.resourceRef = undefined;
    }
    if (object.manifest !== undefined && object.manifest !== null) {
      message.manifest = String(object.manifest);
    } else {
      message.manifest = "";
    }
    return message;
  },

  toJSON(message: GetResourcesResponse): unknown {
    const obj: any = {};
    message.resourceRef !== undefined &&
      (obj.resourceRef = message.resourceRef ? ResourceRef.toJSON(message.resourceRef) : undefined);
    message.manifest !== undefined && (obj.manifest = message.manifest);
    return obj;
  },

  fromPartial(object: DeepPartial<GetResourcesResponse>): GetResourcesResponse {
    const message = { ...baseGetResourcesResponse } as GetResourcesResponse;
    if (object.resourceRef !== undefined && object.resourceRef !== null) {
      message.resourceRef = ResourceRef.fromPartial(object.resourceRef);
    } else {
      message.resourceRef = undefined;
    }
    if (object.manifest !== undefined && object.manifest !== null) {
      message.manifest = object.manifest;
    } else {
      message.manifest = "";
    }
    return message;
  },
};

const baseGetServiceAccountNamesRequest: object = {};

export const GetServiceAccountNamesRequest = {
  encode(
    message: GetServiceAccountNamesRequest,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.context !== undefined) {
      Context.encode(message.context, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetServiceAccountNamesRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseGetServiceAccountNamesRequest,
    } as GetServiceAccountNamesRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.context = Context.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): GetServiceAccountNamesRequest {
    const message = {
      ...baseGetServiceAccountNamesRequest,
    } as GetServiceAccountNamesRequest;
    if (object.context !== undefined && object.context !== null) {
      message.context = Context.fromJSON(object.context);
    } else {
      message.context = undefined;
    }
    return message;
  },

  toJSON(message: GetServiceAccountNamesRequest): unknown {
    const obj: any = {};
    message.context !== undefined &&
      (obj.context = message.context ? Context.toJSON(message.context) : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<GetServiceAccountNamesRequest>): GetServiceAccountNamesRequest {
    const message = {
      ...baseGetServiceAccountNamesRequest,
    } as GetServiceAccountNamesRequest;
    if (object.context !== undefined && object.context !== null) {
      message.context = Context.fromPartial(object.context);
    } else {
      message.context = undefined;
    }
    return message;
  },
};

const baseGetServiceAccountNamesResponse: object = { serviceaccountNames: "" };

export const GetServiceAccountNamesResponse = {
  encode(
    message: GetServiceAccountNamesResponse,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    for (const v of message.serviceaccountNames) {
      writer.uint32(10).string(v!);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetServiceAccountNamesResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseGetServiceAccountNamesResponse,
    } as GetServiceAccountNamesResponse;
    message.serviceaccountNames = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.serviceaccountNames.push(reader.string());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): GetServiceAccountNamesResponse {
    const message = {
      ...baseGetServiceAccountNamesResponse,
    } as GetServiceAccountNamesResponse;
    message.serviceaccountNames = [];
    if (object.serviceaccountNames !== undefined && object.serviceaccountNames !== null) {
      for (const e of object.serviceaccountNames) {
        message.serviceaccountNames.push(String(e));
      }
    }
    return message;
  },

  toJSON(message: GetServiceAccountNamesResponse): unknown {
    const obj: any = {};
    if (message.serviceaccountNames) {
      obj.serviceaccountNames = message.serviceaccountNames.map(e => e);
    } else {
      obj.serviceaccountNames = [];
    }
    return obj;
  },

  fromPartial(object: DeepPartial<GetServiceAccountNamesResponse>): GetServiceAccountNamesResponse {
    const message = {
      ...baseGetServiceAccountNamesResponse,
    } as GetServiceAccountNamesResponse;
    message.serviceaccountNames = [];
    if (object.serviceaccountNames !== undefined && object.serviceaccountNames !== null) {
      for (const e of object.serviceaccountNames) {
        message.serviceaccountNames.push(e);
      }
    }
    return message;
  },
};

const baseGetNamespaceNamesRequest: object = { cluster: "" };

export const GetNamespaceNamesRequest = {
  encode(message: GetNamespaceNamesRequest, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.cluster !== "") {
      writer.uint32(10).string(message.cluster);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetNamespaceNamesRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseGetNamespaceNamesRequest,
    } as GetNamespaceNamesRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.cluster = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): GetNamespaceNamesRequest {
    const message = {
      ...baseGetNamespaceNamesRequest,
    } as GetNamespaceNamesRequest;
    if (object.cluster !== undefined && object.cluster !== null) {
      message.cluster = String(object.cluster);
    } else {
      message.cluster = "";
    }
    return message;
  },

  toJSON(message: GetNamespaceNamesRequest): unknown {
    const obj: any = {};
    message.cluster !== undefined && (obj.cluster = message.cluster);
    return obj;
  },

  fromPartial(object: DeepPartial<GetNamespaceNamesRequest>): GetNamespaceNamesRequest {
    const message = {
      ...baseGetNamespaceNamesRequest,
    } as GetNamespaceNamesRequest;
    if (object.cluster !== undefined && object.cluster !== null) {
      message.cluster = object.cluster;
    } else {
      message.cluster = "";
    }
    return message;
  },
};

const baseGetNamespaceNamesResponse: object = { namespaceNames: "" };

export const GetNamespaceNamesResponse = {
  encode(message: GetNamespaceNamesResponse, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    for (const v of message.namespaceNames) {
      writer.uint32(10).string(v!);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetNamespaceNamesResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseGetNamespaceNamesResponse,
    } as GetNamespaceNamesResponse;
    message.namespaceNames = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.namespaceNames.push(reader.string());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): GetNamespaceNamesResponse {
    const message = {
      ...baseGetNamespaceNamesResponse,
    } as GetNamespaceNamesResponse;
    message.namespaceNames = [];
    if (object.namespaceNames !== undefined && object.namespaceNames !== null) {
      for (const e of object.namespaceNames) {
        message.namespaceNames.push(String(e));
      }
    }
    return message;
  },

  toJSON(message: GetNamespaceNamesResponse): unknown {
    const obj: any = {};
    if (message.namespaceNames) {
      obj.namespaceNames = message.namespaceNames.map(e => e);
    } else {
      obj.namespaceNames = [];
    }
    return obj;
  },

  fromPartial(object: DeepPartial<GetNamespaceNamesResponse>): GetNamespaceNamesResponse {
    const message = {
      ...baseGetNamespaceNamesResponse,
    } as GetNamespaceNamesResponse;
    message.namespaceNames = [];
    if (object.namespaceNames !== undefined && object.namespaceNames !== null) {
      for (const e of object.namespaceNames) {
        message.namespaceNames.push(e);
      }
    }
    return message;
  },
};

const baseCreateNamespaceRequest: object = {};

export const CreateNamespaceRequest = {
  encode(message: CreateNamespaceRequest, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.context !== undefined) {
      Context.encode(message.context, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): CreateNamespaceRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseCreateNamespaceRequest } as CreateNamespaceRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.context = Context.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): CreateNamespaceRequest {
    const message = { ...baseCreateNamespaceRequest } as CreateNamespaceRequest;
    if (object.context !== undefined && object.context !== null) {
      message.context = Context.fromJSON(object.context);
    } else {
      message.context = undefined;
    }
    return message;
  },

  toJSON(message: CreateNamespaceRequest): unknown {
    const obj: any = {};
    message.context !== undefined &&
      (obj.context = message.context ? Context.toJSON(message.context) : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<CreateNamespaceRequest>): CreateNamespaceRequest {
    const message = { ...baseCreateNamespaceRequest } as CreateNamespaceRequest;
    if (object.context !== undefined && object.context !== null) {
      message.context = Context.fromPartial(object.context);
    } else {
      message.context = undefined;
    }
    return message;
  },
};

const baseCreateNamespaceResponse: object = {};

export const CreateNamespaceResponse = {
  encode(_: CreateNamespaceResponse, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): CreateNamespaceResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseCreateNamespaceResponse,
    } as CreateNamespaceResponse;
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

  fromJSON(_: any): CreateNamespaceResponse {
    const message = {
      ...baseCreateNamespaceResponse,
    } as CreateNamespaceResponse;
    return message;
  },

  toJSON(_: CreateNamespaceResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(_: DeepPartial<CreateNamespaceResponse>): CreateNamespaceResponse {
    const message = {
      ...baseCreateNamespaceResponse,
    } as CreateNamespaceResponse;
    return message;
  },
};

const baseCheckNamespaceExistsRequest: object = {};

export const CheckNamespaceExistsRequest = {
  encode(
    message: CheckNamespaceExistsRequest,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.context !== undefined) {
      Context.encode(message.context, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): CheckNamespaceExistsRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseCheckNamespaceExistsRequest,
    } as CheckNamespaceExistsRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.context = Context.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): CheckNamespaceExistsRequest {
    const message = {
      ...baseCheckNamespaceExistsRequest,
    } as CheckNamespaceExistsRequest;
    if (object.context !== undefined && object.context !== null) {
      message.context = Context.fromJSON(object.context);
    } else {
      message.context = undefined;
    }
    return message;
  },

  toJSON(message: CheckNamespaceExistsRequest): unknown {
    const obj: any = {};
    message.context !== undefined &&
      (obj.context = message.context ? Context.toJSON(message.context) : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<CheckNamespaceExistsRequest>): CheckNamespaceExistsRequest {
    const message = {
      ...baseCheckNamespaceExistsRequest,
    } as CheckNamespaceExistsRequest;
    if (object.context !== undefined && object.context !== null) {
      message.context = Context.fromPartial(object.context);
    } else {
      message.context = undefined;
    }
    return message;
  },
};

const baseCheckNamespaceExistsResponse: object = { exists: false };

export const CheckNamespaceExistsResponse = {
  encode(
    message: CheckNamespaceExistsResponse,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.exists === true) {
      writer.uint32(8).bool(message.exists);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): CheckNamespaceExistsResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseCheckNamespaceExistsResponse,
    } as CheckNamespaceExistsResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.exists = reader.bool();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): CheckNamespaceExistsResponse {
    const message = {
      ...baseCheckNamespaceExistsResponse,
    } as CheckNamespaceExistsResponse;
    if (object.exists !== undefined && object.exists !== null) {
      message.exists = Boolean(object.exists);
    } else {
      message.exists = false;
    }
    return message;
  },

  toJSON(message: CheckNamespaceExistsResponse): unknown {
    const obj: any = {};
    message.exists !== undefined && (obj.exists = message.exists);
    return obj;
  },

  fromPartial(object: DeepPartial<CheckNamespaceExistsResponse>): CheckNamespaceExistsResponse {
    const message = {
      ...baseCheckNamespaceExistsResponse,
    } as CheckNamespaceExistsResponse;
    if (object.exists !== undefined && object.exists !== null) {
      message.exists = object.exists;
    } else {
      message.exists = false;
    }
    return message;
  },
};

const baseCreateSecretRequest: object = { type: 0, name: "" };

export const CreateSecretRequest = {
  encode(message: CreateSecretRequest, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.context !== undefined) {
      Context.encode(message.context, writer.uint32(10).fork()).ldelim();
    }
    if (message.type !== 0) {
      writer.uint32(16).int32(message.type);
    }
    if (message.name !== "") {
      writer.uint32(26).string(message.name);
    }
    Object.entries(message.stringData).forEach(([key, value]) => {
      CreateSecretRequest_StringDataEntry.encode(
        { key: key as any, value },
        writer.uint32(34).fork(),
      ).ldelim();
    });
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): CreateSecretRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseCreateSecretRequest } as CreateSecretRequest;
    message.stringData = {};
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.context = Context.decode(reader, reader.uint32());
          break;
        case 2:
          message.type = reader.int32() as any;
          break;
        case 3:
          message.name = reader.string();
          break;
        case 4:
          const entry4 = CreateSecretRequest_StringDataEntry.decode(reader, reader.uint32());
          if (entry4.value !== undefined) {
            message.stringData[entry4.key] = entry4.value;
          }
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): CreateSecretRequest {
    const message = { ...baseCreateSecretRequest } as CreateSecretRequest;
    message.stringData = {};
    if (object.context !== undefined && object.context !== null) {
      message.context = Context.fromJSON(object.context);
    } else {
      message.context = undefined;
    }
    if (object.type !== undefined && object.type !== null) {
      message.type = secretTypeFromJSON(object.type);
    } else {
      message.type = 0;
    }
    if (object.name !== undefined && object.name !== null) {
      message.name = String(object.name);
    } else {
      message.name = "";
    }
    if (object.stringData !== undefined && object.stringData !== null) {
      Object.entries(object.stringData).forEach(([key, value]) => {
        message.stringData[key] = String(value);
      });
    }
    return message;
  },

  toJSON(message: CreateSecretRequest): unknown {
    const obj: any = {};
    message.context !== undefined &&
      (obj.context = message.context ? Context.toJSON(message.context) : undefined);
    message.type !== undefined && (obj.type = secretTypeToJSON(message.type));
    message.name !== undefined && (obj.name = message.name);
    obj.stringData = {};
    if (message.stringData) {
      Object.entries(message.stringData).forEach(([k, v]) => {
        obj.stringData[k] = v;
      });
    }
    return obj;
  },

  fromPartial(object: DeepPartial<CreateSecretRequest>): CreateSecretRequest {
    const message = { ...baseCreateSecretRequest } as CreateSecretRequest;
    message.stringData = {};
    if (object.context !== undefined && object.context !== null) {
      message.context = Context.fromPartial(object.context);
    } else {
      message.context = undefined;
    }
    if (object.type !== undefined && object.type !== null) {
      message.type = object.type;
    } else {
      message.type = 0;
    }
    if (object.name !== undefined && object.name !== null) {
      message.name = object.name;
    } else {
      message.name = "";
    }
    if (object.stringData !== undefined && object.stringData !== null) {
      Object.entries(object.stringData).forEach(([key, value]) => {
        if (value !== undefined) {
          message.stringData[key] = String(value);
        }
      });
    }
    return message;
  },
};

const baseCreateSecretRequest_StringDataEntry: object = { key: "", value: "" };

export const CreateSecretRequest_StringDataEntry = {
  encode(
    message: CreateSecretRequest_StringDataEntry,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.key !== "") {
      writer.uint32(10).string(message.key);
    }
    if (message.value !== "") {
      writer.uint32(18).string(message.value);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): CreateSecretRequest_StringDataEntry {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseCreateSecretRequest_StringDataEntry,
    } as CreateSecretRequest_StringDataEntry;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.key = reader.string();
          break;
        case 2:
          message.value = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): CreateSecretRequest_StringDataEntry {
    const message = {
      ...baseCreateSecretRequest_StringDataEntry,
    } as CreateSecretRequest_StringDataEntry;
    if (object.key !== undefined && object.key !== null) {
      message.key = String(object.key);
    } else {
      message.key = "";
    }
    if (object.value !== undefined && object.value !== null) {
      message.value = String(object.value);
    } else {
      message.value = "";
    }
    return message;
  },

  toJSON(message: CreateSecretRequest_StringDataEntry): unknown {
    const obj: any = {};
    message.key !== undefined && (obj.key = message.key);
    message.value !== undefined && (obj.value = message.value);
    return obj;
  },

  fromPartial(
    object: DeepPartial<CreateSecretRequest_StringDataEntry>,
  ): CreateSecretRequest_StringDataEntry {
    const message = {
      ...baseCreateSecretRequest_StringDataEntry,
    } as CreateSecretRequest_StringDataEntry;
    if (object.key !== undefined && object.key !== null) {
      message.key = object.key;
    } else {
      message.key = "";
    }
    if (object.value !== undefined && object.value !== null) {
      message.value = object.value;
    } else {
      message.value = "";
    }
    return message;
  },
};

const baseCreateSecretResponse: object = {};

export const CreateSecretResponse = {
  encode(_: CreateSecretResponse, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): CreateSecretResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseCreateSecretResponse } as CreateSecretResponse;
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

  fromJSON(_: any): CreateSecretResponse {
    const message = { ...baseCreateSecretResponse } as CreateSecretResponse;
    return message;
  },

  toJSON(_: CreateSecretResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(_: DeepPartial<CreateSecretResponse>): CreateSecretResponse {
    const message = { ...baseCreateSecretResponse } as CreateSecretResponse;
    return message;
  },
};

const baseGetSecretNamesRequest: object = {};

export const GetSecretNamesRequest = {
  encode(message: GetSecretNamesRequest, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.context !== undefined) {
      Context.encode(message.context, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetSecretNamesRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseGetSecretNamesRequest } as GetSecretNamesRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.context = Context.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): GetSecretNamesRequest {
    const message = { ...baseGetSecretNamesRequest } as GetSecretNamesRequest;
    if (object.context !== undefined && object.context !== null) {
      message.context = Context.fromJSON(object.context);
    } else {
      message.context = undefined;
    }
    return message;
  },

  toJSON(message: GetSecretNamesRequest): unknown {
    const obj: any = {};
    message.context !== undefined &&
      (obj.context = message.context ? Context.toJSON(message.context) : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<GetSecretNamesRequest>): GetSecretNamesRequest {
    const message = { ...baseGetSecretNamesRequest } as GetSecretNamesRequest;
    if (object.context !== undefined && object.context !== null) {
      message.context = Context.fromPartial(object.context);
    } else {
      message.context = undefined;
    }
    return message;
  },
};

const baseGetSecretNamesResponse: object = {};

export const GetSecretNamesResponse = {
  encode(message: GetSecretNamesResponse, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    Object.entries(message.secretNames).forEach(([key, value]) => {
      GetSecretNamesResponse_SecretNamesEntry.encode(
        { key: key as any, value },
        writer.uint32(10).fork(),
      ).ldelim();
    });
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetSecretNamesResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseGetSecretNamesResponse } as GetSecretNamesResponse;
    message.secretNames = {};
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          const entry1 = GetSecretNamesResponse_SecretNamesEntry.decode(reader, reader.uint32());
          if (entry1.value !== undefined) {
            message.secretNames[entry1.key] = entry1.value;
          }
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): GetSecretNamesResponse {
    const message = { ...baseGetSecretNamesResponse } as GetSecretNamesResponse;
    message.secretNames = {};
    if (object.secretNames !== undefined && object.secretNames !== null) {
      Object.entries(object.secretNames).forEach(([key, value]) => {
        message.secretNames[key] = value as number;
      });
    }
    return message;
  },

  toJSON(message: GetSecretNamesResponse): unknown {
    const obj: any = {};
    obj.secretNames = {};
    if (message.secretNames) {
      Object.entries(message.secretNames).forEach(([k, v]) => {
        obj.secretNames[k] = secretTypeToJSON(v);
      });
    }
    return obj;
  },

  fromPartial(object: DeepPartial<GetSecretNamesResponse>): GetSecretNamesResponse {
    const message = { ...baseGetSecretNamesResponse } as GetSecretNamesResponse;
    message.secretNames = {};
    if (object.secretNames !== undefined && object.secretNames !== null) {
      Object.entries(object.secretNames).forEach(([key, value]) => {
        if (value !== undefined) {
          message.secretNames[key] = value as number;
        }
      });
    }
    return message;
  },
};

const baseGetSecretNamesResponse_SecretNamesEntry: object = {
  key: "",
  value: 0,
};

export const GetSecretNamesResponse_SecretNamesEntry = {
  encode(
    message: GetSecretNamesResponse_SecretNamesEntry,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.key !== "") {
      writer.uint32(10).string(message.key);
    }
    if (message.value !== 0) {
      writer.uint32(16).int32(message.value);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetSecretNamesResponse_SecretNamesEntry {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseGetSecretNamesResponse_SecretNamesEntry,
    } as GetSecretNamesResponse_SecretNamesEntry;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.key = reader.string();
          break;
        case 2:
          message.value = reader.int32() as any;
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): GetSecretNamesResponse_SecretNamesEntry {
    const message = {
      ...baseGetSecretNamesResponse_SecretNamesEntry,
    } as GetSecretNamesResponse_SecretNamesEntry;
    if (object.key !== undefined && object.key !== null) {
      message.key = String(object.key);
    } else {
      message.key = "";
    }
    if (object.value !== undefined && object.value !== null) {
      message.value = secretTypeFromJSON(object.value);
    } else {
      message.value = 0;
    }
    return message;
  },

  toJSON(message: GetSecretNamesResponse_SecretNamesEntry): unknown {
    const obj: any = {};
    message.key !== undefined && (obj.key = message.key);
    message.value !== undefined && (obj.value = secretTypeToJSON(message.value));
    return obj;
  },

  fromPartial(
    object: DeepPartial<GetSecretNamesResponse_SecretNamesEntry>,
  ): GetSecretNamesResponse_SecretNamesEntry {
    const message = {
      ...baseGetSecretNamesResponse_SecretNamesEntry,
    } as GetSecretNamesResponse_SecretNamesEntry;
    if (object.key !== undefined && object.key !== null) {
      message.key = object.key;
    } else {
      message.key = "";
    }
    if (object.value !== undefined && object.value !== null) {
      message.value = object.value;
    } else {
      message.value = 0;
    }
    return message;
  },
};

/**
 * ResourcesService
 *
 * The Resources service is a plugin that enables some limited access to Kubernetes
 * resources on the cluster, using the user credentials sent with the request.
 */
export interface ResourcesService {
  GetResources(
    request: DeepPartial<GetResourcesRequest>,
    metadata?: grpc.Metadata,
  ): Observable<GetResourcesResponse>;
  GetServiceAccountNames(
    request: DeepPartial<GetServiceAccountNamesRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetServiceAccountNamesResponse>;
  GetNamespaceNames(
    request: DeepPartial<GetNamespaceNamesRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetNamespaceNamesResponse>;
  CreateNamespace(
    request: DeepPartial<CreateNamespaceRequest>,
    metadata?: grpc.Metadata,
  ): Promise<CreateNamespaceResponse>;
  CheckNamespaceExists(
    request: DeepPartial<CheckNamespaceExistsRequest>,
    metadata?: grpc.Metadata,
  ): Promise<CheckNamespaceExistsResponse>;
  GetSecretNames(
    request: DeepPartial<GetSecretNamesRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetSecretNamesResponse>;
  CreateSecret(
    request: DeepPartial<CreateSecretRequest>,
    metadata?: grpc.Metadata,
  ): Promise<CreateSecretResponse>;
}

export class ResourcesServiceClientImpl implements ResourcesService {
  private readonly rpc: Rpc;

  constructor(rpc: Rpc) {
    this.rpc = rpc;
    this.GetResources = this.GetResources.bind(this);
    this.GetServiceAccountNames = this.GetServiceAccountNames.bind(this);
    this.GetNamespaceNames = this.GetNamespaceNames.bind(this);
    this.CreateNamespace = this.CreateNamespace.bind(this);
    this.CheckNamespaceExists = this.CheckNamespaceExists.bind(this);
    this.GetSecretNames = this.GetSecretNames.bind(this);
    this.CreateSecret = this.CreateSecret.bind(this);
  }

  GetResources(
    request: DeepPartial<GetResourcesRequest>,
    metadata?: grpc.Metadata,
  ): Observable<GetResourcesResponse> {
    return this.rpc.invoke(
      ResourcesServiceGetResourcesDesc,
      GetResourcesRequest.fromPartial(request),
      metadata,
    );
  }

  GetServiceAccountNames(
    request: DeepPartial<GetServiceAccountNamesRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetServiceAccountNamesResponse> {
    return this.rpc.unary(
      ResourcesServiceGetServiceAccountNamesDesc,
      GetServiceAccountNamesRequest.fromPartial(request),
      metadata,
    );
  }

  GetNamespaceNames(
    request: DeepPartial<GetNamespaceNamesRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetNamespaceNamesResponse> {
    return this.rpc.unary(
      ResourcesServiceGetNamespaceNamesDesc,
      GetNamespaceNamesRequest.fromPartial(request),
      metadata,
    );
  }

  CreateNamespace(
    request: DeepPartial<CreateNamespaceRequest>,
    metadata?: grpc.Metadata,
  ): Promise<CreateNamespaceResponse> {
    return this.rpc.unary(
      ResourcesServiceCreateNamespaceDesc,
      CreateNamespaceRequest.fromPartial(request),
      metadata,
    );
  }

  CheckNamespaceExists(
    request: DeepPartial<CheckNamespaceExistsRequest>,
    metadata?: grpc.Metadata,
  ): Promise<CheckNamespaceExistsResponse> {
    return this.rpc.unary(
      ResourcesServiceCheckNamespaceExistsDesc,
      CheckNamespaceExistsRequest.fromPartial(request),
      metadata,
    );
  }

  GetSecretNames(
    request: DeepPartial<GetSecretNamesRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetSecretNamesResponse> {
    return this.rpc.unary(
      ResourcesServiceGetSecretNamesDesc,
      GetSecretNamesRequest.fromPartial(request),
      metadata,
    );
  }

  CreateSecret(
    request: DeepPartial<CreateSecretRequest>,
    metadata?: grpc.Metadata,
  ): Promise<CreateSecretResponse> {
    return this.rpc.unary(
      ResourcesServiceCreateSecretDesc,
      CreateSecretRequest.fromPartial(request),
      metadata,
    );
  }
}

export const ResourcesServiceDesc = {
  serviceName: "kubeappsapis.plugins.resources.v1alpha1.ResourcesService",
};

export const ResourcesServiceGetResourcesDesc: UnaryMethodDefinitionish = {
  methodName: "GetResources",
  service: ResourcesServiceDesc,
  requestStream: false,
  responseStream: true,
  requestType: {
    serializeBinary() {
      return GetResourcesRequest.encode(this).finish();
    },
  } as any,
  responseType: {
    deserializeBinary(data: Uint8Array) {
      return {
        ...GetResourcesResponse.decode(data),
        toObject() {
          return this;
        },
      };
    },
  } as any,
};

export const ResourcesServiceGetServiceAccountNamesDesc: UnaryMethodDefinitionish = {
  methodName: "GetServiceAccountNames",
  service: ResourcesServiceDesc,
  requestStream: false,
  responseStream: false,
  requestType: {
    serializeBinary() {
      return GetServiceAccountNamesRequest.encode(this).finish();
    },
  } as any,
  responseType: {
    deserializeBinary(data: Uint8Array) {
      return {
        ...GetServiceAccountNamesResponse.decode(data),
        toObject() {
          return this;
        },
      };
    },
  } as any,
};

export const ResourcesServiceGetNamespaceNamesDesc: UnaryMethodDefinitionish = {
  methodName: "GetNamespaceNames",
  service: ResourcesServiceDesc,
  requestStream: false,
  responseStream: false,
  requestType: {
    serializeBinary() {
      return GetNamespaceNamesRequest.encode(this).finish();
    },
  } as any,
  responseType: {
    deserializeBinary(data: Uint8Array) {
      return {
        ...GetNamespaceNamesResponse.decode(data),
        toObject() {
          return this;
        },
      };
    },
  } as any,
};

export const ResourcesServiceCreateNamespaceDesc: UnaryMethodDefinitionish = {
  methodName: "CreateNamespace",
  service: ResourcesServiceDesc,
  requestStream: false,
  responseStream: false,
  requestType: {
    serializeBinary() {
      return CreateNamespaceRequest.encode(this).finish();
    },
  } as any,
  responseType: {
    deserializeBinary(data: Uint8Array) {
      return {
        ...CreateNamespaceResponse.decode(data),
        toObject() {
          return this;
        },
      };
    },
  } as any,
};

export const ResourcesServiceCheckNamespaceExistsDesc: UnaryMethodDefinitionish = {
  methodName: "CheckNamespaceExists",
  service: ResourcesServiceDesc,
  requestStream: false,
  responseStream: false,
  requestType: {
    serializeBinary() {
      return CheckNamespaceExistsRequest.encode(this).finish();
    },
  } as any,
  responseType: {
    deserializeBinary(data: Uint8Array) {
      return {
        ...CheckNamespaceExistsResponse.decode(data),
        toObject() {
          return this;
        },
      };
    },
  } as any,
};

export const ResourcesServiceGetSecretNamesDesc: UnaryMethodDefinitionish = {
  methodName: "GetSecretNames",
  service: ResourcesServiceDesc,
  requestStream: false,
  responseStream: false,
  requestType: {
    serializeBinary() {
      return GetSecretNamesRequest.encode(this).finish();
    },
  } as any,
  responseType: {
    deserializeBinary(data: Uint8Array) {
      return {
        ...GetSecretNamesResponse.decode(data),
        toObject() {
          return this;
        },
      };
    },
  } as any,
};

export const ResourcesServiceCreateSecretDesc: UnaryMethodDefinitionish = {
  methodName: "CreateSecret",
  service: ResourcesServiceDesc,
  requestStream: false,
  responseStream: false,
  requestType: {
    serializeBinary() {
      return CreateSecretRequest.encode(this).finish();
    },
  } as any,
  responseType: {
    deserializeBinary(data: Uint8Array) {
      return {
        ...CreateSecretResponse.decode(data),
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
  invoke<T extends UnaryMethodDefinitionish>(
    methodDesc: T,
    request: any,
    metadata: grpc.Metadata | undefined,
  ): Observable<any>;
}

export class GrpcWebImpl {
  private host: string;
  private options: {
    transport?: grpc.TransportFactory;
    streamingTransport?: grpc.TransportFactory;
    debug?: boolean;
    metadata?: grpc.Metadata;
  };

  constructor(
    host: string,
    options: {
      transport?: grpc.TransportFactory;
      streamingTransport?: grpc.TransportFactory;
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

  invoke<T extends UnaryMethodDefinitionish>(
    methodDesc: T,
    _request: any,
    metadata: grpc.Metadata | undefined,
  ): Observable<any> {
    // Status Response Codes (https://developers.google.com/maps-booking/reference/grpc-api/status_codes)
    const upStreamCodes = [2, 4, 8, 9, 10, 13, 14, 15];
    const DEFAULT_TIMEOUT_TIME: number = 3_000;
    const request = { ..._request, ...methodDesc.requestType };
    const maybeCombinedMetadata =
      metadata && this.options.metadata
        ? new BrowserHeaders({
            ...this.options?.metadata.headersMap,
            ...metadata?.headersMap,
          })
        : metadata || this.options.metadata;
    return new Observable(observer => {
      const upStream = () => {
        const client = grpc.invoke(methodDesc, {
          host: this.host,
          request,
          transport: this.options.streamingTransport || this.options.transport,
          metadata: maybeCombinedMetadata,
          debug: this.options.debug,
          onMessage: next => observer.next(next),
          onEnd: (code: grpc.Code, message: string) => {
            if (code === 0) {
              observer.complete();
            } else if (upStreamCodes.includes(code)) {
              setTimeout(upStream, DEFAULT_TIMEOUT_TIME);
            } else {
              observer.error(new Error(`Error ${code} ${message}`));
            }
          },
        });
        observer.add(() => client.close());
      };
      upStream();
    }).pipe(share());
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
