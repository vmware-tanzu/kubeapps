import { Dispatch } from "redux";
import { createAction, getReturnOfExpression } from "typesafe-actions";

import { IClusterServiceClass } from "../shared/ClusterServiceClass";
import { IServiceBinding, ServiceBinding } from "../shared/ServiceBinding";
import { IServiceBroker, IServicePlan, ServiceCatalog } from "../shared/ServiceCatalog";
import { IServiceInstance, ServiceInstance } from "../shared/ServiceInstance";
import { IStoreState } from "../shared/types";

export const checkCatalogInstall = createAction("CHECK_INSTALL");
export const installed = createAction("INSTALLED");
export const notInstalled = createAction("_NOT_INSTALLED");
export const requestBrokers = createAction("REQUEST_BROKERS");
export const receiveBrokers = createAction("RECEIVE_BROKERS", (brokers: IServiceBroker[]) => ({
  brokers,
  type: "RECEIVE_BROKERS",
}));
export const requestPlans = createAction("REQUEST_PLANS");
export const receivePlans = createAction("RECEIVE_PLANS", (plans: IServicePlan[]) => ({
  plans,
  type: "RECEIVE_PLANS",
}));
export const requestInstances = createAction("REQUEST_INSTANCES");
export const receiveInstances = createAction(
  "RECEIVE_INSTANCES",
  (instances: IServiceInstance[]) => ({ type: "RECEIVE_INSTANCES", instances }),
);
export const requestBindings = createAction("REQUEST_BINDINGS");
export const receiveBindings = createAction("RECEIVE_BINDINGS", (bindings: IServiceBinding[]) => ({
  bindings,
  type: "RECEIVE_BINDINGS",
}));
export const requestClasses = createAction("REQUEST_PLANS");
export const receiveClasses = createAction(
  "RECEIVE_CLASSES",
  (classes: IClusterServiceClass[]) => ({
    classes,
    type: "RECEIVE_CLASSES",
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
  requestBindings,
  receiveBindings,
  requestClasses,
  receiveClasses,
].map(getReturnOfExpression);

export function provision(
  releaseName: string,
  namespace: string,
  className: string,
  planName: string,
  parameters: {},
) {
  return async (dispatch: Dispatch<IStoreState>) => {
    return ServiceInstance.create(releaseName, namespace, className, planName, parameters);
  };
}

export function deprovision(instance: IServiceInstance) {
  return async (dispatch: Dispatch<IStoreState>) => {
    return ServiceCatalog.deprovisionInstance(instance);
  };
}

export function sync(broker: IServiceBroker) {
  return async (dispatch: Dispatch<IStoreState>) => {
    return ServiceCatalog.syncBroker(broker);
  };
}

export type ServiceCatalogAction = typeof actions[number];

export function getBindings() {
  return async (dispatch: Dispatch<IStoreState>) => {
    dispatch(requestBindings());
    const bindings = await ServiceBinding.list();
    dispatch(receiveBindings(bindings));
    return bindings;
  };
}

export function getBrokers() {
  return async (dispatch: Dispatch<IStoreState>) => {
    dispatch(requestBrokers());
    const brokers = await ServiceCatalog.getServiceBrokers();
    dispatch(receiveBrokers(brokers));
    return brokers;
  };
}

export function getClasses() {
  return async (dispatch: Dispatch<IStoreState>) => {
    dispatch(requestClasses());
    const classes = await ServiceCatalog.getServiceClasses();
    dispatch(receiveClasses(classes));
    return classes;
  };
}

export function getInstances() {
  return async (dispatch: Dispatch<IStoreState>) => {
    dispatch(requestInstances());
    const instances = await ServiceInstance.list();
    dispatch(receiveInstances(instances));
    return instances;
  };
}

export function getPlans() {
  return async (dispatch: Dispatch<IStoreState>) => {
    dispatch(requestPlans());
    const plans = await ServiceCatalog.getServicePlans();
    dispatch(receivePlans(plans));
    return plans;
  };
}

export function getCatalog() {
  return async (dispatch: Dispatch<IStoreState>) => {
    dispatch(getBindings());
    dispatch(getBrokers());
    dispatch(getClasses());
    dispatch(getInstances());
    dispatch(getPlans());
  };
}

export function checkCatalogInstalled() {
  return async (dispatch: Dispatch<IStoreState>) => {
    const isInstalled = await ServiceCatalog.isCatalogInstalled();
    isInstalled ? dispatch(installed()) : dispatch(notInstalled());
    return isInstalled;
  };
}
