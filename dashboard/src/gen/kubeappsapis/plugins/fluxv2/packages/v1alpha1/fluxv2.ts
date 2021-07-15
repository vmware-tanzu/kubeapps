/* eslint-disable */
import Long from "long";
import { grpc } from "@improbable-eng/grpc-web";
import _m0 from "protobufjs/minimal";
import {
  Context,
  GetAvailablePackageSummariesRequest,
  GetAvailablePackageDetailRequest,
  GetAvailablePackageVersionsRequest,
  GetAvailablePackageSummariesResponse,
  GetAvailablePackageDetailResponse,
  GetAvailablePackageVersionsResponse,
} from "../../../../../kubeappsapis/core/packages/v1alpha1/packages";
import { Plugin } from "../../../../../kubeappsapis/core/plugins/v1alpha1/plugins";
import { BrowserHeaders } from "browser-headers";

export const protobufPackage = "kubeappsapis.plugins.fluxv2.packages.v1alpha1";

/**
 * GetPackageRepositories
 *
 * Request for GetPackageRepositories
 */
export interface GetPackageRepositoriesRequest {
  /** The context (cluster/namespace) for the request */
  context?: Context;
}

/**
 * GetPackageRepositories
 *
 * Response for GetPackageRepositories
 */
export interface GetPackageRepositoriesResponse {
  /**
   * Repositories
   *
   * List of PackageRepository
   */
  repositories: PackageRepository[];
}

/**
 * PackageRepository
 *
 * A PackageRepository defines a repository of packages for installation.
 */
export interface PackageRepository {
  /**
   * Package repository name
   *
   * The name identifying package repository on the cluster.
   */
  name: string;
  /**
   * Package repository namespace
   *
   * An optional namespace for namespaced package repositories.
   */
  namespace: string;
  /**
   * Package repository URL
   *
   * A url identifying the package repository location.
   */
  url: string;
  /**
   * Package repository plugin
   *
   * The plugin used to interact with this package repository.
   */
  plugin?: Plugin;
}

const baseGetPackageRepositoriesRequest: object = {};

