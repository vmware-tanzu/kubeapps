/* eslint-disable */
import { grpc } from "@improbable-eng/grpc-web";
import { BrowserHeaders } from "browser-headers";
import _m0 from "protobufjs/minimal";
import { Observable } from "rxjs";
import { share } from "rxjs/operators";
import {
  Context,
  InstalledPackageReference,
  ResourceRef,
} from "../../../core/packages/v1alpha1/packages";

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
    case SecretType.UNRECOGNIZED:
    default:
      return "UNRECOGNIZED";
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
  /**
   * Labels
   *
   * The additional labels added to the namespace at creation time
   */
  labels: { [key: string]: string };
}

export interface CreateNamespaceRequest_LabelsEntry {
  key: string;
  value: string;
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

/**
 * CanIRequest
 *
 * Request for CanI operation
 */
export interface CanIRequest {
  /**
   * The context (cluster/namespace) for the can-i request
   * "" (empty) namespace means "all"
   */
  context?: Context;
  /**
   * Group API Group of the Resource.  "*" means all.
   * +optional
   */
  group: string;
  /**
   * Resource is one of the existing resource types.  "*" means all.
   * +optional
   */
  resource: string;
  /**
   * Verb is a kubernetes resource API verb, like: get, list, watch, create, update, delete, proxy.  "*" means all.
   * +optional
   */
  verb: string;
}

/**
 * CanIResponse
 *
 * Response for CanI operation
 */
export interface CanIResponse {
  /**
   * allowed
   *
   * True if operation is allowed
   */
  allowed: boolean;
}

function createBaseGetResourcesRequest(): GetResourcesRequest {
  return { installedPackageRef: undefined, resourceRefs: [], watch: false };
}

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
    const message = createBaseGetResourcesRequest();
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
    return {
      installedPackageRef: isSet(object.installedPackageRef)
        ? InstalledPackageReference.fromJSON(object.installedPackageRef)
        : undefined,
      resourceRefs: Array.isArray(object?.resourceRefs)
        ? object.resourceRefs.map((e: any) => ResourceRef.fromJSON(e))
        : [],
      watch: isSet(object.watch) ? Boolean(object.watch) : false,
    };
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

  fromPartial<I extends Exact<DeepPartial<GetResourcesRequest>, I>>(
    object: I,
  ): GetResourcesRequest {
    const message = createBaseGetResourcesRequest();
    message.installedPackageRef =
      object.installedPackageRef !== undefined && object.installedPackageRef !== null
        ? InstalledPackageReference.fromPartial(object.installedPackageRef)
        : undefined;
    message.resourceRefs = object.resourceRefs?.map(e => ResourceRef.fromPartial(e)) || [];
    message.watch = object.watch ?? false;
    return message;
  },
};

function createBaseGetResourcesResponse(): GetResourcesResponse {
  return { resourceRef: undefined, manifest: "" };
}

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
    const message = createBaseGetResourcesResponse();
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
    return {
      resourceRef: isSet(object.resourceRef) ? ResourceRef.fromJSON(object.resourceRef) : undefined,
      manifest: isSet(object.manifest) ? String(object.manifest) : "",
    };
  },

  toJSON(message: GetResourcesResponse): unknown {
    const obj: any = {};
    message.resourceRef !== undefined &&
      (obj.resourceRef = message.resourceRef ? ResourceRef.toJSON(message.resourceRef) : undefined);
    message.manifest !== undefined && (obj.manifest = message.manifest);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<GetResourcesResponse>, I>>(
    object: I,
  ): GetResourcesResponse {
    const message = createBaseGetResourcesResponse();
    message.resourceRef =
      object.resourceRef !== undefined && object.resourceRef !== null
        ? ResourceRef.fromPartial(object.resourceRef)
        : undefined;
    message.manifest = object.manifest ?? "";
    return message;
  },
};

function createBaseGetServiceAccountNamesRequest(): GetServiceAccountNamesRequest {
  return { context: undefined };
}

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
    const message = createBaseGetServiceAccountNamesRequest();
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
    return { context: isSet(object.context) ? Context.fromJSON(object.context) : undefined };
  },

  toJSON(message: GetServiceAccountNamesRequest): unknown {
    const obj: any = {};
    message.context !== undefined &&
      (obj.context = message.context ? Context.toJSON(message.context) : undefined);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<GetServiceAccountNamesRequest>, I>>(
    object: I,
  ): GetServiceAccountNamesRequest {
    const message = createBaseGetServiceAccountNamesRequest();
    message.context =
      object.context !== undefined && object.context !== null
        ? Context.fromPartial(object.context)
        : undefined;
    return message;
  },
};

