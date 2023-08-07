// Copyright 2020-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import actions from "actions";
import AlertGroup from "components/AlertGroup";
import Column from "components/Column";
import OperatorSummary from "components/OperatorSummary/OperatorSummary";
import OperatorHeader from "components/OperatorView/OperatorHeader";
import Row from "components/Row";
import { usePush } from "hooks/push";
import placeholder from "icons/placeholder.svg";
import { get } from "lodash";
import { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { useParams } from "react-router-dom";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import {
  IClusterServiceVersion,
  IClusterServiceVersionCRD,
  IResource,
  IStoreState,
} from "shared/types";
import * as url from "shared/url";
import { parseToString } from "shared/yamlUtils";
import OperatorInstanceFormBody from "../OperatorInstanceFormBody/OperatorInstanceFormBody";

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
            setDefaultValues(parseToString(example));
          }
        });
      }
    }
  });
}

export default function DeploymentFormBody() {
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();
  const [defaultValues, setDefaultValues] = useState("");
  const [crd, setCRD] = useState(undefined as IClusterServiceVersionCRD | undefined);
  const [icon, setIcon] = useState(placeholder);

  const {
    operators: {
      csv,
      isFetching,
      errors: {
        csv: { fetch: fetchError },
        resource: { create: createError },
      },
    },
    clusters: { currentCluster: cluster, clusters },
  } = useSelector((state: IStoreState) => state);
  const namespace = clusters[cluster].currentNamespace;
  const push = usePush();

  type IDeploymentFormBodyParams = {
    csv: string;
    crd: string;
  };
  const { csv: csvName, crd: crdName } = useParams<IDeploymentFormBodyParams>();

  useEffect(() => {
    // Clean up component state
    setDefaultValues("");
    setCRD(undefined);
    setIcon(placeholder);
    dispatch(actions.operators.getCSV(cluster, namespace, csvName || ""));
  }, [cluster, dispatch, namespace, csvName]);

  useEffect(() => {
    if (csv) {
      parseCSV(csv, crdName || "", setIcon, setCRD, setDefaultValues);
    }
  }, [csv, crdName]);

  if (!fetchError && !isFetching && !crd) {
    return (
      <AlertGroup status="danger">
        The CRD "{crdName}" was not found in the definition of "{csvName}".
      </AlertGroup>
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
      push(
        url.app.operatorInstances.view(
          cluster,
          namespace,
          csv.metadata.name,
          crd.name,
          resource.metadata.name,
        ),
      );
    }
  };

  if (fetchError) {
    return (
      <AlertGroup status="danger">
        An error occurred while fetching the ClusterServiceVersion: {fetchError.message}.
      </AlertGroup>
    );
  }
  return (
    <section>
      <OperatorHeader title={`Create ${crd?.kind}`} icon={icon} />
      <section>
        {createError && (
          <AlertGroup status="danger">
            An error occurred while creating the instance: {createError.message}.
          </AlertGroup>
        )}
        <Row>
          <Column span={3}>
            <OperatorSummary />
          </Column>
          <Column span={9}>
            <p>{crd?.description}</p>
            <OperatorInstanceFormBody
              isFetching={isFetching}
              handleDeploy={handleDeploy}
              defaultValues={defaultValues}
            />
          </Column>
        </Row>
      </section>
    </section>
  );
}
