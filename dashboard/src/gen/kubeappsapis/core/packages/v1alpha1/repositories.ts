/* eslint-disable */
import { grpc } from "@improbable-eng/grpc-web";
import { BrowserHeaders } from "browser-headers";
import _m0 from "protobufjs/minimal";
import { Any } from "../../../../google/protobuf/any";
import { Plugin } from "../../plugins/v1alpha1/plugins";
import { Context } from "./packages";

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
   * - for direct helm plug-in valid values are: "helm" and "oci"
   * - for flux plug-in valid values are: "helm" and "oci". In the
   *   future, we may add support for git and/or AWS s3-style buckets
   */
  type: string;
  /**
   * A URL identifying the package repository location. Must contain at
   * least a protocol and host
   */
  url: string;
  /**
   * The interval at which to check the upstream for updates (in time+unit)
   * Optional. Defaults to 10m if not specified
   */
  interval: string;
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
 * Authentication/authorization to provide client's identity when connecting
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
  /** SSH credentials */
  sshCreds?: SshCredentials | undefined;
  /** opaque credentials */
  opaqueCreds?: OpaqueCredentials | undefined;
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
  /** PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH - uses UsernamePassword */
  PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH = 1,
  /** PACKAGE_REPOSITORY_AUTH_TYPE_TLS - uses TlsCertKey */
  PACKAGE_REPOSITORY_AUTH_TYPE_TLS = 2,
  /** PACKAGE_REPOSITORY_AUTH_TYPE_BEARER - uses header */
  PACKAGE_REPOSITORY_AUTH_TYPE_BEARER = 3,
  /** PACKAGE_REPOSITORY_AUTH_TYPE_AUTHORIZATION_HEADER - uses header */
  PACKAGE_REPOSITORY_AUTH_TYPE_AUTHORIZATION_HEADER = 4,
  /** PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON - uses DockerCredentials */
  PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON = 5,
  /** PACKAGE_REPOSITORY_AUTH_TYPE_SSH - uses SshCredentials */
  PACKAGE_REPOSITORY_AUTH_TYPE_SSH = 6,
  /** PACKAGE_REPOSITORY_AUTH_TYPE_OPAQUE - uses OpaqueCredentials */
  PACKAGE_REPOSITORY_AUTH_TYPE_OPAQUE = 7,
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
    case "PACKAGE_REPOSITORY_AUTH_TYPE_AUTHORIZATION_HEADER":
      return PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_AUTHORIZATION_HEADER;
    case 5:
    case "PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON":
      return PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON;
    case 6:
    case "PACKAGE_REPOSITORY_AUTH_TYPE_SSH":
      return PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_SSH;
    case 7:
    case "PACKAGE_REPOSITORY_AUTH_TYPE_OPAQUE":
      return PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_OPAQUE;
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
    case PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_AUTHORIZATION_HEADER:
      return "PACKAGE_REPOSITORY_AUTH_TYPE_AUTHORIZATION_HEADER";
    case PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON:
      return "PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON";
    case PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_SSH:
      return "PACKAGE_REPOSITORY_AUTH_TYPE_SSH";
    case PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_OPAQUE:
      return "PACKAGE_REPOSITORY_AUTH_TYPE_OPAQUE";
    case PackageRepositoryAuth_PackageRepositoryAuthType.UNRECOGNIZED:
    default:
      return "UNRECOGNIZED";
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

/** SshCredentials */
export interface SshCredentials {
  /** private key */
  privateKey: string;
  /** known hosts. */
  knownHosts: string;
}

/** OpaqueCredentials */
export interface OpaqueCredentials {
  /** fields */
  data: { [key: string]: string };
}