function createBaseGetServiceAccountNamesResponse(): GetServiceAccountNamesResponse {
  return { serviceaccountNames: [] };
}

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
    const message = createBaseGetServiceAccountNamesResponse();
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
    return {
      serviceaccountNames: Array.isArray(object?.serviceaccountNames)
        ? object.serviceaccountNames.map((e: any) => String(e))
        : [],
    };
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

  fromPartial<I extends Exact<DeepPartial<GetServiceAccountNamesResponse>, I>>(
    object: I,
  ): GetServiceAccountNamesResponse {
    const message = createBaseGetServiceAccountNamesResponse();
    message.serviceaccountNames = object.serviceaccountNames?.map(e => e) || [];
    return message;
  },
};

function createBaseGetNamespaceNamesRequest(): GetNamespaceNamesRequest {
  return { cluster: "" };
}

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
    const message = createBaseGetNamespaceNamesRequest();
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
    return { cluster: isSet(object.cluster) ? String(object.cluster) : "" };
  },

  toJSON(message: GetNamespaceNamesRequest): unknown {
    const obj: any = {};
    message.cluster !== undefined && (obj.cluster = message.cluster);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<GetNamespaceNamesRequest>, I>>(
    object: I,
  ): GetNamespaceNamesRequest {
    const message = createBaseGetNamespaceNamesRequest();
    message.cluster = object.cluster ?? "";
    return message;
  },
};

function createBaseGetNamespaceNamesResponse(): GetNamespaceNamesResponse {
  return { namespaceNames: [] };
}

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
    const message = createBaseGetNamespaceNamesResponse();
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
    return {
      namespaceNames: Array.isArray(object?.namespaceNames)
        ? object.namespaceNames.map((e: any) => String(e))
        : [],
    };
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

  fromPartial<I extends Exact<DeepPartial<GetNamespaceNamesResponse>, I>>(
    object: I,
  ): GetNamespaceNamesResponse {
    const message = createBaseGetNamespaceNamesResponse();
    message.namespaceNames = object.namespaceNames?.map(e => e) || [];
    return message;
  },
};

function createBaseCreateNamespaceRequest(): CreateNamespaceRequest {
  return { context: undefined, labels: {} };
}

export const CreateNamespaceRequest = {
  encode(message: CreateNamespaceRequest, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.context !== undefined) {
      Context.encode(message.context, writer.uint32(10).fork()).ldelim();
    }
    Object.entries(message.labels).forEach(([key, value]) => {
      CreateNamespaceRequest_LabelsEntry.encode(
        { key: key as any, value },
        writer.uint32(18).fork(),
      ).ldelim();
    });
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): CreateNamespaceRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseCreateNamespaceRequest();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.context = Context.decode(reader, reader.uint32());
          break;
        case 2:
          const entry2 = CreateNamespaceRequest_LabelsEntry.decode(reader, reader.uint32());
          if (entry2.value !== undefined) {
            message.labels[entry2.key] = entry2.value;
          }
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): CreateNamespaceRequest {
    return {
      context: isSet(object.context) ? Context.fromJSON(object.context) : undefined,
      labels: isObject(object.labels)
        ? Object.entries(object.labels).reduce<{ [key: string]: string }>((acc, [key, value]) => {
            acc[key] = String(value);
            return acc;
          }, {})
        : {},
    };
  },

  toJSON(message: CreateNamespaceRequest): unknown {
    const obj: any = {};
    message.context !== undefined &&
      (obj.context = message.context ? Context.toJSON(message.context) : undefined);
    obj.labels = {};
    if (message.labels) {
      Object.entries(message.labels).forEach(([k, v]) => {
        obj.labels[k] = v;
      });
    }
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<CreateNamespaceRequest>, I>>(
    object: I,
  ): CreateNamespaceRequest {
    const message = createBaseCreateNamespaceRequest();
    message.context =
      object.context !== undefined && object.context !== null
        ? Context.fromPartial(object.context)
        : undefined;
    message.labels = Object.entries(object.labels ?? {}).reduce<{ [key: string]: string }>(
      (acc, [key, value]) => {
        if (value !== undefined) {
          acc[key] = String(value);
        }
        return acc;
      },
      {},
    );
    return message;
  },
};

function createBaseCreateNamespaceRequest_LabelsEntry(): CreateNamespaceRequest_LabelsEntry {
  return { key: "", value: "" };
}

