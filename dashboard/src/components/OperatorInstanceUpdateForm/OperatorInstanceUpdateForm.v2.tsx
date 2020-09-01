import { push } from "connected-react-router";
import * as yaml from "js-yaml";
import React, { useEffect, useState } from "react";

import actions from "actions";
import Alert from "components/js/Alert";
import Column from "components/js/Column";
import Row from "components/js/Row";
import { parseCSV } from "components/OperatorInstanceForm/OperatorInstanceForm.v2";
import OperatorNotSupported from "components/OperatorList/OperatorsNotSupported.v2";
import OperatorSummary from "components/OperatorSummary/OperatorSummary";
import OperatorHeader from "components/OperatorView/OperatorHeader.v2";
import { useDispatch, useSelector } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import * as url from "shared/url";
import placeholder from "../../placeholder.png";
import { IClusterServiceVersionCRD, IResource, IStoreState } from "../../shared/types";
import OperatorInstanceFormBody from "../OperatorInstanceFormBody/OperatorInstanceFormBody.v2";

export interface IOperatorInstanceUpgradeFormProps {
  csvName: string;
  crdName: string;
  cluster: string;
  namespace: string;
  resourceName: string;
}

function OperatorInstanceUpdateForm({
  csvName,
  crdName,
  cluster,
  namespace,
  resourceName,
}: IOperatorInstanceUpgradeFormProps) {
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();
  const [defaultValues, setDefaultValues] = useState("");
  const [currentValues, setCurrentValues] = useState("");
  const [crd, setCRD] = useState(undefined as IClusterServiceVersionCRD | undefined);
  const [icon, setIcon] = useState(placeholder);

  useEffect(() => {
    dispatch(actions.operators.getResource(cluster, namespace, csvName, crdName, resourceName));
    dispatch(actions.operators.getCSV(cluster, namespace, csvName));
  }, [dispatch, cluster, namespace, csvName, crdName, resourceName]);

  const {
    operators: {
      isFetching,
      csv,
      resource,
      errors: {
        resource: { fetch: fetchError, update: updateError },
      },
    },
    config: { kubeappsCluster },
  } = useSelector((state: IStoreState) => state);

  useEffect(() => {
    if (resource) {
      setCurrentValues(yaml.safeDump(resource));
    }
  }, [resource]);

  useEffect(() => {
    if (csv) {
      parseCSV(csv, crdName, setIcon, setCRD, setDefaultValues);
    }
  }, [csv, crdName]);

  if (cluster !== kubeappsCluster) {
    return <OperatorNotSupported kubeappsCluster={kubeappsCluster} namespace={namespace} />;
  }

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
              namespace={namespace}
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
