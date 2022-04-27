/* eslint-disable */
import Long from "long";
import { grpc } from "@improbable-eng/grpc-web";
import _m0 from "protobufjs/minimal";
import { Context } from "../../../../kubeappsapis/core/packages/v1alpha1/packages";
import { Plugin } from "../../../../kubeappsapis/core/plugins/v1alpha1/plugins";
import { Any } from "../../../../google/protobuf/any";
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
  /** authentication parameters for connecting to a repository. Optional */
  auth?: PackageRepositoryAuth;
  /**
   * The plugin used to interact with this package repository.
   * This field should be omitted when the request is in the context of a
   * specific plugin.
   */
  plugin?: Plugin;
  /**
   * Custom data added by the plugin
   * A plugin can define custom details for data which is not yet, or
   * never will be specified in the core AddPackageRepositoryRequest
   * fields. The use of an `Any` field means that each plugin can define
   * the structure of this message as required, while still satisfying the
   * core interface.
   * See https://developers.google.com/protocol-buffers/docs/proto3#any
   * Just for reference, some of the examples that have been chosen not to
   * be part of the core API but rather plugin-specific details are:
   *   direct-helm:
   *      - image pull secrets
   *      - list of oci repositories
   *      - filter rules
   *      - sync job pod template
   */
  customDetail?: Any;
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

/**
 * PackageRepositoryAuth
 *
 * Authentication/authorization to provide client’s identity when connecting
 * to a package repository.
 * There are 6 total distinct use cases we may support:
 * 1) None (Public)
 * 2) Basic Auth
 * 3) Bearer Token
 * 4) Custom Authorization Header
 * 5) Docker Registry Credentials (for OCI only)
 * 6) TLS certificate and key
 *
 * Note that (1)-(4) may be done over HTTP or HTTPs without any custom
 * certificates or certificate authority
 * (1) is handled by not not having PackageRepositoryAuth field on
 *     the parent object
 * a given plug-in may or may not support a given authentication type.
 * For example
 *  - direct-helm plug-in does not currently support (6), while flux does
 *  - flux plug-in does not support (3) or (4) while direct-helm does
 */
export interface PackageRepositoryAuth {
  type: PackageRepositoryAuth_PackageRepositoryAuthType;
  /** username and plain text password */
  usernamePassword?: UsernamePassword | undefined;
  /** certificate and key for TLS-based authentication */
  tlsCertKey?: TlsCertKey | undefined;
  /** docker credentials */
  dockerCreds?: DockerCredentials | undefined;
  /**
   * for Bearer Auth token value
   * for Custom Auth, complete value of "Authorization" header
   */
  header: string | undefined;
  /** a reference to an existing secret */
  secretRef?: SecretKeyReference | undefined;
  /**
   * pass_credentials allows the credentials from the SecretRef to be passed
   * on to a host that does not match the host as defined in URL.
   * This flag controls whether or not it is allowed to passing credentials
   * with requests to other domains linked from the repository.
   * This may be needed if the host of the advertised chart URLs in the
   * index differs from the defined URL. Optional
   */
  passCredentials: boolean;
}

export enum PackageRepositoryAuth_PackageRepositoryAuthType {
  PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED = 0,
  PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH = 1,
  PACKAGE_REPOSITORY_AUTH_TYPE_TLS = 2,
  PACKAGE_REPOSITORY_AUTH_TYPE_BEARER = 3,
  PACKAGE_REPOSITORY_AUTH_TYPE_CUSTOM = 4,
  PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON = 5,
  UNRECOGNIZED = -1,
}

export function packageRepositoryAuth_PackageRepositoryAuthTypeFromJSON(
  object: any,
): PackageRepositoryAuth_PackageRepositoryAuthType {
  switch (object) {
    case 0:
    case "PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED":
      return PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED;
    case 1:
    case "PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH":
      return PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH;
    case 2:
    case "PACKAGE_REPOSITORY_AUTH_TYPE_TLS":
      return PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_TLS;
    case 3:
    case "PACKAGE_REPOSITORY_AUTH_TYPE_BEARER":
      return PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BEARER;
    case 4:
    case "PACKAGE_REPOSITORY_AUTH_TYPE_CUSTOM":
      return PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_CUSTOM;
    case 5:
    case "PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON":
      return PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON;
    case -1:
    case "UNRECOGNIZED":
    default:
      return PackageRepositoryAuth_PackageRepositoryAuthType.UNRECOGNIZED;
  }
}

export function packageRepositoryAuth_PackageRepositoryAuthTypeToJSON(
  object: PackageRepositoryAuth_PackageRepositoryAuthType,
): string {
  switch (object) {
    case PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED:
      return "PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED";
    case PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH:
      return "PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH";
    case PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_TLS:
      return "PACKAGE_REPOSITORY_AUTH_TYPE_TLS";
    case PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_BEARER:
      return "PACKAGE_REPOSITORY_AUTH_TYPE_BEARER";
    case PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_CUSTOM:
      return "PACKAGE_REPOSITORY_AUTH_TYPE_CUSTOM";
    case PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON:
      return "PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON";
    default:
      return "UNKNOWN";
  }
}

/** UsernamePassword */
export interface UsernamePassword {
  /** In clear text */
  username: string;
  /** In clear text */
  password: string;
}

/** TlsCertKey */
export interface TlsCertKey {
  /** certificate (identity). In clear text */
  cert: string;
  /** certificate key. In clear text */
  key: string;
}

/** DockerCredentials */
export interface DockerCredentials {
  /** server name */
  server: string;
  /** username. */
  username: string;
  /** password. In clear text */
  password: string;
  /** email address */
  email: string;
}

