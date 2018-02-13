import { IServiceCatalogState } from "../reducers/catalog";
import { IAppRepositoryState } from "../reducers/repos";
import { hapi } from "./hapi/release";

export interface IRepo {
  name: string;
  url: string;
}

export interface IChartVersion {
  id: string;
  attributes: IChartVersionAttributes;
  relationships: {
    chart: {
      data: IChartAttributes;
    };
  };
}

export interface IChartVersionAttributes {
  version: string;
  app_version: string;
  created: string;
}

export interface IChart {
  id: string;
  attributes: IChartAttributes;
  relationships: {
    latestChartVersion: {
      data: IChartVersionAttributes;
    };
  };
}

export interface IChartAttributes {
  name: string;
  description: string;
  home?: string;
  icon?: string;
  keywords: string[];
  maintainers: Array<{
    name: string;
    email?: string;
  }>;
  repo: IRepo;
  sources: string[];
}

export interface IChartState {
  isFetching: boolean;
  selected: {
    version?: IChartVersion;
    versions: IChartVersion[];
    readme?: string;
    values?: string;
  };
  items: IChart[];
}

export interface IDeployment {
  metadata: {
    name: string;
    namespace: string;
  };
}

export interface IServiceSpec {
  ports: IPort[];
  clusterIP: string;
  type: string;
}

export interface IPort {
  name: string;
  port: number;
  protocol: string;
  targetPort: string;
  nodePort: string;
}

export interface IResource {
  apiVersion: string;
  kind: string;
  type: string;
  spec: any;
  status: any;
  metadata: {
    name: string;
    namespace: string;
    annotations: string;
    creationTimestamp: string;
  };
}

export interface IApp {
  type: string;
  data: hapi.release.Release;
  repo?: IRepo;
}

export interface IAppState {
  isFetching: boolean;
  // currently items are always Helm releases
  items: IApp[];
  selected?: IApp;
}

export interface IStoreState {
  catalog: IServiceCatalogState;
  apps: IAppState;
  charts: IChartState;
  repos: IAppRepositoryState;
  deployment: IDeployment;
}

interface IK8sResource {
  apiVersion: string;
  kind: string;
}

/** @see https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#objects */
export interface IK8sObject<M, SP, ST> extends IK8sResource {
  metadata: {
    annotations?: { [key: string]: string };
    creationTimestamp?: string;
    deletionTimestamp?: string | null;
    generation?: number;
    labels?: { [key: string]: string };
    name: string;
    namespace: string;
    resourceVersion?: string;
    uid: string;
    selfLink?: string; // Not in docs, but seems to exist everywhere
  } & M;
  spec?: SP;
  status?: ST;
}

/** @see https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#lists-and-simple-kinds */
export interface IK8sList<I, M> extends IK8sResource {
  items: I[];
  metadata?: {
    resourceVersion?: string;
    selfLink?: string; // Not in docs, but seems to exist everywhere
  } & M;
}

export interface IAppRepository
  extends IK8sObject<
      {
        clusterName: string;
        creationTimestamp: string;
        deletionGracePeriodSeconds: string | null;
        deletionTimestamp: string | null;
        resourceVersion: string;
        selfLink: string;
      },
      { type: string; url: string },
      undefined
    > {}

export interface IAppRepositoryList
  extends IK8sList<
      IAppRepository,
      {
        continue: string;
        resourceVersion: string;
        selfLink: string;
      }
    > {}

export interface IClusterServicePlan
  extends IK8sObject<
      {
        namespace: undefined;
        selfLink: string;
        resourceVersion: string;
        creationTimestamp: string;
      },
      {
        clusterServiceBrokerName: string;
        externalName: string;
        externalID: string;
        description: string;
        free: boolean;
        clusterServiceClassRef: { [key: string]: string };
      },
      { removedFromBrokerCatalog: boolean }
    > {}

export interface IClusterServicePlanList
  extends IK8sList<IClusterServicePlan, { selfLink: string; resourceVersion: string }> {}

export interface IClusterServiceClass
  extends IK8sObject<
      {
        creationTimestamp: string;
        namespace: undefined;
        resourceVersion: string;
        selfLink: string;
      },
      {
        clusterServiceBrokerName: string;
        externalName: string;
        externalID: string;
        description: string;
        bindable: boolean;
        binding_retrievable: string;
        planUpdatable: string;
        tags: string[];
      },
      { removedFromBrokerCatalog: boolean }
    > {}

export interface IClusterServiceClassList
  extends IK8sList<IClusterServiceClass, { selfLink: string; resourceVersion: string }> {}

interface ICondition {
  type: string;
  status: string;
  lastTransitionTime: string;
  reason: string;
  message: string;
}

/**
 * @property authInfo Unknown what kind of auth types their are
 */
export interface IClusterServiceBroker
  extends IK8sObject<
      {
        resourceVersion: string;
        generation: number;
        creationTimestamp: string;
        finalizers: string[];
      },
      {
        url: string;
        authInfo: any;
        relistBehavior: string;
        relistDuration: string;
        relistRequests: number;
      },
      {
        conditions: ICondition[];
        reconciledGeneration: number;
        lastCatalogRetrievalTime: string;
      }
    > {}

export interface IClusterServiceBrokerList
  extends IK8sList<IClusterServiceBroker, { selfLink: string; resourceVersion: string }> {}

/** @see https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#response-status-kind */
export interface IStatus extends IK8sResource {
  kind: "Status";
  status: "Success" | "Failure";
  message: string;
  reason:
    | "BadRequest"
    | "Unauthorized"
    | "Forbidden"
    | "NotFound"
    | "AlreadyExists"
    | "Conflict"
    | "Invalid"
    | "Timeout"
    | "ServerTimeout"
    | "MethodNotAllowed"
    | "InternalError";
  details?: {
    kind?: string;
    name?: string;
    causes?: IStatusCause[] | string;
  };
}

interface IStatusCause {
  field: string;
  message: string;
  reason: string;
}

// Representation of the HelmRelease CRD
export interface IHelmRelease {
  metadata: {
    annotations: {
      "apprepositories.kubeapps.com/repo-name"?: string;
    };
    name: string;
    namespace: string;
  };
  spec: {
    repoUrl: string;
  };
}

// Representation of the ConfigMaps Helm uses to store releases
export interface IHelmReleaseConfigMap {
  metadata: {
    labels: {
      NAME: string;
      VERSION: string;
    };
  };
  data: {
    release: string;
  };
}
