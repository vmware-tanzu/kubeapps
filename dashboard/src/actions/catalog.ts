import { ThunkAction } from "redux-thunk";
import { ActionType, createActionDeprecated } from "typesafe-actions";

import { IClusterServiceClass } from "../shared/ClusterServiceClass";
import { definedNamespaces } from "../shared/Namespace";
import { IServiceBindingWithSecret, ServiceBinding } from "../shared/ServiceBinding";
import { IServiceBroker, IServicePlan, ServiceCatalog } from "../shared/ServiceCatalog";
import { IServiceInstance, ServiceInstance } from "../shared/ServiceInstance";
import { IStoreState } from "../shared/types";

export const checkCatalogInstall = createActionDeprecated("CHECK_INSTALL");
export const installed = createActionDeprecated("INSTALLED");
export const notInstalled = createActionDeprecated("_NOT_INSTALLED");
export const requestBrokers = createActionDeprecated("REQUEST_BROKERS");
export const receiveBrokers = createActionDeprecated(
  "RECEIVE_BROKERS",
  (brokers: IServiceBroker[]) => ({
    brokers,
    type: "RECEIVE_BROKERS",
  }),
);
export const requestPlans = createActionDeprecated("REQUEST_PLANS");
export const receivePlans = createActionDeprecated("RECEIVE_PLANS", (plans: IServicePlan[]) => ({
  plans,
  type: "RECEIVE_PLANS",
}));
export const requestInstances = createActionDeprecated("REQUEST_INSTANCES");
export const receiveInstances = createActionDeprecated(
  "RECEIVE_INSTANCES",
  (instances: IServiceInstance[]) => ({ type: "RECEIVE_INSTANCES", instances }),
);
export const requestBindingsWithSecrets = createActionDeprecated("REQUEST_BINDINGS_WITH_SECRETS");
export const receiveBindingsWithSecrets = createActionDeprecated(
  "RECEIVE_BINDINGS_WITH_SECRETS",
  (bindingsWithSecrets: IServiceBindingWithSecret[]) => ({
    bindingsWithSecrets,
    type: "RECEIVE_BINDINGS_WITH_SECRETS",
  }),
);
export const requestClasses = createActionDeprecated("REQUEST_PLANS");
export const receiveClasses = createActionDeprecated(
  "RECEIVE_CLASSES",
  (classes: IClusterServiceClass[]) => ({
    classes,
    type: "RECEIVE_CLASSES",
  }),
);
export const errorCatalog = createActionDeprecated(
  "ERROR_CATALOG",
  (err: Error, op: "fetch" | "create" | "delete" | "deprovision" | "update") => ({
    err,
    op,
    type: "ERROR_CATALOG",
  }),
);

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
  return async dispatch => {
    try {
      await ServiceInstance.create(releaseName, namespace, className, planName, parameters);
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
  return async dispatch => {
    try {
      await ServiceBinding.create(bindingName, instanceName, namespace, parameters);
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
  return async dispatch => {
    try {
      await ServiceBinding.delete(name, namespace);
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
  return async dispatch => {
    try {
      await ServiceCatalog.deprovisionInstance(instance);
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
  return async dispatch => {
    try {
      await ServiceCatalog.syncBroker(broker);
    } catch (e) {
      dispatch(errorCatalog(e, "update"));
    }
  };
}

export function getBindings(
  ns?: string,
): ThunkAction<Promise<void>, IStoreState, null, ServiceCatalogAction> {
  return async dispatch => {
    if (ns && ns === definedNamespaces.all) {
      ns = undefined;
    }
    dispatch(requestBindingsWithSecrets());
    try {
      const bindingsWithSecrets = await ServiceBinding.list(ns);
      dispatch(receiveBindingsWithSecrets(bindingsWithSecrets));
    } catch (e) {
      dispatch(errorCatalog(e, "fetch"));
    }
  };
}

export function getBrokers(): ThunkAction<Promise<void>, IStoreState, null, ServiceCatalogAction> {
  return async dispatch => {
    dispatch(requestBrokers());
    try {
      const brokers = await ServiceCatalog.getServiceBrokers();
      dispatch(receiveBrokers(brokers));
    } catch (e) {
      dispatch(errorCatalog(e, "fetch"));
    }
  };
}

export function getClasses(): ThunkAction<Promise<void>, IStoreState, null, ServiceCatalogAction> {
  return async dispatch => {
    dispatch(requestClasses());
    try {
      const classes = await ServiceCatalog.getServiceClasses();
      dispatch(receiveClasses(classes));
    } catch (e) {
      dispatch(errorCatalog(e, "fetch"));
    }
  };
}

export function getInstances(
  ns?: string,
): ThunkAction<Promise<void>, IStoreState, null, ServiceCatalogAction> {
  return async dispatch => {
    if (ns && ns === definedNamespaces.all) {
      ns = undefined;
    }
    dispatch(requestInstances());
    try {
      const instances = await ServiceInstance.list(ns);
      dispatch(receiveInstances(instances));
    } catch (e) {
      dispatch(errorCatalog(e, "fetch"));
    }
  };
}

export function getPlans(): ThunkAction<Promise<void>, IStoreState, null, ServiceCatalogAction> {
  return async dispatch => {
    dispatch(requestPlans());
    try {
      const plans = await ServiceCatalog.getServicePlans();
      dispatch(receivePlans(plans));
    } catch (e) {
      dispatch(errorCatalog(e, "fetch"));
    }
  };
}

export function getCatalog(
  ns?: string,
): ThunkAction<Promise<void>, IStoreState, null, ServiceCatalogAction> {
  return async dispatch => {
    dispatch(getBindings(ns));
    dispatch(getBrokers());
    dispatch(getClasses());
    dispatch(getInstances(ns));
    dispatch(getPlans());
  };
}

export function checkCatalogInstalled(): ThunkAction<
  Promise<boolean>,
  IStoreState,
  null,
  ServiceCatalogAction
> {
  return async dispatch => {
    const isInstalled = await ServiceCatalog.isCatalogInstalled();
    isInstalled ? dispatch(installed()) : dispatch(notInstalled());
    return isInstalled;
  };
}
