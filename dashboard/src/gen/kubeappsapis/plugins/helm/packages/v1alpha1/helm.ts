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
  InstalledPackageReference,
  UpdateInstalledPackageRequest,
  UpdateInstalledPackageResponse,
} from "../../../../core/packages/v1alpha1/packages";
import {
  AddPackageRepositoryRequest,
  AddPackageRepositoryResponse,
  DeletePackageRepositoryRequest,
  DeletePackageRepositoryResponse,
  DockerCredentials,
  GetPackageRepositoryDetailRequest,
  GetPackageRepositoryDetailResponse,
  GetPackageRepositorySummariesRequest,
  GetPackageRepositorySummariesResponse,
  UpdatePackageRepositoryRequest,
  UpdatePackageRepositoryResponse,
} from "../../../../core/packages/v1alpha1/repositories";

export const protobufPackage = "kubeappsapis.plugins.helm.packages.v1alpha1";

/**
 * InstalledPackageDetailCustomDataHelm
 *
 * InstalledPackageDetailCustomDataHelm is a message type used for the
 * InstalledPackageDetail.CustomDetail field by the helm plugin.
 */
export interface InstalledPackageDetailCustomDataHelm {
  /**
   * ReleaseRevision
   *
   * A number identifying the Helm revision
   */
  releaseRevision: number;
}

export interface RollbackInstalledPackageRequest {
  /**
   * Installed package reference
   *
   * A reference uniquely identifying the installed package.
   */
  installedPackageRef?: InstalledPackageReference;
  /**
   * ReleaseRevision
   *
   * A number identifying the Helm revision to which to rollback.
   */
  releaseRevision: number;
}

/**
 * RollbackInstalledPackageResponse
 *
 * Response for RollbackInstalledPackage
 */
export interface RollbackInstalledPackageResponse {
  /**
   * TODO: add example for API docs
   * option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_schema) = {
   *   example: '{"installed_package_ref": {}}'
   * };
   */
  installedPackageRef?: InstalledPackageReference;
}

export interface SetUserManagedSecretsRequest {
  value: boolean;
}

export interface SetUserManagedSecretsResponse {
  value: boolean;
}

export interface ImagesPullSecret {
  /** docker credentials secret reference */
  secretRef: string | undefined;
  /** docker credentials data */
  credentials?: DockerCredentials | undefined;
}

/**
 * HelmPackageRepositoryCustomDetail
 *
 * Custom details for a Helm repository
 */
export interface HelmPackageRepositoryCustomDetail {
  /** docker registry credentials for pull secrets */
  imagesPullSecret?: ImagesPullSecret;
  /** list of oci repositories */
  ociRepositories: string[];
  /** filter rule to apply to the repository */
  filterRule?: RepositoryFilterRule;
  /** whether to perform validation on the repository */
  performValidation: boolean;
}

/**
 * RepositoryFilterRule
 *
 * JQ expression for filtering packages
 */
export interface RepositoryFilterRule {
  /** jq string expression */
  jq: string;
  /** map of variables */
  variables: { [key: string]: string };
}

export interface RepositoryFilterRule_VariablesEntry {
  key: string;
  value: string;
}

function createBaseInstalledPackageDetailCustomDataHelm(): InstalledPackageDetailCustomDataHelm {
  return { releaseRevision: 0 };
}

