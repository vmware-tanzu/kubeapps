/* eslint-disable */
import Long from "long";
import { grpc } from "@improbable-eng/grpc-web";
import _m0 from "protobufjs/minimal";
import {
  GetAvailablePackageSummariesRequest,
  GetAvailablePackageDetailRequest,
  GetAvailablePackageVersionsRequest,
  GetInstalledPackageSummariesRequest,
  GetInstalledPackageDetailRequest,
  CreateInstalledPackageRequest,
  UpdateInstalledPackageRequest,
  DeleteInstalledPackageRequest,
  GetInstalledPackageResourceRefsRequest,
  GetAvailablePackageSummariesResponse,
  GetAvailablePackageDetailResponse,
  GetAvailablePackageVersionsResponse,
  GetInstalledPackageSummariesResponse,
  GetInstalledPackageDetailResponse,
  CreateInstalledPackageResponse,
  UpdateInstalledPackageResponse,
  DeleteInstalledPackageResponse,
  GetInstalledPackageResourceRefsResponse,
} from "../../../../../kubeappsapis/core/packages/v1alpha1/packages";
import {
  AddPackageRepositoryRequest,
  GetPackageRepositoryDetailRequest,
  GetPackageRepositorySummariesRequest,
  UpdatePackageRepositoryRequest,
  DeletePackageRepositoryRequest,
  AddPackageRepositoryResponse,
  GetPackageRepositoryDetailResponse,
  GetPackageRepositorySummariesResponse,
  UpdatePackageRepositoryResponse,
  DeletePackageRepositoryResponse,
} from "../../../../../kubeappsapis/core/packages/v1alpha1/repositories";
import { BrowserHeaders } from "browser-headers";

export const protobufPackage = "kubeappsapis.plugins.fluxv2.packages.v1alpha1";

export interface SetUserManagedSecretsRequest {
  value: boolean;
}

export interface SetUserManagedSecretsResponse {
  value: boolean;
}

const baseSetUserManagedSecretsRequest: object = { value: false };

export const SetUserManagedSecretsRequest = {
  encode(
    message: SetUserManagedSecretsRequest,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.value === true) {
      writer.uint32(8).bool(message.value);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): SetUserManagedSecretsRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseSetUserManagedSecretsRequest,
    } as SetUserManagedSecretsRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.value = reader.bool();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): SetUserManagedSecretsRequest {
    const message = {
      ...baseSetUserManagedSecretsRequest,
    } as SetUserManagedSecretsRequest;
    if (object.value !== undefined && object.value !== null) {
      message.value = Boolean(object.value);
    } else {
      message.value = false;
    }
    return message;
  },

  toJSON(message: SetUserManagedSecretsRequest): unknown {
    const obj: any = {};
    message.value !== undefined && (obj.value = message.value);
    return obj;
  },

  fromPartial(object: DeepPartial<SetUserManagedSecretsRequest>): SetUserManagedSecretsRequest {
    const message = {
      ...baseSetUserManagedSecretsRequest,
    } as SetUserManagedSecretsRequest;
    if (object.value !== undefined && object.value !== null) {
      message.value = object.value;
    } else {
      message.value = false;
    }
    return message;
  },
};

const baseSetUserManagedSecretsResponse: object = { value: false };