export interface OpaqueCredentials_DataEntry {
  key: string;
  value: string;
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
   *   must of of type "kubernetes.io/dockerconfigjson"
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
   * The interval at which to check the upstream for updates (in time+unit)
   * Optional. Defaults to 10m if not specified
   */
  interval: string;
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
    case PackageRepositoryStatus_StatusReason.UNRECOGNIZED:
    default:
      return "UNRECOGNIZED";
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
  /** The interval at which to check the upstream for updates (in time+unit) */
  interval: string;
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
  /** existence of any authentication parameters for connecting to a repository. */
  requiresAuth: boolean;
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

function createBaseAddPackageRepositoryRequest(): AddPackageRepositoryRequest {
  return {
    context: undefined,
    name: "",
    description: "",
    namespaceScoped: false,
    type: "",
    url: "",
    interval: "",
    tlsConfig: undefined,
    auth: undefined,
    plugin: undefined,
    customDetail: undefined,
  };
}

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
    if (message.interval !== "") {
      writer.uint32(58).string(message.interval);
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
    const message = createBaseAddPackageRepositoryRequest();
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
          message.interval = reader.string();
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
    return {
      context: isSet(object.context) ? Context.fromJSON(object.context) : undefined,
      name: isSet(object.name) ? String(object.name) : "",
      description: isSet(object.description) ? String(object.description) : "",
      namespaceScoped: isSet(object.namespaceScoped) ? Boolean(object.namespaceScoped) : false,
      type: isSet(object.type) ? String(object.type) : "",
      url: isSet(object.url) ? String(object.url) : "",
      interval: isSet(object.interval) ? String(object.interval) : "",
      tlsConfig: isSet(object.tlsConfig)
        ? PackageRepositoryTlsConfig.fromJSON(object.tlsConfig)
        : undefined,
      auth: isSet(object.auth) ? PackageRepositoryAuth.fromJSON(object.auth) : undefined,
      plugin: isSet(object.plugin) ? Plugin.fromJSON(object.plugin) : undefined,
      customDetail: isSet(object.customDetail) ? Any.fromJSON(object.customDetail) : undefined,
    };
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

  fromPartial<I extends Exact<DeepPartial<AddPackageRepositoryRequest>, I>>(
    object: I,
  ): AddPackageRepositoryRequest {
    const message = createBaseAddPackageRepositoryRequest();
    message.context =
      object.context !== undefined && object.context !== null
        ? Context.fromPartial(object.context)
        : undefined;
    message.name = object.name ?? "";
    message.description = object.description ?? "";
    message.namespaceScoped = object.namespaceScoped ?? false;
    message.type = object.type ?? "";
    message.url = object.url ?? "";
    message.interval = object.interval ?? "";
    message.tlsConfig =
      object.tlsConfig !== undefined && object.tlsConfig !== null
        ? PackageRepositoryTlsConfig.fromPartial(object.tlsConfig)
        : undefined;
    message.auth =
      object.auth !== undefined && object.auth !== null
        ? PackageRepositoryAuth.fromPartial(object.auth)
        : undefined;
    message.plugin =
      object.plugin !== undefined && object.plugin !== null
        ? Plugin.fromPartial(object.plugin)
        : undefined;
    message.customDetail =
      object.customDetail !== undefined && object.customDetail !== null
        ? Any.fromPartial(object.customDetail)
        : undefined;
    return message;
  },
};

function createBasePackageRepositoryTlsConfig(): PackageRepositoryTlsConfig {
  return { insecureSkipVerify: false, certAuthority: undefined, secretRef: undefined };
}

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
    const message = createBasePackageRepositoryTlsConfig();
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
    return {
      insecureSkipVerify: isSet(object.insecureSkipVerify)
        ? Boolean(object.insecureSkipVerify)
        : false,
      certAuthority: isSet(object.certAuthority) ? String(object.certAuthority) : undefined,
      secretRef: isSet(object.secretRef)
        ? SecretKeyReference.fromJSON(object.secretRef)
        : undefined,
    };
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

  fromPartial<I extends Exact<DeepPartial<PackageRepositoryTlsConfig>, I>>(
    object: I,
  ): PackageRepositoryTlsConfig {
    const message = createBasePackageRepositoryTlsConfig();
    message.insecureSkipVerify = object.insecureSkipVerify ?? false;
    message.certAuthority = object.certAuthority ?? undefined;
    message.secretRef =
      object.secretRef !== undefined && object.secretRef !== null
        ? SecretKeyReference.fromPartial(object.secretRef)
        : undefined;
    return message;
  },
};