export const InstalledPackageDetailCustomDataHelm = {
  encode(
    message: InstalledPackageDetailCustomDataHelm,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.releaseRevision !== 0) {
      writer.uint32(16).int32(message.releaseRevision);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): InstalledPackageDetailCustomDataHelm {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseInstalledPackageDetailCustomDataHelm();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 2:
          message.releaseRevision = reader.int32();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): InstalledPackageDetailCustomDataHelm {
    return { releaseRevision: isSet(object.releaseRevision) ? Number(object.releaseRevision) : 0 };
  },

  toJSON(message: InstalledPackageDetailCustomDataHelm): unknown {
    const obj: any = {};
    message.releaseRevision !== undefined &&
      (obj.releaseRevision = Math.round(message.releaseRevision));
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<InstalledPackageDetailCustomDataHelm>, I>>(
    object: I,
  ): InstalledPackageDetailCustomDataHelm {
    const message = createBaseInstalledPackageDetailCustomDataHelm();
    message.releaseRevision = object.releaseRevision ?? 0;
    return message;
  },
};

function createBaseRollbackInstalledPackageRequest(): RollbackInstalledPackageRequest {
  return { installedPackageRef: undefined, releaseRevision: 0 };
}

export const RollbackInstalledPackageRequest = {
  encode(
    message: RollbackInstalledPackageRequest,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.installedPackageRef !== undefined) {
      InstalledPackageReference.encode(
        message.installedPackageRef,
        writer.uint32(10).fork(),
      ).ldelim();
    }
    if (message.releaseRevision !== 0) {
      writer.uint32(16).int32(message.releaseRevision);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): RollbackInstalledPackageRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseRollbackInstalledPackageRequest();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.installedPackageRef = InstalledPackageReference.decode(reader, reader.uint32());
          break;
        case 2:
          message.releaseRevision = reader.int32();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): RollbackInstalledPackageRequest {
    return {
      installedPackageRef: isSet(object.installedPackageRef)
        ? InstalledPackageReference.fromJSON(object.installedPackageRef)
        : undefined,
      releaseRevision: isSet(object.releaseRevision) ? Number(object.releaseRevision) : 0,
    };
  },

  toJSON(message: RollbackInstalledPackageRequest): unknown {
    const obj: any = {};
    message.installedPackageRef !== undefined &&
      (obj.installedPackageRef = message.installedPackageRef
        ? InstalledPackageReference.toJSON(message.installedPackageRef)
        : undefined);
    message.releaseRevision !== undefined &&
      (obj.releaseRevision = Math.round(message.releaseRevision));
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<RollbackInstalledPackageRequest>, I>>(
    object: I,
  ): RollbackInstalledPackageRequest {
    const message = createBaseRollbackInstalledPackageRequest();
    message.installedPackageRef =
      object.installedPackageRef !== undefined && object.installedPackageRef !== null
        ? InstalledPackageReference.fromPartial(object.installedPackageRef)
        : undefined;
    message.releaseRevision = object.releaseRevision ?? 0;
    return message;
  },
};

function createBaseRollbackInstalledPackageResponse(): RollbackInstalledPackageResponse {
  return { installedPackageRef: undefined };
}

export const RollbackInstalledPackageResponse = {
  encode(
    message: RollbackInstalledPackageResponse,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.installedPackageRef !== undefined) {
      InstalledPackageReference.encode(
        message.installedPackageRef,
        writer.uint32(10).fork(),
      ).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): RollbackInstalledPackageResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseRollbackInstalledPackageResponse();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.installedPackageRef = InstalledPackageReference.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): RollbackInstalledPackageResponse {
    return {
      installedPackageRef: isSet(object.installedPackageRef)
        ? InstalledPackageReference.fromJSON(object.installedPackageRef)
        : undefined,
    };
  },

  toJSON(message: RollbackInstalledPackageResponse): unknown {
    const obj: any = {};
    message.installedPackageRef !== undefined &&
      (obj.installedPackageRef = message.installedPackageRef
        ? InstalledPackageReference.toJSON(message.installedPackageRef)
        : undefined);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<RollbackInstalledPackageResponse>, I>>(
    object: I,
  ): RollbackInstalledPackageResponse {
    const message = createBaseRollbackInstalledPackageResponse();
    message.installedPackageRef =
      object.installedPackageRef !== undefined && object.installedPackageRef !== null
        ? InstalledPackageReference.fromPartial(object.installedPackageRef)
        : undefined;
    return message;
  },
};

function createBaseSetUserManagedSecretsRequest(): SetUserManagedSecretsRequest {
  return { value: false };
}

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
    const message = createBaseSetUserManagedSecretsRequest();
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
    return { value: isSet(object.value) ? Boolean(object.value) : false };
  },

  toJSON(message: SetUserManagedSecretsRequest): unknown {
    const obj: any = {};
    message.value !== undefined && (obj.value = message.value);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<SetUserManagedSecretsRequest>, I>>(
    object: I,
  ): SetUserManagedSecretsRequest {
    const message = createBaseSetUserManagedSecretsRequest();
    message.value = object.value ?? false;
    return message;
  },
};

