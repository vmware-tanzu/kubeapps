// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import actions from "actions";
import Alert from "components/js/Alert";
import ErrorAlert from "components/ErrorAlert";
import Column from "components/js/Column";
import Row from "components/js/Row";
import PageHeader from "components/PageHeader/PageHeader";
import {
  InstalledPackageReference,
  ResourceRef,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { InstalledPackage } from "shared/InstalledPackage";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import * as yaml from "js-yaml";
import { useEffect, useMemo, useState } from "react";
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
import { grpc } from "@improbable-eng/grpc-web";

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

  const [fetchError, setFetchError] = useState(error);
  const [pluginObj] = useState({ name: pluginName, version: pluginVersion } as Plugin);
  const [resourceRefs, setResourceRefs] = useState([] as ResourceRef[]);

  // useMemo used so that when installedPkgRef is a dependency of other effects,
  // it does not trigger the effect on every render.
  const installedPkgRef = useMemo(() => {
    return {
      context: { cluster, namespace },
      identifier: releaseName,
      plugin: pluginObj,
    } as InstalledPackageReference;
  }, [cluster, namespace, releaseName, pluginObj]);

  useEffect(() => {
    dispatch(actions.installedpackages.getInstalledPackage(installedPkgRef));
  }, [dispatch, installedPkgRef]);

  useEffect(() => {
    // TODO(minelson): currently it is not possible for a client to determine
    // whether resource refs are unavailable because the package is being
    // installed (ie.  Package is "Pending" the actual installation) or the
    // package is currently unable to be installed because the RBAC isn't yet
    // correct (ie. Package is "Pending" required RBAC). The work-around here
    // is to continue polling for the resource refs every two seconds as long
    // as a `NotFound` is returned.
    // See https://github.com/kubeapps/kubeapps/issues/4337
    let abort = false;
    const fetchResourceRefs = async () => {
      while (!abort) {
        try {
          const response = await InstalledPackage.GetInstalledPackageResourceRefs(installedPkgRef);
          if (abort) {
            return;
          }
          setResourceRefs(response.resourceRefs);
          return;
        } catch (e: any) {
          if (e.code !== grpc.Code.NotFound) {
            // If we get any other error, we want the user to know about it.
            setFetchError(new FetchError("unable to fetch resource references", [e]));
            return;
          }
          await new Promise(r => setTimeout(r, 2000));
        }
      }
    };
    fetchResourceRefs();

    // Ensure we abort fetching resource refs when unmounted.
    return () => {
      abort = true;
    };
  }, [installedPkgRef]);

  useEffect(() => {
    if (resourceRefs.length === 0) {
      return () => {};
    }

    const parsedRefs = parseResources(resourceRefs);
    setAppViewResourceRefs(parsedRefs);
    return () => {};
  }, [resourceRefs]);

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
    dispatch(actions.installedpackages.clearErrorInstalledPackage());
    dispatch(actions.installedpackages.getInstalledPackage(installedPkgRef));
  };

  if (fetchError) {
    if (fetchError.constructor === FetchError) {
      return (
        <ErrorAlert error={fetchError}>
          <CdsButton size="sm" action="flat" onClick={forceRetry} type="button">
            {" "}
            Try again{" "}
          </CdsButton>
        </ErrorAlert>
      );
    }
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