export const GetPackageRepositoriesRequest = {
  encode(
    message: GetPackageRepositoriesRequest,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.context !== undefined) {
      Context.encode(message.context, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetPackageRepositoriesRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseGetPackageRepositoriesRequest,
    } as GetPackageRepositoriesRequest;
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

  fromJSON(object: any): GetPackageRepositoriesRequest {
    const message = {
      ...baseGetPackageRepositoriesRequest,
    } as GetPackageRepositoriesRequest;
    if (object.context !== undefined && object.context !== null) {
      message.context = Context.fromJSON(object.context);
    } else {
      message.context = undefined;
    }
    return message;
  },

  toJSON(message: GetPackageRepositoriesRequest): unknown {
    const obj: any = {};
    message.context !== undefined &&
      (obj.context = message.context ? Context.toJSON(message.context) : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<GetPackageRepositoriesRequest>): GetPackageRepositoriesRequest {
    const message = {
      ...baseGetPackageRepositoriesRequest,
    } as GetPackageRepositoriesRequest;
    if (object.context !== undefined && object.context !== null) {
      message.context = Context.fromPartial(object.context);
    } else {
      message.context = undefined;
    }
    return message;
  },
};

const baseGetPackageRepositoriesResponse: object = {};

export const GetPackageRepositoriesResponse = {
  encode(
    message: GetPackageRepositoriesResponse,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    for (const v of message.repositories) {
      PackageRepository.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetPackageRepositoriesResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseGetPackageRepositoriesResponse,
    } as GetPackageRepositoriesResponse;
    message.repositories = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.repositories.push(PackageRepository.decode(reader, reader.uint32()));
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): GetPackageRepositoriesResponse {
    const message = {
      ...baseGetPackageRepositoriesResponse,
    } as GetPackageRepositoriesResponse;
    message.repositories = [];
    if (object.repositories !== undefined && object.repositories !== null) {
      for (const e of object.repositories) {
        message.repositories.push(PackageRepository.fromJSON(e));
      }
    }
    return message;
  },

  toJSON(message: GetPackageRepositoriesResponse): unknown {
    const obj: any = {};
    if (message.repositories) {
      obj.repositories = message.repositories.map(e =>
        e ? PackageRepository.toJSON(e) : undefined,
      );
    } else {
      obj.repositories = [];
    }
    return obj;
  },

  fromPartial(object: DeepPartial<GetPackageRepositoriesResponse>): GetPackageRepositoriesResponse {
    const message = {
      ...baseGetPackageRepositoriesResponse,
    } as GetPackageRepositoriesResponse;
    message.repositories = [];
    if (object.repositories !== undefined && object.repositories !== null) {
      for (const e of object.repositories) {
        message.repositories.push(PackageRepository.fromPartial(e));
      }
    }
    return message;
  },
};

const basePackageRepository: object = { name: "", namespace: "", url: "" };

export const PackageRepository = {
  encode(message: PackageRepository, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.name !== "") {
      writer.uint32(10).string(message.name);
    }
    if (message.namespace !== "") {
      writer.uint32(18).string(message.namespace);
    }
    if (message.url !== "") {
      writer.uint32(26).string(message.url);
    }
    if (message.plugin !== undefined) {
      Plugin.encode(message.plugin, writer.uint32(34).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): PackageRepository {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...basePackageRepository } as PackageRepository;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.name = reader.string();
          break;
        case 2:
          message.namespace = reader.string();
          break;
        case 3:
          message.url = reader.string();
          break;
        case 4:
          message.plugin = Plugin.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): PackageRepository {
    const message = { ...basePackageRepository } as PackageRepository;
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
    if (object.url !== undefined && object.url !== null) {
      message.url = String(object.url);
    } else {
      message.url = "";
    }
    if (object.plugin !== undefined && object.plugin !== null) {
      message.plugin = Plugin.fromJSON(object.plugin);
    } else {
      message.plugin = undefined;
    }
    return message;
  },

  toJSON(message: PackageRepository): unknown {
    const obj: any = {};
    message.name !== undefined && (obj.name = message.name);
    message.namespace !== undefined && (obj.namespace = message.namespace);
    message.url !== undefined && (obj.url = message.url);
    message.plugin !== undefined &&
      (obj.plugin = message.plugin ? Plugin.toJSON(message.plugin) : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<PackageRepository>): PackageRepository {
    const message = { ...basePackageRepository } as PackageRepository;
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
    if (object.url !== undefined && object.url !== null) {
      message.url = object.url;
    } else {
      message.url = "";
    }
    if (object.plugin !== undefined && object.plugin !== null) {
      message.plugin = Plugin.fromPartial(object.plugin);
    } else {
      message.plugin = undefined;
    }
    return message;
  },
};

export interface FluxV2PackagesService {
  /** GetAvailablePackageSummaries returns the available packages managed by the 'fluxv2' plugin */
  GetAvailablePackageSummaries(
    request: DeepPartial<GetAvailablePackageSummariesRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetAvailablePackageSummariesResponse>;
  /** GetAvailablePackageDetail returns the package metadata managed by the 'fluxv2' plugin */
  GetAvailablePackageDetail(
    request: DeepPartial<GetAvailablePackageDetailRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetAvailablePackageDetailResponse>;
  /** GetAvailablePackageVersions returns the package versions managed by the 'fluxv2' plugin */
  GetAvailablePackageVersions(
    request: DeepPartial<GetAvailablePackageVersionsRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetAvailablePackageVersionsResponse>;
  /** GetPackageRepositories returns the repositories managed by the 'fluxv2' plugin */
  GetPackageRepositories(
    request: DeepPartial<GetPackageRepositoriesRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetPackageRepositoriesResponse>;
}

export class FluxV2PackagesServiceClientImpl implements FluxV2PackagesService {
  private readonly rpc: Rpc;

  constructor(rpc: Rpc) {
    this.rpc = rpc;
    this.GetAvailablePackageSummaries = this.GetAvailablePackageSummaries.bind(this);
    this.GetAvailablePackageDetail = this.GetAvailablePackageDetail.bind(this);
    this.GetAvailablePackageVersions = this.GetAvailablePackageVersions.bind(this);
    this.GetPackageRepositories = this.GetPackageRepositories.bind(this);
  }

  GetAvailablePackageSummaries(
    request: DeepPartial<GetAvailablePackageSummariesRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetAvailablePackageSummariesResponse> {
    return this.rpc.unary(
      FluxV2PackagesServiceGetAvailablePackageSummariesDesc,
      GetAvailablePackageSummariesRequest.fromPartial(request),
      metadata,
    );
  }

  GetAvailablePackageDetail(
    request: DeepPartial<GetAvailablePackageDetailRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetAvailablePackageDetailResponse> {
    return this.rpc.unary(
      FluxV2PackagesServiceGetAvailablePackageDetailDesc,
      GetAvailablePackageDetailRequest.fromPartial(request),
      metadata,
    );
  }

  GetAvailablePackageVersions(
    request: DeepPartial<GetAvailablePackageVersionsRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetAvailablePackageVersionsResponse> {
    return this.rpc.unary(
      FluxV2PackagesServiceGetAvailablePackageVersionsDesc,
      GetAvailablePackageVersionsRequest.fromPartial(request),
      metadata,
    );
  }

  GetPackageRepositories(
    request: DeepPartial<GetPackageRepositoriesRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetPackageRepositoriesResponse> {
    return this.rpc.unary(
      FluxV2PackagesServiceGetPackageRepositoriesDesc,
      GetPackageRepositoriesRequest.fromPartial(request),
      metadata,
    );
  }
}

export const FluxV2PackagesServiceDesc = {
  serviceName: "kubeappsapis.plugins.fluxv2.packages.v1alpha1.FluxV2PackagesService",
};

export const FluxV2PackagesServiceGetAvailablePackageSummariesDesc: UnaryMethodDefinitionish = {
  methodName: "GetAvailablePackageSummaries",
  service: FluxV2PackagesServiceDesc,
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

export const FluxV2PackagesServiceGetAvailablePackageDetailDesc: UnaryMethodDefinitionish = {
  methodName: "GetAvailablePackageDetail",
  service: FluxV2PackagesServiceDesc,
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

export const FluxV2PackagesServiceGetAvailablePackageVersionsDesc: UnaryMethodDefinitionish = {
  methodName: "GetAvailablePackageVersions",
  service: FluxV2PackagesServiceDesc,
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

export const FluxV2PackagesServiceGetPackageRepositoriesDesc: UnaryMethodDefinitionish = {
  methodName: "GetPackageRepositories",
  service: FluxV2PackagesServiceDesc,
  requestStream: false,
  responseStream: false,
  requestType: {
    serializeBinary() {
      return GetPackageRepositoriesRequest.encode(this).finish();
    },
  } as any,
  responseType: {
    deserializeBinary(data: Uint8Array) {
      return {
        ...GetPackageRepositoriesResponse.decode(data),
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
