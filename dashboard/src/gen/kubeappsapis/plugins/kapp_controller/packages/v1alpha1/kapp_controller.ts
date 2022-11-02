/* eslint-disable */
import { grpc } from "@improbable-eng/grpc-web";
import { BrowserHeaders } from "browser-headers";
import _m0 from "protobufjs/minimal";
import {
  CreateInstalledPackageRequest,
  CreateInstalledPackageResponse,
  DeleteInstalledPackageRequest,
  DeleteInstalledPackageResponse,
  GetAvailablePackageDetailRequest,
  GetAvailablePackageDetailResponse,
  GetAvailablePackageSummariesRequest,
  GetAvailablePackageSummariesResponse,
  GetAvailablePackageVersionsRequest,
  GetAvailablePackageVersionsResponse,
  GetInstalledPackageDetailRequest,
  GetInstalledPackageDetailResponse,
  GetInstalledPackageResourceRefsRequest,
  GetInstalledPackageResourceRefsResponse,
  GetInstalledPackageSummariesRequest,
  GetInstalledPackageSummariesResponse,
  UpdateInstalledPackageRequest,
  UpdateInstalledPackageResponse,
} from "../../../../core/packages/v1alpha1/packages";
import {
  AddPackageRepositoryRequest,
  AddPackageRepositoryResponse,
  DeletePackageRepositoryRequest,
  DeletePackageRepositoryResponse,
  GetPackageRepositoryDetailRequest,
  GetPackageRepositoryDetailResponse,
  GetPackageRepositorySummariesRequest,
  GetPackageRepositorySummariesResponse,
  UpdatePackageRepositoryRequest,
  UpdatePackageRepositoryResponse,
} from "../../../../core/packages/v1alpha1/repositories";

export const protobufPackage = "kubeappsapis.plugins.kapp_controller.packages.v1alpha1";

/**
 * KappControllerPackageRepositoryCustomDetail
 *
 * custom fields to support other carvel repository types
 * this is mirror from https://github.com/vmware-tanzu/carvel-kapp-controller/blob/develop/pkg/apis/kappctrl/v1alpha1/generated.proto
 * todo -> find a way to define those messages by referencing proto files from kapp_controller rather than duplication
 */
export interface KappControllerPackageRepositoryCustomDetail {
  fetch?: PackageRepositoryFetch;
}

export interface PackageRepositoryFetch {
  imgpkgBundle?: PackageRepositoryImgpkg;
  image?: PackageRepositoryImage;
  git?: PackageRepositoryGit;
  http?: PackageRepositoryHttp;
  inline?: PackageRepositoryInline;
}

export interface PackageRepositoryImgpkg {
  tagSelection?: VersionSelection;
}

export interface PackageRepositoryImage {
  tagSelection?: VersionSelection;
  subPath: string;
}

export interface PackageRepositoryGit {
  ref: string;
  refSelection?: VersionSelection;
  subPath: string;
  lfsSkipSmudge: boolean;
}

export interface PackageRepositoryHttp {
  subPath: string;
  sha256: string;
}

export interface PackageRepositoryInline {
  paths: { [key: string]: string };
  pathsFrom: PackageRepositoryInline_Source[];
}

export interface PackageRepositoryInline_SourceRef {
  name: string;
  directoryPath: string;
}

export interface PackageRepositoryInline_Source {
  secretRef?: PackageRepositoryInline_SourceRef;
  configMapRef?: PackageRepositoryInline_SourceRef;
}

export interface PackageRepositoryInline_PathsEntry {
  key: string;
  value: string;
}

export interface VersionSelection {
  semver?: VersionSelectionSemver;
}

export interface VersionSelectionSemver {
  constraints: string;
  prereleases?: VersionSelectionSemverPrereleases;
}

export interface VersionSelectionSemverPrereleases {
  identifiers: string[];
}

function createBaseKappControllerPackageRepositoryCustomDetail(): KappControllerPackageRepositoryCustomDetail {
  return { fetch: undefined };
}

