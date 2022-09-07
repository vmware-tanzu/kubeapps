// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { JSONSchemaType } from "ajv";
import { RouterState } from "connected-react-router";
import {
  AvailablePackageDetail,
  AvailablePackageSummary,
  GetAvailablePackageSummariesResponse,
  InstalledPackageDetail,
  InstalledPackageSummary,
  PackageAppVersion,
  ResourceRef,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import {
  DockerCredentials,
  OpaqueCredentials,
  PackageRepositoryAuth_PackageRepositoryAuthType,
  SshCredentials,
  TlsCertKey,
  UsernamePassword,
} from "gen/kubeappsapis/core/packages/v1alpha1/repositories";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import { HelmPackageRepositoryCustomDetail } from "gen/kubeappsapis/plugins/helm/packages/v1alpha1/helm";
import { KappControllerPackageRepositoryCustomDetail } from "gen/kubeappsapis/plugins/kapp_controller/packages/v1alpha1/kapp_controller";
import { IOperatorsState } from "reducers/operators";
import { Subscription } from "rxjs";
import { IAuthState } from "../reducers/auth";
import { IClustersState } from "../reducers/cluster";
import { IConfigState } from "../reducers/config";
import { IPackageRepositoryState } from "../reducers/repos";
import { RpcError } from "./RpcError";

export class CustomError extends Error {
  public causes: Error[] | undefined;
  // The constructor is defined so we can later on compare the returned object
  // via err.constructor  == FOO
  constructor(message?: string, causes?: Error[]) {
    super(message);
    Object.setPrototypeOf(this, new.target.prototype);
    this.causes = causes;
    this.checkCauses();
  }
  // Workaround used until RPC code (unary) throws a custom rpc error
  // Check if any RPC error is among the causes
  private checkCauses() {
    if (!this.causes) return;
    for (let i = 0; i < this.causes.length; i++) {
      const cause = this.causes[i];
      if (RpcError.isRpcError(cause)) {
        this.causes[i] = new RpcError(cause);
      }
    }
  }
}

// For 4XX HTTP-alike errors
export class ClientNetworkError extends CustomError {}
// 400
export class BadRequestNetworkError extends ClientNetworkError {}
// 401
export class UnauthorizedNetworkError extends ClientNetworkError {}
// 403
export class ForbiddenNetworkError extends ClientNetworkError {}
// 404
export class NotFoundNetworkError extends ClientNetworkError {}
// 408
export class RequestTimeoutNetworkError extends ClientNetworkError {}
// 409
export class ConflictNetworkError extends ClientNetworkError {}
// 422
export class UnprocessableEntityError extends ClientNetworkError {}
// 429
export class TooManyRequestsNetworkError extends ClientNetworkError {}

// For 5XX HTTP-alike errors
export class ServerNetworkError extends CustomError {}
// 500
export class InternalServerNetworkError extends ServerNetworkError {}
// 501
export class NotImplementedNetworkError extends ServerNetworkError {}
// 503
export class ServerUnavailableNetworkError extends ServerNetworkError {}
// 504
export class GatewayTimeoutNetworkError extends ServerNetworkError {}

// Application-level errors
export class FetchError extends CustomError {}

export class FetchWarning extends CustomError {}

export class CreateError extends CustomError {}

export class UpgradeError extends CustomError {}

export class RollbackError extends CustomError {}

export class DeleteError extends CustomError {}

export type DeploymentEvent = "install" | "upgrade";

export interface IReceivePackagesActionPayload {
  response: GetAvailablePackageSummariesResponse;
  paginationToken: string;
}

export interface IPackageState {
  isFetching: boolean;
  hasFinishedFetching: boolean;
  nextPageToken: string;
  selected: {
    error?: FetchError | Error;
    availablePackageDetail?: AvailablePackageDetail;
    pkgVersion?: string;
    appVersion?: string;
    versions: PackageAppVersion[];
    readme?: string;
    readmeError?: string;
    values?: string;
    schema?: JSONSchemaType<any>;
  };
  items: AvailablePackageSummary[];
  categories: string[];
  size: number;
}

export interface IServiceSpec {
  ports: IPort[];
  clusterIP: string;
  type: string;
}

export interface IServiceStatus {
  loadBalancer: {
    ingress?: Array<{ ip?: string; hostname?: string }>;
  };
}

export interface IPort {
  name: string;
  port: number;
  protocol: string;
  targetPort: string;
  nodePort: string;
}

export interface IHTTPIngressPath {
  path: string;
}

export interface IIngressHTTP {
  paths: IHTTPIngressPath[];
}

export interface IIngressRule {
  host: string;
  http: IIngressHTTP;
}

export interface IIngressTLS {
  hosts: string[];
}

export interface IIngressSpec {
  rules: IIngressRule[];
  tls?: IIngressTLS[];
  backend?: {
    serviceName: string;
    servicePort: number;
  };
}

export interface IResourceMetadata {
  name: string;
  namespace: string;
  annotations: { [key: string]: string };
  ownerReferences?: Array<{
    apiVersion: string;
    blockOwnerDeletion: string;
    kind: string;
    name: string;
    uid: string;
  }>;
  creationTimestamp: string;
  selfLink: string;
  resourceVersion: string;
  deletionTimestamp?: string;
  uid: string;
}

export interface IResource {
  apiVersion: string;
  kind: string;
  type: string;
  spec: any;
  status: any;
  metadata: IResourceMetadata;
}

export interface ISecret {
  apiVersion: string;
  kind: string;
  type: string;
  data: { [s: string]: string };
  metadata: IResourceMetadata;
}

export interface IPackageManifestChannel {
  name: string;
  currentCSV: string;
  currentCSVDesc: {
    annotations: {
      "alm-examples": string;
      capabilities: string;
      categories: string;
      certified: string;
      containerImage: string;
      createdAt: string;
      description: string;
      repository: string;
      support: string;
    };
    description: string;
    displayName: string;
    provider: {
      name: string;
    };
    version: string;
    installModes: [
      { supported: boolean; type: "OwnNamespace" },
      { supported: boolean; type: "SingleNamespace" },
      { supported: boolean; type: "MultiNamespace" },
      { supported: boolean; type: "AllNamespaces" },
    ];
    customresourcedefinitions: {
      owned: Array<{
        description: string;
        displayName: string;
        kind: string;
        name: string;
        version: string;
      }>;
    };
  };
}

export interface IPackageManifestStatus {
  catalogSource: string;
  catalogSourceDisplayName: string;
  catalogSourceNamespace: string;
  catalogSourcePublisher: string;
  provider: {
    name: string;
  };
  defaultChannel: string;
  channels: IPackageManifestChannel[];
}

export interface IPackageManifest extends IResource {
  status: IPackageManifestStatus;
}

export interface IClusterServiceVersionCRDResource {
  kind: string;
  name: string;
  version: string;
}

export interface IClusterServiceVersionCRD {
  description: string;
  displayName: string;
  kind: string;
  name: string;
  version: string;
  resources: IClusterServiceVersionCRDResource[];
  specDescriptors: Array<{
    description: string;
    displayName: string;
    path: string;
    "x-descriptors": string[];
  }>;
  statusDescriptors: Array<{
    description: string;
    displayName: string;
    path: string;
    "x-descriptors": string[];
  }>;
}

export interface IClusterServiceVersionSpec {
  apiservicedefinitions: any;
  customresourcedefinitions: {
    owned?: IClusterServiceVersionCRD[];
  };
  description: string;
  displayName: string;
  icon: Array<{
    base64data: string;
    mediatype: string;
  }>;
  install: any;
  installModes: [
    { supported: boolean; type: "OwnNamespace" },
    { supported: boolean; type: "SingleNamespace" },
    { supported: boolean; type: "MultiNamespace" },
    { supported: boolean; type: "AllNamespaces" },
  ];
  keywords: string[];
  labels: any;
  links: Array<{
    name: string;
    url: string;
  }>;
  maintainers: Array<{
    email: string;
    name: string;
  }>;
  maturity: string;
  provider: {
    name: string;
  };
  selector: any;
  version: string;
}

export interface IClusterServiceVersion extends IResource {
  spec: IClusterServiceVersionSpec;
}

export interface IPkgRepositoryFilter {
  jq: string;
  variables?: { [key: string]: string };
}

export interface IInstalledPackageState {
  isFetching: boolean;
  error?: FetchError | CreateError | UpgradeError | RollbackError | DeleteError;
  items: InstalledPackageDetail[];
  listOverview?: InstalledPackageSummary[];
  selected?: CustomInstalledPackageDetail;
  // TODO(agamez): add tests for this new state field
  selectedDetails?: AvailablePackageDetail;
  resourceRefs?: ResourceRef[];
}

export interface IStoreState {
  router: RouterState;
  apps: IInstalledPackageState;
  auth: IAuthState;
  packages: IPackageState;
  config: IConfigState;
  kube: IKubeState;
  repos: IPackageRepositoryState;
  clusters: IClustersState;
  operators: IOperatorsState;
}

interface IK8sResource {
  apiVersion: string;
  kind: string;
}

/** @see https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#lists-and-simple-kinds */
export interface IK8sList<I, M> extends IK8sResource {
  items: I[];
  metadata?: {
    resourceVersion?: string;
    selfLink?: string; // Not in docs, but seems to exist everywhere
  } & M;
}

export interface IRBACRole {
  apiGroup: string;
  namespace?: string;
  clusterWide?: boolean;
  resource: string;
  verbs: string[];
}

export interface IKubeItem<T> {
  isFetching: boolean;
  item?: T;
  error?: Error;
}

export interface IKind {
  apiVersion: string;
  plural: string;
  namespaced: boolean;
}

export interface IKubeState {
  items: { [s: string]: IKubeItem<IResource | IK8sList<IResource, {}>> };
  subscriptions: { [s: string]: Subscription };
  // TODO(minelson): Remove kinds and kindsError once the operator support is
  // removed from the dashboard or replaced with a plugin.
  kinds: { [kind: string]: IKind };
  kindsError?: Error;
}

export interface IBasicFormParam {
  path: string;
  type?: "string" | "number" | "integer" | "boolean" | "object" | "array" | "null" | "any";
  value?: any;
  title?: string;
  minimum?: number;
  maximum?: number;
  render?: string;
  description?: string;
  customComponent?: object;
  enum?: string[];
  hidden?:
    | {
        event: DeploymentEvent;
        path: string;
        value: string;
        conditions: Array<{
          event: DeploymentEvent;
          path: string;
          value: string;
        }>;
        operator: string;
      }
    | string;
  children?: IBasicFormParam[];
}
export interface IBasicFormSliderParam extends IBasicFormParam {
  sliderMin?: number;
  sliderMax?: number;
  sliderStep?: number;
  sliderUnit?: string;
}

export interface CustomInstalledPackageDetail extends InstalledPackageDetail {
  revision: number;
}

//  enum for the type of package repository storage
export enum RepositoryStorageTypes {
  PACKAGE_REPOSITORY_STORAGE_HELM = "helm",
  PACKAGE_REPOSITORY_STORAGE_OCI = "oci",
  PACKAGE_REPOSITORY_STORAGE_CARVEL_INLINE = "inline",
  PACKAGE_REPOSITORY_STORAGE_CARVEL_IMAGE = "image",
  PACKAGE_REPOSITORY_STORAGE_CARVEL_IMGPKGBUNDLE = "imgpkgBundle",
  PACKAGE_REPOSITORY_STORAGE_CARVEL_HTTP = "http",
  PACKAGE_REPOSITORY_STORAGE_CARVEL_GIT = "git",
}

export enum PluginNames {
  PACKAGES_HELM = "helm.packages",
  PACKAGES_FLUX = "fluxv2.packages",
  PACKAGES_KAPP = "kapp_controller.packages",
}

export interface IPkgRepoFormData {
  authMethod: PackageRepositoryAuth_PackageRepositoryAuthType;
  // kubeapps-managed secrets
  authHeader: string; // used if type == PACKAGE_REPOSITORY_AUTH_TYPE_BEARER || type == PACKAGE_REPOSITORY_AUTH_TYPE_AUTHORIZATION_HEADER
  basicAuth: UsernamePassword; // used if type == PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH
  dockerRegCreds: DockerCredentials; // used if type == PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON
  sshCreds: SshCredentials; // used if type == PACKAGE_REPOSITORY_AUTH_TYPE_SSH
  opaqueCreds: OpaqueCredentials; // used if type == PACKAGE_REPOSITORY_AUTH_TYPE_OPAQUE
  tlsCertKey: TlsCertKey; // used if type == PACKAGE_REPOSITORY_AUTH_TYPE_TLS
  // user-managed secrets
  secretAuthName: string;
  secretTLSName: string;
  // rest of the parameters
  customCA: string;
  description: string;
  interval: string;
  name: string;
  passCredentials: boolean;
  plugin: Plugin;
  skipTLS: boolean;
  type: string;
  url: string;
  // add more types if necessary
  customDetail?: Partial<
    HelmPackageRepositoryCustomDetail | KappControllerPackageRepositoryCustomDetail
  >;
  namespace: string;
  isNamespaceScoped: boolean;
}
