import { ThunkAction } from "redux-thunk";
import { ActionType, createAction } from "typesafe-actions";

import { Operators } from "../shared/Operators";
import { IClusterServiceVersion, IPackageManifest, IResource, IStoreState } from "../shared/types";

export const checkingOLM = createAction("CHECKING_OLM");
export const OLMInstalled = createAction("OLM_INSTALLED");
export const errorOLMCheck = createAction("ERROR_OLM_CHECK", resolve => {
  return (err: Error) => resolve(err);
});

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

export const updatingResource = createAction("UPDATING_RESOURCE");
export const resourceUpdated = createAction("RESOURCE_UPDATED", resolve => {
  return (resource: IResource) => resolve(resource);
});

export const errorResourceUpdate = createAction("ERROR_RESOURCE_UPDATE", resolve => {
  return (err: Error) => resolve(err);
});

export const deletingResource = createAction("DELETING_RESOURCE");
export const resourceDeleted = createAction("RESOURCE_DELETED");
export const errorResourceDelete = createAction("ERROR_RESOURCE_DELETE", resolve => {
  return (err: Error) => resolve(err);
});

export const requestCustomResources = createAction("REQUEST_CUSTOM_RESOURCES");
export const receiveCustomResources = createAction("RECEIVE_CUSTOM_RESOURCES", resolve => {
  return (resources: IResource[]) => resolve(resources);
});

export const errorCustomResource = createAction("ERROR_CUSTOM_RESOURCE", resolve => {
  return (err: Error) => resolve(err);
});

export const requestCustomResource = createAction("REQUEST_CUSTOM_RESOURCE");
export const receiveCustomResource = createAction("RECEIVE_CUSTOM_RESOURCE", resolve => {
  return (resource: IResource) => resolve(resource);
});

export const creatingOperator = createAction("CREATING_OPERATOR");
export const operatorCreated = createAction("OPERATOR_CREATED", resolve => {
  return (resource: IResource) => resolve(resource);
});
export const errorOperatorCreate = createAction("ERROR_OPERATOR_CREATE", resolve => {
  return (err: Error) => resolve(err);
});

const actions = [
  checkingOLM,
  OLMInstalled,
  errorOLMCheck,
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
  updatingResource,
  resourceUpdated,
  errorResourceUpdate,
  requestCustomResources,
  receiveCustomResources,
  errorCustomResource,
  requestCustomResource,
  receiveCustomResource,
  deletingResource,
  resourceDeleted,
  errorResourceDelete,
  creatingOperator,
  operatorCreated,
  errorOperatorCreate,
];

export type OperatorAction = ActionType<typeof actions[number]>;

export function checkOLMInstalled(
  namespace: string,
): ThunkAction<Promise<boolean>, IStoreState, null, OperatorAction> {
  return async dispatch => {
    dispatch(checkingOLM());
    try {
      const installed = await Operators.isOLMInstalled(namespace);
      if (installed) {
        dispatch(OLMInstalled());
      }
      return installed;
    } catch (e) {
      dispatch(errorOLMCheck(e));
      return false;
    }
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
      return sortedCSVs;
    } catch (e) {
      dispatch(errorCSVs(e));
      return [];
    }
  };
}

export function getCSV(
  namespace: string,
  name: string,
): ThunkAction<Promise<IClusterServiceVersion | undefined>, IStoreState, null, OperatorAction> {
  return async dispatch => {
    dispatch(requestCSV());
    try {
      const csv = await Operators.getCSV(namespace, name);
      dispatch(receiveCSV(csv));
      return csv;
    } catch (e) {
      dispatch(errorCSVs(e));
      return;
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

export function deleteResource(
  namespace: string,
  plural: string,
  resource: IResource,
): ThunkAction<Promise<boolean>, IStoreState, null, OperatorAction> {
  return async dispatch => {
    dispatch(deletingResource());
    try {
      await Operators.deleteResource(
        namespace,
        resource.apiVersion,
        plural,
        resource.metadata.name,
      );
      dispatch(resourceDeleted());
      return true;
    } catch (e) {
      dispatch(errorResourceDelete(e));
      return false;
    }
  };
}

export function updateResource(
  namespace: string,
  apiVersion: string,
  resource: string,
  name: string,
  body: object,
): ThunkAction<Promise<boolean>, IStoreState, null, OperatorAction> {
  return async dispatch => {
    dispatch(updatingResource());
    try {
      const r = await Operators.updateResource(namespace, apiVersion, resource, name, body);
      dispatch(resourceUpdated(r));
      return true;
    } catch (e) {
      dispatch(errorResourceUpdate(e));
      return false;
    }
  };
}

function parseCRD(crdName: string) {
  const parsedCRD = crdName.split(".");
  const plural = parsedCRD[0];
  const group = parsedCRD.slice(1).join(".");
  return { plural, group };
}

export function getResources(
  namespace: string,
): ThunkAction<Promise<IResource[]>, IStoreState, null, OperatorAction> {
  return async dispatch => {
    dispatch(requestCustomResources());
    const csvs = await dispatch(getCSVs(namespace));
    let resources: IResource[] = [];
    const csvPromises = csvs.map(async csv => {
      const crdPromises = csv.spec.customresourcedefinitions.owned.map(async crd => {
        const { plural, group } = parseCRD(crd.name);
        try {
          const csvResources = await Operators.listResources(
            namespace,
            `${group}/${crd.version}`,
            plural,
          );
          resources = resources.concat(csvResources.items);
        } catch (e) {
          dispatch(errorCustomResource(e));
        }
      });
      await Promise.all(crdPromises);
    });
    await Promise.all(csvPromises);
    dispatch(receiveCustomResources(resources));
    return resources;
  };
}

export function getResource(
  namespace: string,
  csvName: string,
  crdName: string,
  resourceName: string,
): ThunkAction<Promise<void>, IStoreState, null, OperatorAction> {
  return async dispatch => {
    dispatch(requestCustomResource());
    const csv = await dispatch(getCSV(namespace, csvName));
    if (csv) {
      const crd = csv.spec.customresourcedefinitions.owned.find(c => c.name === crdName);
      if (crd) {
        const { plural, group } = parseCRD(crd.name);
        try {
          const resource = await Operators.getResource(
            namespace,
            `${group}/${crd.version}`,
            plural,
            resourceName,
          );
          dispatch(receiveCustomResource(resource));
        } catch (e) {
          dispatch(errorCustomResource(e));
        }
      } else {
        dispatch(
          errorCustomResource(
            new Error(`Not found a valid CRD definition for ${csvName}/${crdName}`),
          ),
        );
      }
    } else {
      dispatch(errorCustomResource(new Error(`CSV ${csvName} not found in ${namespace}`)));
    }
  };
}

export function createOperator(
  namespace: string,
  name: string,
  channel: string,
  installPlanApproval: string,
  csv: string,
): ThunkAction<Promise<boolean>, IStoreState, null, OperatorAction> {
  return async dispatch => {
    dispatch(creatingOperator());
    try {
      const r = await Operators.createOperator(namespace, name, channel, installPlanApproval, csv);
      dispatch(operatorCreated(r));
      return true;
    } catch (e) {
      dispatch(errorOperatorCreate(e));
      return false;
    }
  };
}