/** SecretKeyReference */
export interface SecretKeyReference {
  /**
   * The name of an existing secret in the same namespace as the object
   * that refers to it (e.g. PackageRepository), containing authentication
   * credentials for the said package repository.
   * - For HTTP/S basic auth the secret must be of type
   *   "kubernetes.io/basic-auth" or opaque and contain username and
   *   password fields
   * - For TLS the secret must be of type "kubernetes.io/tls" or opaque
   *   and contain a certFile and keyFile, and/or
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
 * GetPackageRepositoryDetailRequest
 *
 * Request for GetPackageRepositoryDetail
 */
export interface GetPackageRepositoryDetailRequest {
  packageRepoRef?: PackageRepositoryReference;
}

/**
 * GetPackageRepositorySummariesRequest
 *
 * Request for PackageRepositorySummary
 */
export interface GetPackageRepositorySummariesRequest {
  /** The context (cluster/namespace) for the request. */
  context?: Context;
}

/**
 * UpdatePackageRepositoryRequest
 *
 * Request for UpdatePackageRepository
 */
export interface UpdatePackageRepositoryRequest {
  /**
   * A reference uniquely identifying the package repository being updated.
   * The only required field
   */
  packageRepoRef?: PackageRepositoryReference;
  /** URL identifying the package repository location. */
  url: string;
  /** A user-provided description. */
  description: string;
  /**
   * The interval at which to check the upstream for updates (in seconds)
   * Optional. Defaults to 10m if not specified
   */
  interval: number;
  /** TLS-specific parameters for connecting to a repository. Optional */
  tlsConfig?: PackageRepositoryTlsConfig;
  /** authentication parameters for connecting to a repository. Optional */
  auth?: PackageRepositoryAuth;
  /**
   * Custom data added by the plugin
   * A plugin can define custom details for data which is not yet, or
   * never will be specified in the core AddPackageRepositoryRequest
   * fields. The use of an `Any` field means that each plugin can define
   * the structure of this message as required, while still satisfying the
   * core interface.
   * See https://developers.google.com/protocol-buffers/docs/proto3#any
   * Just for reference, some of the examples that have been chosen not to
   * be part of the core API but rather plugin-specific details are:
   *   direct-helm:
   *      - image pull secrets
   *      - list of oci repositories
   *      - filter rules
   *      - sync job pod template
   */
  customDetail?: Any;
}

/**
 * DeletePackageRepositoryRequest
 *
 * Request for DeletePackageRepository
 */
export interface DeletePackageRepositoryRequest {
  packageRepoRef?: PackageRepositoryReference;
}

/**
 * PackageRepositoryReference
 *
 * A PackageRepositoryReference has the minimum information required to
 * uniquely identify a package repository.
 */
export interface PackageRepositoryReference {
  /** The context (cluster/namespace) for the repository. */
  context?: Context;
  /**
   * The fully qualified identifier for the repository
   * (i.e. a unique name for the context).
   */
  identifier: string;
  /**
   * The plugin used to interact with this available package.
   * This field should be omitted when the request is in the context of a
   * specific plugin.
   */
  plugin?: Plugin;
}

/**
 * AddPackageRepositoryResponse
 *
 * Response for AddPackageRepositoryRequest
 */
export interface AddPackageRepositoryResponse {
  /**
   * TODO: add example for API docs
   * option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_schema) = {
   *   example: '{"package_repo_ref": {}}'
   * };
   */
  packageRepoRef?: PackageRepositoryReference;
}

/**
 * PackageRepositoryStatus
 *
 * A PackageRepositoryStatus reports on the current status of the repository.
 */
export interface PackageRepositoryStatus {
  /**
   * Ready
   *
   * An indication of whether the repository is ready or not
   */
  ready: boolean;
  /**
   * Reason
   *
   * An enum indicating the reason for the current status.
   */
  reason: PackageRepositoryStatus_StatusReason;
  /**
   * UserReason
   *
   * Optional text to return for user context, which may be plugin specific.
   */
  userReason: string;
}

/**
 * StatusReason
 *
 * Generic reasons why a package repository may be ready or not.
 * These should make sense across different packaging plugins.
 */
export enum PackageRepositoryStatus_StatusReason {
  STATUS_REASON_UNSPECIFIED = 0,
  STATUS_REASON_SUCCESS = 1,
  STATUS_REASON_FAILED = 2,
  STATUS_REASON_PENDING = 3,
  UNRECOGNIZED = -1,
}

export function packageRepositoryStatus_StatusReasonFromJSON(
  object: any,
): PackageRepositoryStatus_StatusReason {
  switch (object) {
    case 0:
    case "STATUS_REASON_UNSPECIFIED":
      return PackageRepositoryStatus_StatusReason.STATUS_REASON_UNSPECIFIED;
    case 1:
    case "STATUS_REASON_SUCCESS":
      return PackageRepositoryStatus_StatusReason.STATUS_REASON_SUCCESS;
    case 2:
    case "STATUS_REASON_FAILED":
      return PackageRepositoryStatus_StatusReason.STATUS_REASON_FAILED;
    case 3:
    case "STATUS_REASON_PENDING":
      return PackageRepositoryStatus_StatusReason.STATUS_REASON_PENDING;
    case -1:
    case "UNRECOGNIZED":
    default:
      return PackageRepositoryStatus_StatusReason.UNRECOGNIZED;
  }
}

export function packageRepositoryStatus_StatusReasonToJSON(
  object: PackageRepositoryStatus_StatusReason,
): string {
  switch (object) {
    case PackageRepositoryStatus_StatusReason.STATUS_REASON_UNSPECIFIED:
      return "STATUS_REASON_UNSPECIFIED";
    case PackageRepositoryStatus_StatusReason.STATUS_REASON_SUCCESS:
      return "STATUS_REASON_SUCCESS";
    case PackageRepositoryStatus_StatusReason.STATUS_REASON_FAILED:
      return "STATUS_REASON_FAILED";
    case PackageRepositoryStatus_StatusReason.STATUS_REASON_PENDING:
      return "STATUS_REASON_PENDING";
    default:
      return "UNKNOWN";
  }
}

/** PackageRepositoryDetail */
export interface PackageRepositoryDetail {
  /** A reference uniquely identifying the package repository. */
  packageRepoRef?: PackageRepositoryReference;
  /** A user-provided name for the package repository (e.g. bitnami) */
  name: string;
  /** A user-provided description. */
  description: string;
  /** Whether this repository is global or namespace-scoped. */
  namespaceScoped: boolean;
  /** Package storage type */
  type: string;
  /** A URL identifying the package repository location. */
  url: string;
  /** The interval at which to check the upstream for updates (in seconds) */
  interval: number;
  /**
   * TLS-specific parameters for connecting to a repository.
   * If the cert authority was configured for this repository, then in the context
   * of GetPackageRepositoryDetail() operation, the PackageRepositoryTlsConfig will ALWAYS
   * contain an existing SecretKeyReference, rather that cert_authority field
   */
  tlsConfig?: PackageRepositoryTlsConfig;
  /**
   * authentication parameters for connecting to a repository.
   * If Basic Auth or TLS or Docker Creds Auth was configured for this repository,
   * then in the context of GetPackageRepositoryDetail() operation, the
   * PackageRepositoryAuth will ALWAYS contain an existing SecretKeyReference,
   * rather that string values that may have been used when package repository was created
   * field
   */
  auth?: PackageRepositoryAuth;
  /** Custom data added by the plugin */
  customDetail?: Any;
  /**
   * current status of the repository which can include reconciliation
   * status, where relevant.
   */
  status?: PackageRepositoryStatus;
}

/**
 * GetPackageRepositoryDetailResponse
 *
 * Response for GetPackageRepositoryDetail
 */
export interface GetPackageRepositoryDetailResponse {
  /** package repository detail */
  detail?: PackageRepositoryDetail;
}

/** PackageRepositorySummary */
export interface PackageRepositorySummary {
  /** A reference uniquely identifying the package repository. */
  packageRepoRef?: PackageRepositoryReference;
  /** A user-provided name for the package repository (e.g. bitnami) */
  name: string;
  /** A user-provided description. */
  description: string;
  /** Whether this repository is global or namespace-scoped. */
  namespaceScoped: boolean;
  /** Package storage type */
  type: string;
  /** URL identifying the package repository location. */
  url: string;
  /**
   * current status of the repository which can include reconciliation
   * status, where relevant.
   */
  status?: PackageRepositoryStatus;
}

/**
 * GetPackageRepositorySummariesResponse
 *
 * Response for GetPackageRepositorySummaries
 */
export interface GetPackageRepositorySummariesResponse {
  /** List of PackageRepositorySummary */
  packageRepositorySummaries: PackageRepositorySummary[];
}

/**
 * UpdatePackageRepositoryResponse
 *
 * Response for UpdatePackageRepository
 */
export interface UpdatePackageRepositoryResponse {
  packageRepoRef?: PackageRepositoryReference;
}

/**
 * DeletePackageRepositoryResponse
 *
 * Response for DeletePackageRepository
 */
export interface DeletePackageRepositoryResponse {}

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
    writer: _m0.Writer = _m0.Writer.create(),
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
      PackageRepositoryTlsConfig.encode(message.tlsConfig, writer.uint32(66).fork()).ldelim();
    }
    if (message.auth !== undefined) {
      PackageRepositoryAuth.encode(message.auth, writer.uint32(74).fork()).ldelim();
    }
    if (message.plugin !== undefined) {
      Plugin.encode(message.plugin, writer.uint32(82).fork()).ldelim();
    }
    if (message.customDetail !== undefined) {
      Any.encode(message.customDetail, writer.uint32(90).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): AddPackageRepositoryRequest {
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
          message.tlsConfig = PackageRepositoryTlsConfig.decode(reader, reader.uint32());
          break;
        case 9:
          message.auth = PackageRepositoryAuth.decode(reader, reader.uint32());
          break;
        case 10:
          message.plugin = Plugin.decode(reader, reader.uint32());
          break;
        case 11:
          message.customDetail = Any.decode(reader, reader.uint32());
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
    if (object.namespaceScoped !== undefined && object.namespaceScoped !== null) {
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
    if (object.auth !== undefined && object.auth !== null) {
      message.auth = PackageRepositoryAuth.fromJSON(object.auth);
    } else {
      message.auth = undefined;
    }
    if (object.plugin !== undefined && object.plugin !== null) {
      message.plugin = Plugin.fromJSON(object.plugin);
    } else {
      message.plugin = undefined;
    }
    if (object.customDetail !== undefined && object.customDetail !== null) {
      message.customDetail = Any.fromJSON(object.customDetail);
    } else {
      message.customDetail = undefined;
    }
    return message;
  },

  toJSON(message: AddPackageRepositoryRequest): unknown {
    const obj: any = {};
    message.context !== undefined &&
      (obj.context = message.context ? Context.toJSON(message.context) : undefined);
    message.name !== undefined && (obj.name = message.name);
    message.description !== undefined && (obj.description = message.description);
    message.namespaceScoped !== undefined && (obj.namespaceScoped = message.namespaceScoped);
    message.type !== undefined && (obj.type = message.type);
    message.url !== undefined && (obj.url = message.url);
    message.interval !== undefined && (obj.interval = message.interval);
    message.tlsConfig !== undefined &&
      (obj.tlsConfig = message.tlsConfig
        ? PackageRepositoryTlsConfig.toJSON(message.tlsConfig)
        : undefined);
    message.auth !== undefined &&
      (obj.auth = message.auth ? PackageRepositoryAuth.toJSON(message.auth) : undefined);
    message.plugin !== undefined &&
      (obj.plugin = message.plugin ? Plugin.toJSON(message.plugin) : undefined);
    message.customDetail !== undefined &&
      (obj.customDetail = message.customDetail ? Any.toJSON(message.customDetail) : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<AddPackageRepositoryRequest>): AddPackageRepositoryRequest {
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
    if (object.namespaceScoped !== undefined && object.namespaceScoped !== null) {
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
      message.tlsConfig = PackageRepositoryTlsConfig.fromPartial(object.tlsConfig);
    } else {
      message.tlsConfig = undefined;
    }
    if (object.auth !== undefined && object.auth !== null) {
      message.auth = PackageRepositoryAuth.fromPartial(object.auth);
    } else {
      message.auth = undefined;
    }
    if (object.plugin !== undefined && object.plugin !== null) {
      message.plugin = Plugin.fromPartial(object.plugin);
    } else {
      message.plugin = undefined;
    }
    if (object.customDetail !== undefined && object.customDetail !== null) {
      message.customDetail = Any.fromPartial(object.customDetail);
    } else {
      message.customDetail = undefined;
    }
    return message;
  },
};

const basePackageRepositoryTlsConfig: object = { insecureSkipVerify: false };

export const PackageRepositoryTlsConfig = {
  encode(
    message: PackageRepositoryTlsConfig,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.insecureSkipVerify === true) {
      writer.uint32(8).bool(message.insecureSkipVerify);
    }
    if (message.certAuthority !== undefined) {
      writer.uint32(18).string(message.certAuthority);
    }
    if (message.secretRef !== undefined) {
      SecretKeyReference.encode(message.secretRef, writer.uint32(26).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): PackageRepositoryTlsConfig {
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
          message.secretRef = SecretKeyReference.decode(reader, reader.uint32());
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
    if (object.insecureSkipVerify !== undefined && object.insecureSkipVerify !== null) {
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
    message.certAuthority !== undefined && (obj.certAuthority = message.certAuthority);
    message.secretRef !== undefined &&
      (obj.secretRef = message.secretRef
        ? SecretKeyReference.toJSON(message.secretRef)
        : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<PackageRepositoryTlsConfig>): PackageRepositoryTlsConfig {
    const message = {
      ...basePackageRepositoryTlsConfig,
    } as PackageRepositoryTlsConfig;
    if (object.insecureSkipVerify !== undefined && object.insecureSkipVerify !== null) {
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

const basePackageRepositoryAuth: object = { type: 0, passCredentials: false };

export const PackageRepositoryAuth = {
  encode(message: PackageRepositoryAuth, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.type !== 0) {
      writer.uint32(8).int32(message.type);
    }
    if (message.usernamePassword !== undefined) {
      UsernamePassword.encode(message.usernamePassword, writer.uint32(18).fork()).ldelim();
    }
    if (message.tlsCertKey !== undefined) {
      TlsCertKey.encode(message.tlsCertKey, writer.uint32(26).fork()).ldelim();
    }
    if (message.dockerCreds !== undefined) {
      DockerCredentials.encode(message.dockerCreds, writer.uint32(34).fork()).ldelim();
    }
    if (message.header !== undefined) {
      writer.uint32(42).string(message.header);
    }
    if (message.secretRef !== undefined) {
      SecretKeyReference.encode(message.secretRef, writer.uint32(50).fork()).ldelim();
    }
    if (message.passCredentials === true) {
      writer.uint32(56).bool(message.passCredentials);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): PackageRepositoryAuth {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...basePackageRepositoryAuth } as PackageRepositoryAuth;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.type = reader.int32() as any;
          break;
        case 2:
          message.usernamePassword = UsernamePassword.decode(reader, reader.uint32());
          break;
        case 3:
          message.tlsCertKey = TlsCertKey.decode(reader, reader.uint32());
          break;
        case 4:
          message.dockerCreds = DockerCredentials.decode(reader, reader.uint32());
          break;
        case 5:
          message.header = reader.string();
          break;
        case 6:
          message.secretRef = SecretKeyReference.decode(reader, reader.uint32());
          break;
        case 7:
          message.passCredentials = reader.bool();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): PackageRepositoryAuth {
    const message = { ...basePackageRepositoryAuth } as PackageRepositoryAuth;
    if (object.type !== undefined && object.type !== null) {
      message.type = packageRepositoryAuth_PackageRepositoryAuthTypeFromJSON(object.type);
    } else {
      message.type = 0;
    }
    if (object.usernamePassword !== undefined && object.usernamePassword !== null) {
      message.usernamePassword = UsernamePassword.fromJSON(object.usernamePassword);
    } else {
      message.usernamePassword = undefined;
    }
    if (object.tlsCertKey !== undefined && object.tlsCertKey !== null) {
      message.tlsCertKey = TlsCertKey.fromJSON(object.tlsCertKey);
    } else {
      message.tlsCertKey = undefined;
    }
    if (object.dockerCreds !== undefined && object.dockerCreds !== null) {
      message.dockerCreds = DockerCredentials.fromJSON(object.dockerCreds);
    } else {
      message.dockerCreds = undefined;
    }
    if (object.header !== undefined && object.header !== null) {
      message.header = String(object.header);
    } else {
      message.header = undefined;
    }
    if (object.secretRef !== undefined && object.secretRef !== null) {
      message.secretRef = SecretKeyReference.fromJSON(object.secretRef);
    } else {
      message.secretRef = undefined;
    }
    if (object.passCredentials !== undefined && object.passCredentials !== null) {
      message.passCredentials = Boolean(object.passCredentials);
    } else {
      message.passCredentials = false;
    }
    return message;
  },

  toJSON(message: PackageRepositoryAuth): unknown {
    const obj: any = {};
    message.type !== undefined &&
      (obj.type = packageRepositoryAuth_PackageRepositoryAuthTypeToJSON(message.type));
    message.usernamePassword !== undefined &&
      (obj.usernamePassword = message.usernamePassword
        ? UsernamePassword.toJSON(message.usernamePassword)
        : undefined);
    message.tlsCertKey !== undefined &&
      (obj.tlsCertKey = message.tlsCertKey ? TlsCertKey.toJSON(message.tlsCertKey) : undefined);
    message.dockerCreds !== undefined &&
      (obj.dockerCreds = message.dockerCreds
        ? DockerCredentials.toJSON(message.dockerCreds)
        : undefined);
    message.header !== undefined && (obj.header = message.header);
    message.secretRef !== undefined &&
      (obj.secretRef = message.secretRef
        ? SecretKeyReference.toJSON(message.secretRef)
        : undefined);
    message.passCredentials !== undefined && (obj.passCredentials = message.passCredentials);
    return obj;
  },

  fromPartial(object: DeepPartial<PackageRepositoryAuth>): PackageRepositoryAuth {
    const message = { ...basePackageRepositoryAuth } as PackageRepositoryAuth;
    if (object.type !== undefined && object.type !== null) {
      message.type = object.type;
    } else {
      message.type = 0;
    }
    if (object.usernamePassword !== undefined && object.usernamePassword !== null) {
      message.usernamePassword = UsernamePassword.fromPartial(object.usernamePassword);
    } else {
      message.usernamePassword = undefined;
    }
    if (object.tlsCertKey !== undefined && object.tlsCertKey !== null) {
      message.tlsCertKey = TlsCertKey.fromPartial(object.tlsCertKey);
    } else {
      message.tlsCertKey = undefined;
    }
    if (object.dockerCreds !== undefined && object.dockerCreds !== null) {
      message.dockerCreds = DockerCredentials.fromPartial(object.dockerCreds);
    } else {
      message.dockerCreds = undefined;
    }
    if (object.header !== undefined && object.header !== null) {
      message.header = object.header;
    } else {
      message.header = undefined;
    }
    if (object.secretRef !== undefined && object.secretRef !== null) {
      message.secretRef = SecretKeyReference.fromPartial(object.secretRef);
    } else {
      message.secretRef = undefined;
    }
    if (object.passCredentials !== undefined && object.passCredentials !== null) {
      message.passCredentials = object.passCredentials;
    } else {
      message.passCredentials = false;
    }
    return message;
  },
};

const baseUsernamePassword: object = { username: "", password: "" };

export const UsernamePassword = {
  encode(message: UsernamePassword, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.username !== "") {
      writer.uint32(10).string(message.username);
    }
    if (message.password !== "") {
      writer.uint32(18).string(message.password);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): UsernamePassword {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseUsernamePassword } as UsernamePassword;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.username = reader.string();
          break;
        case 2:
          message.password = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): UsernamePassword {
    const message = { ...baseUsernamePassword } as UsernamePassword;
    if (object.username !== undefined && object.username !== null) {
      message.username = String(object.username);
    } else {
      message.username = "";
    }
    if (object.password !== undefined && object.password !== null) {
      message.password = String(object.password);
    } else {
      message.password = "";
    }
    return message;
  },

  toJSON(message: UsernamePassword): unknown {
    const obj: any = {};
    message.username !== undefined && (obj.username = message.username);
    message.password !== undefined && (obj.password = message.password);
    return obj;
  },

  fromPartial(object: DeepPartial<UsernamePassword>): UsernamePassword {
    const message = { ...baseUsernamePassword } as UsernamePassword;
    if (object.username !== undefined && object.username !== null) {
      message.username = object.username;
    } else {
      message.username = "";
    }
    if (object.password !== undefined && object.password !== null) {
      message.password = object.password;
    } else {
      message.password = "";
    }
    return message;
  },
};

const baseTlsCertKey: object = { cert: "", key: "" };

export const TlsCertKey = {
  encode(message: TlsCertKey, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.cert !== "") {
      writer.uint32(10).string(message.cert);
    }
    if (message.key !== "") {
      writer.uint32(18).string(message.key);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): TlsCertKey {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseTlsCertKey } as TlsCertKey;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.cert = reader.string();
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

  fromJSON(object: any): TlsCertKey {
    const message = { ...baseTlsCertKey } as TlsCertKey;
    if (object.cert !== undefined && object.cert !== null) {
      message.cert = String(object.cert);
    } else {
      message.cert = "";
    }
    if (object.key !== undefined && object.key !== null) {
      message.key = String(object.key);
    } else {
      message.key = "";
    }
    return message;
  },

  toJSON(message: TlsCertKey): unknown {
    const obj: any = {};
    message.cert !== undefined && (obj.cert = message.cert);
    message.key !== undefined && (obj.key = message.key);
    return obj;
  },

  fromPartial(object: DeepPartial<TlsCertKey>): TlsCertKey {
    const message = { ...baseTlsCertKey } as TlsCertKey;
    if (object.cert !== undefined && object.cert !== null) {
      message.cert = object.cert;
    } else {
      message.cert = "";
    }
    if (object.key !== undefined && object.key !== null) {
      message.key = object.key;
    } else {
      message.key = "";
    }
    return message;
  },
};

const baseDockerCredentials: object = {
  server: "",
  username: "",
  password: "",
  email: "",
};

export const DockerCredentials = {
  encode(message: DockerCredentials, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.server !== "") {
      writer.uint32(10).string(message.server);
    }
    if (message.username !== "") {
      writer.uint32(18).string(message.username);
    }
    if (message.password !== "") {
      writer.uint32(26).string(message.password);
    }
    if (message.email !== "") {
      writer.uint32(34).string(message.email);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): DockerCredentials {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseDockerCredentials } as DockerCredentials;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.server = reader.string();
          break;
        case 2:
          message.username = reader.string();
          break;
        case 3:
          message.password = reader.string();
          break;
        case 4:
          message.email = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): DockerCredentials {
    const message = { ...baseDockerCredentials } as DockerCredentials;
    if (object.server !== undefined && object.server !== null) {
      message.server = String(object.server);
    } else {
      message.server = "";
    }
    if (object.username !== undefined && object.username !== null) {
      message.username = String(object.username);
    } else {
      message.username = "";
    }
    if (object.password !== undefined && object.password !== null) {
      message.password = String(object.password);
    } else {
      message.password = "";
    }
    if (object.email !== undefined && object.email !== null) {
      message.email = String(object.email);
    } else {
      message.email = "";
    }
    return message;
  },

  toJSON(message: DockerCredentials): unknown {
    const obj: any = {};
    message.server !== undefined && (obj.server = message.server);
    message.username !== undefined && (obj.username = message.username);
    message.password !== undefined && (obj.password = message.password);
    message.email !== undefined && (obj.email = message.email);
    return obj;
  },

  fromPartial(object: DeepPartial<DockerCredentials>): DockerCredentials {
    const message = { ...baseDockerCredentials } as DockerCredentials;
    if (object.server !== undefined && object.server !== null) {
      message.server = object.server;
    } else {
      message.server = "";
    }
    if (object.username !== undefined && object.username !== null) {
      message.username = object.username;
    } else {
      message.username = "";
    }
    if (object.password !== undefined && object.password !== null) {
      message.password = object.password;
    } else {
      message.password = "";
    }
    if (object.email !== undefined && object.email !== null) {
      message.email = object.email;
    } else {
      message.email = "";
    }
    return message;
  },
};

const baseSecretKeyReference: object = { name: "", key: "" };

export const SecretKeyReference = {
  encode(message: SecretKeyReference, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
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

const baseGetPackageRepositoryDetailRequest: object = {};

export const GetPackageRepositoryDetailRequest = {
  encode(
    message: GetPackageRepositoryDetailRequest,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.packageRepoRef !== undefined) {
      PackageRepositoryReference.encode(message.packageRepoRef, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetPackageRepositoryDetailRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseGetPackageRepositoryDetailRequest,
    } as GetPackageRepositoryDetailRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.packageRepoRef = PackageRepositoryReference.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): GetPackageRepositoryDetailRequest {
    const message = {
      ...baseGetPackageRepositoryDetailRequest,
    } as GetPackageRepositoryDetailRequest;
    if (object.packageRepoRef !== undefined && object.packageRepoRef !== null) {
      message.packageRepoRef = PackageRepositoryReference.fromJSON(object.packageRepoRef);
    } else {
      message.packageRepoRef = undefined;
    }
    return message;
  },

  toJSON(message: GetPackageRepositoryDetailRequest): unknown {
    const obj: any = {};
    message.packageRepoRef !== undefined &&
      (obj.packageRepoRef = message.packageRepoRef
        ? PackageRepositoryReference.toJSON(message.packageRepoRef)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<GetPackageRepositoryDetailRequest>,
  ): GetPackageRepositoryDetailRequest {
    const message = {
      ...baseGetPackageRepositoryDetailRequest,
    } as GetPackageRepositoryDetailRequest;
    if (object.packageRepoRef !== undefined && object.packageRepoRef !== null) {
      message.packageRepoRef = PackageRepositoryReference.fromPartial(object.packageRepoRef);
    } else {
      message.packageRepoRef = undefined;
    }
    return message;
  },
};

const baseGetPackageRepositorySummariesRequest: object = {};

export const GetPackageRepositorySummariesRequest = {
  encode(
    message: GetPackageRepositorySummariesRequest,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.context !== undefined) {
      Context.encode(message.context, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetPackageRepositorySummariesRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseGetPackageRepositorySummariesRequest,
    } as GetPackageRepositorySummariesRequest;
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

  fromJSON(object: any): GetPackageRepositorySummariesRequest {
    const message = {
      ...baseGetPackageRepositorySummariesRequest,
    } as GetPackageRepositorySummariesRequest;
    if (object.context !== undefined && object.context !== null) {
      message.context = Context.fromJSON(object.context);
    } else {
      message.context = undefined;
    }
    return message;
  },

  toJSON(message: GetPackageRepositorySummariesRequest): unknown {
    const obj: any = {};
    message.context !== undefined &&
      (obj.context = message.context ? Context.toJSON(message.context) : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<GetPackageRepositorySummariesRequest>,
  ): GetPackageRepositorySummariesRequest {
    const message = {
      ...baseGetPackageRepositorySummariesRequest,
    } as GetPackageRepositorySummariesRequest;
    if (object.context !== undefined && object.context !== null) {
      message.context = Context.fromPartial(object.context);
    } else {
      message.context = undefined;
    }
    return message;
  },
};

const baseUpdatePackageRepositoryRequest: object = {
  url: "",
  description: "",
  interval: 0,
};

export const UpdatePackageRepositoryRequest = {
  encode(
    message: UpdatePackageRepositoryRequest,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.packageRepoRef !== undefined) {
      PackageRepositoryReference.encode(message.packageRepoRef, writer.uint32(10).fork()).ldelim();
    }
    if (message.url !== "") {
      writer.uint32(18).string(message.url);
    }
    if (message.description !== "") {
      writer.uint32(26).string(message.description);
    }
    if (message.interval !== 0) {
      writer.uint32(32).uint32(message.interval);
    }
    if (message.tlsConfig !== undefined) {
      PackageRepositoryTlsConfig.encode(message.tlsConfig, writer.uint32(42).fork()).ldelim();
    }
    if (message.auth !== undefined) {
      PackageRepositoryAuth.encode(message.auth, writer.uint32(50).fork()).ldelim();
    }
    if (message.customDetail !== undefined) {
      Any.encode(message.customDetail, writer.uint32(90).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): UpdatePackageRepositoryRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseUpdatePackageRepositoryRequest,
    } as UpdatePackageRepositoryRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.packageRepoRef = PackageRepositoryReference.decode(reader, reader.uint32());
          break;
        case 2:
          message.url = reader.string();
          break;
        case 3:
          message.description = reader.string();
          break;
        case 4:
          message.interval = reader.uint32();
          break;
        case 5:
          message.tlsConfig = PackageRepositoryTlsConfig.decode(reader, reader.uint32());
          break;
        case 6:
          message.auth = PackageRepositoryAuth.decode(reader, reader.uint32());
          break;
        case 11:
          message.customDetail = Any.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): UpdatePackageRepositoryRequest {
    const message = {
      ...baseUpdatePackageRepositoryRequest,
    } as UpdatePackageRepositoryRequest;
    if (object.packageRepoRef !== undefined && object.packageRepoRef !== null) {
      message.packageRepoRef = PackageRepositoryReference.fromJSON(object.packageRepoRef);
    } else {
      message.packageRepoRef = undefined;
    }
    if (object.url !== undefined && object.url !== null) {
      message.url = String(object.url);
    } else {
      message.url = "";
    }
    if (object.description !== undefined && object.description !== null) {
      message.description = String(object.description);
    } else {
      message.description = "";
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
    if (object.auth !== undefined && object.auth !== null) {
      message.auth = PackageRepositoryAuth.fromJSON(object.auth);
    } else {
      message.auth = undefined;
    }
    if (object.customDetail !== undefined && object.customDetail !== null) {
      message.customDetail = Any.fromJSON(object.customDetail);
    } else {
      message.customDetail = undefined;
    }
    return message;
  },

  toJSON(message: UpdatePackageRepositoryRequest): unknown {
    const obj: any = {};
    message.packageRepoRef !== undefined &&
      (obj.packageRepoRef = message.packageRepoRef
        ? PackageRepositoryReference.toJSON(message.packageRepoRef)
        : undefined);
    message.url !== undefined && (obj.url = message.url);
    message.description !== undefined && (obj.description = message.description);
    message.interval !== undefined && (obj.interval = message.interval);
    message.tlsConfig !== undefined &&
      (obj.tlsConfig = message.tlsConfig
        ? PackageRepositoryTlsConfig.toJSON(message.tlsConfig)
        : undefined);
    message.auth !== undefined &&
      (obj.auth = message.auth ? PackageRepositoryAuth.toJSON(message.auth) : undefined);
    message.customDetail !== undefined &&
      (obj.customDetail = message.customDetail ? Any.toJSON(message.customDetail) : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<UpdatePackageRepositoryRequest>): UpdatePackageRepositoryRequest {
    const message = {
      ...baseUpdatePackageRepositoryRequest,
    } as UpdatePackageRepositoryRequest;
    if (object.packageRepoRef !== undefined && object.packageRepoRef !== null) {
      message.packageRepoRef = PackageRepositoryReference.fromPartial(object.packageRepoRef);
    } else {
      message.packageRepoRef = undefined;
    }
    if (object.url !== undefined && object.url !== null) {
      message.url = object.url;
    } else {
      message.url = "";
    }
    if (object.description !== undefined && object.description !== null) {
      message.description = object.description;
    } else {
      message.description = "";
    }
    if (object.interval !== undefined && object.interval !== null) {
      message.interval = object.interval;
    } else {
      message.interval = 0;
    }
    if (object.tlsConfig !== undefined && object.tlsConfig !== null) {
      message.tlsConfig = PackageRepositoryTlsConfig.fromPartial(object.tlsConfig);
    } else {
      message.tlsConfig = undefined;
    }
    if (object.auth !== undefined && object.auth !== null) {
      message.auth = PackageRepositoryAuth.fromPartial(object.auth);
    } else {
      message.auth = undefined;
    }
    if (object.customDetail !== undefined && object.customDetail !== null) {
      message.customDetail = Any.fromPartial(object.customDetail);
    } else {
      message.customDetail = undefined;
    }
    return message;
  },
};

const baseDeletePackageRepositoryRequest: object = {};

export const DeletePackageRepositoryRequest = {
  encode(
    message: DeletePackageRepositoryRequest,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.packageRepoRef !== undefined) {
      PackageRepositoryReference.encode(message.packageRepoRef, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): DeletePackageRepositoryRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseDeletePackageRepositoryRequest,
    } as DeletePackageRepositoryRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.packageRepoRef = PackageRepositoryReference.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): DeletePackageRepositoryRequest {
    const message = {
      ...baseDeletePackageRepositoryRequest,
    } as DeletePackageRepositoryRequest;
    if (object.packageRepoRef !== undefined && object.packageRepoRef !== null) {
      message.packageRepoRef = PackageRepositoryReference.fromJSON(object.packageRepoRef);
    } else {
      message.packageRepoRef = undefined;
    }
    return message;
  },

  toJSON(message: DeletePackageRepositoryRequest): unknown {
    const obj: any = {};
    message.packageRepoRef !== undefined &&
      (obj.packageRepoRef = message.packageRepoRef
        ? PackageRepositoryReference.toJSON(message.packageRepoRef)
        : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<DeletePackageRepositoryRequest>): DeletePackageRepositoryRequest {
    const message = {
      ...baseDeletePackageRepositoryRequest,
    } as DeletePackageRepositoryRequest;
    if (object.packageRepoRef !== undefined && object.packageRepoRef !== null) {
      message.packageRepoRef = PackageRepositoryReference.fromPartial(object.packageRepoRef);
    } else {
      message.packageRepoRef = undefined;
    }
    return message;
  },
};

const basePackageRepositoryReference: object = { identifier: "" };

export const PackageRepositoryReference = {
  encode(
    message: PackageRepositoryReference,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.context !== undefined) {
      Context.encode(message.context, writer.uint32(10).fork()).ldelim();
    }
    if (message.identifier !== "") {
      writer.uint32(18).string(message.identifier);
    }
    if (message.plugin !== undefined) {
      Plugin.encode(message.plugin, writer.uint32(26).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): PackageRepositoryReference {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...basePackageRepositoryReference,
    } as PackageRepositoryReference;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.context = Context.decode(reader, reader.uint32());
          break;
        case 2:
          message.identifier = reader.string();
          break;
        case 3:
          message.plugin = Plugin.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): PackageRepositoryReference {
    const message = {
      ...basePackageRepositoryReference,
    } as PackageRepositoryReference;
    if (object.context !== undefined && object.context !== null) {
      message.context = Context.fromJSON(object.context);
    } else {
      message.context = undefined;
    }
    if (object.identifier !== undefined && object.identifier !== null) {
      message.identifier = String(object.identifier);
    } else {
      message.identifier = "";
    }
    if (object.plugin !== undefined && object.plugin !== null) {
      message.plugin = Plugin.fromJSON(object.plugin);
    } else {
      message.plugin = undefined;
    }
    return message;
  },

  toJSON(message: PackageRepositoryReference): unknown {
    const obj: any = {};
    message.context !== undefined &&
      (obj.context = message.context ? Context.toJSON(message.context) : undefined);
    message.identifier !== undefined && (obj.identifier = message.identifier);
    message.plugin !== undefined &&
      (obj.plugin = message.plugin ? Plugin.toJSON(message.plugin) : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<PackageRepositoryReference>): PackageRepositoryReference {
    const message = {
      ...basePackageRepositoryReference,
    } as PackageRepositoryReference;
    if (object.context !== undefined && object.context !== null) {
      message.context = Context.fromPartial(object.context);
    } else {
      message.context = undefined;
    }
    if (object.identifier !== undefined && object.identifier !== null) {
      message.identifier = object.identifier;
    } else {
      message.identifier = "";
    }
    if (object.plugin !== undefined && object.plugin !== null) {
      message.plugin = Plugin.fromPartial(object.plugin);
    } else {
      message.plugin = undefined;
    }
    return message;
  },
};

const baseAddPackageRepositoryResponse: object = {};

export const AddPackageRepositoryResponse = {
  encode(
    message: AddPackageRepositoryResponse,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.packageRepoRef !== undefined) {
      PackageRepositoryReference.encode(message.packageRepoRef, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): AddPackageRepositoryResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseAddPackageRepositoryResponse,
    } as AddPackageRepositoryResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.packageRepoRef = PackageRepositoryReference.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): AddPackageRepositoryResponse {
    const message = {
      ...baseAddPackageRepositoryResponse,
    } as AddPackageRepositoryResponse;
    if (object.packageRepoRef !== undefined && object.packageRepoRef !== null) {
      message.packageRepoRef = PackageRepositoryReference.fromJSON(object.packageRepoRef);
    } else {
      message.packageRepoRef = undefined;
    }
    return message;
  },

  toJSON(message: AddPackageRepositoryResponse): unknown {
    const obj: any = {};
    message.packageRepoRef !== undefined &&
      (obj.packageRepoRef = message.packageRepoRef
        ? PackageRepositoryReference.toJSON(message.packageRepoRef)
        : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<AddPackageRepositoryResponse>): AddPackageRepositoryResponse {
    const message = {
      ...baseAddPackageRepositoryResponse,
    } as AddPackageRepositoryResponse;
    if (object.packageRepoRef !== undefined && object.packageRepoRef !== null) {
      message.packageRepoRef = PackageRepositoryReference.fromPartial(object.packageRepoRef);
    } else {
      message.packageRepoRef = undefined;
    }
    return message;
  },
};

const basePackageRepositoryStatus: object = {
  ready: false,
  reason: 0,
  userReason: "",
};

export const PackageRepositoryStatus = {
  encode(message: PackageRepositoryStatus, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.ready === true) {
      writer.uint32(8).bool(message.ready);
    }
    if (message.reason !== 0) {
      writer.uint32(16).int32(message.reason);
    }
    if (message.userReason !== "") {
      writer.uint32(26).string(message.userReason);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): PackageRepositoryStatus {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...basePackageRepositoryStatus,
    } as PackageRepositoryStatus;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.ready = reader.bool();
          break;
        case 2:
          message.reason = reader.int32() as any;
          break;
        case 3:
          message.userReason = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): PackageRepositoryStatus {
    const message = {
      ...basePackageRepositoryStatus,
    } as PackageRepositoryStatus;
    if (object.ready !== undefined && object.ready !== null) {
      message.ready = Boolean(object.ready);
    } else {
      message.ready = false;
    }
    if (object.reason !== undefined && object.reason !== null) {
      message.reason = packageRepositoryStatus_StatusReasonFromJSON(object.reason);
    } else {
      message.reason = 0;
    }
    if (object.userReason !== undefined && object.userReason !== null) {
      message.userReason = String(object.userReason);
    } else {
      message.userReason = "";
    }
    return message;
  },

  toJSON(message: PackageRepositoryStatus): unknown {
    const obj: any = {};
    message.ready !== undefined && (obj.ready = message.ready);
    message.reason !== undefined &&
      (obj.reason = packageRepositoryStatus_StatusReasonToJSON(message.reason));
    message.userReason !== undefined && (obj.userReason = message.userReason);
    return obj;
  },

  fromPartial(object: DeepPartial<PackageRepositoryStatus>): PackageRepositoryStatus {
    const message = {
      ...basePackageRepositoryStatus,
    } as PackageRepositoryStatus;
    if (object.ready !== undefined && object.ready !== null) {
      message.ready = object.ready;
    } else {
      message.ready = false;
    }
    if (object.reason !== undefined && object.reason !== null) {
      message.reason = object.reason;
    } else {
      message.reason = 0;
    }
    if (object.userReason !== undefined && object.userReason !== null) {
      message.userReason = object.userReason;
    } else {
      message.userReason = "";
    }
    return message;
  },
};

const basePackageRepositoryDetail: object = {
  name: "",
  description: "",
  namespaceScoped: false,
  type: "",
  url: "",
  interval: 0,
};

export const PackageRepositoryDetail = {
  encode(message: PackageRepositoryDetail, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.packageRepoRef !== undefined) {
      PackageRepositoryReference.encode(message.packageRepoRef, writer.uint32(10).fork()).ldelim();
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
      PackageRepositoryTlsConfig.encode(message.tlsConfig, writer.uint32(66).fork()).ldelim();
    }
    if (message.auth !== undefined) {
      PackageRepositoryAuth.encode(message.auth, writer.uint32(74).fork()).ldelim();
    }
    if (message.customDetail !== undefined) {
      Any.encode(message.customDetail, writer.uint32(82).fork()).ldelim();
    }
    if (message.status !== undefined) {
      PackageRepositoryStatus.encode(message.status, writer.uint32(90).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): PackageRepositoryDetail {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...basePackageRepositoryDetail,
    } as PackageRepositoryDetail;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.packageRepoRef = PackageRepositoryReference.decode(reader, reader.uint32());
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
          message.tlsConfig = PackageRepositoryTlsConfig.decode(reader, reader.uint32());
          break;
        case 9:
          message.auth = PackageRepositoryAuth.decode(reader, reader.uint32());
          break;
        case 10:
          message.customDetail = Any.decode(reader, reader.uint32());
          break;
        case 11:
          message.status = PackageRepositoryStatus.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): PackageRepositoryDetail {
    const message = {
      ...basePackageRepositoryDetail,
    } as PackageRepositoryDetail;
    if (object.packageRepoRef !== undefined && object.packageRepoRef !== null) {
      message.packageRepoRef = PackageRepositoryReference.fromJSON(object.packageRepoRef);
    } else {
      message.packageRepoRef = undefined;
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
    if (object.namespaceScoped !== undefined && object.namespaceScoped !== null) {
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
    if (object.auth !== undefined && object.auth !== null) {
      message.auth = PackageRepositoryAuth.fromJSON(object.auth);
    } else {
      message.auth = undefined;
    }
    if (object.customDetail !== undefined && object.customDetail !== null) {
      message.customDetail = Any.fromJSON(object.customDetail);
    } else {
      message.customDetail = undefined;
    }
    if (object.status !== undefined && object.status !== null) {
      message.status = PackageRepositoryStatus.fromJSON(object.status);
    } else {
      message.status = undefined;
    }
    return message;
  },

  toJSON(message: PackageRepositoryDetail): unknown {
    const obj: any = {};
    message.packageRepoRef !== undefined &&
      (obj.packageRepoRef = message.packageRepoRef
        ? PackageRepositoryReference.toJSON(message.packageRepoRef)
        : undefined);
    message.name !== undefined && (obj.name = message.name);
    message.description !== undefined && (obj.description = message.description);
    message.namespaceScoped !== undefined && (obj.namespaceScoped = message.namespaceScoped);
    message.type !== undefined && (obj.type = message.type);
    message.url !== undefined && (obj.url = message.url);
    message.interval !== undefined && (obj.interval = message.interval);
    message.tlsConfig !== undefined &&
      (obj.tlsConfig = message.tlsConfig
        ? PackageRepositoryTlsConfig.toJSON(message.tlsConfig)
        : undefined);
    message.auth !== undefined &&
      (obj.auth = message.auth ? PackageRepositoryAuth.toJSON(message.auth) : undefined);
    message.customDetail !== undefined &&
      (obj.customDetail = message.customDetail ? Any.toJSON(message.customDetail) : undefined);
    message.status !== undefined &&
      (obj.status = message.status ? PackageRepositoryStatus.toJSON(message.status) : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<PackageRepositoryDetail>): PackageRepositoryDetail {
    const message = {
      ...basePackageRepositoryDetail,
    } as PackageRepositoryDetail;
    if (object.packageRepoRef !== undefined && object.packageRepoRef !== null) {
      message.packageRepoRef = PackageRepositoryReference.fromPartial(object.packageRepoRef);
    } else {
      message.packageRepoRef = undefined;
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
    if (object.namespaceScoped !== undefined && object.namespaceScoped !== null) {
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
      message.tlsConfig = PackageRepositoryTlsConfig.fromPartial(object.tlsConfig);
    } else {
      message.tlsConfig = undefined;
    }
    if (object.auth !== undefined && object.auth !== null) {
      message.auth = PackageRepositoryAuth.fromPartial(object.auth);
    } else {
      message.auth = undefined;
    }
    if (object.customDetail !== undefined && object.customDetail !== null) {
      message.customDetail = Any.fromPartial(object.customDetail);
    } else {
      message.customDetail = undefined;
    }
    if (object.status !== undefined && object.status !== null) {
      message.status = PackageRepositoryStatus.fromPartial(object.status);
    } else {
      message.status = undefined;
    }
    return message;
  },
};

const baseGetPackageRepositoryDetailResponse: object = {};

export const GetPackageRepositoryDetailResponse = {
  encode(
    message: GetPackageRepositoryDetailResponse,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.detail !== undefined) {
      PackageRepositoryDetail.encode(message.detail, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetPackageRepositoryDetailResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseGetPackageRepositoryDetailResponse,
    } as GetPackageRepositoryDetailResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.detail = PackageRepositoryDetail.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): GetPackageRepositoryDetailResponse {
    const message = {
      ...baseGetPackageRepositoryDetailResponse,
    } as GetPackageRepositoryDetailResponse;
    if (object.detail !== undefined && object.detail !== null) {
      message.detail = PackageRepositoryDetail.fromJSON(object.detail);
    } else {
      message.detail = undefined;
    }
    return message;
  },

  toJSON(message: GetPackageRepositoryDetailResponse): unknown {
    const obj: any = {};
    message.detail !== undefined &&
      (obj.detail = message.detail ? PackageRepositoryDetail.toJSON(message.detail) : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<GetPackageRepositoryDetailResponse>,
  ): GetPackageRepositoryDetailResponse {
    const message = {
      ...baseGetPackageRepositoryDetailResponse,
    } as GetPackageRepositoryDetailResponse;
    if (object.detail !== undefined && object.detail !== null) {
      message.detail = PackageRepositoryDetail.fromPartial(object.detail);
    } else {
      message.detail = undefined;
    }
    return message;
  },
};

const basePackageRepositorySummary: object = {
  name: "",
  description: "",
  namespaceScoped: false,
  type: "",
  url: "",
};

export const PackageRepositorySummary = {
  encode(message: PackageRepositorySummary, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.packageRepoRef !== undefined) {
      PackageRepositoryReference.encode(message.packageRepoRef, writer.uint32(10).fork()).ldelim();
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
    if (message.status !== undefined) {
      PackageRepositoryStatus.encode(message.status, writer.uint32(58).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): PackageRepositorySummary {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...basePackageRepositorySummary,
    } as PackageRepositorySummary;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.packageRepoRef = PackageRepositoryReference.decode(reader, reader.uint32());
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
          message.status = PackageRepositoryStatus.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): PackageRepositorySummary {
    const message = {
      ...basePackageRepositorySummary,
    } as PackageRepositorySummary;
    if (object.packageRepoRef !== undefined && object.packageRepoRef !== null) {
      message.packageRepoRef = PackageRepositoryReference.fromJSON(object.packageRepoRef);
    } else {
      message.packageRepoRef = undefined;
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
    if (object.namespaceScoped !== undefined && object.namespaceScoped !== null) {
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
    if (object.status !== undefined && object.status !== null) {
      message.status = PackageRepositoryStatus.fromJSON(object.status);
    } else {
      message.status = undefined;
    }
    return message;
  },

  toJSON(message: PackageRepositorySummary): unknown {
    const obj: any = {};
    message.packageRepoRef !== undefined &&
      (obj.packageRepoRef = message.packageRepoRef
        ? PackageRepositoryReference.toJSON(message.packageRepoRef)
        : undefined);
    message.name !== undefined && (obj.name = message.name);
    message.description !== undefined && (obj.description = message.description);
    message.namespaceScoped !== undefined && (obj.namespaceScoped = message.namespaceScoped);
    message.type !== undefined && (obj.type = message.type);
    message.url !== undefined && (obj.url = message.url);
    message.status !== undefined &&
      (obj.status = message.status ? PackageRepositoryStatus.toJSON(message.status) : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<PackageRepositorySummary>): PackageRepositorySummary {
    const message = {
      ...basePackageRepositorySummary,
    } as PackageRepositorySummary;
    if (object.packageRepoRef !== undefined && object.packageRepoRef !== null) {
      message.packageRepoRef = PackageRepositoryReference.fromPartial(object.packageRepoRef);
    } else {
      message.packageRepoRef = undefined;
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
    if (object.namespaceScoped !== undefined && object.namespaceScoped !== null) {
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
    if (object.status !== undefined && object.status !== null) {
      message.status = PackageRepositoryStatus.fromPartial(object.status);
    } else {
      message.status = undefined;
    }
    return message;
  },
};

const baseGetPackageRepositorySummariesResponse: object = {};

export const GetPackageRepositorySummariesResponse = {
  encode(
    message: GetPackageRepositorySummariesResponse,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    for (const v of message.packageRepositorySummaries) {
      PackageRepositorySummary.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): GetPackageRepositorySummariesResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseGetPackageRepositorySummariesResponse,
    } as GetPackageRepositorySummariesResponse;
    message.packageRepositorySummaries = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.packageRepositorySummaries.push(
            PackageRepositorySummary.decode(reader, reader.uint32()),
          );
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): GetPackageRepositorySummariesResponse {
    const message = {
      ...baseGetPackageRepositorySummariesResponse,
    } as GetPackageRepositorySummariesResponse;
    message.packageRepositorySummaries = [];
    if (
      object.packageRepositorySummaries !== undefined &&
      object.packageRepositorySummaries !== null
    ) {
      for (const e of object.packageRepositorySummaries) {
        message.packageRepositorySummaries.push(PackageRepositorySummary.fromJSON(e));
      }
    }
    return message;
  },

  toJSON(message: GetPackageRepositorySummariesResponse): unknown {
    const obj: any = {};
    if (message.packageRepositorySummaries) {
      obj.packageRepositorySummaries = message.packageRepositorySummaries.map(e =>
        e ? PackageRepositorySummary.toJSON(e) : undefined,
      );
    } else {
      obj.packageRepositorySummaries = [];
    }
    return obj;
  },

  fromPartial(
    object: DeepPartial<GetPackageRepositorySummariesResponse>,
  ): GetPackageRepositorySummariesResponse {
    const message = {
      ...baseGetPackageRepositorySummariesResponse,
    } as GetPackageRepositorySummariesResponse;
    message.packageRepositorySummaries = [];
    if (
      object.packageRepositorySummaries !== undefined &&
      object.packageRepositorySummaries !== null
    ) {
      for (const e of object.packageRepositorySummaries) {
        message.packageRepositorySummaries.push(PackageRepositorySummary.fromPartial(e));
      }
    }
    return message;
  },
};

const baseUpdatePackageRepositoryResponse: object = {};

export const UpdatePackageRepositoryResponse = {
  encode(
    message: UpdatePackageRepositoryResponse,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.packageRepoRef !== undefined) {
      PackageRepositoryReference.encode(message.packageRepoRef, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): UpdatePackageRepositoryResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseUpdatePackageRepositoryResponse,
    } as UpdatePackageRepositoryResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.packageRepoRef = PackageRepositoryReference.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): UpdatePackageRepositoryResponse {
    const message = {
      ...baseUpdatePackageRepositoryResponse,
    } as UpdatePackageRepositoryResponse;
    if (object.packageRepoRef !== undefined && object.packageRepoRef !== null) {
      message.packageRepoRef = PackageRepositoryReference.fromJSON(object.packageRepoRef);
    } else {
      message.packageRepoRef = undefined;
    }
    return message;
  },

  toJSON(message: UpdatePackageRepositoryResponse): unknown {
    const obj: any = {};
    message.packageRepoRef !== undefined &&
      (obj.packageRepoRef = message.packageRepoRef
        ? PackageRepositoryReference.toJSON(message.packageRepoRef)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<UpdatePackageRepositoryResponse>,
  ): UpdatePackageRepositoryResponse {
    const message = {
      ...baseUpdatePackageRepositoryResponse,
    } as UpdatePackageRepositoryResponse;
    if (object.packageRepoRef !== undefined && object.packageRepoRef !== null) {
      message.packageRepoRef = PackageRepositoryReference.fromPartial(object.packageRepoRef);
    } else {
      message.packageRepoRef = undefined;
    }
    return message;
  },
};

const baseDeletePackageRepositoryResponse: object = {};

export const DeletePackageRepositoryResponse = {
  encode(_: DeletePackageRepositoryResponse, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): DeletePackageRepositoryResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseDeletePackageRepositoryResponse,
    } as DeletePackageRepositoryResponse;
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

  fromJSON(_: any): DeletePackageRepositoryResponse {
    const message = {
      ...baseDeletePackageRepositoryResponse,
    } as DeletePackageRepositoryResponse;
    return message;
  },

  toJSON(_: DeletePackageRepositoryResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(_: DeepPartial<DeletePackageRepositoryResponse>): DeletePackageRepositoryResponse {
    const message = {
      ...baseDeletePackageRepositoryResponse,
    } as DeletePackageRepositoryResponse;
    return message;
  },
};

/** Each repositories v1alpha1 plugin must implement at least the following rpcs: */
export interface RepositoriesService {
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

export class RepositoriesServiceClientImpl implements RepositoriesService {
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
      RepositoriesServiceAddPackageRepositoryDesc,
      AddPackageRepositoryRequest.fromPartial(request),
      metadata,
    );
  }

  GetPackageRepositoryDetail(
    request: DeepPartial<GetPackageRepositoryDetailRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetPackageRepositoryDetailResponse> {
    return this.rpc.unary(
      RepositoriesServiceGetPackageRepositoryDetailDesc,
      GetPackageRepositoryDetailRequest.fromPartial(request),
      metadata,
    );
  }

  GetPackageRepositorySummaries(
    request: DeepPartial<GetPackageRepositorySummariesRequest>,
    metadata?: grpc.Metadata,
  ): Promise<GetPackageRepositorySummariesResponse> {
    return this.rpc.unary(
      RepositoriesServiceGetPackageRepositorySummariesDesc,
      GetPackageRepositorySummariesRequest.fromPartial(request),
      metadata,
    );
  }

  UpdatePackageRepository(
    request: DeepPartial<UpdatePackageRepositoryRequest>,
    metadata?: grpc.Metadata,
  ): Promise<UpdatePackageRepositoryResponse> {
    return this.rpc.unary(
      RepositoriesServiceUpdatePackageRepositoryDesc,
      UpdatePackageRepositoryRequest.fromPartial(request),
      metadata,
    );
  }

  DeletePackageRepository(
    request: DeepPartial<DeletePackageRepositoryRequest>,
    metadata?: grpc.Metadata,
  ): Promise<DeletePackageRepositoryResponse> {
    return this.rpc.unary(
      RepositoriesServiceDeletePackageRepositoryDesc,
      DeletePackageRepositoryRequest.fromPartial(request),
      metadata,
    );
  }
}

export const RepositoriesServiceDesc = {
  serviceName: "kubeappsapis.core.packages.v1alpha1.RepositoriesService",
};

export const RepositoriesServiceAddPackageRepositoryDesc: UnaryMethodDefinitionish = {
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

export const RepositoriesServiceGetPackageRepositoryDetailDesc: UnaryMethodDefinitionish = {
  methodName: "GetPackageRepositoryDetail",
  service: RepositoriesServiceDesc,
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

export const RepositoriesServiceGetPackageRepositorySummariesDesc: UnaryMethodDefinitionish = {
  methodName: "GetPackageRepositorySummaries",
  service: RepositoriesServiceDesc,
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

export const RepositoriesServiceUpdatePackageRepositoryDesc: UnaryMethodDefinitionish = {
  methodName: "UpdatePackageRepository",
  service: RepositoriesServiceDesc,
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

export const RepositoriesServiceDeletePackageRepositoryDesc: UnaryMethodDefinitionish = {
  methodName: "DeletePackageRepository",
  service: RepositoriesServiceDesc,
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
