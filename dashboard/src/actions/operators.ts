// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { get } from "lodash";
import { ThunkAction } from "redux-thunk";
import { Operators } from "shared/Operators";
import {
  IClusterServiceVersion,
  IClusterServiceVersionCRD,
  IPackageManifest,
  IResource,
  IStoreState,
} from "shared/types";
import { ActionType, deprecated } from "typesafe-actions";

const { createAction } = deprecated;

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

export const requestSubscriptions = createAction("REQUEST_SUBSCRIPTIONS");
export const receiveSubscriptions = createAction("RECEIVE_SUBSCRIPTIONS", resolve => {
  return (resources: IResource[]) => resolve(resources);
});

export const errorSubscriptionList = createAction("ERROR_SUBSCRIPTIONS", resolve => {
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
  requestSubscriptions,
  receiveSubscriptions,
  errorSubscriptionList,
];

export type OperatorAction = ActionType<typeof actions[number]>;

export function checkOLMInstalled(
  cluster: string,
  namespace: string,
): ThunkAction<Promise<boolean>, IStoreState, null, OperatorAction> {
  return async dispatch => {
    dispatch(checkingOLM());
    try {
      const installed = await Operators.isOLMInstalled(cluster, namespace);
      if (installed) {
        dispatch(OLMInstalled());
      }
      return installed;
    } catch (e: any) {
      dispatch(errorOLMCheck(e));
      return false;
    }
  };
}

export function getOperators(
  cluster: string,
  namespace: string,
): ThunkAction<Promise<void>, IStoreState, null, OperatorAction> {
  return async dispatch => {
    dispatch(requestOperators());
    try {
      const operators = await Operators.getOperators(cluster, namespace);
      const sortedOp = operators.sort((o1, o2) => (o1.metadata.name > o2.metadata.name ? 1 : -1));
      dispatch(receiveOperators(sortedOp));
    } catch (e: any) {
      dispatch(errorOperators(e));
    }
  };
}

export function getOperator(
  cluster: string,
  namespace: string,
  operatorName: string,
): ThunkAction<Promise<void>, IStoreState, null, OperatorAction> {
  return async dispatch => {
    dispatch(requestOperator());
    try {
      const operator = await Operators.getOperator(cluster, namespace, operatorName);
      dispatch(receiveOperator(operator));
    } catch (e: any) {
      dispatch(errorOperators(e));
    }
  };
}

export function getCSVs(
  cluster: string,
  namespace: string,
): ThunkAction<Promise<IClusterServiceVersion[]>, IStoreState, null, OperatorAction> {
  return async dispatch => {
    dispatch(requestCSVs());
    try {
      const csvs = await Operators.getCSVs(cluster, namespace);
      const sortedCSVs = csvs.sort((o1, o2) => (o1.metadata.name > o2.metadata.name ? 1 : -1));
      dispatch(receiveCSVs(sortedCSVs));
      return sortedCSVs;
    } catch (e: any) {
      dispatch(errorCSVs(e));
      return [];
    }
  };
}

export function getCSV(
  cluster: string,
  namespace: string,
  name: string,
): ThunkAction<Promise<IClusterServiceVersion | undefined>, IStoreState, null, OperatorAction> {
  return async dispatch => {
    dispatch(requestCSV());
    try {
      const csv = await Operators.getCSV(cluster, namespace, name);
      dispatch(receiveCSV(csv));
      return csv;
    } catch (e: any) {
      dispatch(errorCSVs(e));
      return;
    }
  };
}

export function createResource(
  cluster: string,
  namespace: string,
  apiVersion: string,
  resource: string,
  body: object,
): ThunkAction<Promise<boolean>, IStoreState, null, OperatorAction> {
  return async dispatch => {
    dispatch(creatingResource());
    try {
      const r = await Operators.createResource(cluster, namespace, apiVersion, resource, body);
      dispatch(resourceCreated(r));
      return true;
    } catch (e: any) {
      dispatch(errorResourceCreate(e));
      return false;
    }
  };
}

export function deleteResource(
  cluster: string,
  namespace: string,
  plural: string,
  resource: IResource,
): ThunkAction<Promise<boolean>, IStoreState, null, OperatorAction> {
  return async dispatch => {
    dispatch(deletingResource());
    try {
      await Operators.deleteResource(
        cluster,
        namespace,
        resource.apiVersion,
        plural,
        resource.metadata.name,
      );
      dispatch(resourceDeleted());
      return true;
    } catch (e: any) {
      dispatch(errorResourceDelete(e));
      return false;
    }
  };
}

export function updateResource(
  cluster: string,
  namespace: string,
  apiVersion: string,
  resource: string,
  name: string,
  body: object,
): ThunkAction<Promise<boolean>, IStoreState, null, OperatorAction> {
  return async dispatch => {
    dispatch(updatingResource());
    try {
      const r = await Operators.updateResource(
        cluster,
        namespace,
        apiVersion,
        resource,
        name,
        body,
      );
      dispatch(resourceUpdated(r));
      return true;
    } catch (e: any) {
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
  cluster: string,
  namespace: string,
): ThunkAction<Promise<IResource[]>, IStoreState, null, OperatorAction> {
  return async dispatch => {
    dispatch(requestCustomResources());
    const csvs = await dispatch(getCSVs(cluster, namespace));
    let resources: IResource[] = [];
    const csvPromises = csvs.map(async csv => {
      const crds = get(csv, "spec.customresourcedefinitions.owned", []);
      const crdPromises = crds.map(async (crd: IClusterServiceVersionCRD) => {
        const { plural, group } = parseCRD(crd.name);
        try {
          const csvResources = await Operators.listResources(
            cluster,
            csv.metadata.namespace,
            `${group}/${crd.version}`,
            plural,
          );
          resources = resources.concat(csvResources.items);
        } catch (e: any) {
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
  cluster: string,
  namespace: string,
  csvName: string,
  crdName: string,
  resourceName: string,
): ThunkAction<Promise<void>, IStoreState, null, OperatorAction> {
  return async dispatch => {
    dispatch(requestCustomResource());
    const csv = await dispatch(getCSV(cluster, namespace, csvName));
    if (csv) {
      const crd = csv.spec.customresourcedefinitions.owned?.find(c => c.name === crdName);
      if (crd) {
        const { plural, group } = parseCRD(crd.name);
        try {
          const resource = await Operators.getResource(
            cluster,
            namespace,
            `${group}/${crd.version}`,
            plural,
            resourceName,
          );
          dispatch(receiveCustomResource(resource));
        } catch (e: any) {
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
  cluster: string,
  namespace: string,
  name: string,
  channel: string,
  installPlanApproval: string,
  csv: string,
): ThunkAction<Promise<boolean>, IStoreState, null, OperatorAction> {
  return async dispatch => {
    dispatch(creatingOperator());
    try {
      const r = await Operators.createOperator(
        cluster,
        namespace,
        name,
        channel,
        installPlanApproval,
        csv,
      );
      dispatch(operatorCreated(r));
      return true;
    } catch (e: any) {
      dispatch(errorOperatorCreate(e));
      return false;
    }
  };
}

export function listSubscriptions(
  cluster: string,
  namespace: string,
): ThunkAction<Promise<boolean>, IStoreState, null, OperatorAction> {
  return async dispatch => {
    dispatch(requestSubscriptions());
    try {
      // First, request global subscriptions
      const globalSubs = await Operators.listSubscriptions(cluster, "operators");
      let items = globalSubs.items;
      if (namespace !== "operators") {
        const r = await Operators.listSubscriptions(cluster, namespace);
        items = items.concat(r.items);
      }
      dispatch(receiveSubscriptions(items));
      return true;
    } catch (e: any) {
      dispatch(errorSubscriptionList(e));
      return false;
    }
  };
}
