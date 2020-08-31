import { RouterAction } from "connected-react-router";
import { assignWith } from "lodash";
import { get } from "lodash";
import React, { useEffect, useState } from "react";
import * as yaml from "yaml";
import placeholder from "../../placeholder.png";
import { CdsButton, CdsIcon } from "../Clarity/clarity";

import Alert from "components/js/Alert";
import Column from "components/js/Column";
import Row from "components/js/Row";
import PageHeader from "components/PageHeader/PageHeader.v2";
import { Link } from "react-router-dom";
import ApplicationStatus from "../../containers/ApplicationStatusContainer";
import helmIcon from "../../icons/helm.svg";
import ResourceRef from "../../shared/ResourceRef";
import { IK8sList, IRelease, IResource } from "../../shared/types";
import * as url from "../../shared/url";
import LoadingWrapper from "../LoadingWrapper/LoadingWrapper.v2";
import AccessURLTable from "./AccessURLTable/AccessURLTable.v2";
import DeleteButton from "./AppControls/DeleteButton/DeleteButton.v2";
import RollbackButton from "./AppControls/RollbackButton/RollbackButton.v2";
import AppNotes from "./AppNotes.v2";
import AppSecrets from "./AppSecrets";
import AppValues from "./AppValues/AppValues.v2";
import ChartInfo from "./ChartInfo/ChartInfo.v2";
import ResourceTabs from "./ResourceTabs";

export interface IAppViewProps {
  cluster: string;
  namespace: string;
  releaseName: string;
  app?: IRelease;
  error?: Error;
  deleteError?: Error;
  getAppWithUpdateInfo: (cluster: string, namespace: string, releaseName: string) => void;
  deleteApp: (
    cluster: string,
    namespace: string,
    releaseName: string,
    purge: boolean,
  ) => Promise<boolean>;
  push: (location: string) => RouterAction;
}

export interface IAppViewResourceRefs {
  deployments: ResourceRef[];
  statefulsets: ResourceRef[];
  daemonsets: ResourceRef[];
  services: ResourceRef[];
  ingresses: ResourceRef[];
  secrets: ResourceRef[];
  otherResources: ResourceRef[];
}

function parseResources(
  resources: Array<IResource | IK8sList<IResource, {}>>,
  cluster: string,
  releaseNamespace: string,
) {
  const result: IAppViewResourceRefs = {
    ingresses: [],
    deployments: [],
    statefulsets: [],
    daemonsets: [],
    otherResources: [],
    services: [],
    secrets: [],
  };
  resources.forEach(i => {
    // The item may be a list
    const itemList = i as IK8sList<IResource, {}>;
    if (itemList.items) {
      // If the resource  has a list of items, treat them as a list
      // A List can contain an arbitrary set of resources so we treat them as an
      // additional manifest. We merge the current result with the resources of
      // the List, concatenating items from both.
      assignWith(
        result,
        parseResources((i as IK8sList<IResource, {}>).items, cluster, releaseNamespace),
        // Merge the list with the current result
        (prev, newArray) => prev.concat(newArray),
      );
    } else {
      const item = i as IResource;
      const resource = { isFetching: true, item };
      switch (i.kind) {
        case "Deployment":
          result.deployments.push(new ResourceRef(resource.item, cluster, releaseNamespace));
          break;
        case "StatefulSet":
          result.statefulsets.push(new ResourceRef(resource.item, cluster, releaseNamespace));
          break;
        case "DaemonSet":
          result.daemonsets.push(new ResourceRef(resource.item, cluster, releaseNamespace));
          break;
        case "Service":
          result.services.push(new ResourceRef(resource.item, cluster, releaseNamespace));
          break;
        case "Ingress":
          result.ingresses.push(new ResourceRef(resource.item, cluster, releaseNamespace));
          break;
        case "Secret":
          result.secrets.push(new ResourceRef(resource.item, cluster, releaseNamespace));
          break;
        default:
          result.otherResources.push(new ResourceRef(resource.item, cluster, releaseNamespace));
      }
    }
  });
  return result;
}

