/* eslint-disable */
import Long from "long";
import _m0 from "protobufjs/minimal";
import { ResourceRef } from "../../../../kubeappsapis/core/packages/v1alpha1/packages";
import { Any } from "../../../../google/protobuf/any";

export const protobufPackage = "kubeappsapis.core.resources.v1alpha1";

export interface Resource {
  resourceRef?: ResourceRef;
  manifest?: Any;
}

const baseResource: object = {};

export const Resource = {
  encode(message: Resource, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.resourceRef !== undefined) {
      ResourceRef.encode(message.resourceRef, writer.uint32(10).fork()).ldelim();
    }
    if (message.manifest !== undefined) {
      Any.encode(message.manifest, writer.uint32(18).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): Resource {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseResource } as Resource;
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

  fromJSON(object: any): Resource {
    const message = { ...baseResource } as Resource;
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

  toJSON(message: Resource): unknown {
    const obj: any = {};
    message.resourceRef !== undefined &&
      (obj.resourceRef = message.resourceRef ? ResourceRef.toJSON(message.resourceRef) : undefined);
    message.manifest !== undefined &&
      (obj.manifest = message.manifest ? Any.toJSON(message.manifest) : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<Resource>): Resource {
    const message = { ...baseResource } as Resource;
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
