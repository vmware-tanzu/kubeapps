/* eslint-disable */
import Long from "long";
import { grpc } from "@improbable-eng/grpc-web";
import _m0 from "protobufjs/minimal";
import {
  InstalledPackageReference,
  ResourceRef,
  Context,
} from "../../../../kubeappsapis/core/packages/v1alpha1/packages";
import { Any } from "../../../../google/protobuf/any";
import { Observable } from "rxjs";
import { BrowserHeaders } from "browser-headers";
import { share } from "rxjs/operators";

export const protobufPackage = "kubeappsapis.plugins.resources.v1alpha1";

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
   * The current manifest of the requested resource.
   * Initially the JSON manifest will be returned as an Any, enabling the
   * existing Kubeapps UI to replace its current direct api-server getting and
   * watching of resources, but we may in the future pull out further
   * structured metadata into this message as needed.
   */
  manifest?: Any;
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
export interface CheckNamespaceExistsResponse {}

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
    message.installedPackageRef =
      object.installedPackageRef !== undefined && object.installedPackageRef !== null
        ? InstalledPackageReference.fromJSON(object.installedPackageRef)
        : undefined;
    message.resourceRefs = (object.resourceRefs ?? []).map((e: any) => ResourceRef.fromJSON(e));
    message.watch =
      object.watch !== undefined && object.watch !== null ? Boolean(object.watch) : false;
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

  fromPartial<I extends Exact<DeepPartial<GetResourcesRequest>, I>>(
    object: I,
  ): GetResourcesRequest {
    const message = { ...baseGetResourcesRequest } as GetResourcesRequest;
    message.installedPackageRef =
      object.installedPackageRef !== undefined && object.installedPackageRef !== null
        ? InstalledPackageReference.fromPartial(object.installedPackageRef)
        : undefined;
    message.resourceRefs = object.resourceRefs?.map(e => ResourceRef.fromPartial(e)) || [];
    message.watch = object.watch ?? false;
    return message;
  },
};

const baseGetResourcesResponse: object = {};