function createBasePackageRepositoryAuth(): PackageRepositoryAuth {
  return {
    type: 0,
    usernamePassword: undefined,
    tlsCertKey: undefined,
    dockerCreds: undefined,
    header: undefined,
    secretRef: undefined,
    sshCreds: undefined,
    opaqueCreds: undefined,
    passCredentials: false,
  };
}

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
    if (message.sshCreds !== undefined) {
      SshCredentials.encode(message.sshCreds, writer.uint32(66).fork()).ldelim();
    }
    if (message.opaqueCreds !== undefined) {
      OpaqueCredentials.encode(message.opaqueCreds, writer.uint32(74).fork()).ldelim();
    }
    if (message.passCredentials === true) {
      writer.uint32(56).bool(message.passCredentials);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): PackageRepositoryAuth {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBasePackageRepositoryAuth();
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
        case 8:
          message.sshCreds = SshCredentials.decode(reader, reader.uint32());
          break;
        case 9:
          message.opaqueCreds = OpaqueCredentials.decode(reader, reader.uint32());
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
    return {
      type: isSet(object.type)
        ? packageRepositoryAuth_PackageRepositoryAuthTypeFromJSON(object.type)
        : 0,
      usernamePassword: isSet(object.usernamePassword)
        ? UsernamePassword.fromJSON(object.usernamePassword)
        : undefined,
      tlsCertKey: isSet(object.tlsCertKey) ? TlsCertKey.fromJSON(object.tlsCertKey) : undefined,
      dockerCreds: isSet(object.dockerCreds)
        ? DockerCredentials.fromJSON(object.dockerCreds)
        : undefined,
      header: isSet(object.header) ? String(object.header) : undefined,
      secretRef: isSet(object.secretRef)
        ? SecretKeyReference.fromJSON(object.secretRef)
        : undefined,
      sshCreds: isSet(object.sshCreds) ? SshCredentials.fromJSON(object.sshCreds) : undefined,
      opaqueCreds: isSet(object.opaqueCreds)
        ? OpaqueCredentials.fromJSON(object.opaqueCreds)
        : undefined,
      passCredentials: isSet(object.passCredentials) ? Boolean(object.passCredentials) : false,
    };
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
    message.sshCreds !== undefined &&
      (obj.sshCreds = message.sshCreds ? SshCredentials.toJSON(message.sshCreds) : undefined);
    message.opaqueCreds !== undefined &&
      (obj.opaqueCreds = message.opaqueCreds
        ? OpaqueCredentials.toJSON(message.opaqueCreds)
        : undefined);
    message.passCredentials !== undefined && (obj.passCredentials = message.passCredentials);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<PackageRepositoryAuth>, I>>(
    object: I,
  ): PackageRepositoryAuth {
    const message = createBasePackageRepositoryAuth();
    message.type = object.type ?? 0;
    message.usernamePassword =
      object.usernamePassword !== undefined && object.usernamePassword !== null
        ? UsernamePassword.fromPartial(object.usernamePassword)
        : undefined;
    message.tlsCertKey =
      object.tlsCertKey !== undefined && object.tlsCertKey !== null
        ? TlsCertKey.fromPartial(object.tlsCertKey)
        : undefined;
    message.dockerCreds =
      object.dockerCreds !== undefined && object.dockerCreds !== null
        ? DockerCredentials.fromPartial(object.dockerCreds)
        : undefined;
    message.header = object.header ?? undefined;
    message.secretRef =
      object.secretRef !== undefined && object.secretRef !== null
        ? SecretKeyReference.fromPartial(object.secretRef)
        : undefined;
    message.sshCreds =
      object.sshCreds !== undefined && object.sshCreds !== null
        ? SshCredentials.fromPartial(object.sshCreds)
        : undefined;
    message.opaqueCreds =
      object.opaqueCreds !== undefined && object.opaqueCreds !== null
        ? OpaqueCredentials.fromPartial(object.opaqueCreds)
        : undefined;
    message.passCredentials = object.passCredentials ?? false;
    return message;
  },
};

function createBaseUsernamePassword(): UsernamePassword {
  return { username: "", password: "" };
}

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
    const message = createBaseUsernamePassword();
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
    return {
      username: isSet(object.username) ? String(object.username) : "",
      password: isSet(object.password) ? String(object.password) : "",
    };
  },

  toJSON(message: UsernamePassword): unknown {
    const obj: any = {};
    message.username !== undefined && (obj.username = message.username);
    message.password !== undefined && (obj.password = message.password);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<UsernamePassword>, I>>(object: I): UsernamePassword {
    const message = createBaseUsernamePassword();
    message.username = object.username ?? "";
    message.password = object.password ?? "";
    return message;
  },
};

function createBaseTlsCertKey(): TlsCertKey {
  return { cert: "", key: "" };
}

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
    const message = createBaseTlsCertKey();
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
    return {
      cert: isSet(object.cert) ? String(object.cert) : "",
      key: isSet(object.key) ? String(object.key) : "",
    };
  },

  toJSON(message: TlsCertKey): unknown {
    const obj: any = {};
    message.cert !== undefined && (obj.cert = message.cert);
    message.key !== undefined && (obj.key = message.key);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<TlsCertKey>, I>>(object: I): TlsCertKey {
    const message = createBaseTlsCertKey();
    message.cert = object.cert ?? "";
    message.key = object.key ?? "";
    return message;
  },
};

