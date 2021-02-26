import { ThunkAction } from "redux-thunk";
import { ActionType, deprecated } from "typesafe-actions";

import { IClusterServiceClass } from "../shared/ClusterServiceClass";
import { IServiceBindingWithSecret, ServiceBinding } from "../shared/ServiceBinding";
import { IServiceBroker, IServicePlan, ServiceCatalog } from "../shared/ServiceCatalog";
import { IServiceInstance, ServiceInstance } from "../shared/ServiceInstance";
import { IStoreState } from "../shared/types";
import helpers from "./helpers";

const { createAction } = deprecated;

export const checkCatalogInstall = createAction("CHECK_INSTALL");
export const installed = createAction("INSTALLED");
export const notInstalled = createAction("NOT_INSTALLED");
export const requestBrokers = createAction("REQUEST_BROKERS");
export const receiveBrokers = createAction("RECEIVE_BROKERS", resolve => {
  return (brokers: IServiceBroker[]) => resolve(brokers);
});

export const requestPlans = createAction("REQUEST_PLANS");
export const receivePlans = createAction("RECEIVE_PLANS", resolve => {
  return (plans: IServicePlan[]) => resolve(plans);
});

export const requestInstances = createAction("REQUEST_INSTANCES");
export const receiveInstances = createAction("RECEIVE_INSTANCES", resolve => {
  return (instances: IServiceInstance[]) => resolve(instances);
});

export const requestBindingsWithSecrets = createAction("REQUEST_BINDINGS_WITH_SECRETS");
export const receiveBindingsWithSecrets = createAction("RECEIVE_BINDINGS_WITH_SECRETS", resolve => {
  return (bindingsWithSecrets: IServiceBindingWithSecret[]) => resolve(bindingsWithSecrets);
});

export const requestClasses = createAction("REQUEST_PLANS");
export const receiveClasses = createAction("RECEIVE_CLASSES", resolve => {
  return (classes: IClusterServiceClass[]) => resolve(classes);
});

export const errorCatalog = createAction("ERROR_CATALOG", resolve => {
  return (err: Error, op: "fetch" | "create" | "delete" | "deprovision" | "update") =>
    resolve({ err, op });
});

const actions = [
  checkCatalogInstall,
  installed,
  notInstalled,
  requestBrokers,
  receiveBrokers,
  requestPlans,
  receivePlans,
  requestInstances,
  receiveInstances,
  requestBindingsWithSecrets,
  receiveBindingsWithSecrets,
  requestClasses,
  receiveClasses,
  errorCatalog,
];

export type ServiceCatalogAction = ActionType<typeof actions[number]>;

export function provision(
  releaseName: string,
  namespace: string,
  className: string,
  planName: string,
  parameters: {},
): ThunkAction<Promise<boolean>, IStoreState, null, ServiceCatalogAction> {
  return async (dispatch, getState) => {
    const {
      clusters: { currentCluster },
    } = getState();
    try {
      const filteredParams = helpers.object.removeEmptyFields(parameters);
      await ServiceInstance.create(
        currentCluster,
        releaseName,
        namespace,
        className,
        planName,
        filteredParams,
      );
      return true;
    } catch (e) {
      dispatch(errorCatalog(e, "create"));
      return false;
    }
  };
}

export function addBinding(
  bindingName: string,
  instanceName: string,
  namespace: string,
  parameters: {},
): ThunkAction<Promise<boolean>, IStoreState, null, ServiceCatalogAction> {
  return async (dispatch, getState) => {
    const {
      clusters: { currentCluster },
    } = getState();
    try {
      const filteredParams = helpers.object.removeEmptyFields(parameters);
      await ServiceBinding.create(
        bindingName,
        instanceName,
        currentCluster,
        namespace,
        filteredParams,
      );
      return true;
    } catch (e) {
      dispatch(errorCatalog(e, "create"));
      return false;
    }
  };
}

