import actions from "actions";
import Alert from "components/js/Alert";
import Column from "components/js/Column";
import Row from "components/js/Row";
import PageHeader from "components/PageHeader/PageHeader";
import * as yaml from "js-yaml";
import { assignWith } from "lodash";
import { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import * as ReactRouter from "react-router";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import {
  DeleteError,
  FetchError,
  IK8sList,
  IKubeState,
  IResource,
  IStoreState,
} from "shared/types";
import ApplicationStatus from "../../containers/ApplicationStatusContainer";
import placeholder from "../../placeholder.png";
import ResourceRef from "../../shared/ResourceRef";
import LoadingWrapper from "../LoadingWrapper/LoadingWrapper";
import AccessURLTable from "./AccessURLTable/AccessURLTable";
import DeleteButton from "./AppControls/DeleteButton/DeleteButton";
import RollbackButton from "./AppControls/RollbackButton/RollbackButton";
import UpgradeButton from "./AppControls/UpgradeButton/UpgradeButton";
import AppNotes from "./AppNotes/AppNotes";
import AppSecrets from "./AppSecrets";
import AppValues from "./AppValues/AppValues";
import ChartInfo from "./ChartInfo/ChartInfo";
import ResourceTabs from "./ResourceTabs";

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
  kinds: IKubeState["kinds"],
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
        parseResources((i as IK8sList<IResource, {}>).items, kinds, cluster, releaseNamespace),
        // Merge the list with the current result
        (prev, newArray) => prev.concat(newArray),
      );
    } else {
      const item = i as IResource;
      const resource = { isFetching: true, item };
      const kind = kinds[item.kind] || {};
      switch (i.kind) {
        case "Deployment":
          result.deployments.push(
            new ResourceRef(resource.item, cluster, kind.plural, kind.namespaced, releaseNamespace),
          );
          break;
        case "StatefulSet":
          result.statefulsets.push(
            new ResourceRef(resource.item, cluster, kind.plural, kind.namespaced, releaseNamespace),
          );
          break;
        case "DaemonSet":
          result.daemonsets.push(
            new ResourceRef(resource.item, cluster, kind.plural, kind.namespaced, releaseNamespace),
          );
          break;
        case "Service":
          result.services.push(
            new ResourceRef(resource.item, cluster, kind.plural, kind.namespaced, releaseNamespace),
          );
          break;
        case "Ingress":
          result.ingresses.push(
            new ResourceRef(resource.item, cluster, kind.plural, kind.namespaced, releaseNamespace),
          );
          break;
        case "Secret":
          result.secrets.push(
            new ResourceRef(resource.item, cluster, kind.plural, kind.namespaced, releaseNamespace),
          );
          break;
        default:
          result.otherResources.push(
            new ResourceRef(resource.item, cluster, kind.plural, kind.namespaced, releaseNamespace),
          );
      }
    }
  });
  return result;
}

interface IRouteParams {
  cluster: string;
  namespace: string;
  releaseName: string;
}

export default function AppView() {
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();
  const { cluster, namespace, releaseName } = ReactRouter.useParams() as IRouteParams;
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
    apps: { error, selected: app },
    kube: { kinds },
  } = useSelector((state: IStoreState) => state);
  useEffect(() => {
    dispatch(actions.apps.getApp(cluster, namespace, releaseName));
  }, [cluster, dispatch, namespace, releaseName]);

  useEffect(() => {
    if (!app || !app["manifest"]) {
      return;
    }

    if (Object.values(resourceRefs).some(ref => ref.length)) {
      // Already populated, skip
      return;
    }

    let parsedManifest: IResource[] = yaml
      .loadAll(app["manifest"])
      .map((doc: any) => JSON.parse(doc));
    // Filter out elements in the manifest that does not comply
    // with { kind: foo }
    parsedManifest = parsedManifest.filter(r => r && r.kind);
    const parsedRefs = parseResources(
      parsedManifest,
      kinds,
      cluster,
      app.installedPackageRef?.context?.namespace || "",
    );
    if (Object.values(parsedRefs).some(ref => ref.length)) {
      // Avoid setting refs if the manifest is empty
      setResourceRefs(parsedRefs);
    }
  }, [app, cluster, kinds, resourceRefs]);

  useEffect(() => {
    Object.values(resourceRefs).forEach((refs: ResourceRef[]) => {
      refs.forEach(ref => dispatch(actions.kube.getAndWatchResource(ref)));
    });
    return function cleanup() {
      Object.values(resourceRefs).forEach((refs: ResourceRef[]) => {
        refs.forEach(ref => dispatch(actions.kube.closeWatchResource(ref)));
      });
    };
  }, [dispatch, resourceRefs]);

  if (error && error.constructor === FetchError) {
    return <Alert theme="danger">Application not found. Received: {error.message}</Alert>;
  }
  const { services, ingresses, deployments, statefulsets, daemonsets, secrets, otherResources } =
    resourceRefs;
  // TODO(agamez): get icon from installedPackageDetails once available
  const icon = placeholder;
  const revision = app?.installedPackageRef?.identifier.split("/")[1] ?? "0";
  return (
    <section>
      <PageHeader
        title={releaseName}
        titleSize="md"
        helm={true}
        icon={icon}
        buttons={[
          <UpgradeButton
            key="upgrade-button"
            cluster={cluster}
            namespace={namespace}
            releaseName={releaseName}
            releaseStatus={app?.status}
          />,
          <RollbackButton
            key="rollback-button"
            cluster={cluster}
            namespace={namespace}
            releaseName={releaseName}
            revision={revision}
            releaseStatus={app?.status}
          />,
          <DeleteButton
            key="delete-button"
            cluster={cluster}
            namespace={namespace}
            releaseName={releaseName}
            releaseStatus={app?.status}
          />,
        ]}
      />
      {error &&
        (error.constructor === DeleteError ? (
          <Alert theme="danger">Unable to delete the application. Received: {error.message}</Alert>
        ) : (
          <Alert theme="danger">An error occurred: {error.message}</Alert>
        ))}
      {!app || !app?.status?.userReason ? (
        <LoadingWrapper loadingText={`Loading ${releaseName}...`} />
      ) : (
        <Row>
          <Column span={3}>
            <ChartInfo app={app} cluster={cluster} />
          </Column>
          <Column span={9}>
            <div className="appview-separator">
              <div className="appview-first-row">
                <ApplicationStatus
                  deployRefs={deployments}
                  statefulsetRefs={statefulsets}
                  daemonsetRefs={daemonsets}
                  info={app}
                />
                <AccessURLTable serviceRefs={services} ingressRefs={ingresses} />
                <AppSecrets secretRefs={secrets} />
              </div>
            </div>
            <div className="appview-separator">
              <AppNotes notes={app?.postInstallationNotes} />
            </div>
            <>
              {Object.keys(resourceRefs).every(r => resourceRefs[r].length > 0) && (
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
              )}
            </>
            <div className="appview-separator">
              <AppValues
                values={app?.valuesApplied ? yaml.dump(JSON.parse(app.valuesApplied)) : ""}
              />
            </div>
          </Column>
        </Row>
      )}
    </section>
  );
}