function createBaseDockerCredentials(): DockerCredentials {
  return { server: "", username: "", password: "", email: "" };
}

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
    const message = createBaseDockerCredentials();
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
    return {
      server: isSet(object.server) ? String(object.server) : "",
      username: isSet(object.username) ? String(object.username) : "",
      password: isSet(object.password) ? String(object.password) : "",
      email: isSet(object.email) ? String(object.email) : "",
    };
  },

  toJSON(message: DockerCredentials): unknown {
    const obj: any = {};
    message.server !== undefined && (obj.server = message.server);
    message.username !== undefined && (obj.username = message.username);
    message.password !== undefined && (obj.password = message.password);
    message.email !== undefined && (obj.email = message.email);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<DockerCredentials>, I>>(object: I): DockerCredentials {
    const message = createBaseDockerCredentials();
    message.server = object.server ?? "";
    message.username = object.username ?? "";
    message.password = object.password ?? "";
    message.email = object.email ?? "";
    return message;
  },
};

function createBaseSshCredentials(): SshCredentials {
  return { privateKey: "", knownHosts: "" };
}

export const SshCredentials = {
  encode(message: SshCredentials, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.privateKey !== "") {
      writer.uint32(10).string(message.privateKey);
    }
    if (message.knownHosts !== "") {
      writer.uint32(18).string(message.knownHosts);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): SshCredentials {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseSshCredentials();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.privateKey = reader.string();
          break;
        case 2:
          message.knownHosts = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): SshCredentials {
    return {
      privateKey: isSet(object.privateKey) ? String(object.privateKey) : "",
      knownHosts: isSet(object.knownHosts) ? String(object.knownHosts) : "",
    };
  },

  toJSON(message: SshCredentials): unknown {
    const obj: any = {};
    message.privateKey !== undefined && (obj.privateKey = message.privateKey);
    message.knownHosts !== undefined && (obj.knownHosts = message.knownHosts);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<SshCredentials>, I>>(object: I): SshCredentials {
    const message = createBaseSshCredentials();
    message.privateKey = object.privateKey ?? "";
    message.knownHosts = object.knownHosts ?? "";
    return message;
  },
};

function createBaseOpaqueCredentials(): OpaqueCredentials {
  return { data: {} };
}

export const OpaqueCredentials = {
  encode(message: OpaqueCredentials, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    Object.entries(message.data).forEach(([key, value]) => {
      OpaqueCredentials_DataEntry.encode(
        { key: key as any, value },
        writer.uint32(10).fork(),
      ).ldelim();
    });
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): OpaqueCredentials {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseOpaqueCredentials();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          const entry1 = OpaqueCredentials_DataEntry.decode(reader, reader.uint32());
          if (entry1.value !== undefined) {
            message.data[entry1.key] = entry1.value;
          }
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): OpaqueCredentials {
    return {
      data: isObject(object.data)
        ? Object.entries(object.data).reduce<{ [key: string]: string }>((acc, [key, value]) => {
            acc[key] = String(value);
            return acc;
          }, {})
        : {},
    };
  },

  toJSON(message: OpaqueCredentials): unknown {
    const obj: any = {};
    obj.data = {};
    if (message.data) {
      Object.entries(message.data).forEach(([k, v]) => {
        obj.data[k] = v;
      });
    }
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<OpaqueCredentials>, I>>(object: I): OpaqueCredentials {
    const message = createBaseOpaqueCredentials();
    message.data = Object.entries(object.data ?? {}).reduce<{ [key: string]: string }>(
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

function createBaseOpaqueCredentials_DataEntry(): OpaqueCredentials_DataEntry {
  return { key: "", value: "" };
}

export const OpaqueCredentials_DataEntry = {
  encode(
    message: OpaqueCredentials_DataEntry,
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

  decode(input: _m0.Reader | Uint8Array, length?: number): OpaqueCredentials_DataEntry {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseOpaqueCredentials_DataEntry();
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

  fromJSON(object: any): OpaqueCredentials_DataEntry {
    return {
      key: isSet(object.key) ? String(object.key) : "",
      value: isSet(object.value) ? String(object.value) : "",
    };
  },

  toJSON(message: OpaqueCredentials_DataEntry): unknown {
    const obj: any = {};
    message.key !== undefined && (obj.key = message.key);
    message.value !== undefined && (obj.value = message.value);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<OpaqueCredentials_DataEntry>, I>>(
    object: I,
  ): OpaqueCredentials_DataEntry {
    const message = createBaseOpaqueCredentials_DataEntry();
    message.key = object.key ?? "";
    message.value = object.value ?? "";
    return message;
  },
};

function createBaseSecretKeyReference(): SecretKeyReference {
  return { name: "", key: "" };
}

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
    const message = createBaseSecretKeyReference();
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
    return {
      name: isSet(object.name) ? String(object.name) : "",
      key: isSet(object.key) ? String(object.key) : "",
    };
  },

  toJSON(message: SecretKeyReference): unknown {
    const obj: any = {};
    message.name !== undefined && (obj.name = message.name);
    message.key !== undefined && (obj.key = message.key);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<SecretKeyReference>, I>>(object: I): SecretKeyReference {
    const message = createBaseSecretKeyReference();
    message.name = object.name ?? "";
    message.key = object.key ?? "";
    return message;
  },
};

function createBaseGetPackageRepositoryDetailRequest(): GetPackageRepositoryDetailRequest {
  return { packageRepoRef: undefined };
}

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
    const message = createBaseGetPackageRepositoryDetailRequest();
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
    return {
      packageRepoRef: isSet(object.packageRepoRef)
        ? PackageRepositoryReference.fromJSON(object.packageRepoRef)
        : undefined,
    };
  },

  toJSON(message: GetPackageRepositoryDetailRequest): unknown {
    const obj: any = {};
    message.packageRepoRef !== undefined &&
      (obj.packageRepoRef = message.packageRepoRef
        ? PackageRepositoryReference.toJSON(message.packageRepoRef)
        : undefined);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<GetPackageRepositoryDetailRequest>, I>>(
    object: I,
  ): GetPackageRepositoryDetailRequest {
    const message = createBaseGetPackageRepositoryDetailRequest();
    message.packageRepoRef =
      object.packageRepoRef !== undefined && object.packageRepoRef !== null
        ? PackageRepositoryReference.fromPartial(object.packageRepoRef)
        : undefined;
    return message;
  },
};

function createBaseGetPackageRepositorySummariesRequest(): GetPackageRepositorySummariesRequest {
  return { context: undefined };
}

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
    const message = createBaseGetPackageRepositorySummariesRequest();
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
    return { context: isSet(object.context) ? Context.fromJSON(object.context) : undefined };
  },

  toJSON(message: GetPackageRepositorySummariesRequest): unknown {
    const obj: any = {};
    message.context !== undefined &&
      (obj.context = message.context ? Context.toJSON(message.context) : undefined);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<GetPackageRepositorySummariesRequest>, I>>(
    object: I,
  ): GetPackageRepositorySummariesRequest {
    const message = createBaseGetPackageRepositorySummariesRequest();
    message.context =
      object.context !== undefined && object.context !== null
        ? Context.fromPartial(object.context)
        : undefined;
    return message;
  },
};

function createBaseUpdatePackageRepositoryRequest(): UpdatePackageRepositoryRequest {
  return {
    packageRepoRef: undefined,
    url: "",
    description: "",
    interval: "",
    tlsConfig: undefined,
    auth: undefined,
    customDetail: undefined,
  };
}

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
    if (message.interval !== "") {
      writer.uint32(34).string(message.interval);
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
    const message = createBaseUpdatePackageRepositoryRequest();
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
          message.interval = reader.string();
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
    return {
      packageRepoRef: isSet(object.packageRepoRef)
        ? PackageRepositoryReference.fromJSON(object.packageRepoRef)
        : undefined,
      url: isSet(object.url) ? String(object.url) : "",
      description: isSet(object.description) ? String(object.description) : "",
      interval: isSet(object.interval) ? String(object.interval) : "",
      tlsConfig: isSet(object.tlsConfig)
        ? PackageRepositoryTlsConfig.fromJSON(object.tlsConfig)
        : undefined,
      auth: isSet(object.auth) ? PackageRepositoryAuth.fromJSON(object.auth) : undefined,
      customDetail: isSet(object.customDetail) ? Any.fromJSON(object.customDetail) : undefined,
    };
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

  fromPartial<I extends Exact<DeepPartial<UpdatePackageRepositoryRequest>, I>>(
    object: I,
  ): UpdatePackageRepositoryRequest {
    const message = createBaseUpdatePackageRepositoryRequest();
    message.packageRepoRef =
      object.packageRepoRef !== undefined && object.packageRepoRef !== null
        ? PackageRepositoryReference.fromPartial(object.packageRepoRef)
        : undefined;
    message.url = object.url ?? "";
    message.description = object.description ?? "";
    message.interval = object.interval ?? "";
    message.tlsConfig =
      object.tlsConfig !== undefined && object.tlsConfig !== null
        ? PackageRepositoryTlsConfig.fromPartial(object.tlsConfig)
        : undefined;
    message.auth =
      object.auth !== undefined && object.auth !== null
        ? PackageRepositoryAuth.fromPartial(object.auth)
        : undefined;
    message.customDetail =
      object.customDetail !== undefined && object.customDetail !== null
        ? Any.fromPartial(object.customDetail)
        : undefined;
    return message;
  },
};

function createBaseDeletePackageRepositoryRequest(): DeletePackageRepositoryRequest {
  return { packageRepoRef: undefined };
}

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
    const message = createBaseDeletePackageRepositoryRequest();
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
    return {
      packageRepoRef: isSet(object.packageRepoRef)
        ? PackageRepositoryReference.fromJSON(object.packageRepoRef)
        : undefined,
    };
  },

  toJSON(message: DeletePackageRepositoryRequest): unknown {
    const obj: any = {};
    message.packageRepoRef !== undefined &&
      (obj.packageRepoRef = message.packageRepoRef
        ? PackageRepositoryReference.toJSON(message.packageRepoRef)
        : undefined);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<DeletePackageRepositoryRequest>, I>>(
    object: I,
  ): DeletePackageRepositoryRequest {
    const message = createBaseDeletePackageRepositoryRequest();
    message.packageRepoRef =
      object.packageRepoRef !== undefined && object.packageRepoRef !== null
        ? PackageRepositoryReference.fromPartial(object.packageRepoRef)
        : undefined;
    return message;
  },
};

function createBasePackageRepositoryReference(): PackageRepositoryReference {
  return { context: undefined, identifier: "", plugin: undefined };
}

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
    const message = createBasePackageRepositoryReference();
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
    return {
      context: isSet(object.context) ? Context.fromJSON(object.context) : undefined,
      identifier: isSet(object.identifier) ? String(object.identifier) : "",
      plugin: isSet(object.plugin) ? Plugin.fromJSON(object.plugin) : undefined,
    };
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

  fromPartial<I extends Exact<DeepPartial<PackageRepositoryReference>, I>>(
    object: I,
  ): PackageRepositoryReference {
    const message = createBasePackageRepositoryReference();
    message.context =
      object.context !== undefined && object.context !== null
        ? Context.fromPartial(object.context)
        : undefined;
    message.identifier = object.identifier ?? "";
    message.plugin =
      object.plugin !== undefined && object.plugin !== null
        ? Plugin.fromPartial(object.plugin)
        : undefined;
    return message;
  },
};

function createBaseAddPackageRepositoryResponse(): AddPackageRepositoryResponse {
  return { packageRepoRef: undefined };
}

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
    const message = createBaseAddPackageRepositoryResponse();
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
    return {
      packageRepoRef: isSet(object.packageRepoRef)
        ? PackageRepositoryReference.fromJSON(object.packageRepoRef)
        : undefined,
    };
  },

  toJSON(message: AddPackageRepositoryResponse): unknown {
    const obj: any = {};
    message.packageRepoRef !== undefined &&
      (obj.packageRepoRef = message.packageRepoRef
        ? PackageRepositoryReference.toJSON(message.packageRepoRef)
        : undefined);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<AddPackageRepositoryResponse>, I>>(
    object: I,
  ): AddPackageRepositoryResponse {
    const message = createBaseAddPackageRepositoryResponse();
    message.packageRepoRef =
      object.packageRepoRef !== undefined && object.packageRepoRef !== null
        ? PackageRepositoryReference.fromPartial(object.packageRepoRef)
        : undefined;
    return message;
  },
};

function createBasePackageRepositoryStatus(): PackageRepositoryStatus {
  return { ready: false, reason: 0, userReason: "" };
}

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
    const message = createBasePackageRepositoryStatus();
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
    return {
      ready: isSet(object.ready) ? Boolean(object.ready) : false,
      reason: isSet(object.reason)
        ? packageRepositoryStatus_StatusReasonFromJSON(object.reason)
        : 0,
      userReason: isSet(object.userReason) ? String(object.userReason) : "",
    };
  },

  toJSON(message: PackageRepositoryStatus): unknown {
    const obj: any = {};
    message.ready !== undefined && (obj.ready = message.ready);
    message.reason !== undefined &&
      (obj.reason = packageRepositoryStatus_StatusReasonToJSON(message.reason));
    message.userReason !== undefined && (obj.userReason = message.userReason);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<PackageRepositoryStatus>, I>>(
    object: I,
  ): PackageRepositoryStatus {
    const message = createBasePackageRepositoryStatus();
    message.ready = object.ready ?? false;
    message.reason = object.reason ?? 0;
    message.userReason = object.userReason ?? "";
    return message;
  },
};

function createBasePackageRepositoryDetail(): PackageRepositoryDetail {
  return {
    packageRepoRef: undefined,
    name: "",
    description: "",
    namespaceScoped: false,
    type: "",
    url: "",
    interval: "",
    tlsConfig: undefined,
    auth: undefined,
    customDetail: undefined,
    status: undefined,
  };
}

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
    if (message.interval !== "") {
      writer.uint32(58).string(message.interval);
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
    const message = createBasePackageRepositoryDetail();
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
          message.interval = reader.string();
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
    return {
      packageRepoRef: isSet(object.packageRepoRef)
        ? PackageRepositoryReference.fromJSON(object.packageRepoRef)
        : undefined,
      name: isSet(object.name) ? String(object.name) : "",
      description: isSet(object.description) ? String(object.description) : "",
      namespaceScoped: isSet(object.namespaceScoped) ? Boolean(object.namespaceScoped) : false,
      type: isSet(object.type) ? String(object.type) : "",
      url: isSet(object.url) ? String(object.url) : "",
      interval: isSet(object.interval) ? String(object.interval) : "",
      tlsConfig: isSet(object.tlsConfig)
        ? PackageRepositoryTlsConfig.fromJSON(object.tlsConfig)
        : undefined,
      auth: isSet(object.auth) ? PackageRepositoryAuth.fromJSON(object.auth) : undefined,
      customDetail: isSet(object.customDetail) ? Any.fromJSON(object.customDetail) : undefined,
      status: isSet(object.status) ? PackageRepositoryStatus.fromJSON(object.status) : undefined,
    };
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

  fromPartial<I extends Exact<DeepPartial<PackageRepositoryDetail>, I>>(
    object: I,
  ): PackageRepositoryDetail {
    const message = createBasePackageRepositoryDetail();
    message.packageRepoRef =
      object.packageRepoRef !== undefined && object.packageRepoRef !== null
        ? PackageRepositoryReference.fromPartial(object.packageRepoRef)
        : undefined;
    message.name = object.name ?? "";
    message.description = object.description ?? "";
    message.namespaceScoped = object.namespaceScoped ?? false;
    message.type = object.type ?? "";
    message.url = object.url ?? "";
    message.interval = object.interval ?? "";
    message.tlsConfig =
      object.tlsConfig !== undefined && object.tlsConfig !== null
        ? PackageRepositoryTlsConfig.fromPartial(object.tlsConfig)
        : undefined;
    message.auth =
      object.auth !== undefined && object.auth !== null
        ? PackageRepositoryAuth.fromPartial(object.auth)
        : undefined;
    message.customDetail =
      object.customDetail !== undefined && object.customDetail !== null
        ? Any.fromPartial(object.customDetail)
        : undefined;
    message.status =
      object.status !== undefined && object.status !== null
        ? PackageRepositoryStatus.fromPartial(object.status)
        : undefined;
    return message;
  },
};

function createBaseGetPackageRepositoryDetailResponse(): GetPackageRepositoryDetailResponse {
  return { detail: undefined };
}

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
    const message = createBaseGetPackageRepositoryDetailResponse();
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
    return {
      detail: isSet(object.detail) ? PackageRepositoryDetail.fromJSON(object.detail) : undefined,
    };
  },

  toJSON(message: GetPackageRepositoryDetailResponse): unknown {
    const obj: any = {};
    message.detail !== undefined &&
      (obj.detail = message.detail ? PackageRepositoryDetail.toJSON(message.detail) : undefined);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<GetPackageRepositoryDetailResponse>, I>>(
    object: I,
  ): GetPackageRepositoryDetailResponse {
    const message = createBaseGetPackageRepositoryDetailResponse();
    message.detail =
      object.detail !== undefined && object.detail !== null
        ? PackageRepositoryDetail.fromPartial(object.detail)
        : undefined;
    return message;
  },
};