export function removeBinding(
  name: string,
  namespace: string,
): ThunkAction<Promise<boolean>, IStoreState, null, ServiceCatalogAction> {
  return async (dispatch, getState) => {
    const {
      clusters: { currentCluster },
    } = getState();
    try {
      await ServiceBinding.delete(currentCluster, namespace, name);
      return true;
    } catch (e) {
      dispatch(errorCatalog(e, "delete"));
      return false;
    }
  };
}

export function deprovision(
  instance: IServiceInstance,
): ThunkAction<Promise<boolean>, IStoreState, null, ServiceCatalogAction> {
  return async (dispatch, getState) => {
    const {
      clusters: { currentCluster },
    } = getState();
    try {
      await ServiceCatalog.deprovisionInstance(currentCluster, instance);
      return true;
    } catch (e) {
      dispatch(errorCatalog(e, "deprovision"));
      return false;
    }
  };
}

export function sync(
  broker: IServiceBroker,
): ThunkAction<Promise<void>, IStoreState, null, ServiceCatalogAction> {
  return async (dispatch, getState) => {
    const {
      config: { kubeappsCluster },
    } = getState();
    try {
      await ServiceCatalog.syncBroker(kubeappsCluster, broker);
    } catch (e) {
      dispatch(errorCatalog(e, "update"));
    }
  };
}

export function getBindings(
  ns?: string,
): ThunkAction<Promise<void>, IStoreState, null, ServiceCatalogAction> {
  return async (dispatch, getState) => {
    const {
      config: { kubeappsCluster },
    } = getState();
    dispatch(requestBindingsWithSecrets());
    try {
      const bindingsWithSecrets = await ServiceBinding.list(kubeappsCluster, ns);
      dispatch(receiveBindingsWithSecrets(bindingsWithSecrets));
    } catch (e) {
      dispatch(errorCatalog(e, "fetch"));
    }
  };
}

export function getBrokers(): ThunkAction<Promise<void>, IStoreState, null, ServiceCatalogAction> {
  return async (dispatch, getState) => {
    const {
      clusters: { currentCluster },
    } = getState();
    dispatch(requestBrokers());
    try {
      const brokers = await ServiceCatalog.getServiceBrokers(currentCluster);
      dispatch(receiveBrokers(brokers));
    } catch (e) {
      dispatch(errorCatalog(e, "fetch"));
    }
  };
}

export function getClasses(): ThunkAction<Promise<void>, IStoreState, null, ServiceCatalogAction> {
  return async (dispatch, getState) => {
    const {
      clusters: { currentCluster },
    } = getState();
    dispatch(requestClasses());
    try {
      const classes = await ServiceCatalog.getServiceClasses(currentCluster);
      dispatch(receiveClasses(classes));
    } catch (e) {
      dispatch(errorCatalog(e, "fetch"));
    }
  };
}

export function getInstances(
  ns?: string,
): ThunkAction<Promise<void>, IStoreState, null, ServiceCatalogAction> {
  return async (dispatch, getState) => {
    const {
      clusters: { currentCluster },
    } = getState();
    dispatch(requestInstances());
    try {
      const instances = await ServiceInstance.list(currentCluster, ns);
      dispatch(receiveInstances(instances));
    } catch (e) {
      dispatch(errorCatalog(e, "fetch"));
    }
  };
}

export function getPlans(): ThunkAction<Promise<void>, IStoreState, null, ServiceCatalogAction> {
  return async (dispatch, getState) => {
    const {
      clusters: { currentCluster },
    } = getState();
    dispatch(requestPlans());
    try {
      const plans = await ServiceCatalog.getServicePlans(currentCluster);
      dispatch(receivePlans(plans));
    } catch (e) {
      dispatch(errorCatalog(e, "fetch"));
    }
  };
}

export function checkCatalogInstalled(): ThunkAction<
  Promise<boolean>,
  IStoreState,
  null,
  ServiceCatalogAction
> {
  return async (dispatch, getState) => {
    const {
      clusters: { currentCluster },
    } = getState();
    const isServiceCatalogInstalled = await ServiceCatalog.isCatalogInstalled(currentCluster);
    isServiceCatalogInstalled ? dispatch(installed()) : dispatch(notInstalled());
    return isServiceCatalogInstalled;
  };
}