export const CreateNamespaceRequest_LabelsEntry = {
  encode(
    message: CreateNamespaceRequest_LabelsEntry,
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

  decode(input: _m0.Reader | Uint8Array, length?: number): CreateNamespaceRequest_LabelsEntry {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseCreateNamespaceRequest_LabelsEntry();
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

  fromJSON(object: any): CreateNamespaceRequest_LabelsEntry {
    return {
      key: isSet(object.key) ? String(object.key) : "",
      value: isSet(object.value) ? String(object.value) : "",
    };
  },

  toJSON(message: CreateNamespaceRequest_LabelsEntry): unknown {
    const obj: any = {};
    message.key !== undefined && (obj.key = message.key);
    message.value !== undefined && (obj.value = message.value);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<CreateNamespaceRequest_LabelsEntry>, I>>(
    object: I,
  ): CreateNamespaceRequest_LabelsEntry {
    const message = createBaseCreateNamespaceRequest_LabelsEntry();
    message.key = object.key ?? "";
    message.value = object.value ?? "";
    return message;
  },
};

function createBaseCreateNamespaceResponse(): CreateNamespaceResponse {
  return {};
}

export const CreateNamespaceResponse = {
  encode(_: CreateNamespaceResponse, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): CreateNamespaceResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseCreateNamespaceResponse();
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
    return {};
  },

  toJSON(_: CreateNamespaceResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<CreateNamespaceResponse>, I>>(
    _: I,
  ): CreateNamespaceResponse {
    const message = createBaseCreateNamespaceResponse();
    return message;
  },
};

function createBaseCheckNamespaceExistsRequest(): CheckNamespaceExistsRequest {
  return { context: undefined };
}

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
    const message = createBaseCheckNamespaceExistsRequest();
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
    return { context: isSet(object.context) ? Context.fromJSON(object.context) : undefined };
  },

  toJSON(message: CheckNamespaceExistsRequest): unknown {
    const obj: any = {};
    message.context !== undefined &&
      (obj.context = message.context ? Context.toJSON(message.context) : undefined);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<CheckNamespaceExistsRequest>, I>>(
    object: I,
  ): CheckNamespaceExistsRequest {
    const message = createBaseCheckNamespaceExistsRequest();
    message.context =
      object.context !== undefined && object.context !== null
        ? Context.fromPartial(object.context)
        : undefined;
    return message;
  },
};

function createBaseCheckNamespaceExistsResponse(): CheckNamespaceExistsResponse {
  return { exists: false };
}

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
    const message = createBaseCheckNamespaceExistsResponse();
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
    return { exists: isSet(object.exists) ? Boolean(object.exists) : false };
  },

  toJSON(message: CheckNamespaceExistsResponse): unknown {
    const obj: any = {};
    message.exists !== undefined && (obj.exists = message.exists);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<CheckNamespaceExistsResponse>, I>>(
    object: I,
  ): CheckNamespaceExistsResponse {
    const message = createBaseCheckNamespaceExistsResponse();
    message.exists = object.exists ?? false;
    return message;
  },
};

function createBaseCreateSecretRequest(): CreateSecretRequest {
  return { context: undefined, type: 0, name: "", stringData: {} };
}

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
    const message = createBaseCreateSecretRequest();
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
    return {
      context: isSet(object.context) ? Context.fromJSON(object.context) : undefined,
      type: isSet(object.type) ? secretTypeFromJSON(object.type) : 0,
      name: isSet(object.name) ? String(object.name) : "",
      stringData: isObject(object.stringData)
        ? Object.entries(object.stringData).reduce<{ [key: string]: string }>(
            (acc, [key, value]) => {
              acc[key] = String(value);
              return acc;
            },
            {},
          )
        : {},
    };
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

  fromPartial<I extends Exact<DeepPartial<CreateSecretRequest>, I>>(
    object: I,
  ): CreateSecretRequest {
    const message = createBaseCreateSecretRequest();
    message.context =
      object.context !== undefined && object.context !== null
        ? Context.fromPartial(object.context)
        : undefined;
    message.type = object.type ?? 0;
    message.name = object.name ?? "";
    message.stringData = Object.entries(object.stringData ?? {}).reduce<{ [key: string]: string }>(
      (acc, [key, value]) => {
        if (value !== undefined) {
          acc[key] = String(value);
        }
        return acc;
      },
      {},
    );
    return message;
  },
};

