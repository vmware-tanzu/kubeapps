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
  bindingsWithSecrets: IServiceBindingWithSecret[];
  brokers: IServiceBroker[];
  classes: IClusterServiceClass[];
  errors: {
    create?: Error;
    fetch?: Error;
    delete?: Error;
    deprovision?: Error;
    update?: Error;
  };
  instances: IServiceInstance[];
  isChecking: boolean;
  isInstalled: boolean;
  plans: IServicePlan[];
  isFetching: boolean;
}

const initialState: IServiceCatalogState = {
  bindingsWithSecrets: [],
  brokers: [],
  classes: [],
  errors: {},
  instances: [],
  isChecking: true,
  isInstalled: false,
  isFetching: false,
  plans: [],
};

const catalogReducer = (
  state: IServiceCatalogState = initialState,
  action: ServiceCatalogAction | LocationChangeAction | NamespaceAction,
): IServiceCatalogState => {
  const { catalog } = actions;
  switch (action.type) {
    case getType(catalog.installed):
      return { ...state, isChecking: false, isInstalled: true };
    case getType(catalog.notInstalled):
      return { ...state, isChecking: false, isInstalled: false };
    case getType(catalog.checkCatalogInstall):
      return { ...state, isChecking: true };
    case getType(catalog.receiveBrokers):
      return { ...state, isFetching: false, brokers: action.payload };
    case getType(catalog.receiveBindingsWithSecrets):
      return { ...state, isFetching: false, bindingsWithSecrets: action.payload };
    case getType(catalog.receiveClasses):
      return { ...state, isFetching: false, classes: action.payload };
    case getType(catalog.receiveInstances):
      return { ...state, isFetching: false, instances: action.payload };
    case getType(catalog.receivePlans):
      return { ...state, isFetching: false, plans: action.payload };
    case getType(catalog.errorCatalog):
      return { ...state, errors: { [action.payload.op]: action.payload.err } };
    case getType(catalog.requestClasses):
      return { ...state, isFetching: true };
    case LOCATION_CHANGE:
      return { ...state, errors: {} };
    case getType(actions.namespace.setNamespace):
      return { ...state, errors: {} };
    default:
      return { ...state };
  }
};

export default catalogReducer;