function createBaseSetUserManagedSecretsResponse(): SetUserManagedSecretsResponse {
  return { value: false };
}

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
    const message = createBaseSetUserManagedSecretsResponse();
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
    return { value: isSet(object.value) ? Boolean(object.value) : false };
  },

  toJSON(message: SetUserManagedSecretsResponse): unknown {
    const obj: any = {};
    message.value !== undefined && (obj.value = message.value);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<SetUserManagedSecretsResponse>, I>>(
    object: I,
  ): SetUserManagedSecretsResponse {
    const message = createBaseSetUserManagedSecretsResponse();
    message.value = object.value ?? false;
    return message;
  },
};

function createBaseImagesPullSecret(): ImagesPullSecret {
  return { secretRef: undefined, credentials: undefined };
}

export const ImagesPullSecret = {
  encode(message: ImagesPullSecret, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.secretRef !== undefined) {
      writer.uint32(10).string(message.secretRef);
    }
    if (message.credentials !== undefined) {
      DockerCredentials.encode(message.credentials, writer.uint32(18).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): ImagesPullSecret {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseImagesPullSecret();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.secretRef = reader.string();
          break;
        case 2:
          message.credentials = DockerCredentials.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): ImagesPullSecret {
    return {
      secretRef: isSet(object.secretRef) ? String(object.secretRef) : undefined,
      credentials: isSet(object.credentials)
        ? DockerCredentials.fromJSON(object.credentials)
        : undefined,
    };
  },

  toJSON(message: ImagesPullSecret): unknown {
    const obj: any = {};
    message.secretRef !== undefined && (obj.secretRef = message.secretRef);
    message.credentials !== undefined &&
      (obj.credentials = message.credentials
        ? DockerCredentials.toJSON(message.credentials)
        : undefined);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<ImagesPullSecret>, I>>(object: I): ImagesPullSecret {
    const message = createBaseImagesPullSecret();
    message.secretRef = object.secretRef ?? undefined;
    message.credentials =
      object.credentials !== undefined && object.credentials !== null
        ? DockerCredentials.fromPartial(object.credentials)
        : undefined;
    return message;
  },
};

function createBaseHelmPackageRepositoryCustomDetail(): HelmPackageRepositoryCustomDetail {
  return {
    imagesPullSecret: undefined,
    ociRepositories: [],
    filterRule: undefined,
    performValidation: false,
  };
}

export const HelmPackageRepositoryCustomDetail = {
  encode(
    message: HelmPackageRepositoryCustomDetail,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.imagesPullSecret !== undefined) {
      ImagesPullSecret.encode(message.imagesPullSecret, writer.uint32(10).fork()).ldelim();
    }
    for (const v of message.ociRepositories) {
      writer.uint32(18).string(v!);
    }
    if (message.filterRule !== undefined) {
      RepositoryFilterRule.encode(message.filterRule, writer.uint32(26).fork()).ldelim();
    }
    if (message.performValidation === true) {
      writer.uint32(32).bool(message.performValidation);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): HelmPackageRepositoryCustomDetail {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseHelmPackageRepositoryCustomDetail();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.imagesPullSecret = ImagesPullSecret.decode(reader, reader.uint32());
          break;
        case 2:
          message.ociRepositories.push(reader.string());
          break;
        case 3:
          message.filterRule = RepositoryFilterRule.decode(reader, reader.uint32());
          break;
        case 4:
          message.performValidation = reader.bool();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): HelmPackageRepositoryCustomDetail {
    return {
      imagesPullSecret: isSet(object.imagesPullSecret)
        ? ImagesPullSecret.fromJSON(object.imagesPullSecret)
        : undefined,
      ociRepositories: Array.isArray(object?.ociRepositories)
        ? object.ociRepositories.map((e: any) => String(e))
        : [],
      filterRule: isSet(object.filterRule)
        ? RepositoryFilterRule.fromJSON(object.filterRule)
        : undefined,
      performValidation: isSet(object.performValidation)
        ? Boolean(object.performValidation)
        : false,
    };
  },

  toJSON(message: HelmPackageRepositoryCustomDetail): unknown {
    const obj: any = {};
    message.imagesPullSecret !== undefined &&
      (obj.imagesPullSecret = message.imagesPullSecret
        ? ImagesPullSecret.toJSON(message.imagesPullSecret)
        : undefined);
    if (message.ociRepositories) {
      obj.ociRepositories = message.ociRepositories.map(e => e);
    } else {
      obj.ociRepositories = [];
    }
    message.filterRule !== undefined &&
      (obj.filterRule = message.filterRule
        ? RepositoryFilterRule.toJSON(message.filterRule)
        : undefined);
    message.performValidation !== undefined && (obj.performValidation = message.performValidation);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<HelmPackageRepositoryCustomDetail>, I>>(
    object: I,
  ): HelmPackageRepositoryCustomDetail {
    const message = createBaseHelmPackageRepositoryCustomDetail();
    message.imagesPullSecret =
      object.imagesPullSecret !== undefined && object.imagesPullSecret !== null
        ? ImagesPullSecret.fromPartial(object.imagesPullSecret)
        : undefined;
    message.ociRepositories = object.ociRepositories?.map(e => e) || [];
    message.filterRule =
      object.filterRule !== undefined && object.filterRule !== null
        ? RepositoryFilterRule.fromPartial(object.filterRule)
        : undefined;
    message.performValidation = object.performValidation ?? false;
    return message;
  },
};

function createBaseRepositoryFilterRule(): RepositoryFilterRule {
  return { jq: "", variables: {} };
}

export const RepositoryFilterRule = {
  encode(message: RepositoryFilterRule, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.jq !== "") {
      writer.uint32(10).string(message.jq);
    }
    Object.entries(message.variables).forEach(([key, value]) => {
      RepositoryFilterRule_VariablesEntry.encode(
        { key: key as any, value },
        writer.uint32(34).fork(),
      ).ldelim();
    });
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): RepositoryFilterRule {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseRepositoryFilterRule();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.jq = reader.string();
          break;
        case 4:
          const entry4 = RepositoryFilterRule_VariablesEntry.decode(reader, reader.uint32());
          if (entry4.value !== undefined) {
            message.variables[entry4.key] = entry4.value;
          }
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): RepositoryFilterRule {
    return {
      jq: isSet(object.jq) ? String(object.jq) : "",
      variables: isObject(object.variables)
        ? Object.entries(object.variables).reduce<{ [key: string]: string }>(
            (acc, [key, value]) => {
              acc[key] = String(value);
              return acc;
            },
            {},
          )
        : {},
    };
  },

  toJSON(message: RepositoryFilterRule): unknown {
    const obj: any = {};
    message.jq !== undefined && (obj.jq = message.jq);
    obj.variables = {};
    if (message.variables) {
      Object.entries(message.variables).forEach(([k, v]) => {
        obj.variables[k] = v;
      });
    }
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<RepositoryFilterRule>, I>>(
    object: I,
  ): RepositoryFilterRule {
    const message = createBaseRepositoryFilterRule();
    message.jq = object.jq ?? "";
    message.variables = Object.entries(object.variables ?? {}).reduce<{ [key: string]: string }>(
      (acc, [key, value]) => {
        if (value !== undefined) {
          acc[key] = String(value);
        }
        return acc;
      },
      {},
    );
    return message;
  },
};

function createBaseRepositoryFilterRule_VariablesEntry(): RepositoryFilterRule_VariablesEntry {
  return { key: "", value: "" };
}

export const RepositoryFilterRule_VariablesEntry = {
  encode(
    message: RepositoryFilterRule_VariablesEntry,
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

  decode(input: _m0.Reader | Uint8Array, length?: number): RepositoryFilterRule_VariablesEntry {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseRepositoryFilterRule_VariablesEntry();
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

  fromJSON(object: any): RepositoryFilterRule_VariablesEntry {
    return {
      key: isSet(object.key) ? String(object.key) : "",
      value: isSet(object.value) ? String(object.value) : "",
    };
  },

  toJSON(message: RepositoryFilterRule_VariablesEntry): unknown {
    const obj: any = {};
    message.key !== undefined && (obj.key = message.key);
    message.value !== undefined && (obj.value = message.value);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<RepositoryFilterRule_VariablesEntry>, I>>(
    object: I,
  ): RepositoryFilterRule_VariablesEntry {
    const message = createBaseRepositoryFilterRule_VariablesEntry();
    message.key = object.key ?? "";
    message.value = object.value ?? "";
    return message;
  },
};

export interface HelmPackagesService {
  /** GetAvailablePackageSummaries returns the available packages managed by the 'helm' plugin */
  GetAvailablePackageSummaries(
    request: DeepPartial<GetAvailablePackageSummariesRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetAvailablePackageSummariesResponse>;
  /** GetAvailablePackageDetail returns the package details managed by the 'helm' plugin */
  GetAvailablePackageDetail(
    request: DeepPartial<GetAvailablePackageDetailRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetAvailablePackageDetailResponse>;
  /** GetAvailablePackageVersions returns the package versions managed by the 'helm' plugin */
  GetAvailablePackageVersions(
    request: DeepPartial<GetAvailablePackageVersionsRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetAvailablePackageVersionsResponse>;
  /** GetInstalledPackageSummaries returns the installed packages managed by the 'helm' plugin */
  GetInstalledPackageSummaries(
    request: DeepPartial<GetInstalledPackageSummariesRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetInstalledPackageSummariesResponse>;
  /** GetInstalledPackageDetail returns the requested installed package managed by the 'helm' plugin */
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
  /** RollbackInstalledPackage updates an installed package based on the request. */
  RollbackInstalledPackage(
    request: DeepPartial<RollbackInstalledPackageRequest>,
    metadata?: grpc.Metadata,
  ): Promise<RollbackInstalledPackageResponse>;
  /**
   * GetInstalledPackageResourceRefs returns the references for the Kubernetes resources created by
   * an installed package.
   */
  GetInstalledPackageResourceRefs(
    request: DeepPartial<GetInstalledPackageResourceRefsRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetInstalledPackageResourceRefsResponse>;
}

export class HelmPackagesServiceClientImpl implements HelmPackagesService {
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
    this.RollbackInstalledPackage = this.RollbackInstalledPackage.bind(this);
    this.GetInstalledPackageResourceRefs = this.GetInstalledPackageResourceRefs.bind(this);
  }

  GetAvailablePackageSummaries(
    request: DeepPartial<GetAvailablePackageSummariesRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetAvailablePackageSummariesResponse> {
    return this.rpc.unary(
      HelmPackagesServiceGetAvailablePackageSummariesDesc,
      GetAvailablePackageSummariesRequest.fromPartial(request),
      metadata,
    );
  }

  GetAvailablePackageDetail(
    request: DeepPartial<GetAvailablePackageDetailRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetAvailablePackageDetailResponse> {
    return this.rpc.unary(
      HelmPackagesServiceGetAvailablePackageDetailDesc,
      GetAvailablePackageDetailRequest.fromPartial(request),
      metadata,
    );
  }

  GetAvailablePackageVersions(
    request: DeepPartial<GetAvailablePackageVersionsRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetAvailablePackageVersionsResponse> {
    return this.rpc.unary(
      HelmPackagesServiceGetAvailablePackageVersionsDesc,
      GetAvailablePackageVersionsRequest.fromPartial(request),
      metadata,
    );
  }

  GetInstalledPackageSummaries(
    request: DeepPartial<GetInstalledPackageSummariesRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetInstalledPackageSummariesResponse> {
    return this.rpc.unary(
      HelmPackagesServiceGetInstalledPackageSummariesDesc,
      GetInstalledPackageSummariesRequest.fromPartial(request),
      metadata,
    );
  }

  GetInstalledPackageDetail(
    request: DeepPartial<GetInstalledPackageDetailRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetInstalledPackageDetailResponse> {
    return this.rpc.unary(
      HelmPackagesServiceGetInstalledPackageDetailDesc,
      GetInstalledPackageDetailRequest.fromPartial(request),
      metadata,
    );
  }

  CreateInstalledPackage(
    request: DeepPartial<CreateInstalledPackageRequest>,
    metadata?: grpc.Metadata,
  ): Promise<CreateInstalledPackageResponse> {
    return this.rpc.unary(
      HelmPackagesServiceCreateInstalledPackageDesc,
      CreateInstalledPackageRequest.fromPartial(request),
      metadata,
    );
  }

  UpdateInstalledPackage(
    request: DeepPartial<UpdateInstalledPackageRequest>,
    metadata?: grpc.Metadata,
  ): Promise<UpdateInstalledPackageResponse> {
    return this.rpc.unary(
      HelmPackagesServiceUpdateInstalledPackageDesc,
      UpdateInstalledPackageRequest.fromPartial(request),
      metadata,
    );
  }

  DeleteInstalledPackage(
    request: DeepPartial<DeleteInstalledPackageRequest>,
    metadata?: grpc.Metadata,
  ): Promise<DeleteInstalledPackageResponse> {
    return this.rpc.unary(
      HelmPackagesServiceDeleteInstalledPackageDesc,
      DeleteInstalledPackageRequest.fromPartial(request),
      metadata,
    );
  }

  RollbackInstalledPackage(
    request: DeepPartial<RollbackInstalledPackageRequest>,
    metadata?: grpc.Metadata,
  ): Promise<RollbackInstalledPackageResponse> {
    return this.rpc.unary(
      HelmPackagesServiceRollbackInstalledPackageDesc,
      RollbackInstalledPackageRequest.fromPartial(request),
      metadata,
    );
  }

  GetInstalledPackageResourceRefs(
    request: DeepPartial<GetInstalledPackageResourceRefsRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetInstalledPackageResourceRefsResponse> {
    return this.rpc.unary(
      HelmPackagesServiceGetInstalledPackageResourceRefsDesc,
      GetInstalledPackageResourceRefsRequest.fromPartial(request),
      metadata,
    );
  }
}

export const HelmPackagesServiceDesc = {
  serviceName: "kubeappsapis.plugins.helm.packages.v1alpha1.HelmPackagesService",
};

export const HelmPackagesServiceGetAvailablePackageSummariesDesc: UnaryMethodDefinitionish = {
  methodName: "GetAvailablePackageSummaries",
  service: HelmPackagesServiceDesc,
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

export const HelmPackagesServiceGetAvailablePackageDetailDesc: UnaryMethodDefinitionish = {
  methodName: "GetAvailablePackageDetail",
  service: HelmPackagesServiceDesc,
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

export const HelmPackagesServiceGetAvailablePackageVersionsDesc: UnaryMethodDefinitionish = {
  methodName: "GetAvailablePackageVersions",
  service: HelmPackagesServiceDesc,
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

export const HelmPackagesServiceGetInstalledPackageSummariesDesc: UnaryMethodDefinitionish = {
  methodName: "GetInstalledPackageSummaries",
  service: HelmPackagesServiceDesc,
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

export const HelmPackagesServiceGetInstalledPackageDetailDesc: UnaryMethodDefinitionish = {
  methodName: "GetInstalledPackageDetail",
  service: HelmPackagesServiceDesc,
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

export const HelmPackagesServiceCreateInstalledPackageDesc: UnaryMethodDefinitionish = {
  methodName: "CreateInstalledPackage",
  service: HelmPackagesServiceDesc,
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

export const HelmPackagesServiceUpdateInstalledPackageDesc: UnaryMethodDefinitionish = {
  methodName: "UpdateInstalledPackage",
  service: HelmPackagesServiceDesc,
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

export const HelmPackagesServiceDeleteInstalledPackageDesc: UnaryMethodDefinitionish = {
  methodName: "DeleteInstalledPackage",
  service: HelmPackagesServiceDesc,
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

export const HelmPackagesServiceRollbackInstalledPackageDesc: UnaryMethodDefinitionish = {
  methodName: "RollbackInstalledPackage",
  service: HelmPackagesServiceDesc,
  requestStream: false,
  responseStream: false,
  requestType: {
    serializeBinary() {
      return RollbackInstalledPackageRequest.encode(this).finish();
    },
  } as any,
  responseType: {
    deserializeBinary(data: Uint8Array) {
      return {
        ...RollbackInstalledPackageResponse.decode(data),
        toObject() {
          return this;
        },
      };
    },
  } as any,
};

export const HelmPackagesServiceGetInstalledPackageResourceRefsDesc: UnaryMethodDefinitionish = {
  methodName: "GetInstalledPackageResourceRefs",
  service: HelmPackagesServiceDesc,
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

export interface HelmRepositoriesService {
  /** AddPackageRepository add an existing package repository to the set of ones already managed by the Helm plugin */
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

export class HelmRepositoriesServiceClientImpl implements HelmRepositoriesService {
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
      HelmRepositoriesServiceAddPackageRepositoryDesc,
      AddPackageRepositoryRequest.fromPartial(request),
      metadata,
    );
  }

  GetPackageRepositoryDetail(
    request: DeepPartial<GetPackageRepositoryDetailRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetPackageRepositoryDetailResponse> {
    return this.rpc.unary(
      HelmRepositoriesServiceGetPackageRepositoryDetailDesc,
      GetPackageRepositoryDetailRequest.fromPartial(request),
      metadata,
    );
  }

  GetPackageRepositorySummaries(
    request: DeepPartial<GetPackageRepositorySummariesRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetPackageRepositorySummariesResponse> {
    return this.rpc.unary(
      HelmRepositoriesServiceGetPackageRepositorySummariesDesc,
      GetPackageRepositorySummariesRequest.fromPartial(request),
      metadata,
    );
  }

  UpdatePackageRepository(
    request: DeepPartial<UpdatePackageRepositoryRequest>,
    metadata?: grpc.Metadata,
  ): Promise<UpdatePackageRepositoryResponse> {
    return this.rpc.unary(
      HelmRepositoriesServiceUpdatePackageRepositoryDesc,
      UpdatePackageRepositoryRequest.fromPartial(request),
      metadata,
    );
  }

  DeletePackageRepository(
    request: DeepPartial<DeletePackageRepositoryRequest>,
    metadata?: grpc.Metadata,
  ): Promise<DeletePackageRepositoryResponse> {
    return this.rpc.unary(
      HelmRepositoriesServiceDeletePackageRepositoryDesc,
      DeletePackageRepositoryRequest.fromPartial(request),
      metadata,
    );
  }

  SetUserManagedSecrets(
    request: DeepPartial<SetUserManagedSecretsRequest>,
    metadata?: grpc.Metadata,
  ): Promise<SetUserManagedSecretsResponse> {
    return this.rpc.unary(
      HelmRepositoriesServiceSetUserManagedSecretsDesc,
      SetUserManagedSecretsRequest.fromPartial(request),
      metadata,
    );
  }
}

export const HelmRepositoriesServiceDesc = {
  serviceName: "kubeappsapis.plugins.helm.packages.v1alpha1.HelmRepositoriesService",
};

export const HelmRepositoriesServiceAddPackageRepositoryDesc: UnaryMethodDefinitionish = {
  methodName: "AddPackageRepository",
  service: HelmRepositoriesServiceDesc,
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

export const HelmRepositoriesServiceGetPackageRepositoryDetailDesc: UnaryMethodDefinitionish = {
  methodName: "GetPackageRepositoryDetail",
  service: HelmRepositoriesServiceDesc,
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

export const HelmRepositoriesServiceGetPackageRepositorySummariesDesc: UnaryMethodDefinitionish = {
  methodName: "GetPackageRepositorySummaries",
  service: HelmRepositoriesServiceDesc,
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

export const HelmRepositoriesServiceUpdatePackageRepositoryDesc: UnaryMethodDefinitionish = {
  methodName: "UpdatePackageRepository",
  service: HelmRepositoriesServiceDesc,
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

export const HelmRepositoriesServiceDeletePackageRepositoryDesc: UnaryMethodDefinitionish = {
  methodName: "DeletePackageRepository",
  service: HelmRepositoriesServiceDesc,
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

export const HelmRepositoriesServiceSetUserManagedSecretsDesc: UnaryMethodDefinitionish = {
  methodName: "SetUserManagedSecrets",
  service: HelmRepositoriesServiceDesc,
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
