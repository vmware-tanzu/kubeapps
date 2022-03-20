/* eslint-disable */
import Long from "long";
import { grpc } from "@improbable-eng/grpc-web";
import _m0 from "protobufjs/minimal";
import { BrowserHeaders } from "browser-headers";

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

const baseGetConfiguredPluginsRequest: object = {};

export const GetConfiguredPluginsRequest = {
  encode(_: GetConfiguredPluginsRequest, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetConfiguredPluginsRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseGetConfiguredPluginsRequest,
    } as GetConfiguredPluginsRequest;
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
    const message = {
      ...baseGetConfiguredPluginsRequest,
    } as GetConfiguredPluginsRequest;
    return message;
  },

  toJSON(_: GetConfiguredPluginsRequest): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(_: DeepPartial<GetConfiguredPluginsRequest>): GetConfiguredPluginsRequest {
    const message = {
      ...baseGetConfiguredPluginsRequest,
    } as GetConfiguredPluginsRequest;
    return message;
  },
};

const baseGetConfiguredPluginsResponse: object = {};

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
    const message = {
      ...baseGetConfiguredPluginsResponse,
    } as GetConfiguredPluginsResponse;
    message.plugins = [];
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
    const message = {
      ...baseGetConfiguredPluginsResponse,
    } as GetConfiguredPluginsResponse;
    message.plugins = [];
    if (object.plugins !== undefined && object.plugins !== null) {
      for (const e of object.plugins) {
        message.plugins.push(Plugin.fromJSON(e));
      }
    }
    return message;
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

  fromPartial(object: DeepPartial<GetConfiguredPluginsResponse>): GetConfiguredPluginsResponse {
    const message = {
      ...baseGetConfiguredPluginsResponse,
    } as GetConfiguredPluginsResponse;
    message.plugins = [];
    if (object.plugins !== undefined && object.plugins !== null) {
      for (const e of object.plugins) {
        message.plugins.push(Plugin.fromPartial(e));
      }
    }
    return message;
  },
};

const basePlugin: object = { name: "", version: "" };

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
    const message = { ...basePlugin } as Plugin;
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
    const message = { ...basePlugin } as Plugin;
    if (object.name !== undefined && object.name !== null) {
      message.name = String(object.name);
    } else {
      message.name = "";
    }
    if (object.version !== undefined && object.version !== null) {
      message.version = String(object.version);
    } else {
      message.version = "";
    }
    return message;
  },

  toJSON(message: Plugin): unknown {
    const obj: any = {};
    message.name !== undefined && (obj.name = message.name);
    message.version !== undefined && (obj.version = message.version);
    return obj;
  },

  fromPartial(object: DeepPartial<Plugin>): Plugin {
    const message = { ...basePlugin } as Plugin;
    if (object.name !== undefined && object.name !== null) {
      message.name = object.name;
    } else {
      message.name = "";
    }
    if (object.version !== undefined && object.version !== null) {
      message.version = object.version;
    } else {
      message.version = "";
    }
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
