// Copyright 2020-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import { CdsIcon } from "@cds/react/icon";
import actions from "actions";
import AlertGroup from "components/AlertGroup";
import AppNotes from "components/AppView/AppNotes/AppNotes";
import AppSecrets from "components/AppView/AppSecrets";
import { IAppViewResourceRefs } from "components/AppView/AppView";
import Column from "components/Column";
import LoadingWrapper from "components/LoadingWrapper";
import { parseCSV } from "components/OperatorInstanceForm/OperatorInstanceForm";
import OperatorSummary from "components/OperatorSummary/OperatorSummary";
import OperatorHeader from "components/OperatorView/OperatorHeader";
import Row from "components/Row";
import { usePush } from "hooks/push";
import placeholder from "icons/placeholder.svg";
import { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { useParams } from "react-router-dom";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { fromCRD } from "shared/ResourceRef";
import { IClusterServiceVersionCRD, IKind, IResource, IStoreState } from "shared/types";
import { app } from "shared/url";
import { parseToString } from "shared/yamlUtils";
import AccessURLTable from "../AppView/AccessURLTable/AccessURLTable";
import AppValues from "../AppView/AppValues/AppValues";
import ResourceTabs from "../AppView/ResourceTabs";
import ApplicationStatus from "../ApplicationStatus/ApplicationStatus";
import ConfirmDialog from "../ConfirmDialog/ConfirmDialog";

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
        // TODO(minelson): Audit code to see if we can switch to use the normal ResourceRef
        // here instead, then remove the shared/ResourceRef with the extra fields.
        // https://github.com/vmware-tanzu/kubeapps/issues/6062
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

function OperatorInstance() {
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
    clusters: { currentCluster: cluster, clusters },
    kube: { kinds },
  } = useSelector((state: IStoreState) => state);
  const namespace = clusters[cluster].currentNamespace;

  type IOperatorInstanceParams = {
    csv: string;
    crd: string;
    instanceName: string;
  };
  const { csv: csvName, crd: crdName, instanceName } = useParams<IOperatorInstanceParams>();

  useEffect(() => {
    dispatch(
      actions.operators.getResource(
        cluster,
        namespace,
        csvName || "",
        crdName || "",
        instanceName || "",
      ),
    );
    dispatch(actions.operators.getCSV(cluster, namespace, csvName || ""));
  }, [dispatch, cluster, namespace, csvName, crdName, instanceName]);

  useEffect(() => {
    if (csv) {
      parseCSV(csv, crdName || "", setIcon, setCRD);
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

  const push = usePush();
  const onUpdateClick = () =>
    push(
      app.operatorInstances.update(
        cluster,
        namespace,
        csvName || "",
        crdName || "",
        instanceName || "",
      ),
    );
  const handleDeleteClick = async () => {
    setDeleting(true);
    const deleted = await dispatch(
      actions.operators.deleteResource(cluster, namespace, crd!.name.split(".")[0], resource!),
    );
    setDeleting(false);
    closeModal();
    if (deleted) {
      push(app.apps.list(cluster, namespace));
    }
  };

  if (errors.fetch) {
    return (
      <AlertGroup status="danger">
        An error occurred while fetching the instance: {errors.fetch.message}.
      </AlertGroup>
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
          {error && <AlertGroup status="danger">An error occurred: {error.message}.</AlertGroup>}
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
                    <AppNotes title="Resource Status" notes={parseToString(resource.status)} />
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
                    <AppValues values={parseToString(resource.spec)} />
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
