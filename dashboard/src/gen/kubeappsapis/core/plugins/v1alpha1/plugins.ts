/* eslint-disable */
import { grpc } from "@improbable-eng/grpc-web";
import { BrowserHeaders } from "browser-headers";
import _m0 from "protobufjs/minimal";

export const protobufPackage = "kubeappsapis.core.plugins.v1alpha1";

/**
 * GetConfiguredPluginsRequest
 *
 * Request for GetConfiguredPlugins
 */
export interface GetConfiguredPluginsRequest {}

/**
 * GetConfiguredPluginsResponse
 *
 * Response for GetConfiguredPlugins
 */
export interface GetConfiguredPluginsResponse {
  /**
   * Plugins
   *
   * List of Plugin
   */
  plugins: Plugin[];
}

/**
 * Plugin
 *
 * A plugin can implement multiple services and multiple versions of a service.
 */
export interface Plugin {
  /**
   * Plugin name
   *
   * The name of the plugin, such as `fluxv2.packages` or `kapp_controller.packages`.
   */
  name: string;
  /**
   * Plugin version
   *
   * The version of the plugin, such as v1alpha1
   */
  version: string;
}

function createBaseGetConfiguredPluginsRequest(): GetConfiguredPluginsRequest {
  return {};
}

export const GetConfiguredPluginsRequest = {
  encode(_: GetConfiguredPluginsRequest, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetConfiguredPluginsRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseGetConfiguredPluginsRequest();
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

  fromJSON(_: any): GetConfiguredPluginsRequest {
    return {};
  },

  toJSON(_: GetConfiguredPluginsRequest): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<GetConfiguredPluginsRequest>, I>>(
    _: I,
  ): GetConfiguredPluginsRequest {
    const message = createBaseGetConfiguredPluginsRequest();
    return message;
  },
};

function createBaseGetConfiguredPluginsResponse(): GetConfiguredPluginsResponse {
  return { plugins: [] };
}

export const GetConfiguredPluginsResponse = {
  encode(
    message: GetConfiguredPluginsResponse,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    for (const v of message.plugins) {
      Plugin.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetConfiguredPluginsResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseGetConfiguredPluginsResponse();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.plugins.push(Plugin.decode(reader, reader.uint32()));
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): GetConfiguredPluginsResponse {
    return {
      plugins: Array.isArray(object?.plugins)
        ? object.plugins.map((e: any) => Plugin.fromJSON(e))
        : [],
    };
  },

  toJSON(message: GetConfiguredPluginsResponse): unknown {
    const obj: any = {};
    if (message.plugins) {
      obj.plugins = message.plugins.map(e => (e ? Plugin.toJSON(e) : undefined));
    } else {
      obj.plugins = [];
    }
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<GetConfiguredPluginsResponse>, I>>(
    object: I,
  ): GetConfiguredPluginsResponse {
    const message = createBaseGetConfiguredPluginsResponse();
    message.plugins = object.plugins?.map(e => Plugin.fromPartial(e)) || [];
    return message;
  },
};

function createBasePlugin(): Plugin {
  return { name: "", version: "" };
}

export const Plugin = {
  encode(message: Plugin, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.name !== "") {
      writer.uint32(10).string(message.name);
    }
    if (message.version !== "") {
      writer.uint32(18).string(message.version);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): Plugin {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBasePlugin();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.name = reader.string();
          break;
        case 2:
          message.version = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): Plugin {
    return {
      name: isSet(object.name) ? String(object.name) : "",
      version: isSet(object.version) ? String(object.version) : "",
    };
  },

  toJSON(message: Plugin): unknown {
    const obj: any = {};
    message.name !== undefined && (obj.name = message.name);
    message.version !== undefined && (obj.version = message.version);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<Plugin>, I>>(object: I): Plugin {
    const message = createBasePlugin();
    message.name = object.name ?? "";
    message.version = object.version ?? "";
    return message;
  },
};

export interface PluginsService {
  /** GetConfiguredPlugins returns a map of short and longnames for the configured plugins. */
  GetConfiguredPlugins(
    request: DeepPartial<GetConfiguredPluginsRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetConfiguredPluginsResponse>;
}

export class PluginsServiceClientImpl implements PluginsService {
  private readonly rpc: Rpc;

  constructor(rpc: Rpc) {
    this.rpc = rpc;
    this.GetConfiguredPlugins = this.GetConfiguredPlugins.bind(this);
  }

  GetConfiguredPlugins(
    request: DeepPartial<GetConfiguredPluginsRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetConfiguredPluginsResponse> {
    return this.rpc.unary(
      PluginsServiceGetConfiguredPluginsDesc,
      GetConfiguredPluginsRequest.fromPartial(request),
      metadata,
    );
  }
}

export const PluginsServiceDesc = {
  serviceName: "kubeappsapis.core.plugins.v1alpha1.PluginsService",
};

export const PluginsServiceGetConfiguredPluginsDesc: UnaryMethodDefinitionish = {
  methodName: "GetConfiguredPlugins",
  service: PluginsServiceDesc,
  requestStream: false,
  responseStream: false,
  requestType: {
    serializeBinary() {
      return GetConfiguredPluginsRequest.encode(this).finish();
    },
  } as any,
  responseType: {
    deserializeBinary(data: Uint8Array) {
      return {
        ...GetConfiguredPluginsResponse.decode(data),
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
    upStreamRetryCodes?: number[];
  };

  constructor(
    host: string,
    options: {
      transport?: grpc.TransportFactory;

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

function isSet(value: any): boolean {
  return value !== null && value !== undefined;
}

export class GrpcWebError extends Error {
  constructor(message: string, public code: grpc.Code, public metadata: grpc.Metadata) {
    super(message);
  }
}