function createBaseCreateSecretRequest_StringDataEntry(): CreateSecretRequest_StringDataEntry {
  return { key: "", value: "" };
}

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
    const message = createBaseCreateSecretRequest_StringDataEntry();
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
    return {
      key: isSet(object.key) ? String(object.key) : "",
      value: isSet(object.value) ? String(object.value) : "",
    };
  },

  toJSON(message: CreateSecretRequest_StringDataEntry): unknown {
    const obj: any = {};
    message.key !== undefined && (obj.key = message.key);
    message.value !== undefined && (obj.value = message.value);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<CreateSecretRequest_StringDataEntry>, I>>(
    object: I,
  ): CreateSecretRequest_StringDataEntry {
    const message = createBaseCreateSecretRequest_StringDataEntry();
    message.key = object.key ?? "";
    message.value = object.value ?? "";
    return message;
  },
};

function createBaseCreateSecretResponse(): CreateSecretResponse {
  return {};
}

export const CreateSecretResponse = {
  encode(_: CreateSecretResponse, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): CreateSecretResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseCreateSecretResponse();
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
    return {};
  },

  toJSON(_: CreateSecretResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<CreateSecretResponse>, I>>(_: I): CreateSecretResponse {
    const message = createBaseCreateSecretResponse();
    return message;
  },
};

function createBaseGetSecretNamesRequest(): GetSecretNamesRequest {
  return { context: undefined };
}

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
    const message = createBaseGetSecretNamesRequest();
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
    return { context: isSet(object.context) ? Context.fromJSON(object.context) : undefined };
  },

  toJSON(message: GetSecretNamesRequest): unknown {
    const obj: any = {};
    message.context !== undefined &&
      (obj.context = message.context ? Context.toJSON(message.context) : undefined);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<GetSecretNamesRequest>, I>>(
    object: I,
  ): GetSecretNamesRequest {
    const message = createBaseGetSecretNamesRequest();
    message.context =
      object.context !== undefined && object.context !== null
        ? Context.fromPartial(object.context)
        : undefined;
    return message;
  },
};

function createBaseGetSecretNamesResponse(): GetSecretNamesResponse {
  return { secretNames: {} };
}

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
    const message = createBaseGetSecretNamesResponse();
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
    return {
      secretNames: isObject(object.secretNames)
        ? Object.entries(object.secretNames).reduce<{ [key: string]: SecretType }>(
            (acc, [key, value]) => {
              acc[key] = secretTypeFromJSON(value);
              return acc;
            },
            {},
          )
        : {},
    };
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

  fromPartial<I extends Exact<DeepPartial<GetSecretNamesResponse>, I>>(
    object: I,
  ): GetSecretNamesResponse {
    const message = createBaseGetSecretNamesResponse();
    message.secretNames = Object.entries(object.secretNames ?? {}).reduce<{
      [key: string]: SecretType;
    }>((acc, [key, value]) => {
      if (value !== undefined) {
        acc[key] = value as SecretType;
      }
      return acc;
    }, {});
    return message;
  },
};

function createBaseGetSecretNamesResponse_SecretNamesEntry(): GetSecretNamesResponse_SecretNamesEntry {
  return { key: "", value: 0 };
}

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
    const message = createBaseGetSecretNamesResponse_SecretNamesEntry();
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
    return {
      key: isSet(object.key) ? String(object.key) : "",
      value: isSet(object.value) ? secretTypeFromJSON(object.value) : 0,
    };
  },

  toJSON(message: GetSecretNamesResponse_SecretNamesEntry): unknown {
    const obj: any = {};
    message.key !== undefined && (obj.key = message.key);
    message.value !== undefined && (obj.value = secretTypeToJSON(message.value));
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<GetSecretNamesResponse_SecretNamesEntry>, I>>(
    object: I,
  ): GetSecretNamesResponse_SecretNamesEntry {
    const message = createBaseGetSecretNamesResponse_SecretNamesEntry();
    message.key = object.key ?? "";
    message.value = object.value ?? 0;
    return message;
  },
};

function createBaseCanIRequest(): CanIRequest {
  return { context: undefined, group: "", resource: "", verb: "" };
}

