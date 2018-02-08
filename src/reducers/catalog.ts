import { getType } from "typesafe-actions";

import actions from "../actions";
import { ServiceCatalogAction } from "../actions/catalog";
import {
  IServiceBinding,
  IServiceBroker,
  IServiceClass,
  IServiceInstance,
  IServicePlan,
} from "../shared/ServiceCatalog";

export interface IServiceCatalogState {
  bindings: IServiceBinding[];
  brokers: IServiceBroker[];
  classes: IServiceClass[];
  instances: IServiceInstance[];
  isChecking: boolean;
  isInstalled: boolean;
  plans: IServicePlan[];
}

const initialState: IServiceCatalogState = {
  bindings: [],
  brokers: [],
  classes: [],
  instances: [],
  isChecking: true,
  isInstalled: false,
  plans: [],
};

export const catalogReducer = (
  state: IServiceCatalogState = initialState,
  action: ServiceCatalogAction,
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
    case getType(catalog.receiveBindings):
      const { bindings } = action;
      return { ...state, bindings };
    case getType(catalog.receiveClasses):
      const { classes } = action;
      return { ...state, classes };
    case getType(catalog.receiveInstances):
      const { instances } = action;
      return { ...state, instances };
    case getType(catalog.receivePlans):
      const { plans } = action;
      return { ...state, plans };
    default:
      return { ...state };
  }
};
