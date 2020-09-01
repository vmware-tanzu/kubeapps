import actions from "actions";
import AppNotes from "components/AppView/AppNotes.v2";
import AppSecrets from "components/AppView/AppSecrets";
import { IAppViewResourceRefs } from "components/AppView/AppView.v2";
import { CdsButton, CdsIcon } from "components/Clarity/clarity";
import Alert from "components/js/Alert";
import Column from "components/js/Column";
import Row from "components/js/Row";
import { parseCSV } from "components/OperatorInstanceForm/OperatorInstanceForm.v2";
import OperatorNotSupported from "components/OperatorList/OperatorsNotSupported.v2";
import OperatorSummary from "components/OperatorSummary/OperatorSummary";
import OperatorHeader from "components/OperatorView/OperatorHeader.v2";
import { push } from "connected-react-router";
import * as yaml from "js-yaml";
import React, { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import ApplicationStatus from "../../containers/ApplicationStatusContainer";
import placeholder from "../../placeholder.png";
import { fromCRD } from "../../shared/ResourceRef";
import { IClusterServiceVersionCRD, IResource, IStoreState } from "../../shared/types";
import { app } from "../../shared/url";
import AccessURLTable from "../AppView/AccessURLTable/AccessURLTable.v2";
import AppValues from "../AppView/AppValues/AppValues.v2";
import { IPartialAppViewState } from "../AppView/AppView";
import ResourceTabs from "../AppView/ResourceTabs";
import ConfirmDialog from "../ConfirmDialog/ConfirmDialog.v2";
import LoadingWrapper from "../LoadingWrapper/LoadingWrapper.v2";

export interface IOperatorInstanceProps {
  cluster: string;
  namespace: string;
  csvName: string;
  crdName: string;
  instanceName: string;
}

export interface IOperatorInstanceState {
  modalIsOpen: boolean;
  crd?: IClusterServiceVersionCRD;
  resources?: IPartialAppViewState;
}

function parseResource(
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
          result.deployments.push(fromCRD(r, cluster, namespace, ownerRef));
          break;
        case "StatefulSet":
          result.statefulsets.push(fromCRD(r, cluster, namespace, ownerRef));
          break;
        case "DaemonSet":
          result.daemonsets.push(fromCRD(r, cluster, namespace, ownerRef));
          break;
        case "Service":
          result.services.push(fromCRD(r, cluster, namespace, ownerRef));
          break;
        case "Ingress":
          result.ingresses.push(fromCRD(r, cluster, namespace, ownerRef));
          break;
        case "Secret":
          result.secrets.push(fromCRD(r, cluster, namespace, ownerRef));
          break;
        default:
          result.otherResources.push(fromCRD(r, cluster, namespace, ownerRef));
      }
    });
  } else {
    const emptyCRD = { kind: "", name: "", version: "" };
    // The CRD definition doesn't define any service so pull everything
    result = {
      deployments: [fromCRD({ ...emptyCRD, kind: "Deployment" }, cluster, namespace, ownerRef)],
      ingresses: [fromCRD({ ...emptyCRD, kind: "Ingress" }, cluster, namespace, ownerRef)],
      statefulsets: [fromCRD({ ...emptyCRD, kind: "StatefulSet" }, cluster, namespace, ownerRef)],
      daemonsets: [fromCRD({ ...emptyCRD, kind: "DaemonSet" }, cluster, namespace, ownerRef)],
      services: [fromCRD({ ...emptyCRD, kind: "Service" }, cluster, namespace, ownerRef)],
      secrets: [fromCRD({ ...emptyCRD, kind: "Secret" }, cluster, namespace, ownerRef)],
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
  const {
    services,
    ingresses,
    deployments,
    statefulsets,
    daemonsets,
    secrets,
    otherResources,
  } = resourceRefs;
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
    config: { kubeappsCluster },
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
      setResourceRefs(parseResource(cluster, namespace, resource, crd));
    }
  }, [crd, resource, cluster, namespace]);

  if (cluster !== kubeappsCluster) {
    return <OperatorNotSupported kubeappsCluster={kubeappsCluster} namespace={namespace} />;
  }

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
      dispatch(push(app.apps.list(kubeappsCluster, namespace)));
    }
  };

  const error = errors.fetch || errors.delete || errors.update;
  return (
    <section>
      <OperatorHeader
        title={`${instanceName} (${crd?.kind})`}
        icon={icon}
        version={csv?.spec.version}
      >
        <Row>
          <div className="header-button">
            <CdsButton status="primary" onClick={onUpdateClick}>
              <CdsIcon shape="upload-cloud" inverse={true} /> Update
            </CdsButton>
          </div>
          <div className="header-button">
            <CdsButton status="primary" onClick={openModal}>
              <CdsIcon shape="trash" inverse={true} /> Delete
            </CdsButton>
          </div>
          <ConfirmDialog
            onConfirm={handleDeleteClick}
            modalIsOpen={modalIsOpen}
            loading={deleting}
            confirmationText="Are you sure you want to delete the resource?"
            closeModal={closeModal}
          />
        </Row>
      </OperatorHeader>
      <section>
        <LoadingWrapper loaded={!isFetching}>
          {error && <Alert theme="danger">Found an error: {error.message}</Alert>}
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
                    <AppNotes title="Resource Status" notes={yaml.safeDump(resource.status)} />
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
                    <AppValues values={yaml.safeDump(resource.spec)} />
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
