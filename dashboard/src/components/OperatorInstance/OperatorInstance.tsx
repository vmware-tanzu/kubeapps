// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import { CdsIcon } from "@cds/react/icon";
import actions from "actions";
import AppNotes from "components/AppView/AppNotes/AppNotes";
import AppSecrets from "components/AppView/AppSecrets";
import { IAppViewResourceRefs } from "components/AppView/AppView";
import Alert from "components/js/Alert";
import Column from "components/js/Column";
import Row from "components/js/Row";
import { parseCSV } from "components/OperatorInstanceForm/OperatorInstanceForm";
import OperatorSummary from "components/OperatorSummary/OperatorSummary";
import OperatorHeader from "components/OperatorView/OperatorHeader";
import { push } from "connected-react-router";
import * as yaml from "js-yaml";
import { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { fromCRD } from "shared/ResourceRef";
import { IClusterServiceVersionCRD, IKind, IResource, IStoreState } from "shared/types";
import { app } from "shared/url";
import ApplicationStatus from "../../containers/ApplicationStatusContainer";
import placeholder from "icons/placeholder.svg";
import AccessURLTable from "../AppView/AccessURLTable/AccessURLTable";
import AppValues from "../AppView/AppValues/AppValues";
import ResourceTabs from "../AppView/ResourceTabs";
import ConfirmDialog from "../ConfirmDialog/ConfirmDialog";
import LoadingWrapper from "../LoadingWrapper/LoadingWrapper";

export interface IOperatorInstanceProps {
  cluster: string;
  namespace: string;
  csvName: string;
  crdName: string;
  instanceName: string;
}

function parseResource(
  kind: IKind,
  cluster: string,
  namespace: string,
  resource: IResource,
  crd: IClusterServiceVersionCRD,
) {
  let result: IAppViewResourceRefs = {
    ingresses: [],
    deployments: [],
    statefulsets: [],
    daemonsets: [],
    otherResources: [],
    services: [],
    secrets: [],
  };
  const ownerRef = { name: resource.metadata.name, kind: resource.kind };
  if (crd.resources) {
    crd.resources?.forEach(r => {
      switch (r.kind) {
        case "Deployment":
          result.deployments.push(fromCRD(r, kind, cluster, namespace, ownerRef));
          break;
        case "StatefulSet":
          result.statefulsets.push(fromCRD(r, kind, cluster, namespace, ownerRef));
          break;
        case "DaemonSet":
          result.daemonsets.push(fromCRD(r, kind, cluster, namespace, ownerRef));
          break;
        case "Service":
          result.services.push(fromCRD(r, kind, cluster, namespace, ownerRef));
          break;
        case "Ingress":
          result.ingresses.push(fromCRD(r, kind, cluster, namespace, ownerRef));
          break;
        case "Secret":
          result.secrets.push(fromCRD(r, kind, cluster, namespace, ownerRef));
          break;
        default:
          result.otherResources.push(fromCRD(r, kind, cluster, namespace, ownerRef));
      }
    });
  } else {
    const emptyCRD = { kind: "", name: "", version: "" };
    // The CRD definition doesn't define any service so pull everything
    result = {
      deployments: [
        fromCRD({ ...emptyCRD, kind: "Deployment" }, kind, cluster, namespace, ownerRef),
      ],
      ingresses: [fromCRD({ ...emptyCRD, kind: "Ingress" }, kind, cluster, namespace, ownerRef)],
      statefulsets: [
        fromCRD({ ...emptyCRD, kind: "StatefulSet" }, kind, cluster, namespace, ownerRef),
      ],
      daemonsets: [fromCRD({ ...emptyCRD, kind: "DaemonSet" }, kind, cluster, namespace, ownerRef)],
      services: [fromCRD({ ...emptyCRD, kind: "Service" }, kind, cluster, namespace, ownerRef)],
      secrets: [fromCRD({ ...emptyCRD, kind: "Secret" }, kind, cluster, namespace, ownerRef)],
      otherResources: [],
    };
  }
  return result;
}

function OperatorInstance({
  cluster,
  namespace,
  csvName,
  crdName,
  instanceName,
}: IOperatorInstanceProps) {
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();
  const [crd, setCRD] = useState(undefined as IClusterServiceVersionCRD | undefined);
  const [icon, setIcon] = useState(placeholder);
  const [deleting, setDeleting] = useState(false);
  const [resourceRefs, setResourceRefs] = useState({
    ingresses: [],
    deployments: [],
    statefulsets: [],
    daemonsets: [],
    otherResources: [],
    services: [],
    secrets: [],
  } as IAppViewResourceRefs);
  const { services, ingresses, deployments, statefulsets, daemonsets, secrets, otherResources } =
    resourceRefs;
  const [modalIsOpen, setModalIsOpen] = useState(false);
  const closeModal = () => setModalIsOpen(false);
  const openModal = () => setModalIsOpen(true);

  const {
    operators: {
      isFetching,
      csv,
      resource,
      errors: { resource: errors },
    },
    kube: { kinds },
  } = useSelector((state: IStoreState) => state);

  useEffect(() => {
    dispatch(actions.operators.getResource(cluster, namespace, csvName, crdName, instanceName));
    dispatch(actions.operators.getCSV(cluster, namespace, csvName));
  }, [dispatch, cluster, namespace, csvName, crdName, instanceName]);

  useEffect(() => {
    if (csv) {
      parseCSV(csv, crdName, setIcon, setCRD);
    }
  }, [csv, crdName]);

  useEffect(() => {
    if (crd && resource) {
      const kind = kinds[resource.kind];
      if (kind) {
        setResourceRefs(parseResource(kind, cluster, namespace, resource, crd));
      }
    }
  }, [crd, resource, cluster, kinds, namespace]);

  const onUpdateClick = () =>
    dispatch(
      push(app.operatorInstances.update(cluster, namespace, csvName, crdName, instanceName)),
    );
  const handleDeleteClick = async () => {
    setDeleting(true);
    const deleted = await dispatch(
      actions.operators.deleteResource(cluster, namespace, crd!.name.split(".")[0], resource!),
    );
    setDeleting(false);
    closeModal();
    if (deleted) {
      dispatch(push(app.apps.list(cluster, namespace)));
    }
  };

  if (errors.fetch) {
    return (
      <Alert theme="danger">
        An error occurred while fetching the instance: {errors.fetch.message}
      </Alert>
    );
  }
  const error = errors.delete || errors.update;
  return (
    <section>
      <ConfirmDialog
        onConfirm={handleDeleteClick}
        modalIsOpen={modalIsOpen}
        loading={deleting}
        headerText={"Delete resource"}
        confirmationText="Are you sure you want to delete the resource?"
        closeModal={closeModal}
      />
      <OperatorHeader
        title={`${instanceName} (${crd?.kind})`}
        icon={icon}
        version={csv?.spec.version}
        buttons={[
          <CdsButton key="update-button" status="primary" onClick={onUpdateClick}>
            <CdsIcon shape="upload-cloud" /> Update
          </CdsButton>,
          <CdsButton key="delete-button" status="primary" onClick={openModal}>
            <CdsIcon shape="trash" /> Delete
          </CdsButton>,
        ]}
      />
      <section>
        <LoadingWrapper
          className="margin-t-xxl"
          loadingText={`Fetching ${instanceName}...`}
          loaded={!isFetching}
        >
          {error && <Alert theme="danger">An error occurred: {error.message}</Alert>}
          {resource && (
            <Row>
              <Column span={3}>
                <OperatorSummary />
              </Column>
              <Column span={9}>
                <div className="appview-separator">
                  <div className="appview-first-row">
                    <ApplicationStatus
                      deployRefs={deployments}
                      statefulsetRefs={statefulsets}
                      daemonsetRefs={daemonsets}
                    />
                    <AccessURLTable serviceRefs={services} ingressRefs={ingresses} />
                    <AppSecrets secretRefs={secrets} />
                  </div>
                </div>
                {resource.status && (
                  <div className="appview-separator">
                    <AppNotes title="Resource Status" notes={yaml.dump(resource.status)} />
                  </div>
                )}
                <div className="appview-separator">
                  <ResourceTabs
                    {...{
                      deployments,
                      statefulsets,
                      daemonsets,
                      secrets,
                      services,
                      otherResources,
                    }}
                  />
                </div>
                {resource.spec && (
                  <div className="appview-separator">
                    <AppValues values={yaml.dump(resource.spec)} />
                  </div>
                )}
              </Column>
            </Row>
          )}
        </LoadingWrapper>
      </section>
    </section>
  );
}

export default OperatorInstance;
