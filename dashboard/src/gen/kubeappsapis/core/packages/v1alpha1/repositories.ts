/* eslint-disable */
import Long from "long";
import { grpc } from "@improbable-eng/grpc-web";
import _m0 from "protobufjs/minimal";
import { Context } from "../../../../kubeappsapis/core/packages/v1alpha1/packages";
import { Plugin } from "../../../../kubeappsapis/core/plugins/v1alpha1/plugins";
import { BrowserHeaders } from "browser-headers";

export const protobufPackage = "kubeappsapis.core.packages.v1alpha1";

/**
 * AddPackageRepositoryRequest
 *
 * Request for AddPackageRepository
 */
export interface AddPackageRepositoryRequest {
  /**
   * The target context where the package repository is intended to be
   * installed.
   */
  context?: Context;
  /** A user-provided name for the package repository (e.g. bitnami) */
  name: string;
  /** A user-provided description. Optional */
  description: string;
  /**
   * Whether this repository is global or namespace-scoped. Optional.
   * By default, the value is false, i.e. the repository is global
   */
  namespaceScoped: boolean;
  /**
   * Package storage type
   * In general, each plug-in will define an acceptable set of valid types
   * - for direct helm plug-in valid values are: helm, oci
   * - for flux plug-in currently only supported value is helm. In the
   *   future, we may add support for git and/or AWS s3-style buckets
   */
  type: string;
  /**
   * A URL identifying the package repository location. Must contain at
   * least a protocol and host
   */
  url: string;
  /**
   * The interval at which to check the upstream for updates (in seconds)
   * Optional. Defaults to 10m if not specified
   */
  interval: number;
  /** TLS-specific parameters for connecting to a repository. Optional */
  tlsConfig?: PackageRepositoryTlsConfig;
  /**
   * The plugin used to interact with this package repository.
   * This field should be omitted when the request is in the context of a
   * specific plugin.
   */
  plugin?: Plugin;
}

/** PackageRepositoryTlsConfig */
export interface PackageRepositoryTlsConfig {
  /**
   * whether or not to skip TLS verification
   * note that fluxv2 does not currently support this and will raise an
   * error should this flag be set to true
   */
  insecureSkipVerify: boolean;
  /** certificate authority. Optional */
  certAuthority: string | undefined;
  /** a reference to an existing secret that contains custom CA */
  secretRef?: SecretKeyReference | undefined;
}

/** SecretKeyReference */
export interface SecretKeyReference {
  /**
   * The name of an existing secret in the pod namespace containing
   * authentication credentials for the package repository.
   * - For HTTP/S basic auth the secret must be of type
   *   "kubernetes.io/basic-auth" and contain username and
   *   password fields.
   * - For TLS the secret must be of type "kubernetes.io/tls"
   *   contain a certFile and keyFile, and/or
   *   caCert fields.
   * - For Bearer or Custom Auth, the secret must be opaque, and
   *   the key must be provided
   * - For Docker Registry Credentials (OCI registries) the secret
   *   must of of type "kubernetes.io/dockerconfigjson”
   * For more details, refer to
   * https://kubernetes.io/docs/concepts/configuration/secret/
   */
  name: string;
  /** Optional. Must be provided when name refers to an opaque secret */
  key: string;
}

/**
 * AddPackageRepositoryResponse
 *
 * Response for AddPackageRepositoryRequest
 */
export interface AddPackageRepositoryResponse {}

const baseAddPackageRepositoryRequest: object = {
  name: "",
  description: "",
  namespaceScoped: false,
  type: "",
  url: "",
  interval: 0,
};