function createBasePackageRepositorySummary(): PackageRepositorySummary {
  return {
    packageRepoRef: undefined,
    name: "",
    description: "",
    namespaceScoped: false,
    type: "",
    url: "",
    status: undefined,
    requiresAuth: false,
  };
}

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
    if (message.requiresAuth === true) {
      writer.uint32(64).bool(message.requiresAuth);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): PackageRepositorySummary {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBasePackageRepositorySummary();
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
        case 8:
          message.requiresAuth = reader.bool();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): PackageRepositorySummary {
    return {
      packageRepoRef: isSet(object.packageRepoRef)
        ? PackageRepositoryReference.fromJSON(object.packageRepoRef)
        : undefined,
      name: isSet(object.name) ? String(object.name) : "",
      description: isSet(object.description) ? String(object.description) : "",
      namespaceScoped: isSet(object.namespaceScoped) ? Boolean(object.namespaceScoped) : false,
      type: isSet(object.type) ? String(object.type) : "",
      url: isSet(object.url) ? String(object.url) : "",
      status: isSet(object.status) ? PackageRepositoryStatus.fromJSON(object.status) : undefined,
      requiresAuth: isSet(object.requiresAuth) ? Boolean(object.requiresAuth) : false,
    };
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
    message.requiresAuth !== undefined && (obj.requiresAuth = message.requiresAuth);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<PackageRepositorySummary>, I>>(
    object: I,
  ): PackageRepositorySummary {
    const message = createBasePackageRepositorySummary();
    message.packageRepoRef =
      object.packageRepoRef !== undefined && object.packageRepoRef !== null
        ? PackageRepositoryReference.fromPartial(object.packageRepoRef)
        : undefined;
    message.name = object.name ?? "";
    message.description = object.description ?? "";
    message.namespaceScoped = object.namespaceScoped ?? false;
    message.type = object.type ?? "";
    message.url = object.url ?? "";
    message.status =
      object.status !== undefined && object.status !== null
        ? PackageRepositoryStatus.fromPartial(object.status)
        : undefined;
    message.requiresAuth = object.requiresAuth ?? false;
    return message;
  },
};