export const GetResourcesResponse = {
  encode(message: GetResourcesResponse, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.resourceRef !== undefined) {
      ResourceRef.encode(message.resourceRef, writer.uint32(10).fork()).ldelim();
    }
    if (message.manifest !== undefined) {
      Any.encode(message.manifest, writer.uint32(18).fork()).ldelim();
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
          message.manifest = Any.decode(reader, reader.uint32());
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
    message.resourceRef =
      object.resourceRef !== undefined && object.resourceRef !== null
        ? ResourceRef.fromJSON(object.resourceRef)
        : undefined;
    message.manifest =
      object.manifest !== undefined && object.manifest !== null
        ? Any.fromJSON(object.manifest)
        : undefined;
    return message;
  },

  toJSON(message: GetResourcesResponse): unknown {
    const obj: any = {};
    message.resourceRef !== undefined &&
      (obj.resourceRef = message.resourceRef ? ResourceRef.toJSON(message.resourceRef) : undefined);
    message.manifest !== undefined &&
      (obj.manifest = message.manifest ? Any.toJSON(message.manifest) : undefined);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<GetResourcesResponse>, I>>(
    object: I,
  ): GetResourcesResponse {
    const message = { ...baseGetResourcesResponse } as GetResourcesResponse;
    message.resourceRef =
      object.resourceRef !== undefined && object.resourceRef !== null
        ? ResourceRef.fromPartial(object.resourceRef)
        : undefined;
    message.manifest =
      object.manifest !== undefined && object.manifest !== null
        ? Any.fromPartial(object.manifest)
        : undefined;
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
    message.context =
      object.context !== undefined && object.context !== null
        ? Context.fromJSON(object.context)
        : undefined;
    return message;
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
    const message = {
      ...baseGetServiceAccountNamesRequest,
    } as GetServiceAccountNamesRequest;
    message.context =
      object.context !== undefined && object.context !== null
        ? Context.fromPartial(object.context)
        : undefined;
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
    message.serviceaccountNames = (object.serviceaccountNames ?? []).map((e: any) => String(e));
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

  fromPartial<I extends Exact<DeepPartial<GetServiceAccountNamesResponse>, I>>(
    object: I,
  ): GetServiceAccountNamesResponse {
    const message = {
      ...baseGetServiceAccountNamesResponse,
    } as GetServiceAccountNamesResponse;
    message.serviceaccountNames = object.serviceaccountNames?.map(e => e) || [];
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
    message.cluster =
      object.cluster !== undefined && object.cluster !== null ? String(object.cluster) : "";
    return message;
  },

  toJSON(message: GetNamespaceNamesRequest): unknown {
    const obj: any = {};
    message.cluster !== undefined && (obj.cluster = message.cluster);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<GetNamespaceNamesRequest>, I>>(
    object: I,
  ): GetNamespaceNamesRequest {
    const message = {
      ...baseGetNamespaceNamesRequest,
    } as GetNamespaceNamesRequest;
    message.cluster = object.cluster ?? "";
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
    message.context =
      object.context !== undefined && object.context !== null
        ? Context.fromJSON(object.context)
        : undefined;
    return message;
  },

  toJSON(message: CreateNamespaceRequest): unknown {
    const obj: any = {};
    message.context !== undefined &&
      (obj.context = message.context ? Context.toJSON(message.context) : undefined);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<CreateNamespaceRequest>, I>>(
    object: I,
  ): CreateNamespaceRequest {
    const message = { ...baseCreateNamespaceRequest } as CreateNamespaceRequest;
    message.context =
      object.context !== undefined && object.context !== null
        ? Context.fromPartial(object.context)
        : undefined;
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

  fromPartial<I extends Exact<DeepPartial<CreateNamespaceResponse>, I>>(
    _: I,
  ): CreateNamespaceResponse {
    const message = {
      ...baseCreateNamespaceResponse,
    } as CreateNamespaceResponse;
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
    message.namespaceNames = (object.namespaceNames ?? []).map((e: any) => String(e));
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

  fromPartial<I extends Exact<DeepPartial<GetNamespaceNamesResponse>, I>>(
    object: I,
  ): GetNamespaceNamesResponse {
    const message = {
      ...baseGetNamespaceNamesResponse,
    } as GetNamespaceNamesResponse;
    message.namespaceNames = object.namespaceNames?.map(e => e) || [];
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
    message.context =
      object.context !== undefined && object.context !== null
        ? Context.fromJSON(object.context)
        : undefined;
    return message;
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
    const message = {
      ...baseCheckNamespaceExistsRequest,
    } as CheckNamespaceExistsRequest;
    message.context =
      object.context !== undefined && object.context !== null
        ? Context.fromPartial(object.context)
        : undefined;
    return message;
  },
};

const baseCheckNamespaceExistsResponse: object = {};

export const CheckNamespaceExistsResponse = {
  encode(_: CheckNamespaceExistsResponse, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
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
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(_: any): CheckNamespaceExistsResponse {
    const message = {
      ...baseCheckNamespaceExistsResponse,
    } as CheckNamespaceExistsResponse;
    return message;
  },

  toJSON(_: CheckNamespaceExistsResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<CheckNamespaceExistsResponse>, I>>(
    _: I,
  ): CheckNamespaceExistsResponse {
    const message = {
      ...baseCheckNamespaceExistsResponse,
    } as CheckNamespaceExistsResponse;
    return message;
  },
};

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

type KeysOfUnion<T> = T extends T ? keyof T : never;
export type Exact<P, I extends P> = P extends Builtin
  ? P
  : P & { [K in keyof P]: Exact<P[K], I[K]> } & Record<Exclude<keyof I, KeysOfUnion<P>>, never>;

if (_m0.util.Long !== Long) {
  _m0.util.Long = Long as any;
  _m0.configure();
}