export const CanIRequest = {
  encode(message: CanIRequest, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.context !== undefined) {
      Context.encode(message.context, writer.uint32(10).fork()).ldelim();
    }
    if (message.group !== "") {
      writer.uint32(18).string(message.group);
    }
    if (message.resource !== "") {
      writer.uint32(26).string(message.resource);
    }
    if (message.verb !== "") {
      writer.uint32(34).string(message.verb);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): CanIRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseCanIRequest();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.context = Context.decode(reader, reader.uint32());
          break;
        case 2:
          message.group = reader.string();
          break;
        case 3:
          message.resource = reader.string();
          break;
        case 4:
          message.verb = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): CanIRequest {
    return {
      context: isSet(object.context) ? Context.fromJSON(object.context) : undefined,
      group: isSet(object.group) ? String(object.group) : "",
      resource: isSet(object.resource) ? String(object.resource) : "",
      verb: isSet(object.verb) ? String(object.verb) : "",
    };
  },

  toJSON(message: CanIRequest): unknown {
    const obj: any = {};
    message.context !== undefined &&
      (obj.context = message.context ? Context.toJSON(message.context) : undefined);
    message.group !== undefined && (obj.group = message.group);
    message.resource !== undefined && (obj.resource = message.resource);
    message.verb !== undefined && (obj.verb = message.verb);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<CanIRequest>, I>>(object: I): CanIRequest {
    const message = createBaseCanIRequest();
    message.context =
      object.context !== undefined && object.context !== null
        ? Context.fromPartial(object.context)
        : undefined;
    message.group = object.group ?? "";
    message.resource = object.resource ?? "";
    message.verb = object.verb ?? "";
    return message;
  },
};

function createBaseCanIResponse(): CanIResponse {
  return { allowed: false };
}

export const CanIResponse = {
  encode(message: CanIResponse, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.allowed === true) {
      writer.uint32(8).bool(message.allowed);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): CanIResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseCanIResponse();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.allowed = reader.bool();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): CanIResponse {
    return { allowed: isSet(object.allowed) ? Boolean(object.allowed) : false };
  },

  toJSON(message: CanIResponse): unknown {
    const obj: any = {};
    message.allowed !== undefined && (obj.allowed = message.allowed);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<CanIResponse>, I>>(object: I): CanIResponse {
    const message = createBaseCanIResponse();
    message.allowed = object.allowed ?? false;
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
  CanI(request: DeepPartial<CanIRequest>, metadata?: grpc.Metadata): Promise<CanIResponse>;
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
    this.CanI = this.CanI.bind(this);
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

  CanI(request: DeepPartial<CanIRequest>, metadata?: grpc.Metadata): Promise<CanIResponse> {
    return this.rpc.unary(ResourcesServiceCanIDesc, CanIRequest.fromPartial(request), metadata);
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

export const ResourcesServiceCanIDesc: UnaryMethodDefinitionish = {
  methodName: "CanI",
  service: ResourcesServiceDesc,
  requestStream: false,
  responseStream: false,
  requestType: {
    serializeBinary() {
      return CanIRequest.encode(this).finish();
    },
  } as any,
  responseType: {
    deserializeBinary(data: Uint8Array) {
      return {
        ...CanIResponse.decode(data),
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
    upStreamRetryCodes?: number[];
  };

  constructor(
    host: string,
    options: {
      transport?: grpc.TransportFactory;
      streamingTransport?: grpc.TransportFactory;
      debug?: boolean;
      metadata?: grpc.Metadata;
      upStreamRetryCodes?: number[];
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
        ? new BrowserHeaders({ ...this.options?.metadata.headersMap, ...metadata?.headersMap })
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
            const err = new GrpcWebError(
              response.statusMessage,
              response.status,
              response.trailers,
            );
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
    const upStreamCodes = this.options.upStreamRetryCodes || [];
    const DEFAULT_TIMEOUT_TIME: number = 3_000;
    const request = { ..._request, ...methodDesc.requestType };
    const maybeCombinedMetadata =
      metadata && this.options.metadata
        ? new BrowserHeaders({ ...this.options?.metadata.headersMap, ...metadata?.headersMap })
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
          onEnd: (code: grpc.Code, message: string, trailers: grpc.Metadata) => {
            if (code === 0) {
              observer.complete();
            } else if (upStreamCodes.includes(code)) {
              setTimeout(upStream, DEFAULT_TIMEOUT_TIME);
            } else {
              const err = new Error(message) as any;
              err.code = code;
              err.metadata = trailers;
              observer.error(err);
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

type KeysOfUnion<T> = T extends T ? keyof T : never;
export type Exact<P, I extends P> = P extends Builtin
  ? P
  : P & { [K in keyof P]: Exact<P[K], I[K]> } & { [K in Exclude<keyof I, KeysOfUnion<P>>]: never };

function isObject(value: any): boolean {
  return typeof value === "object" && value !== null;
}

function isSet(value: any): boolean {
  return value !== null && value !== undefined;
}

export class GrpcWebError extends Error {
  constructor(message: string, public code: grpc.Code, public metadata: grpc.Metadata) {
    super(message);
  }
}
