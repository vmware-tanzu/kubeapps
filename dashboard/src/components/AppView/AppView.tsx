import { CdsButton } from "@cds/react/button";
import actions from "actions";
import Alert from "components/js/Alert";
import Column from "components/js/Column";
import Row from "components/js/Row";
import PageHeader from "components/PageHeader/PageHeader";
import {
  InstalledPackageReference,
  ResourceRef,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import * as yaml from "js-yaml";
import { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import * as ReactRouter from "react-router-dom";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import {
  CustomInstalledPackageDetail,
  DeleteError,
  FetchError,
  FetchWarning,
  IStoreState,
} from "shared/types";
import { getPluginsSupportingRollback } from "shared/utils";
import ApplicationStatus from "../../containers/ApplicationStatusContainer";
import placeholder from "../../placeholder.png";
import LoadingWrapper from "../LoadingWrapper/LoadingWrapper";
import AccessURLTable from "./AccessURLTable/AccessURLTable";
import DeleteButton from "./AppControls/DeleteButton/DeleteButton";
import RollbackButton from "./AppControls/RollbackButton/RollbackButton";
import UpgradeButton from "./AppControls/UpgradeButton/UpgradeButton";
import AppNotes from "./AppNotes/AppNotes";
import AppSecrets from "./AppSecrets";
import AppValues from "./AppValues/AppValues";
import CustomAppView from "./CustomAppView";
import PackageInfo from "./PackageInfo/PackageInfo";
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

function parseResources(resourceRefs: Array<ResourceRef>) {
  const result: IAppViewResourceRefs = {
    ingresses: [],
    deployments: [],
    statefulsets: [],
    daemonsets: [],
    otherResources: [],
    services: [],
    secrets: [],
  };
  resourceRefs.forEach(ref => {
    switch (ref.kind) {
      case "Deployment":
        result.deployments.push(ref);
        break;
      case "StatefulSet":
        result.statefulsets.push(ref);
        break;
      case "DaemonSet":
        result.daemonsets.push(ref);
        break;
      case "Service":
        result.services.push(ref);
        break;
      case "Ingress":
        result.ingresses.push(ref);
        break;
      case "Secret":
        result.secrets.push(ref);
        break;
      default:
        result.otherResources.push(ref);
    }
  });
  return result;
}

function getButtons(app: CustomInstalledPackageDetail, error: any, revision: number) {
  if (!app || !app?.installedPackageRef || !app.installedPackageRef.plugin) {
    return [];
  }

  const buttons = [];

  // Upgrade is a core operation, it will always be available
  buttons.push(
    <UpgradeButton
      key="upgrade-button"
      installedPackageRef={app.installedPackageRef}
      releaseStatus={app?.status}
      disabled={error !== undefined}
    />,
  );

  // Rollback is a helm-only operation, it will only be available for helm-plugin packages
  if (getPluginsSupportingRollback().includes(app.installedPackageRef.plugin.name)) {
    buttons.push(
      <RollbackButton
        key="rollback-button"
        installedPackageRef={app.installedPackageRef}
        revision={revision}
        releaseStatus={app?.status}
        disabled={error !== undefined}
      />,
    );
  }

  // Delete is a core operation, it will always be available
  buttons.push(
    <DeleteButton
      key="delete-button"
      installedPackageRef={app.installedPackageRef}
      releaseStatus={app?.status}
    />,
  );

  return buttons;
}

export interface IRouteParams {
  cluster: string;
  namespace: string;
  releaseName: string;
  pluginName: string;
  pluginVersion: string;
}

export default function AppView() {
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();
  const { cluster, namespace, releaseName, pluginName, pluginVersion } =
    ReactRouter.useParams() as IRouteParams;
  const [appViewResourceRefs, setAppViewResourceRefs] = useState({
    ingresses: [],
    deployments: [],
    statefulsets: [],
    daemonsets: [],
    otherResources: [],
    services: [],
    secrets: [],
  } as IAppViewResourceRefs);
  const {
    apps: { error, selected: app, selectedDetails: appDetails },
    config: { customAppViews },
  } = useSelector((state: IStoreState) => state);

  const [pluginObj] = useState({ name: pluginName, version: pluginVersion } as Plugin);

  useEffect(() => {
    dispatch(
      actions.apps.getApp({
        context: { cluster: cluster, namespace: namespace },
        identifier: releaseName,
        plugin: pluginObj,
      } as InstalledPackageReference),
    );
  }, [cluster, dispatch, namespace, releaseName, pluginObj]);

  useEffect(() => {
    if (!app?.apiResourceRefs) {
      return () => {};
    }

    const parsedRefs = parseResources(app.apiResourceRefs);
    setAppViewResourceRefs(parsedRefs);
    return () => {};
  }, [app?.apiResourceRefs]);

  useEffect(() => {
    if (!app?.installedPackageRef) {
      return () => {};
    }
    const installedPackageRef = app.installedPackageRef;
    // Watch Deployments, StatefulSets, DaemonSets, Ingresses and Services.
    const refsToWatch = appViewResourceRefs.deployments.concat(
      appViewResourceRefs.statefulsets,
      appViewResourceRefs.daemonsets,
      appViewResourceRefs.ingresses,
      appViewResourceRefs.services,
    );
    // And just get the rest.
    const refsToGet = appViewResourceRefs.secrets.concat(appViewResourceRefs.otherResources);
    if (refsToGet.length > 0) {
      dispatch(actions.kube.getResources(installedPackageRef, refsToGet, false));
    }
    if (refsToWatch.length > 0) {
      dispatch(actions.kube.getResources(installedPackageRef, refsToWatch, true));
      return function cleanup() {
        dispatch(actions.kube.closeRequestResources(installedPackageRef));
      };
    }
    return () => {};
  }, [dispatch, app?.installedPackageRef, appViewResourceRefs]);

  const forceRetry = () => {
    dispatch(actions.apps.clearErrorApp());
    dispatch(
      actions.apps.getApp({
        context: { cluster: cluster, namespace: namespace },
        identifier: releaseName,
        plugin: pluginObj,
      } as InstalledPackageReference),
    );
  };

  if (error && error.constructor === FetchError) {
    return (
      <Alert theme="danger">
        Application not found: {error.message}
        <CdsButton size="sm" action="flat" onClick={forceRetry} type="button">
          {" "}
          Try again{" "}
        </CdsButton>
      </Alert>
    );
  }
  const { services, ingresses, deployments, statefulsets, daemonsets, secrets, otherResources } =
    appViewResourceRefs;
  const revision = app?.revision ?? 0;
  const icon = appDetails?.iconUrl ?? placeholder;

  // If the package identifier matches the current list of loaded customAppViews,
  // then load the custom view from external bundle instead of the default one.
  const appRepo = app?.availablePackageRef?.identifier.split("/")[0];
  const appName = app?.availablePackageRef?.identifier.split("/")[1];
  const appPlugin = app?.availablePackageRef?.plugin?.name;
  if (
    customAppViews.some(
      entry => entry.name === appName && entry.plugin === appPlugin && entry.repository === appRepo,
    )
  ) {
    return <CustomAppView resourceRefs={appViewResourceRefs} app={app!} appDetails={appDetails!} />;
  }

  return (
    <LoadingWrapper loaded={!!app} loadingText="Retrieving application..." className="margin-t-xl">
      {!app || !app?.installedPackageRef ? (
        <Alert theme="danger">There is a problem with this package</Alert>
      ) : (
        <section>
          <PageHeader
            title={releaseName}
            titleSize="md"
            plugin={app?.availablePackageRef?.plugin}
            icon={icon}
            buttons={getButtons(app, error, revision)}
          />
          {error &&
            (error.constructor === FetchWarning ? (
              <Alert theme="warning">
                There is a problem with this package: {error["message"]}
              </Alert>
            ) : error.constructor === DeleteError ? (
              <Alert theme="danger">
                Unable to delete the application. Received: {error["message"]}
              </Alert>
            ) : (
              <Alert theme="danger">An error occurred: {error["message"]}</Alert>
            ))}
          {!app || !app?.status ? (
            <LoadingWrapper loadingText={`Loading ${releaseName}...`} />
          ) : (
            <Row>
              <Column span={3}>
                <PackageInfo installedPackageDetail={app} availablePackageDetail={appDetails!} />
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
                <div className="appview-separator">
                  <AppValues
                    values={app?.valuesApplied ? yaml.dump(yaml.load(app.valuesApplied)) : ""}
                  />
                </div>
              </Column>
            </Row>
          )}
        </section>
      )}
    </LoadingWrapper>
  );
}