export const AddPackageRepositoryRequest = {
  encode(
    message: AddPackageRepositoryRequest,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.context !== undefined) {
      Context.encode(message.context, writer.uint32(10).fork()).ldelim();
    }
    if (message.name !== "") {
      writer.uint32(18).string(message.name);
    }
    if (message.description !== "") {
      writer.uint32(26).string(message.description);
    }
    if (message.namespaceScoped === true) {
      writer.uint32(32).bool(message.namespaceScoped);
    }
    if (message.type !== "") {
      writer.uint32(42).string(message.type);
    }
    if (message.url !== "") {
      writer.uint32(50).string(message.url);
    }
    if (message.interval !== 0) {
      writer.uint32(56).uint32(message.interval);
    }
    if (message.tlsConfig !== undefined) {
      PackageRepositoryTlsConfig.encode(
        message.tlsConfig,
        writer.uint32(66).fork()
      ).ldelim();
    }
    if (message.plugin !== undefined) {
      Plugin.encode(message.plugin, writer.uint32(82).fork()).ldelim();
    }
    return writer;
  },

  decode(
    input: _m0.Reader | Uint8Array,
    length?: number
  ): AddPackageRepositoryRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseAddPackageRepositoryRequest,
    } as AddPackageRepositoryRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.context = Context.decode(reader, reader.uint32());
          break;
        case 2:
          message.name = reader.string();
          break;
        case 3:
          message.description = reader.string();
          break;
        case 4:
          message.namespaceScoped = reader.bool();
          break;
        case 5:
          message.type = reader.string();
          break;
        case 6:
          message.url = reader.string();
          break;
        case 7:
          message.interval = reader.uint32();
          break;
        case 8:
          message.tlsConfig = PackageRepositoryTlsConfig.decode(
            reader,
            reader.uint32()
          );
          break;
        case 10:
          message.plugin = Plugin.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): AddPackageRepositoryRequest {
    const message = {
      ...baseAddPackageRepositoryRequest,
    } as AddPackageRepositoryRequest;
    if (object.context !== undefined && object.context !== null) {
      message.context = Context.fromJSON(object.context);
    } else {
      message.context = undefined;
    }
    if (object.name !== undefined && object.name !== null) {
      message.name = String(object.name);
    } else {
      message.name = "";
    }
    if (object.description !== undefined && object.description !== null) {
      message.description = String(object.description);
    } else {
      message.description = "";
    }
    if (
      object.namespaceScoped !== undefined &&
      object.namespaceScoped !== null
    ) {
      message.namespaceScoped = Boolean(object.namespaceScoped);
    } else {
      message.namespaceScoped = false;
    }
    if (object.type !== undefined && object.type !== null) {
      message.type = String(object.type);
    } else {
      message.type = "";
    }
    if (object.url !== undefined && object.url !== null) {
      message.url = String(object.url);
    } else {
      message.url = "";
    }
    if (object.interval !== undefined && object.interval !== null) {
      message.interval = Number(object.interval);
    } else {
      message.interval = 0;
    }
    if (object.tlsConfig !== undefined && object.tlsConfig !== null) {
      message.tlsConfig = PackageRepositoryTlsConfig.fromJSON(object.tlsConfig);
    } else {
      message.tlsConfig = undefined;
    }
    if (object.plugin !== undefined && object.plugin !== null) {
      message.plugin = Plugin.fromJSON(object.plugin);
    } else {
      message.plugin = undefined;
    }
    return message;
  },

  toJSON(message: AddPackageRepositoryRequest): unknown {
    const obj: any = {};
    message.context !== undefined &&
      (obj.context = message.context
        ? Context.toJSON(message.context)
        : undefined);
    message.name !== undefined && (obj.name = message.name);
    message.description !== undefined &&
      (obj.description = message.description);
    message.namespaceScoped !== undefined &&
      (obj.namespaceScoped = message.namespaceScoped);
    message.type !== undefined && (obj.type = message.type);
    message.url !== undefined && (obj.url = message.url);
    message.interval !== undefined && (obj.interval = message.interval);
    message.tlsConfig !== undefined &&
      (obj.tlsConfig = message.tlsConfig
        ? PackageRepositoryTlsConfig.toJSON(message.tlsConfig)
        : undefined);
    message.plugin !== undefined &&
      (obj.plugin = message.plugin ? Plugin.toJSON(message.plugin) : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<AddPackageRepositoryRequest>
  ): AddPackageRepositoryRequest {
    const message = {
      ...baseAddPackageRepositoryRequest,
    } as AddPackageRepositoryRequest;
    if (object.context !== undefined && object.context !== null) {
      message.context = Context.fromPartial(object.context);
    } else {
      message.context = undefined;
    }
    if (object.name !== undefined && object.name !== null) {
      message.name = object.name;
    } else {
      message.name = "";
    }
    if (object.description !== undefined && object.description !== null) {
      message.description = object.description;
    } else {
      message.description = "";
    }
    if (
      object.namespaceScoped !== undefined &&
      object.namespaceScoped !== null
    ) {
      message.namespaceScoped = object.namespaceScoped;
    } else {
      message.namespaceScoped = false;
    }
    if (object.type !== undefined && object.type !== null) {
      message.type = object.type;
    } else {
      message.type = "";
    }
    if (object.url !== undefined && object.url !== null) {
      message.url = object.url;
    } else {
      message.url = "";
    }
    if (object.interval !== undefined && object.interval !== null) {
      message.interval = object.interval;
    } else {
      message.interval = 0;
    }
    if (object.tlsConfig !== undefined && object.tlsConfig !== null) {
      message.tlsConfig = PackageRepositoryTlsConfig.fromPartial(
        object.tlsConfig
      );
    } else {
      message.tlsConfig = undefined;
    }
    if (object.plugin !== undefined && object.plugin !== null) {
      message.plugin = Plugin.fromPartial(object.plugin);
    } else {
      message.plugin = undefined;
    }
    return message;
  },
};

