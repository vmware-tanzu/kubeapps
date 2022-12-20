/* eslint-disable */
import { grpc } from "@improbable-eng/grpc-web";
import { BrowserHeaders } from "browser-headers";
import Long from "long";
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
  GetPackageRepositoryPermissionsRequest,
  GetPackageRepositoryPermissionsResponse,
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
  /** the query options for the proxy call */
  proxyOptions?: ProxyOptions;
  /** selector which must be true for the pod to fit on a node */
  nodeSelector: { [key: string]: string };
  /** set of Pod's Tolerations */
  tolerations: Toleration[];
  /** defines the security options the container should be run with. */
  securityContext?: PodSecurityContext;
}

export interface HelmPackageRepositoryCustomDetail_NodeSelectorEntry {
  key: string;
  value: string;
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

/**
 * ProxyOptions
 *
 * query options for a proxy call
 */
export interface ProxyOptions {
  /** if true, the proxy options will be taken into account */
  enabled: boolean;
  /** value for the HTTP_PROXY env variable passed to the Pod */
  httpProxy: string;
  /** value for the HTTPS_PROXY env variable passed to the Pod */
  httpsProxy: string;
  /** value for the NO_PROXY env variable passed to the Pod */
  noProxy: string;
}

/**
 * Toleration
 *
 * Extracted from the K8s API to avoid a dependency on the K8s API
 * https://github.com/kubernetes/api/blob/master/core/v1/generated.proto
 */
export interface Toleration {
  key?: string | undefined;
  operator?: string | undefined;
  value?: string | undefined;
  effect?: string | undefined;
  tolerationSeconds?: number | undefined;
}

/**
 * PodSecurityContext
 *
 * Extracted from the K8s API to avoid a dependency on the K8s API
 * https://github.com/kubernetes/api/blob/master/core/v1/generated.proto
 */
export interface PodSecurityContext {
  runAsUser?: number | undefined;
  runAsGroup?: number | undefined;
  runAsNonRoot?: boolean | undefined;
  supplementalGroups: number[];
  fSGroup?: number | undefined;
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
    proxyOptions: undefined,
    nodeSelector: {},
    tolerations: [],
    securityContext: undefined,
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
    if (message.proxyOptions !== undefined) {
      ProxyOptions.encode(message.proxyOptions, writer.uint32(42).fork()).ldelim();
    }
    Object.entries(message.nodeSelector).forEach(([key, value]) => {
      HelmPackageRepositoryCustomDetail_NodeSelectorEntry.encode(
        { key: key as any, value },
        writer.uint32(50).fork(),
      ).ldelim();
    });
    for (const v of message.tolerations) {
      Toleration.encode(v!, writer.uint32(58).fork()).ldelim();
    }
    if (message.securityContext !== undefined) {
      PodSecurityContext.encode(message.securityContext, writer.uint32(66).fork()).ldelim();
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
        case 5:
          message.proxyOptions = ProxyOptions.decode(reader, reader.uint32());
          break;
        case 6:
          const entry6 = HelmPackageRepositoryCustomDetail_NodeSelectorEntry.decode(
            reader,
            reader.uint32(),
          );
          if (entry6.value !== undefined) {
            message.nodeSelector[entry6.key] = entry6.value;
          }
          break;
        case 7:
          message.tolerations.push(Toleration.decode(reader, reader.uint32()));
          break;
        case 8:
          message.securityContext = PodSecurityContext.decode(reader, reader.uint32());
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
      proxyOptions: isSet(object.proxyOptions)
        ? ProxyOptions.fromJSON(object.proxyOptions)
        : undefined,
      nodeSelector: isObject(object.nodeSelector)
        ? Object.entries(object.nodeSelector).reduce<{ [key: string]: string }>(
            (acc, [key, value]) => {
              acc[key] = String(value);
              return acc;
            },
            {},
          )
        : {},
      tolerations: Array.isArray(object?.tolerations)
        ? object.tolerations.map((e: any) => Toleration.fromJSON(e))
        : [],
      securityContext: isSet(object.securityContext)
        ? PodSecurityContext.fromJSON(object.securityContext)
        : undefined,
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
    message.proxyOptions !== undefined &&
      (obj.proxyOptions = message.proxyOptions
        ? ProxyOptions.toJSON(message.proxyOptions)
        : undefined);
    obj.nodeSelector = {};
    if (message.nodeSelector) {
      Object.entries(message.nodeSelector).forEach(([k, v]) => {
        obj.nodeSelector[k] = v;
      });
    }
    if (message.tolerations) {
      obj.tolerations = message.tolerations.map(e => (e ? Toleration.toJSON(e) : undefined));
    } else {
      obj.tolerations = [];
    }
    message.securityContext !== undefined &&
      (obj.securityContext = message.securityContext
        ? PodSecurityContext.toJSON(message.securityContext)
        : undefined);
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
    message.proxyOptions =
      object.proxyOptions !== undefined && object.proxyOptions !== null
        ? ProxyOptions.fromPartial(object.proxyOptions)
        : undefined;
    message.nodeSelector = Object.entries(object.nodeSelector ?? {}).reduce<{
      [key: string]: string;
    }>((acc, [key, value]) => {
      if (value !== undefined) {
        acc[key] = String(value);
      }
      return acc;
    }, {});
    message.tolerations = object.tolerations?.map(e => Toleration.fromPartial(e)) || [];
    message.securityContext =
      object.securityContext !== undefined && object.securityContext !== null
        ? PodSecurityContext.fromPartial(object.securityContext)
        : undefined;
    return message;
  },
};

function createBaseHelmPackageRepositoryCustomDetail_NodeSelectorEntry(): HelmPackageRepositoryCustomDetail_NodeSelectorEntry {
  return { key: "", value: "" };
}

export const HelmPackageRepositoryCustomDetail_NodeSelectorEntry = {
  encode(
    message: HelmPackageRepositoryCustomDetail_NodeSelectorEntry,
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

  decode(
    input: _m0.Reader | Uint8Array,
    length?: number,
  ): HelmPackageRepositoryCustomDetail_NodeSelectorEntry {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseHelmPackageRepositoryCustomDetail_NodeSelectorEntry();
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

  fromJSON(object: any): HelmPackageRepositoryCustomDetail_NodeSelectorEntry {
    return {
      key: isSet(object.key) ? String(object.key) : "",
      value: isSet(object.value) ? String(object.value) : "",
    };
  },

  toJSON(message: HelmPackageRepositoryCustomDetail_NodeSelectorEntry): unknown {
    const obj: any = {};
    message.key !== undefined && (obj.key = message.key);
    message.value !== undefined && (obj.value = message.value);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<HelmPackageRepositoryCustomDetail_NodeSelectorEntry>, I>>(
    object: I,
  ): HelmPackageRepositoryCustomDetail_NodeSelectorEntry {
    const message = createBaseHelmPackageRepositoryCustomDetail_NodeSelectorEntry();
    message.key = object.key ?? "";
    message.value = object.value ?? "";
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

function createBaseProxyOptions(): ProxyOptions {
  return { enabled: false, httpProxy: "", httpsProxy: "", noProxy: "" };
}

export const ProxyOptions = {
  encode(message: ProxyOptions, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.enabled === true) {
      writer.uint32(8).bool(message.enabled);
    }
    if (message.httpProxy !== "") {
      writer.uint32(18).string(message.httpProxy);
    }
    if (message.httpsProxy !== "") {
      writer.uint32(26).string(message.httpsProxy);
    }
    if (message.noProxy !== "") {
      writer.uint32(34).string(message.noProxy);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): ProxyOptions {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseProxyOptions();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.enabled = reader.bool();
          break;
        case 2:
          message.httpProxy = reader.string();
          break;
        case 3:
          message.httpsProxy = reader.string();
          break;
        case 4:
          message.noProxy = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): ProxyOptions {
    return {
      enabled: isSet(object.enabled) ? Boolean(object.enabled) : false,
      httpProxy: isSet(object.httpProxy) ? String(object.httpProxy) : "",
      httpsProxy: isSet(object.httpsProxy) ? String(object.httpsProxy) : "",
      noProxy: isSet(object.noProxy) ? String(object.noProxy) : "",
    };
  },

  toJSON(message: ProxyOptions): unknown {
    const obj: any = {};
    message.enabled !== undefined && (obj.enabled = message.enabled);
    message.httpProxy !== undefined && (obj.httpProxy = message.httpProxy);
    message.httpsProxy !== undefined && (obj.httpsProxy = message.httpsProxy);
    message.noProxy !== undefined && (obj.noProxy = message.noProxy);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<ProxyOptions>, I>>(object: I): ProxyOptions {
    const message = createBaseProxyOptions();
    message.enabled = object.enabled ?? false;
    message.httpProxy = object.httpProxy ?? "";
    message.httpsProxy = object.httpsProxy ?? "";
    message.noProxy = object.noProxy ?? "";
    return message;
  },
};

function createBaseToleration(): Toleration {
  return {
    key: undefined,
    operator: undefined,
    value: undefined,
    effect: undefined,
    tolerationSeconds: undefined,
  };
}

export const Toleration = {
  encode(message: Toleration, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.key !== undefined) {
      writer.uint32(10).string(message.key);
    }
    if (message.operator !== undefined) {
      writer.uint32(18).string(message.operator);
    }
    if (message.value !== undefined) {
      writer.uint32(26).string(message.value);
    }
    if (message.effect !== undefined) {
      writer.uint32(34).string(message.effect);
    }
    if (message.tolerationSeconds !== undefined) {
      writer.uint32(40).int64(message.tolerationSeconds);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): Toleration {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseToleration();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.key = reader.string();
          break;
        case 2:
          message.operator = reader.string();
          break;
        case 3:
          message.value = reader.string();
          break;
        case 4:
          message.effect = reader.string();
          break;
        case 5:
          message.tolerationSeconds = longToNumber(reader.int64() as Long);
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): Toleration {
    return {
      key: isSet(object.key) ? String(object.key) : undefined,
      operator: isSet(object.operator) ? String(object.operator) : undefined,
      value: isSet(object.value) ? String(object.value) : undefined,
      effect: isSet(object.effect) ? String(object.effect) : undefined,
      tolerationSeconds: isSet(object.tolerationSeconds)
        ? Number(object.tolerationSeconds)
        : undefined,
    };
  },

  toJSON(message: Toleration): unknown {
    const obj: any = {};
    message.key !== undefined && (obj.key = message.key);
    message.operator !== undefined && (obj.operator = message.operator);
    message.value !== undefined && (obj.value = message.value);
    message.effect !== undefined && (obj.effect = message.effect);
    message.tolerationSeconds !== undefined &&
      (obj.tolerationSeconds = Math.round(message.tolerationSeconds));
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<Toleration>, I>>(object: I): Toleration {
    const message = createBaseToleration();
    message.key = object.key ?? undefined;
    message.operator = object.operator ?? undefined;
    message.value = object.value ?? undefined;
    message.effect = object.effect ?? undefined;
    message.tolerationSeconds = object.tolerationSeconds ?? undefined;
    return message;
  },
};

function createBasePodSecurityContext(): PodSecurityContext {
  return {
    runAsUser: undefined,
    runAsGroup: undefined,
    runAsNonRoot: undefined,
    supplementalGroups: [],
    fSGroup: undefined,
  };
}

export const PodSecurityContext = {
  encode(message: PodSecurityContext, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.runAsUser !== undefined) {
      writer.uint32(8).int64(message.runAsUser);
    }
    if (message.runAsGroup !== undefined) {
      writer.uint32(48).int64(message.runAsGroup);
    }
    if (message.runAsNonRoot !== undefined) {
      writer.uint32(24).bool(message.runAsNonRoot);
    }
    writer.uint32(34).fork();
    for (const v of message.supplementalGroups) {
      writer.int64(v);
    }
    writer.ldelim();
    if (message.fSGroup !== undefined) {
      writer.uint32(40).int64(message.fSGroup);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): PodSecurityContext {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBasePodSecurityContext();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.runAsUser = longToNumber(reader.int64() as Long);
          break;
        case 6:
          message.runAsGroup = longToNumber(reader.int64() as Long);
          break;
        case 3:
          message.runAsNonRoot = reader.bool();
          break;
        case 4:
          if ((tag & 7) === 2) {
            const end2 = reader.uint32() + reader.pos;
            while (reader.pos < end2) {
              message.supplementalGroups.push(longToNumber(reader.int64() as Long));
            }
          } else {
            message.supplementalGroups.push(longToNumber(reader.int64() as Long));
          }
          break;
        case 5:
          message.fSGroup = longToNumber(reader.int64() as Long);
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): PodSecurityContext {
    return {
      runAsUser: isSet(object.runAsUser) ? Number(object.runAsUser) : undefined,
      runAsGroup: isSet(object.runAsGroup) ? Number(object.runAsGroup) : undefined,
      runAsNonRoot: isSet(object.runAsNonRoot) ? Boolean(object.runAsNonRoot) : undefined,
      supplementalGroups: Array.isArray(object?.supplementalGroups)
        ? object.supplementalGroups.map((e: any) => Number(e))
        : [],
      fSGroup: isSet(object.fSGroup) ? Number(object.fSGroup) : undefined,
    };
  },

  toJSON(message: PodSecurityContext): unknown {
    const obj: any = {};
    message.runAsUser !== undefined && (obj.runAsUser = Math.round(message.runAsUser));
    message.runAsGroup !== undefined && (obj.runAsGroup = Math.round(message.runAsGroup));
    message.runAsNonRoot !== undefined && (obj.runAsNonRoot = message.runAsNonRoot);
    if (message.supplementalGroups) {
      obj.supplementalGroups = message.supplementalGroups.map(e => Math.round(e));
    } else {
      obj.supplementalGroups = [];
    }
    message.fSGroup !== undefined && (obj.fSGroup = Math.round(message.fSGroup));
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<PodSecurityContext>, I>>(object: I): PodSecurityContext {
    const message = createBasePodSecurityContext();
    message.runAsUser = object.runAsUser ?? undefined;
    message.runAsGroup = object.runAsGroup ?? undefined;
    message.runAsNonRoot = object.runAsNonRoot ?? undefined;
    message.supplementalGroups = object.supplementalGroups?.map(e => e) || [];
    message.fSGroup = object.fSGroup ?? undefined;
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
  GetPackageRepositoryPermissions(
    request: DeepPartial<GetPackageRepositoryPermissionsRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetPackageRepositoryPermissionsResponse>;
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
    this.GetPackageRepositoryPermissions = this.GetPackageRepositoryPermissions.bind(this);
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

  GetPackageRepositoryPermissions(
    request: DeepPartial<GetPackageRepositoryPermissionsRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetPackageRepositoryPermissionsResponse> {
    return this.rpc.unary(
      HelmRepositoriesServiceGetPackageRepositoryPermissionsDesc,
      GetPackageRepositoryPermissionsRequest.fromPartial(request),
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

export const HelmRepositoriesServiceGetPackageRepositoryPermissionsDesc: UnaryMethodDefinitionish =
  {
    methodName: "GetPackageRepositoryPermissions",
    service: HelmRepositoriesServiceDesc,
    requestStream: false,
    responseStream: false,
    requestType: {
      serializeBinary() {
        return GetPackageRepositoryPermissionsRequest.encode(this).finish();
      },
    } as any,
    responseType: {
      deserializeBinary(data: Uint8Array) {
        return {
          ...GetPackageRepositoryPermissionsResponse.decode(data),
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

declare var self: any | undefined;
declare var window: any | undefined;
declare var global: any | undefined;
var globalThis: any = (() => {
  if (typeof globalThis !== "undefined") {
    return globalThis;
  }
  if (typeof self !== "undefined") {
    return self;
  }
  if (typeof window !== "undefined") {
    return window;
  }
  if (typeof global !== "undefined") {
    return global;
  }
  throw "Unable to locate global object";
})();

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

function longToNumber(long: Long): number {
  if (long.gt(Number.MAX_SAFE_INTEGER)) {
    throw new globalThis.Error("Value is larger than Number.MAX_SAFE_INTEGER");
  }
  return long.toNumber();
}

if (_m0.util.Long !== Long) {
  _m0.util.Long = Long as any;
  _m0.configure();
}

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