export default function AppView({
  cluster,
  namespace,
  releaseName,
  app,
  error,
  deleteError,
  getAppWithUpdateInfo,
  deleteApp,
  push,
}: IAppViewProps) {
  const [resourceRefs, setResourceRefs] = useState({
    ingresses: [],
    deployments: [],
    statefulsets: [],
    daemonsets: [],
    otherResources: [],
    services: [],
    secrets: [],
  } as IAppViewResourceRefs);

  useEffect(() => {
    getAppWithUpdateInfo(cluster, namespace, releaseName);
  }, [getAppWithUpdateInfo, cluster, namespace, releaseName]);

  useEffect(() => {
    if (!app?.manifest) {
      return;
    }

    let parsedManifest: IResource[] = yaml
      .parseAllDocuments(app.manifest)
      .map((doc: yaml.ast.Document) => doc.toJSON());
    // Filter out elements in the manifest that does not comply
    // with { kind: foo }
    parsedManifest = parsedManifest.filter(r => r && r.kind);
    setResourceRefs(parseResources(parsedManifest, cluster, app.namespace));
  }, [app, cluster]);

  const {
    services,
    ingresses,
    deployments,
    statefulsets,
    daemonsets,
    secrets,
    otherResources,
  } = resourceRefs;
  const icon = get(app, "chart.metadata.icon", placeholder);
  return (
    <section>
      <PageHeader>
        <div className="kubeapps-header-content">
          <Row>
            <Column span={7}>
              <Row>
                <img src={icon} alt="app-icon" />
                <div className="kubeapps-title-block">
                  <h3>{releaseName}</h3>
                  <div className="kubeapps-header-subtitle">
                    <img src={helmIcon} alt="helm-icon" />
                    <span>Helm Chart</span>
                  </div>
                </div>
              </Row>
            </Column>
            <Column span={5}>
              <div className="control-buttons">
                <Row>
                  <div className="header-button">
                    <Link to={url.app.apps.upgrade(cluster, namespace, releaseName)}>
                      <CdsButton status="primary">
                        <CdsIcon shape="upload-cloud" inverse={true} /> Upgrade
                      </CdsButton>
                    </Link>
                  </div>
                  <div className="header-button">
                    <RollbackButton
                      cluster={cluster}
                      namespace={namespace}
                      releaseName={releaseName}
                      revision={app?.version || 0}
                    />
                  </div>
                  <div className="header-button">
                    <DeleteButton
                      cluster={cluster}
                      namespace={namespace}
                      releaseName={releaseName}
                    />
                  </div>
                </Row>
              </div>
            </Column>
          </Row>
        </div>
      </PageHeader>

      {error && <Alert theme="danger">Found an error: {error.message}</Alert>}
      {deleteError && (
        <Alert theme="danger">
          Unable to delete the application. Received: {deleteError.message}
        </Alert>
      )}
      {!app || !app.info ? (
        <LoadingWrapper />
      ) : (
        <Row>
          <Column span={3}>
            <ChartInfo app={app} />
          </Column>
          <Column span={9}>
            <div className="appview-separator">
              <div className="appview-first-row">
                <ApplicationStatus
                  deployRefs={deployments}
                  statefulsetRefs={statefulsets}
                  daemonsetRefs={daemonsets}
                  info={app.info}
                />
                <AccessURLTable serviceRefs={services} ingressRefs={ingresses} />
                <AppSecrets secretRefs={secrets} />
              </div>
            </div>
            <div className="appview-separator">
              <AppNotes notes={app.info && app.info.status && app.info.status.notes} />
            </div>
            <div className="appview-separator">
              <ResourceTabs
                {...{ deployments, statefulsets, daemonsets, secrets, services, otherResources }}
              />
            </div>
            <div className="appview-separator">
              <AppValues values={(app.config && app.config.raw) || ""} />
            </div>
          </Column>
        </Row>
      )}
    </section>
  );
}