export const SetUserManagedSecretsResponse = {
  encode(
    message: SetUserManagedSecretsResponse,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.value === true) {
      writer.uint32(8).bool(message.value);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): SetUserManagedSecretsResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseSetUserManagedSecretsResponse,
    } as SetUserManagedSecretsResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.value = reader.bool();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): SetUserManagedSecretsResponse {
    const message = {
      ...baseSetUserManagedSecretsResponse,
    } as SetUserManagedSecretsResponse;
    if (object.value !== undefined && object.value !== null) {
      message.value = Boolean(object.value);
    } else {
      message.value = false;
    }
    return message;
  },

  toJSON(message: SetUserManagedSecretsResponse): unknown {
    const obj: any = {};
    message.value !== undefined && (obj.value = message.value);
    return obj;
  },

  fromPartial(object: DeepPartial<SetUserManagedSecretsResponse>): SetUserManagedSecretsResponse {
    const message = {
      ...baseSetUserManagedSecretsResponse,
    } as SetUserManagedSecretsResponse;
    if (object.value !== undefined && object.value !== null) {
      message.value = object.value;
    } else {
      message.value = false;
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
  /** GetInstalledPackageSummaries returns the installed packages managed by the 'fluxv2' plugin */
  GetInstalledPackageSummaries(
    request: DeepPartial<GetInstalledPackageSummariesRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetInstalledPackageSummariesResponse>;
  /** GetInstalledPackageDetail returns the requested installed package managed by the 'fluxv2' plugin */
  GetInstalledPackageDetail(
    request: DeepPartial<GetInstalledPackageDetailRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetInstalledPackageDetailResponse>;
  /** CreateInstalledPackage creates an installed package based on the request. */
  CreateInstalledPackage(
    request: DeepPartial<CreateInstalledPackageRequest>,
    metadata?: grpc.Metadata,
  ): Promise<CreateInstalledPackageResponse>;
  /** UpdateInstalledPackage updates an installed package based on the request. */
  UpdateInstalledPackage(
    request: DeepPartial<UpdateInstalledPackageRequest>,
    metadata?: grpc.Metadata,
  ): Promise<UpdateInstalledPackageResponse>;
  /** DeleteInstalledPackage deletes an installed package based on the request. */
  DeleteInstalledPackage(
    request: DeepPartial<DeleteInstalledPackageRequest>,
    metadata?: grpc.Metadata,
  ): Promise<DeleteInstalledPackageResponse>;
  /**
   * GetInstalledPackageResourceRefs returns the references for the Kubernetes
   * resources created by an installed package.
   */
  GetInstalledPackageResourceRefs(
    request: DeepPartial<GetInstalledPackageResourceRefsRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetInstalledPackageResourceRefsResponse>;
}

export class FluxV2PackagesServiceClientImpl implements FluxV2PackagesService {
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

  GetInstalledPackageSummaries(
    request: DeepPartial<GetInstalledPackageSummariesRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetInstalledPackageSummariesResponse> {
    return this.rpc.unary(
      FluxV2PackagesServiceGetInstalledPackageSummariesDesc,
      GetInstalledPackageSummariesRequest.fromPartial(request),
      metadata,
    );
  }

  GetInstalledPackageDetail(
    request: DeepPartial<GetInstalledPackageDetailRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetInstalledPackageDetailResponse> {
    return this.rpc.unary(
      FluxV2PackagesServiceGetInstalledPackageDetailDesc,
      GetInstalledPackageDetailRequest.fromPartial(request),
      metadata,
    );
  }

  CreateInstalledPackage(
    request: DeepPartial<CreateInstalledPackageRequest>,
    metadata?: grpc.Metadata,
  ): Promise<CreateInstalledPackageResponse> {
    return this.rpc.unary(
      FluxV2PackagesServiceCreateInstalledPackageDesc,
      CreateInstalledPackageRequest.fromPartial(request),
      metadata,
    );
  }

  UpdateInstalledPackage(
    request: DeepPartial<UpdateInstalledPackageRequest>,
    metadata?: grpc.Metadata,
  ): Promise<UpdateInstalledPackageResponse> {
    return this.rpc.unary(
      FluxV2PackagesServiceUpdateInstalledPackageDesc,
      UpdateInstalledPackageRequest.fromPartial(request),
      metadata,
    );
  }

  DeleteInstalledPackage(
    request: DeepPartial<DeleteInstalledPackageRequest>,
    metadata?: grpc.Metadata,
  ): Promise<DeleteInstalledPackageResponse> {
    return this.rpc.unary(
      FluxV2PackagesServiceDeleteInstalledPackageDesc,
      DeleteInstalledPackageRequest.fromPartial(request),
      metadata,
    );
  }

  GetInstalledPackageResourceRefs(
    request: DeepPartial<GetInstalledPackageResourceRefsRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetInstalledPackageResourceRefsResponse> {
    return this.rpc.unary(
      FluxV2PackagesServiceGetInstalledPackageResourceRefsDesc,
      GetInstalledPackageResourceRefsRequest.fromPartial(request),
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

export const FluxV2PackagesServiceGetInstalledPackageSummariesDesc: UnaryMethodDefinitionish = {
  methodName: "GetInstalledPackageSummaries",
  service: FluxV2PackagesServiceDesc,
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

export const FluxV2PackagesServiceGetInstalledPackageDetailDesc: UnaryMethodDefinitionish = {
  methodName: "GetInstalledPackageDetail",
  service: FluxV2PackagesServiceDesc,
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

export const FluxV2PackagesServiceCreateInstalledPackageDesc: UnaryMethodDefinitionish = {
  methodName: "CreateInstalledPackage",
  service: FluxV2PackagesServiceDesc,
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

export const FluxV2PackagesServiceUpdateInstalledPackageDesc: UnaryMethodDefinitionish = {
  methodName: "UpdateInstalledPackage",
  service: FluxV2PackagesServiceDesc,
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

export const FluxV2PackagesServiceDeleteInstalledPackageDesc: UnaryMethodDefinitionish = {
  methodName: "DeleteInstalledPackage",
  service: FluxV2PackagesServiceDesc,
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

export const FluxV2PackagesServiceGetInstalledPackageResourceRefsDesc: UnaryMethodDefinitionish = {
  methodName: "GetInstalledPackageResourceRefs",
  service: FluxV2PackagesServiceDesc,
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

export interface FluxV2RepositoriesService {
  /**
   * AddPackageRepository add an existing package repository to the set of ones already managed by the
   * 'fluxv2' plugin
   */
  AddPackageRepository(
    request: DeepPartial<AddPackageRepositoryRequest>,
    metadata?: grpc.Metadata,
  ): Promise<AddPackageRepositoryResponse>;
  GetPackageRepositoryDetail(
    request: DeepPartial<GetPackageRepositoryDetailRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetPackageRepositoryDetailResponse>;
  GetPackageRepositorySummaries(
    request: DeepPartial<GetPackageRepositorySummariesRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetPackageRepositorySummariesResponse>;
  UpdatePackageRepository(
    request: DeepPartial<UpdatePackageRepositoryRequest>,
    metadata?: grpc.Metadata,
  ): Promise<UpdatePackageRepositoryResponse>;
  DeletePackageRepository(
    request: DeepPartial<DeletePackageRepositoryRequest>,
    metadata?: grpc.Metadata,
  ): Promise<DeletePackageRepositoryResponse>;
  /** this endpoint only exists for the purpose of integration tests */
  SetUserManagedSecrets(
    request: DeepPartial<SetUserManagedSecretsRequest>,
    metadata?: grpc.Metadata,
  ): Promise<SetUserManagedSecretsResponse>;
}

export class FluxV2RepositoriesServiceClientImpl implements FluxV2RepositoriesService {
  private readonly rpc: Rpc;

  constructor(rpc: Rpc) {
    this.rpc = rpc;
    this.AddPackageRepository = this.AddPackageRepository.bind(this);
    this.GetPackageRepositoryDetail = this.GetPackageRepositoryDetail.bind(this);
    this.GetPackageRepositorySummaries = this.GetPackageRepositorySummaries.bind(this);
    this.UpdatePackageRepository = this.UpdatePackageRepository.bind(this);
    this.DeletePackageRepository = this.DeletePackageRepository.bind(this);
    this.SetUserManagedSecrets = this.SetUserManagedSecrets.bind(this);
  }

  AddPackageRepository(
    request: DeepPartial<AddPackageRepositoryRequest>,
    metadata?: grpc.Metadata,
  ): Promise<AddPackageRepositoryResponse> {
    return this.rpc.unary(
      FluxV2RepositoriesServiceAddPackageRepositoryDesc,
      AddPackageRepositoryRequest.fromPartial(request),
      metadata,
    );
  }

  GetPackageRepositoryDetail(
    request: DeepPartial<GetPackageRepositoryDetailRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetPackageRepositoryDetailResponse> {
    return this.rpc.unary(
      FluxV2RepositoriesServiceGetPackageRepositoryDetailDesc,
      GetPackageRepositoryDetailRequest.fromPartial(request),
      metadata,
    );
  }

  GetPackageRepositorySummaries(
    request: DeepPartial<GetPackageRepositorySummariesRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetPackageRepositorySummariesResponse> {
    return this.rpc.unary(
      FluxV2RepositoriesServiceGetPackageRepositorySummariesDesc,
      GetPackageRepositorySummariesRequest.fromPartial(request),
      metadata,
    );
  }

  UpdatePackageRepository(
    request: DeepPartial<UpdatePackageRepositoryRequest>,
    metadata?: grpc.Metadata,
  ): Promise<UpdatePackageRepositoryResponse> {
    return this.rpc.unary(
      FluxV2RepositoriesServiceUpdatePackageRepositoryDesc,
      UpdatePackageRepositoryRequest.fromPartial(request),
      metadata,
    );
  }

  DeletePackageRepository(
    request: DeepPartial<DeletePackageRepositoryRequest>,
    metadata?: grpc.Metadata,
  ): Promise<DeletePackageRepositoryResponse> {
    return this.rpc.unary(
      FluxV2RepositoriesServiceDeletePackageRepositoryDesc,
      DeletePackageRepositoryRequest.fromPartial(request),
      metadata,
    );
  }

  SetUserManagedSecrets(
    request: DeepPartial<SetUserManagedSecretsRequest>,
    metadata?: grpc.Metadata,
  ): Promise<SetUserManagedSecretsResponse> {
    return this.rpc.unary(
      FluxV2RepositoriesServiceSetUserManagedSecretsDesc,
      SetUserManagedSecretsRequest.fromPartial(request),
      metadata,
    );
  }
}

export const FluxV2RepositoriesServiceDesc = {
  serviceName: "kubeappsapis.plugins.fluxv2.packages.v1alpha1.FluxV2RepositoriesService",
};

export const FluxV2RepositoriesServiceAddPackageRepositoryDesc: UnaryMethodDefinitionish = {
  methodName: "AddPackageRepository",
  service: FluxV2RepositoriesServiceDesc,
  requestStream: false,
  responseStream: false,
  requestType: {
    serializeBinary() {
      return AddPackageRepositoryRequest.encode(this).finish();
    },
  } as any,
  responseType: {
    deserializeBinary(data: Uint8Array) {
      return {
        ...AddPackageRepositoryResponse.decode(data),
        toObject() {
          return this;
        },
      };
    },
  } as any,
};

export const FluxV2RepositoriesServiceGetPackageRepositoryDetailDesc: UnaryMethodDefinitionish = {
  methodName: "GetPackageRepositoryDetail",
  service: FluxV2RepositoriesServiceDesc,
  requestStream: false,
  responseStream: false,
  requestType: {
    serializeBinary() {
      return GetPackageRepositoryDetailRequest.encode(this).finish();
    },
  } as any,
  responseType: {
    deserializeBinary(data: Uint8Array) {
      return {
        ...GetPackageRepositoryDetailResponse.decode(data),
        toObject() {
          return this;
        },
      };
    },
  } as any,
};

export const FluxV2RepositoriesServiceGetPackageRepositorySummariesDesc: UnaryMethodDefinitionish =
  {
    methodName: "GetPackageRepositorySummaries",
    service: FluxV2RepositoriesServiceDesc,
    requestStream: false,
    responseStream: false,
    requestType: {
      serializeBinary() {
        return GetPackageRepositorySummariesRequest.encode(this).finish();
      },
    } as any,
    responseType: {
      deserializeBinary(data: Uint8Array) {
        return {
          ...GetPackageRepositorySummariesResponse.decode(data),
          toObject() {
            return this;
          },
        };
      },
    } as any,
  };

export const FluxV2RepositoriesServiceUpdatePackageRepositoryDesc: UnaryMethodDefinitionish = {
  methodName: "UpdatePackageRepository",
  service: FluxV2RepositoriesServiceDesc,
  requestStream: false,
  responseStream: false,
  requestType: {
    serializeBinary() {
      return UpdatePackageRepositoryRequest.encode(this).finish();
    },
  } as any,
  responseType: {
    deserializeBinary(data: Uint8Array) {
      return {
        ...UpdatePackageRepositoryResponse.decode(data),
        toObject() {
          return this;
        },
      };
    },
  } as any,
};

export const FluxV2RepositoriesServiceDeletePackageRepositoryDesc: UnaryMethodDefinitionish = {
  methodName: "DeletePackageRepository",
  service: FluxV2RepositoriesServiceDesc,
  requestStream: false,
  responseStream: false,
  requestType: {
    serializeBinary() {
      return DeletePackageRepositoryRequest.encode(this).finish();
    },
  } as any,
  responseType: {
    deserializeBinary(data: Uint8Array) {
      return {
        ...DeletePackageRepositoryResponse.decode(data),
        toObject() {
          return this;
        },
      };
    },
  } as any,
};

export const FluxV2RepositoriesServiceSetUserManagedSecretsDesc: UnaryMethodDefinitionish = {
  methodName: "SetUserManagedSecrets",
  service: FluxV2RepositoriesServiceDesc,
  requestStream: false,
  responseStream: false,
  requestType: {
    serializeBinary() {
      return SetUserManagedSecretsRequest.encode(this).finish();
    },
  } as any,
  responseType: {
    deserializeBinary(data: Uint8Array) {
      return {
        ...SetUserManagedSecretsResponse.decode(data),
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
