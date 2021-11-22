/* eslint-disable */
import Long from "long";
import { grpc } from "@improbable-eng/grpc-web";
import _m0 from "protobufjs/minimal";
import {
  InstalledPackageReference,
  ResourceRef,
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
    if (object.installedPackageRef !== undefined && object.installedPackageRef !== null) {
      message.installedPackageRef = InstalledPackageReference.fromJSON(object.installedPackageRef);
    } else {
      message.installedPackageRef = undefined;
    }
    message.resourceRefs = (object.resourceRefs ?? []).map((e: any) => ResourceRef.fromJSON(e));
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
    if (object.installedPackageRef !== undefined && object.installedPackageRef !== null) {
      message.installedPackageRef = InstalledPackageReference.fromPartial(
        object.installedPackageRef,
      );
    } else {
      message.installedPackageRef = undefined;
    }
    message.resourceRefs = (object.resourceRefs ?? []).map(e => ResourceRef.fromPartial(e));
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
    if (object.resourceRef !== undefined && object.resourceRef !== null) {
      message.resourceRef = ResourceRef.fromJSON(object.resourceRef);
    } else {
      message.resourceRef = undefined;
    }
    if (object.manifest !== undefined && object.manifest !== null) {
      message.manifest = Any.fromJSON(object.manifest);
    } else {
      message.manifest = undefined;
    }
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

  fromPartial(object: DeepPartial<GetResourcesResponse>): GetResourcesResponse {
    const message = { ...baseGetResourcesResponse } as GetResourcesResponse;
    if (object.resourceRef !== undefined && object.resourceRef !== null) {
      message.resourceRef = ResourceRef.fromPartial(object.resourceRef);
    } else {
      message.resourceRef = undefined;
    }
    if (object.manifest !== undefined && object.manifest !== null) {
      message.manifest = Any.fromPartial(object.manifest);
    } else {
      message.manifest = undefined;
    }
    return message;
  },
};

export interface ResourcesService {
  GetResources(
    request: DeepPartial<GetResourcesRequest>,
    metadata?: grpc.Metadata,
  ): Observable<GetResourcesResponse>;
}

export class ResourcesServiceClientImpl implements ResourcesService {
  private readonly rpc: Rpc;

  constructor(rpc: Rpc) {
    this.rpc = rpc;
    this.GetResources = this.GetResources.bind(this);
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