function createBaseGetPackageRepositorySummariesResponse(): GetPackageRepositorySummariesResponse {
  return { packageRepositorySummaries: [] };
}

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
    const message = createBaseGetPackageRepositorySummariesResponse();
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
    return {
      packageRepositorySummaries: Array.isArray(object?.packageRepositorySummaries)
        ? object.packageRepositorySummaries.map((e: any) => PackageRepositorySummary.fromJSON(e))
        : [],
    };
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

  fromPartial<I extends Exact<DeepPartial<GetPackageRepositorySummariesResponse>, I>>(
    object: I,
  ): GetPackageRepositorySummariesResponse {
    const message = createBaseGetPackageRepositorySummariesResponse();
    message.packageRepositorySummaries =
      object.packageRepositorySummaries?.map(e => PackageRepositorySummary.fromPartial(e)) || [];
    return message;
  },
};

function createBaseUpdatePackageRepositoryResponse(): UpdatePackageRepositoryResponse {
  return { packageRepoRef: undefined };
}

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
    const message = createBaseUpdatePackageRepositoryResponse();
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
    return {
      packageRepoRef: isSet(object.packageRepoRef)
        ? PackageRepositoryReference.fromJSON(object.packageRepoRef)
        : undefined,
    };
  },

  toJSON(message: UpdatePackageRepositoryResponse): unknown {
    const obj: any = {};
    message.packageRepoRef !== undefined &&
      (obj.packageRepoRef = message.packageRepoRef
        ? PackageRepositoryReference.toJSON(message.packageRepoRef)
        : undefined);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<UpdatePackageRepositoryResponse>, I>>(
    object: I,
  ): UpdatePackageRepositoryResponse {
    const message = createBaseUpdatePackageRepositoryResponse();
    message.packageRepoRef =
      object.packageRepoRef !== undefined && object.packageRepoRef !== null
        ? PackageRepositoryReference.fromPartial(object.packageRepoRef)
        : undefined;
    return message;
  },
};

function createBaseDeletePackageRepositoryResponse(): DeletePackageRepositoryResponse {
  return {};
}

export const DeletePackageRepositoryResponse = {
  encode(_: DeletePackageRepositoryResponse, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): DeletePackageRepositoryResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseDeletePackageRepositoryResponse();
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
    return {};
  },

  toJSON(_: DeletePackageRepositoryResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<DeletePackageRepositoryResponse>, I>>(
    _: I,
  ): DeletePackageRepositoryResponse {
    const message = createBaseDeletePackageRepositoryResponse();
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
