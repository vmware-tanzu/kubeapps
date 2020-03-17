import { ThunkAction } from "redux-thunk";
import { ActionType, createAction } from "typesafe-actions";

import { Kube } from "../shared/Kube";
import { Operators } from "../shared/Operators";
import { IClusterServiceVersion, IPackageManifest, IResource, IStoreState } from "../shared/types";

export const checkingOLM = createAction("CHECKING_OLM");
export const OLMInstalled = createAction("OLM_INSTALLED");
export const OLMNotInstalled = createAction("OLM_NOT_INSTALLED");

export const requestOperators = createAction("REQUEST_OPERATORS");
export const receiveOperators = createAction("RECEIVE_OPERATORS", resolve => {
  return (operators: IPackageManifest[]) => resolve(operators);
});

export const requestOperator = createAction("REQUEST_OPERATOR");
export const receiveOperator = createAction("RECEIVE_OPERATOR", resolve => {
  return (operator: IPackageManifest) => resolve(operator);
});

export const errorOperators = createAction("ERROR_OPERATORS", resolve => {
  return (err: Error) => resolve(err);
});

export const requestCSVs = createAction("REQUEST_CSVS");
export const receiveCSVs = createAction("RECEIVE_CSVS", resolve => {
  return (csvs: IClusterServiceVersion[]) => resolve(csvs);
});

export const errorCSVs = createAction("ERROR_CSVS", resolve => {
  return (err: Error) => resolve(err);
});

export const requestCSV = createAction("REQUEST_CSV");
export const receiveCSV = createAction("RECEIVE_CSV", resolve => {
  return (csv: IClusterServiceVersion) => resolve(csv);
});

export const creatingResource = createAction("CREATING_RESOURCE");
export const resourceCreated = createAction("RESOURCE_CREATED", resolve => {
  return (resource: IResource) => resolve(resource);
});

export const errorResourceCreate = createAction("ERROR_RESOURCE_CREATE", resolve => {
  return (err: Error) => resolve(err);
});

export const requestCustomResources = createAction("REQUEST_CUSTOM_RESOURCES");
export const receiveCustomResources = createAction("RECEIVE_CUSTOM_RESOURCES", resolve => {
  return (resources: IResource[]) => resolve(resources);
});

export const errorCustomResource = createAction("ERROR_CUSTOM_RESOURCE", resolve => {
  return (err: Error) => resolve(err);
});

const actions = [
  checkingOLM,
  OLMInstalled,
  OLMNotInstalled,
  requestOperators,
  receiveOperators,
  errorOperators,
  requestOperator,
  receiveOperator,
  requestCSVs,
  receiveCSVs,
  errorCSVs,
  requestCSV,
  receiveCSV,
  creatingResource,
  resourceCreated,
  errorResourceCreate,
  requestCustomResources,
  receiveCustomResources,
  errorCustomResource,
];

export type OperatorAction = ActionType<typeof actions[number]>;

export function checkOLMInstalled(): ThunkAction<
  Promise<boolean>,
  IStoreState,
  null,
  OperatorAction
> {
  return async dispatch => {
    dispatch(checkingOLM());
    const installed = await Operators.isOLMInstalled();
    installed ? dispatch(OLMInstalled()) : dispatch(OLMNotInstalled());
    return installed;
  };
}

export function getOperators(
  namespace: string,
): ThunkAction<Promise<void>, IStoreState, null, OperatorAction> {
  return async dispatch => {
    dispatch(requestOperators());
    try {
      const operators = await Operators.getOperators(namespace);
      const sortedOp = operators.sort((o1, o2) => (o1.metadata.name > o2.metadata.name ? 1 : -1));
      dispatch(receiveOperators(sortedOp));
    } catch (e) {
      dispatch(errorOperators(e));
    }
  };
}

export function getOperator(
  namespace: string,
  operatorName: string,
): ThunkAction<Promise<void>, IStoreState, null, OperatorAction> {
  return async dispatch => {
    dispatch(requestOperator());
    try {
      const operator = await Operators.getOperator(namespace, operatorName);
      dispatch(receiveOperator(operator));
    } catch (e) {
      dispatch(errorOperators(e));
    }
  };
}

export function getCSVs(
  namespace: string,
): ThunkAction<Promise<IClusterServiceVersion[]>, IStoreState, null, OperatorAction> {
  return async dispatch => {
    dispatch(requestCSVs());
    try {
      const csvs = await Operators.getCSVs(namespace);
      const sortedCSVs = csvs.sort((o1, o2) => (o1.metadata.name > o2.metadata.name ? 1 : -1));
      dispatch(receiveCSVs(sortedCSVs));
      return csvs;
    } catch (e) {
      dispatch(errorCSVs(e));
      return [];
    }
  };
}

export function getCSV(
  namespace: string,
  name: string,
): ThunkAction<Promise<void>, IStoreState, null, OperatorAction> {
  return async dispatch => {
    dispatch(requestCSV());
    try {
      const csv = await Operators.getCSV(namespace, name);
      dispatch(receiveCSV(csv));
    } catch (e) {
      dispatch(errorCSVs(e));
    }
  };
}

export function createResource(
  namespace: string,
  apiVersion: string,
  resource: string,
  body: object,
): ThunkAction<Promise<boolean>, IStoreState, null, OperatorAction> {
  return async dispatch => {
    dispatch(creatingResource());
    try {
      const r = await Operators.createResource(namespace, apiVersion, resource, body);
      dispatch(resourceCreated(r));
      return true;
    } catch (e) {
      dispatch(errorResourceCreate(e));
      return false;
    }
  };
}

export function getResources(
  namespace: string,
): ThunkAction<Promise<void>, IStoreState, null, OperatorAction> {
  return async dispatch => {
    dispatch(requestCustomResources());
    const csvs = await dispatch(getCSVs(namespace));
    let resources: IResource[] = [];
    const csvPromises = csvs.map(async csv => {
      const crdPromises = csv.spec.customresourcedefinitions.owned.map(async crd => {
        const parsedCRD = crd.name.split(".");
        const name = parsedCRD[0];
        const group = parsedCRD.slice(1).join(".");
        try {
          const resourceGroup = await Kube.getAPIGroup(group);
          try {
            const csvResources = await Operators.listResources(
              namespace,
              resourceGroup.preferredVersion.groupVersion,
              name,
            );
            resources = resources.concat(csvResources.items);
          } catch (e) {
            dispatch(errorCustomResource(e));
          }
        } catch (e) {
          dispatch(
            errorCustomResource(
              new Error(`Unable to find resource group for ${crd.name}. Got ${e.message}`),
            ),
          );
        }
      });
      await Promise.all(crdPromises);
    });
    await Promise.all(csvPromises);
    if (resources.length) {
      dispatch(receiveCustomResources(resources));
    }
  };
}
