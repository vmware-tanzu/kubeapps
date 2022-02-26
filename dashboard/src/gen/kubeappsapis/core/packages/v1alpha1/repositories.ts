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
  /** authentication parameters for connecting to a repository. Optional */
  auth?: PackageRepositoryAuth;
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

const baseAddPackageRepositoryResponse: object = {};

export const AddPackageRepositoryResponse = {
  encode(_: AddPackageRepositoryResponse, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
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

  fromPartial(_: DeepPartial<AddPackageRepositoryResponse>): AddPackageRepositoryResponse {
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
    metadata?: grpc.Metadata,
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
    metadata?: grpc.Metadata,
  ): Promise<AddPackageRepositoryResponse> {
    return this.rpc.unary(
      RepositoriesServiceAddPackageRepositoryDesc,
      AddPackageRepositoryRequest.fromPartial(request),
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
