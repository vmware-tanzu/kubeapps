import { LOCATION_CHANGE, LocationChangeAction } from "react-router-redux";
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
}

const initialState: IServiceCatalogState = {
  bindingsWithSecrets: [],
  brokers: [],
  classes: [],
  errors: {},
  instances: [],
  isChecking: true,
  isInstalled: false,
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
      const { brokers } = action;
      return { ...state, brokers };
    case getType(catalog.receiveBindingsWithSecrets):
      const { bindingsWithSecrets } = action;
      return { ...state, bindingsWithSecrets };
    case getType(catalog.receiveClasses):
      const { classes } = action;
      return { ...state, classes };
    case getType(catalog.receiveInstances):
      const { instances } = action;
      return { ...state, instances };
    case getType(catalog.receivePlans):
      const { plans } = action;
      return { ...state, plans };
    case getType(catalog.errorCatalog):
      return { ...state, errors: { [action.op]: action.err } };
    case LOCATION_CHANGE:
      return { ...state, errors: {} };
    case getType(actions.namespace.setNamespace):
      return { ...state, errors: {} };
    default:
      return { ...state };
  }
};

export default catalogReducer;
