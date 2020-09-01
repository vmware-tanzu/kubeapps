import { push } from "connected-react-router";
import * as yaml from "js-yaml";
import { get } from "lodash";
import React, { useEffect, useState } from "react";

import actions from "actions";
import Alert from "components/js/Alert";
import Column from "components/js/Column";
import Row from "components/js/Row";
import OperatorNotSupported from "components/OperatorList/OperatorsNotSupported.v2";
import OperatorSummary from "components/OperatorSummary/OperatorSummary";
import OperatorHeader from "components/OperatorView/OperatorHeader.v2";
import { useDispatch, useSelector } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import * as url from "shared/url";
import placeholder from "../../placeholder.png";
import {
  IClusterServiceVersion,
  IClusterServiceVersionCRD,
  IResource,
  IStoreState,
} from "../../shared/types";
import OperatorInstanceFormBody from "../OperatorInstanceFormBody/OperatorInstanceFormBody.v2";

export interface IOperatorInstanceFormProps {
  csvName: string;
  crdName: string;
  cluster: string;
  namespace: string;
}

export interface IOperatorInstanceFormBodyState {
  defaultValues: string;
  crd?: IClusterServiceVersionCRD;
}

export function parseCSV(
  csv: IClusterServiceVersion,
  crdName: string,
  setIcon: (icon: string) => void,
  setCRD: (crd: IClusterServiceVersionCRD) => void,
  setDefaultValues?: (v: string) => void,
) {
  const ownedCRDs = get(
    csv,
    "spec.customresourcedefinitions.owned",
    [],
  ) as IClusterServiceVersionCRD[];
  const csvIcon = get(csv, "spec.icon[0]");
  if (csvIcon) {
    setIcon(`data:${csvIcon.mediatype};base64,${csvIcon.base64data}`);
  }
  ownedCRDs.forEach(ownedCRD => {
    if (ownedCRD.name === crdName) {
      setCRD(ownedCRD);
      // Got the target CRD, extract the example
      if (setDefaultValues) {
        const kind = ownedCRD.kind;
        const rawExamples = get(csv, 'metadata.annotations["alm-examples"]', "[]");
        const examples = JSON.parse(rawExamples) as IResource[];
        examples.forEach(example => {
          if (example.kind === kind) {
            // Found the example, set the default values
            setDefaultValues(yaml.safeDump(example));
          }
        });
      }
    }
  });
}

export default function DeploymentFormBody({
  csvName,
  crdName,
  cluster,
  namespace,
}: IOperatorInstanceFormProps) {
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();
  const [defaultValues, setDefaultValues] = useState("");
  const [crd, setCRD] = useState(undefined as IClusterServiceVersionCRD | undefined);
  const [icon, setIcon] = useState(placeholder);

  useEffect(() => {
    dispatch(actions.operators.getCSV(cluster, namespace, csvName));
  }, [cluster, dispatch, namespace, csvName]);

  const {
    operators: {
      csv,
      isFetching,
      errors: {
        csv: { fetch: fetchError },
        resource: { create: createError },
      },
    },
    config: { kubeappsCluster },
  } = useSelector((state: IStoreState) => state);

  useEffect(() => {
    if (csv) {
      parseCSV(csv, crdName, setIcon, setCRD, setDefaultValues);
    }
  }, [csv, crdName]);

  if (cluster !== kubeappsCluster) {
    return <OperatorNotSupported kubeappsCluster={kubeappsCluster} namespace={namespace} />;
  }
  if (!fetchError && !isFetching && !crd) {
    return (
      <Alert theme="danger">
        {crdName} not found in the definition of {csvName}
      </Alert>
    );
  }

  const handleDeploy = async (resource: IResource) => {
    if (!crd || !csv) {
      // Unexpected error, CRD and CSV should have been previously populated
      throw new Error(`Missing CRD (${JSON.stringify(crd)}) or CSV (${JSON.stringify(csv)})`);
    }
    const resourceType = crd.name.split(".")[0];
    const created = await dispatch(
      actions.operators.createResource(
        cluster,
        namespace,
        resource.apiVersion,
        resourceType,
        resource,
      ),
    );
    if (created) {
      dispatch(
        push(
          url.app.operatorInstances.view(
            cluster,
            namespace,
            csv.metadata.name,
            crd.name,
            resource.metadata.name,
          ),
        ),
      );
    }
  };

  return (
    <section>
      <OperatorHeader title={`Create ${crd?.kind}`} icon={icon} />
      <section>
        {fetchError && (
          <Alert theme="danger">
            Found an error fetching the ClusterServiceVersion: {fetchError.message}
          </Alert>
        )}
        {createError && (
          <Alert theme="danger">Found an error creating the instance: {createError.message}</Alert>
        )}
        <Row>
          <Column span={3}>
            <OperatorSummary />
          </Column>
          <Column span={9}>
            <p>{crd?.description}</p>
            <OperatorInstanceFormBody
              isFetching={isFetching}
              namespace={namespace}
              handleDeploy={handleDeploy}
              defaultValues={defaultValues}
            />
          </Column>
        </Row>
      </section>
    </section>
  );
}
