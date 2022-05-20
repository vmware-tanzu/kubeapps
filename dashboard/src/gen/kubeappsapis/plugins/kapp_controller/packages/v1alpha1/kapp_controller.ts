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

export const protobufPackage = "kubeappsapis.plugins.kapp_controller.packages.v1alpha1";

export interface KappControllerPackagesService {
  /** GetAvailablePackageSummaries returns the available packages managed by the 'kapp_controller' plugin */
  GetAvailablePackageSummaries(
    request: DeepPartial<GetAvailablePackageSummariesRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetAvailablePackageSummariesResponse>;
  /** GetAvailablePackageDetail returns the package details managed by the 'kapp_controller' plugin */
  GetAvailablePackageDetail(
    request: DeepPartial<GetAvailablePackageDetailRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetAvailablePackageDetailResponse>;
  /** GetAvailablePackageVersions returns the package versions managed by the 'kapp_controller' plugin */
  GetAvailablePackageVersions(
    request: DeepPartial<GetAvailablePackageVersionsRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetAvailablePackageVersionsResponse>;
  /** GetInstalledPackageSummaries returns the installed packages managed by the 'kapp_controller' plugin */
  GetInstalledPackageSummaries(
    request: DeepPartial<GetInstalledPackageSummariesRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetInstalledPackageSummariesResponse>;
  /** GetInstalledPackageDetail returns the requested installed package managed by the 'kapp_controller' plugin */
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
   * GetInstalledPackageResourceRefs returns the references for the Kubernetes resources created by
   * an installed package.
   */
  GetInstalledPackageResourceRefs(
    request: DeepPartial<GetInstalledPackageResourceRefsRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetInstalledPackageResourceRefsResponse>;
}

export class KappControllerPackagesServiceClientImpl implements KappControllerPackagesService {
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
      KappControllerPackagesServiceGetAvailablePackageSummariesDesc,
      GetAvailablePackageSummariesRequest.fromPartial(request),
      metadata,
    );
  }

  GetAvailablePackageDetail(
    request: DeepPartial<GetAvailablePackageDetailRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetAvailablePackageDetailResponse> {
    return this.rpc.unary(
      KappControllerPackagesServiceGetAvailablePackageDetailDesc,
      GetAvailablePackageDetailRequest.fromPartial(request),
      metadata,
    );
  }

  GetAvailablePackageVersions(
    request: DeepPartial<GetAvailablePackageVersionsRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetAvailablePackageVersionsResponse> {
    return this.rpc.unary(
      KappControllerPackagesServiceGetAvailablePackageVersionsDesc,
      GetAvailablePackageVersionsRequest.fromPartial(request),
      metadata,
    );
  }

  GetInstalledPackageSummaries(
    request: DeepPartial<GetInstalledPackageSummariesRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetInstalledPackageSummariesResponse> {
    return this.rpc.unary(
      KappControllerPackagesServiceGetInstalledPackageSummariesDesc,
      GetInstalledPackageSummariesRequest.fromPartial(request),
      metadata,
    );
  }

  GetInstalledPackageDetail(
    request: DeepPartial<GetInstalledPackageDetailRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetInstalledPackageDetailResponse> {
    return this.rpc.unary(
      KappControllerPackagesServiceGetInstalledPackageDetailDesc,
      GetInstalledPackageDetailRequest.fromPartial(request),
      metadata,
    );
  }

  CreateInstalledPackage(
    request: DeepPartial<CreateInstalledPackageRequest>,
    metadata?: grpc.Metadata,
  ): Promise<CreateInstalledPackageResponse> {
    return this.rpc.unary(
      KappControllerPackagesServiceCreateInstalledPackageDesc,
      CreateInstalledPackageRequest.fromPartial(request),
      metadata,
    );
  }

  UpdateInstalledPackage(
    request: DeepPartial<UpdateInstalledPackageRequest>,
    metadata?: grpc.Metadata,
  ): Promise<UpdateInstalledPackageResponse> {
    return this.rpc.unary(
      KappControllerPackagesServiceUpdateInstalledPackageDesc,
      UpdateInstalledPackageRequest.fromPartial(request),
      metadata,
    );
  }

  DeleteInstalledPackage(
    request: DeepPartial<DeleteInstalledPackageRequest>,
    metadata?: grpc.Metadata,
  ): Promise<DeleteInstalledPackageResponse> {
    return this.rpc.unary(
      KappControllerPackagesServiceDeleteInstalledPackageDesc,
      DeleteInstalledPackageRequest.fromPartial(request),
      metadata,
    );
  }

  GetInstalledPackageResourceRefs(
    request: DeepPartial<GetInstalledPackageResourceRefsRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetInstalledPackageResourceRefsResponse> {
    return this.rpc.unary(
      KappControllerPackagesServiceGetInstalledPackageResourceRefsDesc,
      GetInstalledPackageResourceRefsRequest.fromPartial(request),
      metadata,
    );
  }
}

export const KappControllerPackagesServiceDesc = {
  serviceName:
    "kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService",
};

export const KappControllerPackagesServiceGetAvailablePackageSummariesDesc: UnaryMethodDefinitionish =
  {
    methodName: "GetAvailablePackageSummaries",
    service: KappControllerPackagesServiceDesc,
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

export const KappControllerPackagesServiceGetAvailablePackageDetailDesc: UnaryMethodDefinitionish =
  {
    methodName: "GetAvailablePackageDetail",
    service: KappControllerPackagesServiceDesc,
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

export const KappControllerPackagesServiceGetAvailablePackageVersionsDesc: UnaryMethodDefinitionish =
  {
    methodName: "GetAvailablePackageVersions",
    service: KappControllerPackagesServiceDesc,
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

export const KappControllerPackagesServiceGetInstalledPackageSummariesDesc: UnaryMethodDefinitionish =
  {
    methodName: "GetInstalledPackageSummaries",
    service: KappControllerPackagesServiceDesc,
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

export const KappControllerPackagesServiceGetInstalledPackageDetailDesc: UnaryMethodDefinitionish =
  {
    methodName: "GetInstalledPackageDetail",
    service: KappControllerPackagesServiceDesc,
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

export const KappControllerPackagesServiceCreateInstalledPackageDesc: UnaryMethodDefinitionish = {
  methodName: "CreateInstalledPackage",
  service: KappControllerPackagesServiceDesc,
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

export const KappControllerPackagesServiceUpdateInstalledPackageDesc: UnaryMethodDefinitionish = {
  methodName: "UpdateInstalledPackage",
  service: KappControllerPackagesServiceDesc,
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

export const KappControllerPackagesServiceDeleteInstalledPackageDesc: UnaryMethodDefinitionish = {
  methodName: "DeleteInstalledPackage",
  service: KappControllerPackagesServiceDesc,
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

export const KappControllerPackagesServiceGetInstalledPackageResourceRefsDesc: UnaryMethodDefinitionish =
  {
    methodName: "GetInstalledPackageResourceRefs",
    service: KappControllerPackagesServiceDesc,
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

export interface KappControllerRepositoriesService {
  /** AddPackageRepository add an existing package repository to the set of ones already managed by the 'kapp_controller' plugin */
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
}

export class KappControllerRepositoriesServiceClientImpl
  implements KappControllerRepositoriesService
{
  private readonly rpc: Rpc;

  constructor(rpc: Rpc) {
    this.rpc = rpc;
    this.AddPackageRepository = this.AddPackageRepository.bind(this);
    this.GetPackageRepositoryDetail = this.GetPackageRepositoryDetail.bind(this);
    this.GetPackageRepositorySummaries = this.GetPackageRepositorySummaries.bind(this);
    this.UpdatePackageRepository = this.UpdatePackageRepository.bind(this);
    this.DeletePackageRepository = this.DeletePackageRepository.bind(this);
  }

  AddPackageRepository(
    request: DeepPartial<AddPackageRepositoryRequest>,
    metadata?: grpc.Metadata,
  ): Promise<AddPackageRepositoryResponse> {
    return this.rpc.unary(
      KappControllerRepositoriesServiceAddPackageRepositoryDesc,
      AddPackageRepositoryRequest.fromPartial(request),
      metadata,
    );
  }

  GetPackageRepositoryDetail(
    request: DeepPartial<GetPackageRepositoryDetailRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetPackageRepositoryDetailResponse> {
    return this.rpc.unary(
      KappControllerRepositoriesServiceGetPackageRepositoryDetailDesc,
      GetPackageRepositoryDetailRequest.fromPartial(request),
      metadata,
    );
  }

  GetPackageRepositorySummaries(
    request: DeepPartial<GetPackageRepositorySummariesRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetPackageRepositorySummariesResponse> {
    return this.rpc.unary(
      KappControllerRepositoriesServiceGetPackageRepositorySummariesDesc,
      GetPackageRepositorySummariesRequest.fromPartial(request),
      metadata,
    );
  }

  UpdatePackageRepository(
    request: DeepPartial<UpdatePackageRepositoryRequest>,
    metadata?: grpc.Metadata,
  ): Promise<UpdatePackageRepositoryResponse> {
    return this.rpc.unary(
      KappControllerRepositoriesServiceUpdatePackageRepositoryDesc,
      UpdatePackageRepositoryRequest.fromPartial(request),
      metadata,
    );
  }

  DeletePackageRepository(
    request: DeepPartial<DeletePackageRepositoryRequest>,
    metadata?: grpc.Metadata,
  ): Promise<DeletePackageRepositoryResponse> {
    return this.rpc.unary(
      KappControllerRepositoriesServiceDeletePackageRepositoryDesc,
      DeletePackageRepositoryRequest.fromPartial(request),
      metadata,
    );
  }
}

export const KappControllerRepositoriesServiceDesc = {
  serviceName:
    "kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerRepositoriesService",
};

export const KappControllerRepositoriesServiceAddPackageRepositoryDesc: UnaryMethodDefinitionish = {
  methodName: "AddPackageRepository",
  service: KappControllerRepositoriesServiceDesc,
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

export const KappControllerRepositoriesServiceGetPackageRepositoryDetailDesc: UnaryMethodDefinitionish =
  {
    methodName: "GetPackageRepositoryDetail",
    service: KappControllerRepositoriesServiceDesc,
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

export const KappControllerRepositoriesServiceGetPackageRepositorySummariesDesc: UnaryMethodDefinitionish =
  {
    methodName: "GetPackageRepositorySummaries",
    service: KappControllerRepositoriesServiceDesc,
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

export const KappControllerRepositoriesServiceUpdatePackageRepositoryDesc: UnaryMethodDefinitionish =
  {
    methodName: "UpdatePackageRepository",
    service: KappControllerRepositoriesServiceDesc,
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

export const KappControllerRepositoriesServiceDeletePackageRepositoryDesc: UnaryMethodDefinitionish =
  {
    methodName: "DeletePackageRepository",
    service: KappControllerRepositoriesServiceDesc,
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