export const KappControllerPackageRepositoryCustomDetail = {
  encode(
    message: KappControllerPackageRepositoryCustomDetail,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.fetch !== undefined) {
      PackageRepositoryFetch.encode(message.fetch, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(
    input: _m0.Reader | Uint8Array,
    length?: number,
  ): KappControllerPackageRepositoryCustomDetail {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseKappControllerPackageRepositoryCustomDetail();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.fetch = PackageRepositoryFetch.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): KappControllerPackageRepositoryCustomDetail {
    return {
      fetch: isSet(object.fetch) ? PackageRepositoryFetch.fromJSON(object.fetch) : undefined,
    };
  },

  toJSON(message: KappControllerPackageRepositoryCustomDetail): unknown {
    const obj: any = {};
    message.fetch !== undefined &&
      (obj.fetch = message.fetch ? PackageRepositoryFetch.toJSON(message.fetch) : undefined);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<KappControllerPackageRepositoryCustomDetail>, I>>(
    object: I,
  ): KappControllerPackageRepositoryCustomDetail {
    const message = createBaseKappControllerPackageRepositoryCustomDetail();
    message.fetch =
      object.fetch !== undefined && object.fetch !== null
        ? PackageRepositoryFetch.fromPartial(object.fetch)
        : undefined;
    return message;
  },
};

function createBasePackageRepositoryFetch(): PackageRepositoryFetch {
  return {
    imgpkgBundle: undefined,
    image: undefined,
    git: undefined,
    http: undefined,
    inline: undefined,
  };
}

export const PackageRepositoryFetch = {
  encode(message: PackageRepositoryFetch, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.imgpkgBundle !== undefined) {
      PackageRepositoryImgpkg.encode(message.imgpkgBundle, writer.uint32(10).fork()).ldelim();
    }
    if (message.image !== undefined) {
      PackageRepositoryImage.encode(message.image, writer.uint32(18).fork()).ldelim();
    }
    if (message.git !== undefined) {
      PackageRepositoryGit.encode(message.git, writer.uint32(26).fork()).ldelim();
    }
    if (message.http !== undefined) {
      PackageRepositoryHttp.encode(message.http, writer.uint32(34).fork()).ldelim();
    }
    if (message.inline !== undefined) {
      PackageRepositoryInline.encode(message.inline, writer.uint32(42).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): PackageRepositoryFetch {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBasePackageRepositoryFetch();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.imgpkgBundle = PackageRepositoryImgpkg.decode(reader, reader.uint32());
          break;
        case 2:
          message.image = PackageRepositoryImage.decode(reader, reader.uint32());
          break;
        case 3:
          message.git = PackageRepositoryGit.decode(reader, reader.uint32());
          break;
        case 4:
          message.http = PackageRepositoryHttp.decode(reader, reader.uint32());
          break;
        case 5:
          message.inline = PackageRepositoryInline.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): PackageRepositoryFetch {
    return {
      imgpkgBundle: isSet(object.imgpkgBundle)
        ? PackageRepositoryImgpkg.fromJSON(object.imgpkgBundle)
        : undefined,
      image: isSet(object.image) ? PackageRepositoryImage.fromJSON(object.image) : undefined,
      git: isSet(object.git) ? PackageRepositoryGit.fromJSON(object.git) : undefined,
      http: isSet(object.http) ? PackageRepositoryHttp.fromJSON(object.http) : undefined,
      inline: isSet(object.inline) ? PackageRepositoryInline.fromJSON(object.inline) : undefined,
    };
  },

  toJSON(message: PackageRepositoryFetch): unknown {
    const obj: any = {};
    message.imgpkgBundle !== undefined &&
      (obj.imgpkgBundle = message.imgpkgBundle
        ? PackageRepositoryImgpkg.toJSON(message.imgpkgBundle)
        : undefined);
    message.image !== undefined &&
      (obj.image = message.image ? PackageRepositoryImage.toJSON(message.image) : undefined);
    message.git !== undefined &&
      (obj.git = message.git ? PackageRepositoryGit.toJSON(message.git) : undefined);
    message.http !== undefined &&
      (obj.http = message.http ? PackageRepositoryHttp.toJSON(message.http) : undefined);
    message.inline !== undefined &&
      (obj.inline = message.inline ? PackageRepositoryInline.toJSON(message.inline) : undefined);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<PackageRepositoryFetch>, I>>(
    object: I,
  ): PackageRepositoryFetch {
    const message = createBasePackageRepositoryFetch();
    message.imgpkgBundle =
      object.imgpkgBundle !== undefined && object.imgpkgBundle !== null
        ? PackageRepositoryImgpkg.fromPartial(object.imgpkgBundle)
        : undefined;
    message.image =
      object.image !== undefined && object.image !== null
        ? PackageRepositoryImage.fromPartial(object.image)
        : undefined;
    message.git =
      object.git !== undefined && object.git !== null
        ? PackageRepositoryGit.fromPartial(object.git)
        : undefined;
    message.http =
      object.http !== undefined && object.http !== null
        ? PackageRepositoryHttp.fromPartial(object.http)
        : undefined;
    message.inline =
      object.inline !== undefined && object.inline !== null
        ? PackageRepositoryInline.fromPartial(object.inline)
        : undefined;
    return message;
  },
};

function createBasePackageRepositoryImgpkg(): PackageRepositoryImgpkg {
  return { tagSelection: undefined };
}

export const PackageRepositoryImgpkg = {
  encode(message: PackageRepositoryImgpkg, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.tagSelection !== undefined) {
      VersionSelection.encode(message.tagSelection, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): PackageRepositoryImgpkg {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBasePackageRepositoryImgpkg();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.tagSelection = VersionSelection.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): PackageRepositoryImgpkg {
    return {
      tagSelection: isSet(object.tagSelection)
        ? VersionSelection.fromJSON(object.tagSelection)
        : undefined,
    };
  },

  toJSON(message: PackageRepositoryImgpkg): unknown {
    const obj: any = {};
    message.tagSelection !== undefined &&
      (obj.tagSelection = message.tagSelection
        ? VersionSelection.toJSON(message.tagSelection)
        : undefined);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<PackageRepositoryImgpkg>, I>>(
    object: I,
  ): PackageRepositoryImgpkg {
    const message = createBasePackageRepositoryImgpkg();
    message.tagSelection =
      object.tagSelection !== undefined && object.tagSelection !== null
        ? VersionSelection.fromPartial(object.tagSelection)
        : undefined;
    return message;
  },
};

function createBasePackageRepositoryImage(): PackageRepositoryImage {
  return { tagSelection: undefined, subPath: "" };
}

export const PackageRepositoryImage = {
  encode(message: PackageRepositoryImage, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.tagSelection !== undefined) {
      VersionSelection.encode(message.tagSelection, writer.uint32(10).fork()).ldelim();
    }
    if (message.subPath !== "") {
      writer.uint32(18).string(message.subPath);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): PackageRepositoryImage {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBasePackageRepositoryImage();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.tagSelection = VersionSelection.decode(reader, reader.uint32());
          break;
        case 2:
          message.subPath = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): PackageRepositoryImage {
    return {
      tagSelection: isSet(object.tagSelection)
        ? VersionSelection.fromJSON(object.tagSelection)
        : undefined,
      subPath: isSet(object.subPath) ? String(object.subPath) : "",
    };
  },

  toJSON(message: PackageRepositoryImage): unknown {
    const obj: any = {};
    message.tagSelection !== undefined &&
      (obj.tagSelection = message.tagSelection
        ? VersionSelection.toJSON(message.tagSelection)
        : undefined);
    message.subPath !== undefined && (obj.subPath = message.subPath);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<PackageRepositoryImage>, I>>(
    object: I,
  ): PackageRepositoryImage {
    const message = createBasePackageRepositoryImage();
    message.tagSelection =
      object.tagSelection !== undefined && object.tagSelection !== null
        ? VersionSelection.fromPartial(object.tagSelection)
        : undefined;
    message.subPath = object.subPath ?? "";
    return message;
  },
};

function createBasePackageRepositoryGit(): PackageRepositoryGit {
  return { ref: "", refSelection: undefined, subPath: "", lfsSkipSmudge: false };
}

export const PackageRepositoryGit = {
  encode(message: PackageRepositoryGit, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.ref !== "") {
      writer.uint32(10).string(message.ref);
    }
    if (message.refSelection !== undefined) {
      VersionSelection.encode(message.refSelection, writer.uint32(18).fork()).ldelim();
    }
    if (message.subPath !== "") {
      writer.uint32(26).string(message.subPath);
    }
    if (message.lfsSkipSmudge === true) {
      writer.uint32(32).bool(message.lfsSkipSmudge);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): PackageRepositoryGit {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBasePackageRepositoryGit();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.ref = reader.string();
          break;
        case 2:
          message.refSelection = VersionSelection.decode(reader, reader.uint32());
          break;
        case 3:
          message.subPath = reader.string();
          break;
        case 4:
          message.lfsSkipSmudge = reader.bool();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): PackageRepositoryGit {
    return {
      ref: isSet(object.ref) ? String(object.ref) : "",
      refSelection: isSet(object.refSelection)
        ? VersionSelection.fromJSON(object.refSelection)
        : undefined,
      subPath: isSet(object.subPath) ? String(object.subPath) : "",
      lfsSkipSmudge: isSet(object.lfsSkipSmudge) ? Boolean(object.lfsSkipSmudge) : false,
    };
  },

  toJSON(message: PackageRepositoryGit): unknown {
    const obj: any = {};
    message.ref !== undefined && (obj.ref = message.ref);
    message.refSelection !== undefined &&
      (obj.refSelection = message.refSelection
        ? VersionSelection.toJSON(message.refSelection)
        : undefined);
    message.subPath !== undefined && (obj.subPath = message.subPath);
    message.lfsSkipSmudge !== undefined && (obj.lfsSkipSmudge = message.lfsSkipSmudge);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<PackageRepositoryGit>, I>>(
    object: I,
  ): PackageRepositoryGit {
    const message = createBasePackageRepositoryGit();
    message.ref = object.ref ?? "";
    message.refSelection =
      object.refSelection !== undefined && object.refSelection !== null
        ? VersionSelection.fromPartial(object.refSelection)
        : undefined;
    message.subPath = object.subPath ?? "";
    message.lfsSkipSmudge = object.lfsSkipSmudge ?? false;
    return message;
  },
};

function createBasePackageRepositoryHttp(): PackageRepositoryHttp {
  return { subPath: "", sha256: "" };
}

export const PackageRepositoryHttp = {
  encode(message: PackageRepositoryHttp, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.subPath !== "") {
      writer.uint32(10).string(message.subPath);
    }
    if (message.sha256 !== "") {
      writer.uint32(18).string(message.sha256);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): PackageRepositoryHttp {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBasePackageRepositoryHttp();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.subPath = reader.string();
          break;
        case 2:
          message.sha256 = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): PackageRepositoryHttp {
    return {
      subPath: isSet(object.subPath) ? String(object.subPath) : "",
      sha256: isSet(object.sha256) ? String(object.sha256) : "",
    };
  },

  toJSON(message: PackageRepositoryHttp): unknown {
    const obj: any = {};
    message.subPath !== undefined && (obj.subPath = message.subPath);
    message.sha256 !== undefined && (obj.sha256 = message.sha256);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<PackageRepositoryHttp>, I>>(
    object: I,
  ): PackageRepositoryHttp {
    const message = createBasePackageRepositoryHttp();
    message.subPath = object.subPath ?? "";
    message.sha256 = object.sha256 ?? "";
    return message;
  },
};

function createBasePackageRepositoryInline(): PackageRepositoryInline {
  return { paths: {}, pathsFrom: [] };
}

export const PackageRepositoryInline = {
  encode(message: PackageRepositoryInline, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    Object.entries(message.paths).forEach(([key, value]) => {
      PackageRepositoryInline_PathsEntry.encode(
        { key: key as any, value },
        writer.uint32(10).fork(),
      ).ldelim();
    });
    for (const v of message.pathsFrom) {
      PackageRepositoryInline_Source.encode(v!, writer.uint32(18).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): PackageRepositoryInline {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBasePackageRepositoryInline();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          const entry1 = PackageRepositoryInline_PathsEntry.decode(reader, reader.uint32());
          if (entry1.value !== undefined) {
            message.paths[entry1.key] = entry1.value;
          }
          break;
        case 2:
          message.pathsFrom.push(PackageRepositoryInline_Source.decode(reader, reader.uint32()));
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): PackageRepositoryInline {
    return {
      paths: isObject(object.paths)
        ? Object.entries(object.paths).reduce<{ [key: string]: string }>((acc, [key, value]) => {
            acc[key] = String(value);
            return acc;
          }, {})
        : {},
      pathsFrom: Array.isArray(object?.pathsFrom)
        ? object.pathsFrom.map((e: any) => PackageRepositoryInline_Source.fromJSON(e))
        : [],
    };
  },

  toJSON(message: PackageRepositoryInline): unknown {
    const obj: any = {};
    obj.paths = {};
    if (message.paths) {
      Object.entries(message.paths).forEach(([k, v]) => {
        obj.paths[k] = v;
      });
    }
    if (message.pathsFrom) {
      obj.pathsFrom = message.pathsFrom.map(e =>
        e ? PackageRepositoryInline_Source.toJSON(e) : undefined,
      );
    } else {
      obj.pathsFrom = [];
    }
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<PackageRepositoryInline>, I>>(
    object: I,
  ): PackageRepositoryInline {
    const message = createBasePackageRepositoryInline();
    message.paths = Object.entries(object.paths ?? {}).reduce<{ [key: string]: string }>(
      (acc, [key, value]) => {
        if (value !== undefined) {
          acc[key] = String(value);
        }
        return acc;
      },
      {},
    );
    message.pathsFrom =
      object.pathsFrom?.map(e => PackageRepositoryInline_Source.fromPartial(e)) || [];
    return message;
  },
};

function createBasePackageRepositoryInline_SourceRef(): PackageRepositoryInline_SourceRef {
  return { name: "", directoryPath: "" };
}

export const PackageRepositoryInline_SourceRef = {
  encode(
    message: PackageRepositoryInline_SourceRef,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.name !== "") {
      writer.uint32(10).string(message.name);
    }
    if (message.directoryPath !== "") {
      writer.uint32(18).string(message.directoryPath);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): PackageRepositoryInline_SourceRef {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBasePackageRepositoryInline_SourceRef();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.name = reader.string();
          break;
        case 2:
          message.directoryPath = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): PackageRepositoryInline_SourceRef {
    return {
      name: isSet(object.name) ? String(object.name) : "",
      directoryPath: isSet(object.directoryPath) ? String(object.directoryPath) : "",
    };
  },

  toJSON(message: PackageRepositoryInline_SourceRef): unknown {
    const obj: any = {};
    message.name !== undefined && (obj.name = message.name);
    message.directoryPath !== undefined && (obj.directoryPath = message.directoryPath);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<PackageRepositoryInline_SourceRef>, I>>(
    object: I,
  ): PackageRepositoryInline_SourceRef {
    const message = createBasePackageRepositoryInline_SourceRef();
    message.name = object.name ?? "";
    message.directoryPath = object.directoryPath ?? "";
    return message;
  },
};

function createBasePackageRepositoryInline_Source(): PackageRepositoryInline_Source {
  return { secretRef: undefined, configMapRef: undefined };
}

export const PackageRepositoryInline_Source = {
  encode(
    message: PackageRepositoryInline_Source,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.secretRef !== undefined) {
      PackageRepositoryInline_SourceRef.encode(
        message.secretRef,
        writer.uint32(10).fork(),
      ).ldelim();
    }
    if (message.configMapRef !== undefined) {
      PackageRepositoryInline_SourceRef.encode(
        message.configMapRef,
        writer.uint32(18).fork(),
      ).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): PackageRepositoryInline_Source {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBasePackageRepositoryInline_Source();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.secretRef = PackageRepositoryInline_SourceRef.decode(reader, reader.uint32());
          break;
        case 2:
          message.configMapRef = PackageRepositoryInline_SourceRef.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): PackageRepositoryInline_Source {
    return {
      secretRef: isSet(object.secretRef)
        ? PackageRepositoryInline_SourceRef.fromJSON(object.secretRef)
        : undefined,
      configMapRef: isSet(object.configMapRef)
        ? PackageRepositoryInline_SourceRef.fromJSON(object.configMapRef)
        : undefined,
    };
  },

  toJSON(message: PackageRepositoryInline_Source): unknown {
    const obj: any = {};
    message.secretRef !== undefined &&
      (obj.secretRef = message.secretRef
        ? PackageRepositoryInline_SourceRef.toJSON(message.secretRef)
        : undefined);
    message.configMapRef !== undefined &&
      (obj.configMapRef = message.configMapRef
        ? PackageRepositoryInline_SourceRef.toJSON(message.configMapRef)
        : undefined);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<PackageRepositoryInline_Source>, I>>(
    object: I,
  ): PackageRepositoryInline_Source {
    const message = createBasePackageRepositoryInline_Source();
    message.secretRef =
      object.secretRef !== undefined && object.secretRef !== null
        ? PackageRepositoryInline_SourceRef.fromPartial(object.secretRef)
        : undefined;
    message.configMapRef =
      object.configMapRef !== undefined && object.configMapRef !== null
        ? PackageRepositoryInline_SourceRef.fromPartial(object.configMapRef)
        : undefined;
    return message;
  },
};

function createBasePackageRepositoryInline_PathsEntry(): PackageRepositoryInline_PathsEntry {
  return { key: "", value: "" };
}

export const PackageRepositoryInline_PathsEntry = {
  encode(
    message: PackageRepositoryInline_PathsEntry,
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

  decode(input: _m0.Reader | Uint8Array, length?: number): PackageRepositoryInline_PathsEntry {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBasePackageRepositoryInline_PathsEntry();
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

  fromJSON(object: any): PackageRepositoryInline_PathsEntry {
    return {
      key: isSet(object.key) ? String(object.key) : "",
      value: isSet(object.value) ? String(object.value) : "",
    };
  },

  toJSON(message: PackageRepositoryInline_PathsEntry): unknown {
    const obj: any = {};
    message.key !== undefined && (obj.key = message.key);
    message.value !== undefined && (obj.value = message.value);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<PackageRepositoryInline_PathsEntry>, I>>(
    object: I,
  ): PackageRepositoryInline_PathsEntry {
    const message = createBasePackageRepositoryInline_PathsEntry();
    message.key = object.key ?? "";
    message.value = object.value ?? "";
    return message;
  },
};

function createBaseVersionSelection(): VersionSelection {
  return { semver: undefined };
}

export const VersionSelection = {
  encode(message: VersionSelection, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.semver !== undefined) {
      VersionSelectionSemver.encode(message.semver, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): VersionSelection {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseVersionSelection();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.semver = VersionSelectionSemver.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): VersionSelection {
    return {
      semver: isSet(object.semver) ? VersionSelectionSemver.fromJSON(object.semver) : undefined,
    };
  },

  toJSON(message: VersionSelection): unknown {
    const obj: any = {};
    message.semver !== undefined &&
      (obj.semver = message.semver ? VersionSelectionSemver.toJSON(message.semver) : undefined);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<VersionSelection>, I>>(object: I): VersionSelection {
    const message = createBaseVersionSelection();
    message.semver =
      object.semver !== undefined && object.semver !== null
        ? VersionSelectionSemver.fromPartial(object.semver)
        : undefined;
    return message;
  },
};

function createBaseVersionSelectionSemver(): VersionSelectionSemver {
  return { constraints: "", prereleases: undefined };
}

export const VersionSelectionSemver = {
  encode(message: VersionSelectionSemver, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.constraints !== "") {
      writer.uint32(10).string(message.constraints);
    }
    if (message.prereleases !== undefined) {
      VersionSelectionSemverPrereleases.encode(
        message.prereleases,
        writer.uint32(18).fork(),
      ).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): VersionSelectionSemver {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseVersionSelectionSemver();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.constraints = reader.string();
          break;
        case 2:
          message.prereleases = VersionSelectionSemverPrereleases.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): VersionSelectionSemver {
    return {
      constraints: isSet(object.constraints) ? String(object.constraints) : "",
      prereleases: isSet(object.prereleases)
        ? VersionSelectionSemverPrereleases.fromJSON(object.prereleases)
        : undefined,
    };
  },

  toJSON(message: VersionSelectionSemver): unknown {
    const obj: any = {};
    message.constraints !== undefined && (obj.constraints = message.constraints);
    message.prereleases !== undefined &&
      (obj.prereleases = message.prereleases
        ? VersionSelectionSemverPrereleases.toJSON(message.prereleases)
        : undefined);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<VersionSelectionSemver>, I>>(
    object: I,
  ): VersionSelectionSemver {
    const message = createBaseVersionSelectionSemver();
    message.constraints = object.constraints ?? "";
    message.prereleases =
      object.prereleases !== undefined && object.prereleases !== null
        ? VersionSelectionSemverPrereleases.fromPartial(object.prereleases)
        : undefined;
    return message;
  },
};

function createBaseVersionSelectionSemverPrereleases(): VersionSelectionSemverPrereleases {
  return { identifiers: [] };
}

export const VersionSelectionSemverPrereleases = {
  encode(
    message: VersionSelectionSemverPrereleases,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    for (const v of message.identifiers) {
      writer.uint32(10).string(v!);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): VersionSelectionSemverPrereleases {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseVersionSelectionSemverPrereleases();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.identifiers.push(reader.string());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): VersionSelectionSemverPrereleases {
    return {
      identifiers: Array.isArray(object?.identifiers)
        ? object.identifiers.map((e: any) => String(e))
        : [],
    };
  },

  toJSON(message: VersionSelectionSemverPrereleases): unknown {
    const obj: any = {};
    if (message.identifiers) {
      obj.identifiers = message.identifiers.map(e => e);
    } else {
      obj.identifiers = [];
    }
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<VersionSelectionSemverPrereleases>, I>>(
    object: I,
  ): VersionSelectionSemverPrereleases {
    const message = createBaseVersionSelectionSemverPrereleases();
    message.identifiers = object.identifiers?.map(e => e) || [];
    return message;
  },
};

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
