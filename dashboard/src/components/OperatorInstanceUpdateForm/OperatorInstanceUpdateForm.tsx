// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import actions from "actions";
import Alert from "components/js/Alert";
import Column from "components/js/Column";
import Row from "components/js/Row";
import { parseCSV } from "components/OperatorInstanceForm/OperatorInstanceForm";
import OperatorSummary from "components/OperatorSummary/OperatorSummary";
import OperatorHeader from "components/OperatorView/OperatorHeader";
import { push } from "connected-react-router";
import placeholder from "icons/placeholder.svg";
import { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { IClusterServiceVersionCRD, IResource, IStoreState } from "shared/types";
import * as url from "shared/url";
import { parseToString } from "shared/yamlUtils";
import OperatorInstanceFormBody from "../OperatorInstanceFormBody/OperatorInstanceFormBody";
import { useParams } from "react-router-dom";

function OperatorInstanceUpdateForm() {
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();
  const [defaultValues, setDefaultValues] = useState("");
  const [currentValues, setCurrentValues] = useState("");
  const [crd, setCRD] = useState(undefined as IClusterServiceVersionCRD | undefined);
  const [icon, setIcon] = useState(placeholder);

  type IOperatorInstanceUpdateFormParams = {
    csv: string;
    crd: string;
    instanceName: string;
  };
  const {
    csv: csvName,
    crd: crdName,
    instanceName: resourceName,
  } = useParams<IOperatorInstanceUpdateFormParams>();
  const {
    operators: {
      isFetching,
      csv,
      resource,
      errors: {
        resource: { fetch: fetchError, update: updateError },
      },
    },
    clusters: { currentCluster: cluster, clusters },
  } = useSelector((state: IStoreState) => state);
  const namespace = clusters[cluster].currentNamespace;

  useEffect(() => {
    // Clean up component state
    setDefaultValues("");
    setCurrentValues("");
    setCRD(undefined);
    setIcon(placeholder);
    dispatch(actions.operators.getResource(cluster, namespace, csvName, crdName, resourceName));
    dispatch(actions.operators.getCSV(cluster, namespace, csvName));
  }, [dispatch, cluster, namespace, csvName, crdName, resourceName]);

  useEffect(() => {
    if (resource) {
      setCurrentValues(parseToString(resource));
    }
  }, [resource]);

  useEffect(() => {
    if (csv) {
      parseCSV(csv, crdName, setIcon, setCRD, setDefaultValues);
    }
  }, [csv, crdName]);

  if (!fetchError && !isFetching && !resource) {
    return <Alert>Resource {resourceName} not found</Alert>;
  }

  const handleDeploy = async (updatedResource: IResource) => {
    const created = await dispatch(
      actions.operators.updateResource(
        cluster,
        namespace,
        updatedResource.apiVersion,
        crdName.split(".")[0],
        resourceName,
        updatedResource,
      ),
    );
    if (created) {
      dispatch(
        push(url.app.operatorInstances.view(cluster, namespace, csvName, crdName, resourceName)),
      );
    }
  };

  if (fetchError) {
    return (
      <Alert theme="danger">
        An error occurred while fetching the ClusterServiceVersion: {fetchError.message}
      </Alert>
    );
  }
  return (
    <section>
      <OperatorHeader title={`Update ${resourceName}`} icon={icon} />
      <section>
        {updateError && (
          <Alert theme="danger">Found an error updating the instance: {updateError.message}</Alert>
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
              deployedValues={currentValues}
            />
          </Column>
        </Row>
      </section>
    </section>
  );
}

export default OperatorInstanceUpdateForm;