const basePackageRepositoryTlsConfig: object = { insecureSkipVerify: false };

export const PackageRepositoryTlsConfig = {
  encode(
    message: PackageRepositoryTlsConfig,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.insecureSkipVerify === true) {
      writer.uint32(8).bool(message.insecureSkipVerify);
    }
    if (message.certAuthority !== undefined) {
      writer.uint32(18).string(message.certAuthority);
    }
    if (message.secretRef !== undefined) {
      SecretKeyReference.encode(
        message.secretRef,
        writer.uint32(26).fork()
      ).ldelim();
    }
    return writer;
  },

  decode(
    input: _m0.Reader | Uint8Array,
    length?: number
  ): PackageRepositoryTlsConfig {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...basePackageRepositoryTlsConfig,
    } as PackageRepositoryTlsConfig;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.insecureSkipVerify = reader.bool();
          break;
        case 2:
          message.certAuthority = reader.string();
          break;
        case 3:
          message.secretRef = SecretKeyReference.decode(
            reader,
            reader.uint32()
          );
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): PackageRepositoryTlsConfig {
    const message = {
      ...basePackageRepositoryTlsConfig,
    } as PackageRepositoryTlsConfig;
    if (
      object.insecureSkipVerify !== undefined &&
      object.insecureSkipVerify !== null
    ) {
      message.insecureSkipVerify = Boolean(object.insecureSkipVerify);
    } else {
      message.insecureSkipVerify = false;
    }
    if (object.certAuthority !== undefined && object.certAuthority !== null) {
      message.certAuthority = String(object.certAuthority);
    } else {
      message.certAuthority = undefined;
    }
    if (object.secretRef !== undefined && object.secretRef !== null) {
      message.secretRef = SecretKeyReference.fromJSON(object.secretRef);
    } else {
      message.secretRef = undefined;
    }
    return message;
  },

  toJSON(message: PackageRepositoryTlsConfig): unknown {
    const obj: any = {};
    message.insecureSkipVerify !== undefined &&
      (obj.insecureSkipVerify = message.insecureSkipVerify);
    message.certAuthority !== undefined &&
      (obj.certAuthority = message.certAuthority);
    message.secretRef !== undefined &&
      (obj.secretRef = message.secretRef
        ? SecretKeyReference.toJSON(message.secretRef)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<PackageRepositoryTlsConfig>
  ): PackageRepositoryTlsConfig {
    const message = {
      ...basePackageRepositoryTlsConfig,
    } as PackageRepositoryTlsConfig;
    if (
      object.insecureSkipVerify !== undefined &&
      object.insecureSkipVerify !== null
    ) {
      message.insecureSkipVerify = object.insecureSkipVerify;
    } else {
      message.insecureSkipVerify = false;
    }
    if (object.certAuthority !== undefined && object.certAuthority !== null) {
      message.certAuthority = object.certAuthority;
    } else {
      message.certAuthority = undefined;
    }
    if (object.secretRef !== undefined && object.secretRef !== null) {
      message.secretRef = SecretKeyReference.fromPartial(object.secretRef);
    } else {
      message.secretRef = undefined;
    }
    return message;
  },
};

const baseSecretKeyReference: object = { name: "", key: "" };

