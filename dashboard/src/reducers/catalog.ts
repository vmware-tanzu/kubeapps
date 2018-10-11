import { LOCATION_CHANGE, LocationChangeAction } from "connected-react-router";
import { getType } from "typesafe-actions";

import actions from "../actions";
import { ServiceCatalogAction } from "../actions/catalog";
import { NamespaceAction } from "../actions/namespace";
import { IClusterServiceClass } from "../shared/ClusterServiceClass";
import { IServiceBindingWithSecret } from "../shared/ServiceBinding";
import { IServiceBroker, IServicePlan } from "../shared/ServiceCatalog";
import { IServiceInstance } from "../shared/ServiceInstance";

export interface IServiceCatalogState {
  bindingsWithSecrets: {
    isFetching: boolean;
    list: IServiceBindingWithSecret[];
  };
  brokers: {
    isFetching: boolean;
    list: IServiceBroker[];
  };
  classes: {
    isFetching: boolean;
    list: IClusterServiceClass[];
  };
  errors: {
    create?: Error;
    fetch?: Error;
    delete?: Error;
    deprovision?: Error;
    update?: Error;
  };
  instances: {
    isFetching: boolean;
    list: IServiceInstance[];
  };
  isChecking: boolean;
  isServiceCatalogInstalled: boolean;
  plans: {
    isFetching: boolean;
    list: IServicePlan[];
  };
}

const initialState: IServiceCatalogState = {
  bindingsWithSecrets: { isFetching: false, list: [] },
  brokers: { isFetching: false, list: [] },
  classes: { isFetching: false, list: [] },
  errors: {},
  instances: { isFetching: false, list: [] },
  isChecking: true,
  isServiceCatalogInstalled: false,
  plans: { isFetching: false, list: [] },
};

const catalogReducer = (
  state: IServiceCatalogState = initialState,
  action: ServiceCatalogAction | LocationChangeAction | NamespaceAction,
): IServiceCatalogState => {
  const { catalog } = actions;
  let list = [];
  switch (action.type) {
    case getType(catalog.installed):
      return { ...state, isChecking: false, isServiceCatalogInstalled: true };
    case getType(catalog.notInstalled):
      return { ...state, isChecking: false, isServiceCatalogInstalled: false };
    case getType(catalog.checkCatalogInstall):
      return { ...state, isChecking: true };
    case getType(catalog.requestBrokers):
      list = state.brokers.list;
      return { ...state, brokers: { isFetching: true, list } };
    case getType(catalog.receiveBrokers):
      return { ...state, brokers: { isFetching: false, list: action.payload } };
    case getType(catalog.requestBindingsWithSecrets):
      list = state.bindingsWithSecrets.list;
      return { ...state, bindingsWithSecrets: { isFetching: true, list } };
    case getType(catalog.receiveBindingsWithSecrets):
      return { ...state, bindingsWithSecrets: { isFetching: false, list: action.payload } };
    case getType(catalog.requestClasses):
      list = state.classes.list;
      return { ...state, classes: { isFetching: true, list } };
    case getType(catalog.receiveClasses):
      return { ...state, classes: { isFetching: false, list: action.payload } };
    case getType(catalog.requestInstances):
      list = state.instances.list;
      return { ...state, instances: { isFetching: true, list } };
    case getType(catalog.receiveInstances):
      return { ...state, instances: { isFetching: false, list: action.payload } };
    case getType(catalog.requestPlans):
      list = state.plans.list;
      return { ...state, plans: { isFetching: true, list } };
    case getType(catalog.receivePlans):
      return { ...state, plans: { isFetching: false, list: action.payload } };
    case getType(catalog.errorCatalog):
      const brokers = { ...state.brokers, isFetching: false };
      const bindingsWithSecrets = { ...state.bindingsWithSecrets, isFetching: false };
      const classes = { ...state.classes, isFetching: false };
      const instances = { ...state.instances, isFetching: false };
      const plans = { ...state.plans, isFetching: false };
      return {
        ...state,
        brokers,
        bindingsWithSecrets,
        classes,
        instances,
        plans,
        errors: { [action.payload.op]: action.payload.err },
      };
    case LOCATION_CHANGE:
      return { ...state, errors: {} };
    case getType(actions.namespace.setNamespace):
      return { ...state, errors: {} };
    default:
      return { ...state };
  }
};

export default catalogReducer;