export const SecretKeyReference = {
  encode(
    message: SecretKeyReference,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.name !== "") {
      writer.uint32(10).string(message.name);
    }
    if (message.key !== "") {
      writer.uint32(18).string(message.key);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): SecretKeyReference {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseSecretKeyReference } as SecretKeyReference;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.name = reader.string();
          break;
        case 2:
          message.key = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): SecretKeyReference {
    const message = { ...baseSecretKeyReference } as SecretKeyReference;
    if (object.name !== undefined && object.name !== null) {
      message.name = String(object.name);
    } else {
      message.name = "";
    }
    if (object.key !== undefined && object.key !== null) {
      message.key = String(object.key);
    } else {
      message.key = "";
    }
    return message;
  },

  toJSON(message: SecretKeyReference): unknown {
    const obj: any = {};
    message.name !== undefined && (obj.name = message.name);
    message.key !== undefined && (obj.key = message.key);
    return obj;
  },

  fromPartial(object: DeepPartial<SecretKeyReference>): SecretKeyReference {
    const message = { ...baseSecretKeyReference } as SecretKeyReference;
    if (object.name !== undefined && object.name !== null) {
      message.name = object.name;
    } else {
      message.name = "";
    }
    if (object.key !== undefined && object.key !== null) {
      message.key = object.key;
    } else {
      message.key = "";
    }
    return message;
  },
};

const baseAddPackageRepositoryResponse: object = {};

export const AddPackageRepositoryResponse = {
  encode(
    _: AddPackageRepositoryResponse,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    return writer;
  },

  decode(
    input: _m0.Reader | Uint8Array,
    length?: number
  ): AddPackageRepositoryResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseAddPackageRepositoryResponse,
    } as AddPackageRepositoryResponse;
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

  fromJSON(_: any): AddPackageRepositoryResponse {
    const message = {
      ...baseAddPackageRepositoryResponse,
    } as AddPackageRepositoryResponse;
    return message;
  },

  toJSON(_: AddPackageRepositoryResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(
    _: DeepPartial<AddPackageRepositoryResponse>
  ): AddPackageRepositoryResponse {
    const message = {
      ...baseAddPackageRepositoryResponse,
    } as AddPackageRepositoryResponse;
    return message;
  },
};

/** Each repositories v1alpha1 plugin must implement at least the following rpcs: */
export interface RepositoriesService {
  AddPackageRepository(
    request: DeepPartial<AddPackageRepositoryRequest>,
    metadata?: grpc.Metadata
  ): Promise<AddPackageRepositoryResponse>;
}

export class RepositoriesServiceClientImpl implements RepositoriesService {
  private readonly rpc: Rpc;

  constructor(rpc: Rpc) {
    this.rpc = rpc;
    this.AddPackageRepository = this.AddPackageRepository.bind(this);
  }

  AddPackageRepository(
    request: DeepPartial<AddPackageRepositoryRequest>,
    metadata?: grpc.Metadata
  ): Promise<AddPackageRepositoryResponse> {
    return this.rpc.unary(
      RepositoriesServiceAddPackageRepositoryDesc,
      AddPackageRepositoryRequest.fromPartial(request),
      metadata
    );
  }
}

export const RepositoriesServiceDesc = {
  serviceName: "kubeappsapis.core.packages.v1alpha1.RepositoriesService",
};

export const RepositoriesServiceAddPackageRepositoryDesc: UnaryMethodDefinitionish =
  {
    methodName: "AddPackageRepository",
    service: RepositoriesServiceDesc,
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

interface UnaryMethodDefinitionishR
  extends grpc.UnaryMethodDefinition<any, any> {
  requestStream: any;
  responseStream: any;
}

type UnaryMethodDefinitionish = UnaryMethodDefinitionishR;

interface Rpc {
  unary<T extends UnaryMethodDefinitionish>(
    methodDesc: T,
    request: any,
    metadata: grpc.Metadata | undefined
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
    }
  ) {
    this.host = host;
    this.options = options;
  }

  unary<T extends UnaryMethodDefinitionish>(
    methodDesc: T,
    _request: any,
    metadata: grpc.Metadata | undefined
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

type Builtin =
  | Date
  | Function
  | Uint8Array
  | string
  | number
  | boolean
  | undefined;
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
